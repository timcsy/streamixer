# Phase 0 Research: 字幕字體管理

## R1. 字體檔格式識別

**Decision**：以檔案前 4 bytes 的 magic number 識別字體格式：
- `\x00\x01\x00\x00` → TrueType (.ttf)
- `OTTO` (0x4F54544F) → OpenType with CFF (.otf)
- `ttcf` (0x74746366) → TrueType Collection (.ttc)

若檔頭不符，直接拒絕上傳。

**Rationale**：副檔名不可信；使用者可能將 `.txt` 改名。magic number 是最低成本的檔案內容驗證。完整解析 `name` table 取 family name 也需要先確認是字體。

**Alternatives considered**：
- 呼叫 `fc-query`（fontconfig CLI）驗證：需多一次 subprocess 呼叫，且錯誤訊息依 fontconfig 行為而定。
- 用 Go 字體解析套件（如 `golang.org/x/image/font/sfnt`）：需多引入相依，簡約原則傾向避免。

## R2. 取得字體 family name

**Decision**：使用 `golang.org/x/image/font/sfnt` 解析字體 `name` table 取 family name。此套件屬 Go 官方 extended library，相依簡單。

對 `.ttc`（TrueType Collection），預設取第一個 face 的 family name，將 face 索引記為 `0`。

**Rationale**：libass 與 fontconfig 都用 family name 找字體。正確取出 name 是讓「檔名與 FontName 解耦」的前提。管理員看到的是 family name（例如「思源黑體」），不是檔名。

**Alternatives considered**：
- 呼叫 `fc-scan --format='%{family[0]}'` subprocess：可行但增加外部依賴。
- 僅以檔名作為 family name：使用者體驗差，且無法在 FFmpeg 指定時正確解析。

## R3. 字體儲存位置與持久化

**Decision**：
- 系統字體：從 `/usr/share/fonts/` 遞迴掃描（Docker 映像內預裝）。
- 使用者字體：寫入 `/fonts/user/`，對應一個 Docker named volume `streamixer-fonts`。
- 每筆字體對應一個檔（檔名為 UUID + 原副檔名），同一目錄存一份 JSON metadata 指向 family name、size、upload time。

**Rationale**：
- named volume 重啟持久、與媒體目錄 `/media` 解耦。
- 檔名使用 UUID 避免名稱衝突、遏止路徑穿越。
- family name 存 metadata 避免每次掃描重新 parse 字體，加速列表 API。

**Alternatives considered**：
- 直接用原檔名：路徑穿越風險、name 衝突。
- 字體資料寫進 SQLite：對教會網站 < 10 個字體是過度設計。

## R4. libass fontsdir 多來源合併

**Decision**：FFmpeg `subtitles=` 濾鏡的 `fontsdir` 只接受單一路徑，但 libass 與 fontconfig 會走系統 font path。
實作上使用 symlink：容器啟動時（或 upload 完成後）把 `/fonts/user/*` symlink 到 `/usr/share/fonts/user/`，讓 fontconfig 一次找到系統與使用者字體；FFmpeg 不需顯式 `fontsdir`。

**Rationale**：
- 沿用 fontconfig 的搜尋機制，行為與既有 Noto CJK fallback 一致。
- 不需在 FFmpeg 命令列加字體路徑，保持合成邏輯單純。
- symlink 成本低、刪除字體時移除 symlink 即可。

**Alternatives considered**：
- 把 `/fonts/user` 加到 `fontsdir=/fonts/user`：需要在 FFmpeg 命令組裝時分歧，若想合併系統與使用者需要逗號分隔列表，libass 不支援。
- 定期執行 `fc-cache`：`/usr/share/fonts/` 下的字體 symlink 不需要額外 cache 就能被 fontconfig 發現（我們驗證 Alpine 行為）。若遇到未更新情境，upload 後呼叫 `fc-cache -f /usr/share/fonts/user` 作為保險。

## R5. FontName 在 ASS force_style 的傳遞

**Decision**：`-vf "subtitles=...:force_style='FontName=<name>,MarginV=30,Fontsize=28'"`。
無指定字體時省略 `FontName`，libass 使用預設。

**Rationale**：
- 與現有 `force_style='MarginV=30,Fontsize=28'` 組合方式一致。
- libass 遇到 family name 找不到會自動 fallback 到 Noto Sans CJK（因為 fontconfig 會做 match）。

**Edge case**：family name 含特殊字元（空格、逗號）時，ASS force_style 的分隔符可能衝突。決定：僅允許 family name 字元集 `[A-Za-z0-9 \u4e00-\u9fff-]`，其餘拒絕。

## R6. 使用者字體刪除的被引用檢查

**Decision**：
- Streamixer 後端 `/fonts` DELETE 端點只檢查檔案本身，不知道哪些 WP 素材用它。
- WP 端負責：刪除前呼叫 `WP_Query` 搜 `_streamixer_font` meta 找出指定該字體的素材；若有，在前端以 confirm dialog 列出清單。
- 後端刪除後立即移除檔案與 symlink；WP 端的素材 meta 保持不動（family name 字串），下次合成時後端找不到字體 → libass fallback → 不失敗。

**Rationale**：
- 職責分離：後端不需要知道 WP 資料模型。
- 用「找不到就 fallback」行為避免刪除失敗或合成失敗。

**Alternatives considered**：
- 刪除時自動把相關素材的 meta 改成「使用預設」：需跨 WP↔Streamixer 同步，複雜。
- 禁止刪除有被引用的字體：降低管理員自主權。

## R7. 單檔與總數上限

**Decision**：
- 單檔：10 MB（配合 `r.ParseMultipartForm` 設定）。
- 總數：50 個（硬碼，未來可調為 config）。
- 超過上限時回傳 400 + 可讀錯誤訊息。

**Rationale**：常見字體檔 < 5 MB；教會情境用不到 50 個字體。限制防止 volume 膨脹。

## R8. WP 前後端整合

**Decision**：
- WP 設定頁：新增「字體管理」區塊（字體列表 + 上傳表單 + 全站預設下拉）。表單 submit 時以 `wp_remote_post` 呼叫後端 `/fonts`。
- WP CPT：素材編輯頁新增字體下拉；欄位值存 `_streamixer_font`（字串，空 = 使用全站預設）。
- `sync_composition` 把 `font`（family name）作為 form field POST 給 Streamixer。
- Streamixer `/upload/{id}` 端點偵測到 `font` 欄位時寫入 `/media/{id}/font.txt`（一行 family name）。
- 合成時 `Loader` 讀 `font.txt`，若無則查全站預設（從 Streamixer `/config` 或直接 env）。

**Rationale**：以檔案儲存 family name 最簡單，沿用現行 `subtitle` / `transcript` 模式。

**Alternatives considered**：
- 把 font 編碼進 composition metadata JSON：需要新增 composition 層級 metadata 儲存機制，過度設計。

## R9. 全站預設字體儲存位置

**Decision**：存於 Streamixer 端的 `/fonts/default.txt`（單行 family name）。WP 端的 `/config` 端點擴充回傳 `default_font`；PUT 更新時寫入。

**Rationale**：
- 階段 8.5 已有 `/config` GET/PUT 機制，擴充一個欄位成本低。
- 所有合成路徑統一查 Streamixer 的預設，不依賴 WP 狀態。

**Alternatives considered**：
- 由 WP 端在每次 sync 時把全站預設帶進 composition：重複工作。

## R10. Dockerfile 擴充

**Decision**：保持現有 `font-noto-cjk` 安裝；新增 `/fonts/user/` 目錄（`mkdir -p /fonts/user`）與 symlink 從 `/usr/share/fonts/user` → `/fonts/user`（或直接把 `/fonts/user` 放進 font path）。`docker-compose.yml` 加 named volume `streamixer-fonts` 掛到 `/fonts`。

**Rationale**：不新增字體映像體積；使用者自行決定要上傳哪些字體。

## R11. 測試覆蓋

**Decision**：
- 單元測試：`src/fonts/validate_test.go` — magic number 識別（含三種格式 + 偽造拒絕）。
- 整合測試：`tests/integration/fonts_test.go`
  - GET `/fonts` 回傳包含系統字體
  - POST 合法 ttf → 列表新增
  - POST 偽造 ttf → 400
  - DELETE 存在字體 → 從列表消失
  - DELETE 系統字體 → 403
- 合成側：`tests/integration/font_render_test.go` — 指定字體 + 未指定字體皆能產生分段（不驗證視覺，只驗證 FFmpeg 回 0 exit code）。
