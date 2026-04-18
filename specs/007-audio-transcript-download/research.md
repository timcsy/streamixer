# Phase 0 Research: 音檔與逐字稿下載

## R1. 逐字稿副檔名保留策略

**Decision**: 後端儲存為 `transcript.<原副檔名>`；前端 WP 在 meta 中同時儲存 attachment id 與原檔名 (`_streamixer_transcript_id_filename`)；Streamixer 端以 `filepath.Glob("transcript.*")` 探測。

**Rationale**:
- 使用者可能上傳不同格式（txt/pdf/doc/docx/md），保留原副檔名可避免反向推斷 MIME type。
- 與既有 `audio.*` / `background.*` / `subtitle.*` 探測模式一致（見 `src/media/loader.go`），符合「模組化架構」原則。
- 替換逐字稿（副檔名改變）時，上傳 handler 需先清除既有 `transcript.*`。

**Alternatives considered**:
- 固定為 `transcript.bin` 加 metadata 檔：多一層間接，違反簡約原則。
- 以 hash 命名：下載時還需再查表取得顯示名稱，成本高。

## R2. 中文檔名下載標頭

**Decision**: `Content-Disposition: attachment; filename="<ascii-fallback>"; filename*=UTF-8''<url-encoded-utf8>`（RFC 5987）。

**Rationale**:
- Chrome/Firefox/Safari 皆支援 `filename*=UTF-8''...`；舊瀏覽器 fallback 到 `filename=` 的 ASCII 值。
- 既有 `sanitizeFilename` 只清特殊字元、未處理非 ASCII，需擴充為產出雙欄位。

**Alternatives considered**:
- 純 URL encode `filename=`：Safari 下會把 `%E6...` 顯示為檔名，不友善。
- 只用 `filename*`：IE 與某些老舊 webview 不識別，會省略整個 header。

**Implementation note**: 抽一個 `contentDisposition(name string) string` 共用 helper，`download.go` 的影片下載也一併升級。

## R3. 音檔端點如何找到原始檔

**Decision**: 沿用 `media.Loader.Load(id)` 取得 `MediaComposition.Audio.Path`，由該 path 的副檔名決定下載的 MIME type 與 filename 副檔名。

**Rationale**:
- `Loader` 已封裝素材目錄探測，並處理音檔驗證（存在性、格式）。
- 音檔已經是儲存模態，直接 `http.ServeFile` 即可，不需額外處理。

**Alternatives considered**:
- 重新從 HLS 分段組裝：重度浪費 CPU，且重新編碼損失資訊，違反根公理「轉換成本守恆」。

## R4. 逐字稿在 MediaComposition 的位置

**Decision**: 在 `media/composition.go` 的 `MediaComposition` 新增可選欄位 `Transcript *TranscriptFile`（僅在檔案存在時設定）。`media/types.go` 新增 `TranscriptFile` struct（Path、Ext、Size）。

**Rationale**:
- 與既有 `Subtitle *SubtitleFile` 模式一致（選填媒體）。
- 讓端點 handler 不需自己在 filesystem 上找檔案，統一由 Loader 負責。

**Alternatives considered**:
- 在 handler 內直接 `filepath.Glob`：違反模組化架構原則，重複職責。

## R5. WordPress 同步流程擴充

**Decision**:
- `class-cpt.php`：render_meta_box 增加逐字稿欄位與 JS handler（接受 txt/pdf/doc/docx/md）。允許 doc/docx/md 上傳需擴充 `upload_mimes` filter。
- `class-api.php`：`sync_composition` 偵測 `_streamixer_transcript_id`，若有則 `build_multipart_field` 多帶一個 `transcript` field；`cleanup_local_files` 迴圈擴充為四個欄位。
- `class-frontend.php`：播放頁 toolbar 渲染音檔按鈕恆顯、逐字稿按鈕條件顯示（依 `_streamixer_transcript_id_filename` 是否存在判斷）。

**Rationale**:
- 與階段 6 既有 pattern 完全一致（fields 陣列、sync meta 狀態機），風險低。
- 逐字稿按鈕依「檔名 meta」而非「attachment id」判斷顯示，因為同步後 id 會歸零。

**Alternatives considered**:
- 用 REST + sidebar panel：與現行強制傳統編輯器設定衝突，違反階段 7.5 教訓。

## R6. WordPress upload_mimes 擴充

**Decision**: 在 `streamixer.php` 的 `upload_mimes` filter 內加入：
- `md → text/markdown`
- `doc → application/msword`（WP 預設已支援）
- `docx → application/vnd.openxmlformats-officedocument.wordprocessingml.document`（WP 預設已支援）
- `txt`、`pdf`（WP 預設已支援）

並擴充 `wp_check_filetype_and_ext` filter，將 `md` 加入白名單。

**Rationale**:
- WP 預設拒絕 `md`，與階段 5 字幕（srt/vtt）同樣情況，沿用既有擴充手法。

## R7. 前端下載按鈕與進度條互動

**Decision**:
- 影片下載：沿用現行「點擊 → 輪詢進度 → 觸發下載」邏輯（進度條顯示）。
- 音檔下載：直接 `<a download href>`，不輪詢（音檔已是靜態檔）。
- 逐字稿下載：同音檔，直接連結。
- 三顆按鈕水平排列；窄螢幕（< 480px）自動換行（flex-wrap）。

**Rationale**:
- 音檔與逐字稿不經合成、無進度可言，多加進度條反而誤導。
- 成功標準 SC-006 要求「逐字稿下載點擊數 ≤ 1」，直接連結最直接。

## R8. 批次匯出擴充

**Decision**: 在 `streamixer.php` 的 `bulk_actions-edit-streamixer` filter 新增兩個 action：`streamixer_export_audio`、`streamixer_export_transcript`。handler 依既有模式收集 URL 後以 JS 錯開開新分頁。

**Rationale**: 與影片匯出完全對稱，使用者心智負擔低。

## R9. 路由註冊

**Decision**: 在 `src/handler/stream.go` 的 `SetupRouter` 新增：
```go
r.Get("/audio/{id}", mdh.Audio)
r.Get("/transcript/{id}", mdh.Transcript)
```
`mdh` 為新增的 `MediaDownloadHandler`。

## R10. 測試涵蓋

**Decision**:
- `tests/integration/audio_download_test.go`：上傳音檔 → GET `/audio/{id}` → 比對位元與雜湊。
- `tests/integration/transcript_download_test.go`：上傳 `.pdf` 逐字稿 → GET → 比對；中文 slug `測試/範例` → 驗證下載 URL 無雙重編碼；RFC 5987 header 檢查。
- 404 情境：素材不存在、逐字稿未上傳。
- 大檔記憶體測試可省略，使用 `http.ServeFile` 本身即串流輸出，依賴標準函式庫保證。
