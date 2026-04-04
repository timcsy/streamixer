# Tasks: fMP4 分段格式（消除破音）

**Input**: Design documents from `/specs/003-fmp4-segment-format/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: 包含測試任務（Constitution 要求 TDD）

**Organization**: 任務按 user story 分組。本功能主要是修改既有程式碼，非新增。

## Format: `[ID] [P?] [Story] Description`

- **[P]**: 可平行執行（不同檔案、無相依）
- **[Story]**: 所屬 user story（US1、US2、US3）
- 包含確切的檔案路徑

## Phase 1: Setup

**Purpose**: 無新增相依，確認 FFmpeg fMP4 支援可用

- [x] T001 驗證本地 FFmpeg 支援 `-hls_segment_type fmp4`（執行測試指令確認）

---

## Phase 2: Foundational（阻塞性前置作業）

**Purpose**: 修改核心合成邏輯，從 .ts 改為 fMP4

**⚠️ CRITICAL**: 所有 user story 必須等此階段完成才能開始

- [x] T002 修改 BuildFFmpegArgs：加入 `-hls_segment_type fmp4`、`-hls_fmp4_init_filename init.mp4`、分段副檔名改為 `.m4s` 於 `src/composer/ffmpeg.go`
- [x] T003 修改 BuildSegmentArgs：輸出格式從 `-f mpegts` 改為 `-f mp4 -movflags +frag_keyframe+empty_moov+default_base_moof+cmaf`，副檔名改為 `.m4s` 於 `src/composer/ffmpeg.go`
- [x] T004 修改 GeneratePlaylist：版本升至 7，加入 `#EXT-X-MAP:URI="init.mp4"`，分段名稱改為 `.m4s` 於 `src/composer/playlist.go`
- [x] T005 修改 PregenManager.IsSegmentReady/GetSegmentPath：副檔名從 `.ts` 改為 `.m4s` 於 `src/composer/pregen.go`
- [x] T006 新增 PregenManager.IsInitReady/GetInitPath：檢查 init.mp4 是否存在於 `src/composer/pregen.go`
- [x] T007 新增 GenerateInitSegment：獨立產生 init.mp4（用於按需生成前 init 不存在的情況）於 `src/composer/ffmpeg.go`

**Checkpoint**: 核心合成邏輯已改為 fMP4 格式

---

## Phase 3: User Story 1 — 順序播放無破音 (Priority: P1) 🎯 MVP

**Goal**: 預生成產出 fMP4 分段，順序播放時分段切換無破音

**Independent Test**: 播放 30 秒以上合成串流，聆聽每個分段切換點確認無破音

### Tests for User Story 1 ⚠️

> **NOTE: 先寫這些測試，確認測試失敗後才開始實作**

- [x] T008 [P] [US1] 單元測試：GeneratePlaylist 輸出包含 `#EXT-X-VERSION:7`、`#EXT-X-MAP:URI="init.mp4"` 和 `.m4s` 分段名稱於 `tests/unit/playlist_test.go`
- [x] T009 [P] [US1] 單元測試：BuildFFmpegArgs 包含 `-hls_segment_type fmp4` 和 `-hls_fmp4_init_filename init.mp4` 於 `tests/unit/composer_test.go`
- [x] T010 [P] [US1] 整合測試：預生成完成後 init.mp4 和 .m4s 分段都存在，ffprobe 可驗證格式正確於 `tests/integration/fmp4_test.go`
- [x] T011 [P] [US1] 整合測試：GET /stream/{id}/index.m3u8 回傳的 playlist 包含 EXT-X-MAP 和 .m4s 分段於 `tests/integration/fmp4_test.go`

### Implementation for User Story 1

- [x] T012 [US1] 修改 SegmentHandler：支援 init.mp4 請求（新增路由 `/stream/{id}/init.mp4`）於 `src/handler/stream.go`
- [x] T013 [US1] 修改 SegmentHandler：`.m4s` 分段的 Content-Type 從 `video/mp2t` 改為 `video/mp4` 於 `src/handler/stream.go`
- [x] T014 [US1] 更新路由：加入 `/stream/{id}/init.mp4` handler 於 `src/handler/stream.go`

**Checkpoint**: 預生成的 fMP4 串流可在瀏覽器中無破音播放

---

## Phase 4: User Story 2 — 跳轉後音訊平順 (Priority: P2)

**Goal**: 按需生成的 fMP4 分段與預生成的相容，跳轉後無破音

**Independent Test**: 跳轉到多個時間點，確認音訊平順

### Tests for User Story 2 ⚠️

- [x] T015 [P] [US2] 單元測試：BuildSegmentArgs 包含 fMP4 movflags 和 .m4s 副檔名於 `tests/unit/composer_test.go`
- [x] T016 [P] [US2] 整合測試：按需生成的 .m4s 分段可被 ffprobe 驗證為有效的 fMP4 於 `tests/integration/fmp4_test.go`

### Implementation for User Story 2

- [x] T017 [US2] 修改 SegmentHandler：按需生成前檢查 init.mp4 是否存在，不存在則先產生於 `src/handler/stream.go`

**Checkpoint**: 跳轉到任意位置後音訊平順，無破音

---

## Phase 5: User Story 3 — 主流播放器相容 (Priority: P3)

**Goal**: Chrome（hls.js）、Safari、手機瀏覽器都能正常播放

**Independent Test**: 在不同瀏覽器上播放同一串流

### Tests for User Story 3 ⚠️

- [x] T018 [P] [US3] 整合測試：GET /stream/{id}/init.mp4 回傳 Content-Type video/mp4 於 `tests/integration/fmp4_test.go`
- [x] T019 [P] [US3] 整合測試：GET /stream/{id}/seg_000.m4s 回傳 Content-Type video/mp4 於 `tests/integration/fmp4_test.go`

### Implementation for User Story 3

- [x] T020 [US3] 確認前端 hls.js 不需修改（fMP4 自動支援），在瀏覽器中手動驗證於 `static/index.html`

**Checkpoint**: 所有主流瀏覽器可正常播放 fMP4 串流

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: 清理與驗證

- [x] T021 [P] 更新既有測試中的 .ts 引用改為 .m4s 於 `tests/integration/stream_test.go`、`tests/integration/api_test.go`
- [x] T022 [P] 更新既有測試中的 Content-Type 斷言改為 video/mp4 於 `tests/integration/stream_test.go`
- [x] T023 重建 Docker 映像並驗證於 `Dockerfile`
- [x] T024 執行 quickstart.md 驗證流程

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: 無相依
- **Foundational (Phase 2)**: 相依於 Setup
- **User Story 1 (Phase 3)**: 相依於 Foundational
- **User Story 2 (Phase 4)**: 相依於 Foundational（可與 US1 平行）
- **User Story 3 (Phase 5)**: 相依於 US1（需要可播放的串流來測試瀏覽器相容性）
- **Polish (Phase 6)**: 相依於所有 user story 完成

### Within Each User Story

- 測試 MUST 先寫並確認失敗，然後才開始實作

### Parallel Opportunities

- T008-T011 所有 US1 測試可平行
- T015-T016 所有 US2 測試可平行
- T018-T019 所有 US3 測試可平行
- T021-T022 Polish 任務可平行

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. 完成 Phase 1: Setup
2. 完成 Phase 2: Foundational（核心 fMP4 參數修改）
3. 完成 Phase 3: User Story 1（預生成 fMP4 + handler 修改）
4. **停下驗證**：在瀏覽器中播放，確認無破音

### Incremental Delivery

1. Setup + Foundational → fMP4 合成就緒
2. User Story 1 → 順序播放無破音（MVP!）
3. User Story 2 → 跳轉平順
4. User Story 3 → 瀏覽器相容性確認
5. Polish → 清理與驗證

---

## Notes

- 本功能主要是修改既有程式碼，非新增模組
- .ts 相關的程式碼全部替換為 .m4s
- init.mp4 是新概念，需要新增 handler
