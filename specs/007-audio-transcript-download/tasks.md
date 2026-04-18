# Tasks: 音檔與逐字稿下載

**Feature Branch**: `007-audio-transcript-download`
**Spec**: [spec.md](./spec.md) | **Plan**: [plan.md](./plan.md)

> 路徑皆為相對於專案根目錄 `/Users/timcsy/Documents/Projects/streamixer/`。

## Phase 1: Setup

- [X] T001 確認開發環境：`docker compose up -d streamixer wordpress`、確認 `http://localhost:8080/health` 與 `http://localhost:8081` 正常
- [X] T002 在 `tests/integration/` 下建立空殼檔 `audio_download_test.go`、`transcript_download_test.go`（僅 package 宣告 + 一個 `TestMain` 骨架），確認 `go test ./tests/integration/... -run zzz` 可編譯
- [X] T003 新增 test fixtures：`tests/integration/fixtures/audio.mp3`、`tests/integration/fixtures/transcript.pdf`、`tests/integration/fixtures/transcript.md`（小檔即可）

## Phase 2: Foundational（阻塞所有 user stories）

- [X] T004 在 `src/media/types.go` 新增 `TranscriptFile` struct（欄位 Path、Ext、Size），並加入白名單常數 `AllowedTranscriptExts = {"txt","pdf","doc","docx","md"}`
- [X] T005 在 `src/media/composition.go` 為 `MediaComposition` 新增選填欄位 `Transcript *TranscriptFile`
- [X] T006 在 `src/media/loader.go` 新增 `loadTranscript(compDir string) (*TranscriptFile, error)`，以 `filepath.Glob("transcript.*")` 探測；`Load(id)` 結尾呼叫並賦值；找不到回傳 `nil, nil`
- [X] T007 在 `src/handler/download.go` 抽出共用 helper `contentDisposition(name string) string`，回傳符合 RFC 5987 的 `attachment; filename="..."; filename*=UTF-8''...`，並將現有影片下載的 `Content-Disposition` 組裝改用此 helper
- [X] T008 在 `src/handler/upload.go` 的 multipart 解析新增 `transcript` 欄位：儲存為 `transcript.<原副檔名>`；若舊 `transcript.*` 存在先移除避免殘留
- [X] T009 擴充 `wordpress/streamixer/streamixer.php` 的 `upload_mimes` 與 `wp_check_filetype_and_ext` filter：加入 `md → text/markdown`（txt/pdf/doc/docx 由 WP 預設支援即可）

## Phase 3: User Story 1 — 訪客下載原始音檔 (P1) 🎯 MVP

**Story Goal**：訪客從播放頁點「下載音檔」即可取得原始音檔。

**Independent Test**：上傳 `test.mp3` 到素材 `abc`，`curl -I http://localhost:8080/audio/abc` 回 200、`Content-Type: audio/mpeg`、`Content-Disposition: attachment`；`curl -s http://localhost:8080/audio/abc -o /tmp/out.mp3` 後 `shasum` 與原檔一致。

- [X] T010 [P] [US1] 在 `src/handler/` 新增 `media_download.go`，定義 `MediaDownloadHandler` struct 與 `NewMediaDownloadHandler(cfg, loader)`
- [X] T011 [US1] 在 `src/handler/media_download.go` 實作 `Audio(w, r)`：`loader.Load(id)` → 404 若不存在；`http.ServeFile(w, r, comp.Audio.Path)` 前設定 `Content-Type`（依副檔名）、`Content-Disposition`（用 T007 helper，檔名為 `<id>.<ext>`）、`X-Content-Type-Options: nosniff`
- [X] T012 [US1] 在 `src/handler/stream.go` 的 `SetupRouter` 註冊 `r.Get("/audio/{id}", mdh.Audio)`，並建立 `mdh`
- [X] T013 [P] [US1] 在 `tests/integration/audio_download_test.go` 撰寫 `TestAudioDownload_Basic`：啟動 httptest server、上傳 fixture → GET `/audio/<id>` → 檢查 status、headers、body sha256 與 fixture 一致
- [X] T014 [P] [US1] 在 `tests/integration/audio_download_test.go` 撰寫 `TestAudioDownload_404_MissingComposition` 與 `TestAudioDownload_404_MissingAudio`
- [X] T015 [P] [US1] 在 `tests/integration/audio_download_test.go` 撰寫 `TestAudioDownload_ChineseSlug`：composition id `測試` → URL encoded request → 200，`Content-Disposition` 含 `filename*=UTF-8''%E6%B8%AC%E8%A9%A6.mp3`
- [X] T016 [US1] 在 `wordpress/streamixer/includes/class-api.php` 新增 `get_audio_url($post_id)`：直接拼接 `post_name`（不 encode），參考既有 `get_download_url`
- [X] T017 [US1] 在 `wordpress/streamixer/includes/class-frontend.php` 的 `render_player` 中 toolbar 新增「下載音檔」按鈕（`<a>` 直接連結，無需 JS 進度）；放在影片下載右側
- [X] T018 [US1] 在 `wordpress/streamixer/assets/css/player.css` 確認 toolbar `flex-wrap: wrap` 讓手機版自動換行（若未設定則補上）

**Checkpoint**：此階段完成後即可獨立 demo 音檔下載（MVP）。

## Phase 4: User Story 2 — 訪客下載逐字稿 (P1)

**Story Goal**：訪客從播放頁點「下載逐字稿」取得逐字稿檔；無逐字稿時按鈕不顯示。

**Independent Test**：上傳 `.pdf` 逐字稿至 `abc`，`curl -I /transcript/abc` 回 200 + `Content-Type: application/pdf`；移除後 GET 回 404；WP 播放頁在有逐字稿時才顯示按鈕。

- [X] T019 [US2] 在 `src/handler/media_download.go` 實作 `Transcript(w, r)`：`loader.Load(id)` → 若 `comp.Transcript == nil` 回 404；設定 `Content-Type`（依 Ext 對應 MIME）、`Content-Disposition` attachment、`X-Content-Type-Options: nosniff`；`http.ServeFile(w, r, comp.Transcript.Path)`
- [X] T020 [US2] 在 `src/handler/stream.go` 的 `SetupRouter` 註冊 `r.Get("/transcript/{id}", mdh.Transcript)`
- [X] T021 [P] [US2] 在 `tests/integration/transcript_download_test.go` 撰寫 `TestTranscriptDownload_PDF`：上傳 fixture pdf → GET → 比對 sha256、檢查 `Content-Type: application/pdf`
- [X] T022 [P] [US2] 在 `tests/integration/transcript_download_test.go` 撰寫 `TestTranscriptDownload_Markdown`：`.md` 檔 → `Content-Type: text/markdown; charset=utf-8` 且有 `X-Content-Type-Options: nosniff`
- [X] T023 [P] [US2] 在 `tests/integration/transcript_download_test.go` 撰寫 `TestTranscriptDownload_404_NoTranscript`：僅上傳音檔的素材 → GET `/transcript/...` → 404
- [X] T024 [US2] 在 `wordpress/streamixer/includes/class-cpt.php` 的 `render_meta_box` 新增「逐字稿」欄位（`_streamixer_transcript_id`、hidden input、選擇/清除按鈕、preview），接受 MIME `text/plain,application/pdf,application/msword,application/vnd.openxmlformats-officedocument.wordprocessingml.document,text/markdown`
- [X] T025 [US2] 在 `class-cpt.php` 的 `save_meta` 欄位陣列加入 `streamixer_transcript_id`；`setupMediaButton` JS 呼叫新增對應一組
- [X] T026 [US2] 在 `class-cpt.php` 的檔案顯示邏輯複製 subtitle 模式：`$transcript_display` 處理三態（未選擇 / 已選擇 / 已同步本地清除），並讀 `_streamixer_transcript_id_filename` meta
- [X] T027 [US2] 在 `class-api.php` 的 `sync_composition` multipart body 新增 transcript 欄位（若 `_streamixer_transcript_id` 存在且檔案存在）
- [X] T028 [US2] 在 `class-api.php` 新增 `get_transcript_url($post_id)`：沿 `get_audio_url` 模式
- [X] T029 [US2] 在 `class-frontend.php` 的 `render_player`，判斷「有逐字稿」條件（`_streamixer_transcript_id` 或 `_streamixer_transcript_id_filename` 其一非空），條件性輸出「下載逐字稿」按鈕

**Checkpoint**：完成後訪客可下載音檔與逐字稿（雙 P1 齊備）。

## Phase 5: User Story 3 — 管理員管理逐字稿 (P2)

**Story Goal**：管理員可在素材編輯頁上傳、替換、清除逐字稿；儲存時同步到 Streamixer；覆蓋舊副檔名時後端正確反映。

**Independent Test**：上傳 `.pdf` 存檔 → 後端有 `transcript.pdf`；換成 `.md` 存檔 → 後端有 `transcript.md`、舊 `transcript.pdf` 已被清除；清除欄位後存檔 → 後端無 `transcript.*`。

- [X] T030 [US3] 在 `src/handler/upload.go` 的 transcript 解析前加入「清除既有 `transcript.*`」邏輯（`filepath.Glob` + `os.Remove`），確保副檔名變更時不殘留舊檔
- [X] T031 [US3] 在 `class-api.php` 的 `sync_composition`：若管理員清除了逐字稿欄位（`_streamixer_transcript_id == 0` 且無檔名 meta），上傳 multipart 增加一個特殊欄位 `transcript_delete=1`，讓 Streamixer 後端刪除 `transcript.*`
- [X] T032 [US3] 在 `src/handler/upload.go` 偵測 `transcript_delete` form value，若為 `1` 則刪除 `transcript.*` 並跳過上傳
- [X] T033 [P] [US3] 在 `tests/integration/transcript_download_test.go` 撰寫 `TestTranscriptUpload_ReplaceExtension`：先 POST `/upload/{id}` 帶 `.pdf` → GET 成功；再 POST 帶 `.md` → GET `.md` 成功且 `.pdf` 消失
- [X] T034 [P] [US3] 在 `tests/integration/transcript_download_test.go` 撰寫 `TestTranscriptUpload_Delete`：上傳後再 POST `transcript_delete=1` → GET `/transcript/{id}` 回 404

## Phase 6: User Story 4 — 同步後自動清除本地逐字稿 (P2)

**Story Goal**：啟用自動清除時，逐字稿與其他素材檔同步成功後從 WP 端移除，保留檔名 meta。

**Independent Test**：啟用自動清除 → 上傳並儲存含逐字稿的素材 → 查看 WP 媒體庫無該 attachment、編輯頁顯示原檔名 + 「已同步至 Streamixer，本地已清除」；訪客端 `/transcript/{id}` 仍可下載。

- [X] T035 [US4] 在 `class-api.php` 的 `cleanup_local_files` 的 `$fields` 陣列新增 `_streamixer_transcript_id`（其餘邏輯為迴圈，無需動）
- [X] T036 [US4] 在 `class-cpt.php` 的「已同步本地清除」顯示字串沿用既有三欄位模式、確認 `$transcript_display` 也 cover 這個情境（T026 若已處理可免）
- [X] T037 [US4] 手動驗收清單加入 quickstart.md 追蹤：啟用 `streamixer_auto_cleanup` → 儲存素材 → 確認四個 attachment 皆被刪除、`_filename` meta 保留、Streamixer 後端 `transcript.*` 仍在

## Phase 7: User Story 5 — 後台批次匯出音檔／逐字稿 (P3)

**Story Goal**：管理員在列表頁批次選「匯出音檔」或「匯出逐字稿」，瀏覽器依序下載每個已同步素材的對應檔案。

**Independent Test**：選 3 個已同步素材 → 下拉選「匯出音檔」→ 套用後瀏覽器依序開啟 3 個音檔下載；混入 1 個未同步 → 通知顯示「已匯出 3 個、跳過 1 個」。

- [X] T038 [US5] 在 `wordpress/streamixer/streamixer.php` 的 `bulk_actions-edit-streamixer` filter 新增 `streamixer_export_audio`、`streamixer_export_transcript`
- [X] T039 [US5] 在 `handle_bulk_actions-edit-streamixer` 依 action 類型收集對應 URL（`get_audio_url` / `get_transcript_url`），略過未同步者，並在 redirect query 多傳 `streamixer_export_skipped` 數量
- [X] T040 [US5] 在 `admin_notices` 更新通知字串，顯示已匯出 / 已跳過的數量（兩個數字）

## Phase 8: Polish & Cross-cutting

- [X] T041 [P] 更新 `wordpress/streamixer/streamixer.php` 的 `STREAMIXER_VERSION` 與 Plugin Header 至 `1.6.0`
- [X] T042 [P] 在 `knowledge/vision.md` 把階段 8 標記為進行中或完成（依實際發版時機）
- [X] T043 [P] 在 `knowledge/experience.md` 若合併期間出現新教訓（如 RFC 5987 細節、multipart 刪除語意），新增條目；否則略過
- [X] T044 執行完整測試：`go test ./... -race`、手動跑完 `quickstart.md` 的管理員／訪客流程
- [X] T045 重建 Docker image、清除 tmpfs 快取、在瀏覽器驗收三顆下載按鈕於桌面與手機寬度皆正常

## Dependencies

- Setup（T001–T003）→ Foundational（T004–T009）→ Stories
- US1（T010–T018）獨立，僅依賴 Foundational
- US2（T019–T029）依賴 Foundational 與 US1 的 `MediaDownloadHandler`（T010）
- US3（T030–T034）依賴 US2 的上傳／下載契約
- US4（T035–T037）依賴 US2（欄位存在）
- US5（T038–T040）依賴 US1、US2（URL helpers）
- Polish（T041–T045）在所有 US 完成後

## Parallel Opportunities

- T010、T013、T014、T015 可平行（不同檔案或讀 code 不衝突）
- T021、T022、T023 皆在同檔案 `transcript_download_test.go`，屬 `[P]` 因與 US1 測試檔不同，但建議先完成 T019 再跑測試
- T033、T034 同檔案、與 US2 測試檔同一檔，依序寫即可
- T041、T042、T043 平行（不同檔案）

## Implementation Strategy

1. **MVP**：做完 Phase 1–3（US1），即可交付「訪客下載原始音檔」。
2. **補第二個 P1**：Phase 4（US2），把逐字稿下載與 WP UI 串起來。
3. **管理成熟度**：Phase 5（US3）與 Phase 6（US4）— 替換／清除行為。
4. **批次能力**：Phase 7（US5）。
5. **收斂**：Phase 8 polish、版號、文件、回歸測試。

每個 US 結束均為可交付節點；建議 US1 單獨打 `v1.5.1`，全部完成後發 `v1.6.0`。
