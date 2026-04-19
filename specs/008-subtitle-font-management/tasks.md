# Tasks: 字幕字體管理

**Feature Branch**: `008-subtitle-font-management`
**Spec**: [spec.md](./spec.md) | **Plan**: [plan.md](./plan.md)

> 路徑皆相對於 `/Users/timcsy/Documents/Projects/streamixer/`。

## Phase 1: Setup

- [X] T001 在 `Dockerfile` 建立 `/fonts/user` 目錄與 `/usr/share/fonts/user` symlink；驗證 `font-noto-cjk` 仍安裝
- [X] T002 在 `docker-compose.yml` 為 `streamixer` service 新增 named volume `streamixer-fonts:/fonts`
- [X] T003 新增 test fixture：`tests/integration/fixtures/sample.ttf`（小型合法 ttf，可由 `fc-match` + `ttf2woff2` 產生或從 GoNoto 取最小子集）

## Phase 2: Foundational（阻塞所有 user stories）

- [X] T004 在 `src/config/config.go` 新增欄位：`FontDir`（預設 `/fonts`）、`SystemFontDirs`（預設 `["/usr/share/fonts"]`）、`FontSymlinkDir`（預設 `/usr/share/fonts/user`）、`MaxFontSize`（預設 10MB）、`MaxFontCount`（預設 50）
- [X] T005 在 `src/fonts/types.go` 定義 `Font` struct 與 `FontSource` 列舉（`SourceSystem` / `SourceUser`）
- [X] T006 在 `src/fonts/validate.go` 實作 `SniffFontFormat(r io.Reader) (ext string, err error)`，以 magic number 識別 ttf/otf/ttc；非字體則回傳錯誤
- [X] T007 在 `src/fonts/validate.go` 新增 `ParseFamilyName(path string) (string, error)`，使用 `golang.org/x/image/font/sfnt` 解析 name table
- [X] T008 執行 `go get golang.org/x/image/font/sfnt` 並更新 `go.mod` / `go.sum`
- [X] T009 在 `src/fonts/manager.go` 實作 `FontManager` struct 與 `NewFontManager(cfg)`，含 `List`、`Get`、`Upload`、`Delete`、`DefaultFamily`、`SetDefaultFamily` 方法；檔案佈局依 data-model.md
- [X] T010 在 `src/fonts/manager_test.go` 撰寫單元測試：`TestSniff_TTF`、`TestSniff_OTF`、`TestSniff_TTC`、`TestSniff_Rejects_TextFile`、`TestParseFamilyName_NotoSansTC`

## Phase 3: User Story 1 — 上傳並設為全站預設 (P1) 🎯 MVP

**Story Goal**：管理員可從 WP 設定頁上傳字體並設為全站預設；合成採用該字體。

**Independent Test**：POST `/fonts` 帶 ttf fixture → 201；PUT `/fonts/default` 指定該 family → 200；合成時 `subtitles=` filter 含 `FontName=<family>`；視覺比對字幕風格改變。

- [X] T011 [P] [US1] 在 `src/handler/fonts.go` 新增 `FontHandler` struct 與 `NewFontHandler(manager)`
- [X] T012 [US1] 在 `src/handler/fonts.go` 實作 `List(w, r)`：回 `{default_family, fonts: [...]}`
- [X] T013 [US1] 在 `src/handler/fonts.go` 實作 `Upload(w, r)`：multipart `font` 欄位 → sniff → parse → manager.Upload → 201
- [X] T014 [US1] 在 `src/handler/fonts.go` 實作 `SetDefault(w, r)`：PUT body `{family_name}` → validate in list → manager.SetDefaultFamily
- [X] T015 [US1] 在 `src/handler/stream.go` 的 `SetupRouterWithSweeper` 註冊路由：`GET /fonts`、`POST /fonts`（auth）、`PUT /fonts/default`（auth）；並在 `main.go` 初始化 `FontManager` 傳入
- [X] T016 [US1] 在 `src/composer/ffmpeg.go` 的 `BuildFFmpegArgs` 與 `BuildSegmentArgs` 新增 `resolveFontFamily` 邏輯：先看 `comp.FontFamily`、再看 `manager.DefaultFamily()`，組 `force_style='MarginV=30,Fontsize=28,FontName=<family>'`
- [X] T017 [US1] 在 `src/media/composition.go` 為 `MediaComposition` 新增 `FontFamily string` 欄位；`src/media/loader.go` 讀取 `/media/{id}/font.txt` 填入
- [X] T018 [US1] 在 `src/handler/upload.go` 解析 multipart `font` form 值，寫入 `/media/{id}/font.txt`（空字串則刪除該檔）
- [X] T019 [P] [US1] 在 `tests/integration/fonts_test.go` 撰寫 `TestFonts_ListIncludesSystem`：GET `/fonts` → 至少一筆 `source=system`
- [X] T020 [P] [US1] 在 `tests/integration/fonts_test.go` 撰寫 `TestFonts_UploadAndList`：POST fixture ttf → 201；再 GET 看到該字體
- [X] T021 [P] [US1] 在 `tests/integration/fonts_test.go` 撰寫 `TestFonts_UploadRejectsInvalidHeader`：POST text 偽裝 ttf → 400
- [X] T022 [P] [US1] 在 `tests/integration/fonts_test.go` 撰寫 `TestFonts_UploadRequiresAuth`：POST 不帶 API Key → 401
- [X] T023 [P] [US1] 在 `tests/integration/fonts_test.go` 撰寫 `TestFonts_SetDefault_UpdatesListResponse`：PUT default 後 GET `/fonts` 的 `default_family` 正確
- [X] T024 [P] [US1] 在 `tests/integration/fonts_test.go` 撰寫 `TestSynth_WithFontName_IncludesForceStyle`：以 `BuildFFmpegArgs` + 設定 FontFamily 檢驗輸出包含 `FontName=`
- [X] T025 [US1] 在 `wordpress/streamixer/includes/class-fonts.php` 新增 thin client：`fetch_fonts()`、`upload_font($file)`、`delete_font($id)`、`set_default($family)`；內部使用 `wp_remote_*` 呼叫後端
- [X] T026 [US1] 在 `wordpress/streamixer/includes/class-settings.php` 新增「字體管理」區塊：列表（從 `fetch_fonts()`）、上傳表單、全站預設下拉；表單 submit 直接呼叫後端 `/fonts`，成功後 `update_option('streamixer_default_font', ...)` 與 `redirect`
- [X] T027 [US1] 在 `wordpress/streamixer/includes/class-api.php` 的 `sync_composition` 新增 `font` multipart 欄位（值為 `_streamixer_font` meta，空字串則不送）

**Checkpoint**：完成後可上傳、選預設、合成即採用（MVP）。

## Phase 4: User Story 2 — 系統內建字體可用 (P1)

**Story Goal**：未上傳任何字體時既有素材仍正確顯示中文字幕。

**Independent Test**：全新部署、無 user font → 播放舊素材字幕仍顯示中文；GET `/fonts` 至少一筆系統字體含 CJK。

- [X] T028 [US2] 在 `src/fonts/manager.go` 的 `scanSystemFonts` 遞迴掃 `SystemFontDirs`，識別 ttf/otf/ttc，parse family name；預設排除 `user/` 子目錄避免重複
- [X] T029 [US2] 在 `src/fonts/manager.go` 對系統字體使用穩定 ID（基於檔名 slug），source=system，UploadedAt=zero
- [X] T030 [P] [US2] 在 `tests/integration/fonts_test.go` 撰寫 `TestFonts_SystemFontsCJK`：GET `/fonts` 回的 system 字體至少一個 family 含「Noto」或等效中文支援字體

## Phase 5: User Story 3 — 素材指定字體 (P2)

**Story Goal**：單一素材可覆寫全站預設字體。

**Independent Test**：素材 A 指定 `Special`、B 用預設；A 合成的 subtitles filter 含 `FontName=Special`、B 含 `FontName=<default>`（或省略若 default 空）。

- [X] T031 [US3] 在 `wordpress/streamixer/includes/class-cpt.php` 的 `render_meta_box` 新增「字體」下拉：選項 = 「使用全站預設」+ `fetch_fonts()` 結果；對應 meta `_streamixer_font`
- [X] T032 [US3] 在 `class-cpt.php` 的 `save_meta` 欄位陣列加入 `streamixer_font`（字串，不是 int）
- [X] T033 [P] [US3] 在 `tests/integration/fonts_test.go` 撰寫 `TestSynth_CompositionOverridesDefault`：先 SetDefault=A；composition 設 FontFamily=B → 輸出 `FontName=B`

## Phase 6: User Story 4 — 刪除使用者字體 (P2)

**Story Goal**：使用者字體可刪；系統字體不可；刪除時 WP 前端提示被引用的素材。

**Independent Test**：上傳 X → 指定給素材 M → WP 前端按刪除 → confirm dialog 列出 M；確認後 X 消失；M 下次合成自動 fallback 不失敗。

- [X] T034 [US4] 在 `src/handler/fonts.go` 實作 `Delete(w, r)`：404 若不存在、403 若 source=system；成功後若 default.txt 指向此 family 則清空
- [X] T035 [US4] 在 `src/handler/stream.go` 路由註冊 `DELETE /fonts/{id}`（auth）
- [X] T036 [US4] 在 `src/composer/ffmpeg.go` 的 `resolveFontFamily` 加 fallback：若指定 family 不在 manager.List() 結果中，則視為未指定（空字串）
- [X] T037 [US4] 在 `wordpress/streamixer/includes/class-settings.php` 字體列表每筆的「刪除」按鈕：僅在 source=user 顯示；按下時 `WP_Query` 找 `meta_key=_streamixer_font, meta_value=<family>` 的 posts，以 JS confirm 顯示 title 清單
- [X] T038 [P] [US4] 在 `tests/integration/fonts_test.go` 撰寫 `TestFonts_DeleteUser`：upload → delete → list 不再有
- [X] T039 [P] [US4] 在 `tests/integration/fonts_test.go` 撰寫 `TestFonts_DeleteSystem_Forbidden`：DELETE `/fonts/<system-id>` → 403
- [X] T040 [P] [US4] 在 `tests/integration/fonts_test.go` 撰寫 `TestSynth_MissingFontFallsBack`：composition 指定不存在的 family → BuildFFmpegArgs 輸出不含 `FontName=<missing>`

## Phase 7: User Story 5 — 上傳防呆與容量控制 (P3)

**Story Goal**：拒絕偽造字體檔、限制單檔大小與總數。

**Independent Test**：text 偽裝 ttf → 400；15MB 檔 → 413；上傳第 51 個字體 → 409。

- [X] T041 [US5] 在 `src/fonts/manager.go` 的 `Upload` 前檢查 `List()` 的 user 子集數量 ≥ `MaxFontCount` 則回傳 `ErrFontLimit`；handler 對應 409
- [X] T042 [US5] 在 `src/handler/fonts.go` 的 `Upload` 檢查 multipart 大小 > `MaxFontSize` → 413；拒絕後清理暫存檔
- [X] T043 [P] [US5] 在 `tests/integration/fonts_test.go` 撰寫 `TestFonts_UploadSizeLimit`：超過 10MB → 413
- [X] T044 [P] [US5] 在 `tests/integration/fonts_test.go` 撰寫 `TestFonts_UploadCountLimit`：mock 已有 50 個 user 字體 → 下次 → 409

## Phase 8: Polish & Cross-cutting

- [X] T045 [P] 在 `wordpress/streamixer/streamixer.php` 將 `STREAMIXER_VERSION` 與 Plugin Header 升至 `1.7.0`
- [X] T046 [P] 更新 `knowledge/vision.md` 標記階段 9 完成
- [X] T047 [P] 若產生新教訓（例如 symlink + fontconfig 行為細節）新增 `knowledge/experience.md` 條目
- [X] T048 執行 `go test ./... -race`；手動跑 `quickstart.md` 全流程（管理員 + 訪客）
- [X] T049 重建 Docker image、清 tmpfs 快取、重新合成一個素材觀察字體切換；桌面與手機寬度皆驗收

## Dependencies

- Setup（T001–T003）→ Foundational（T004–T010）→ Stories
- US1（T011–T027）獨立、依賴 Foundational
- US2（T028–T030）可與 US1 並行（掃描邏輯獨立）
- US3（T031–T033）依賴 US1（需要 `fetch_fonts()` thin client 與 FontFamily 欄位）
- US4（T034–T040）依賴 US1（列表端點）與 US3（`_streamixer_font` meta）
- US5（T041–T044）依賴 US1（Upload 路徑）
- Polish（T045–T049）在所有 US 完成後

## Parallel Opportunities

- T019–T024、T030、T033、T038–T040、T043–T044 皆為獨立測試檔案片段，可並行撰寫
- T011、T025、T028 可於 T010 完成後並行（handler struct、WP thin client、系統掃描）
- T045、T046、T047 皆不同檔案，可並行

## Implementation Strategy

1. **MVP**：Phase 1–3（US1）— 上傳字體 + 設全站預設 + 合成採用，交付階段 9 核心價值
2. **補第二個 P1**：Phase 4（US2）確認系統字體路徑
3. **素材粒度**：Phase 5（US3）加素材編輯頁下拉
4. **管理成熟度**：Phase 6（US4）刪除與 fallback
5. **維運邊界**：Phase 7（US5）防呆與上限
6. **收斂**：Phase 8 版號、文件、回歸測試。預計打 `v1.7.0`。
