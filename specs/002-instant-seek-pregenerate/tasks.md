# Tasks: 即時跳轉與背景預生成

**Input**: Design documents from `/specs/002-instant-seek-pregenerate/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: 包含測試任務（Constitution 要求 TDD）

**Organization**: 任務按 user story 分組，確保每個 story 可獨立實作與測試。

## Format: `[ID] [P?] [Story] Description`

- **[P]**: 可平行執行（不同檔案、無相依）
- **[Story]**: 所屬 user story（US1、US2、US3）
- 包含確切的檔案路徑

## Phase 1: Setup

**Purpose**: 新增相依套件

- [x] T001 加入 golang.org/x/sync 相依（`go.mod`）

---

## Phase 2: Foundational（阻塞性前置作業）

**Purpose**: 預生成管理器核心邏輯，所有 user story 共用

**⚠️ CRITICAL**: 所有 user story 必須等此階段完成才能開始

- [x] T002 定義 PregenTask 資料結構（compositionID、status、totalSegments、completedSegments、outputDir、error）於 `src/composer/pregen.go`
- [x] T003 實作 PregenManager（管理預生成任務的啟動、查詢、狀態追蹤）於 `src/composer/pregen.go`
- [x] T004 實作 PregenManager.StartPregen()：背景啟動 FFmpeg 完整 HLS 合成（使用既有 BuildFFmpegArgs），以 singleflight 防止重複啟動，於 `src/composer/pregen.go`
- [x] T005 實作 PregenManager.IsSegmentReady()：檢查特定分段是否已預生成（檔案存在且大小 > 0）於 `src/composer/pregen.go`
- [x] T006 實作 PregenManager.GetStatus()：回傳預生成任務狀態，於 `src/composer/pregen.go`

**Checkpoint**: PregenManager 核心邏輯完成，可被 handler 呼叫

---

## Phase 3: User Story 1 — 順序播放不中斷 (Priority: P1) 🎯 MVP

**Goal**: 首次請求觸發背景預生成，分段優先使用預生成結果，順序播放流暢

**Independent Test**: 播放 30 秒以上合成串流，驗證分段切換無中斷

### Tests for User Story 1 ⚠️

> **NOTE: 先寫這些測試，確認測試失敗後才開始實作**

- [x] T007 [P] [US1] 單元測試：PregenManager 啟動預生成後可查詢狀態於 `tests/unit/pregen_test.go`
- [x] T008 [P] [US1] 單元測試：singleflight 確保同一素材不重複啟動預生成於 `tests/unit/pregen_test.go`
- [x] T009 [P] [US1] 整合測試：請求 playlist 後背景預生成啟動，分段逐漸就緒，順序請求全部成功於 `tests/integration/pregen_test.go`

### Implementation for User Story 1

- [x] T010 [US1] 修改 PlaylistHandler：首次請求時呼叫 PregenManager.StartPregen() 於 `src/handler/stream.go`
- [x] T011 [US1] 修改 SegmentHandler：優先檢查 IsSegmentReady()，已就緒則直接回傳檔案，否則按需生成於 `src/handler/stream.go`
- [x] T012 [US1] 將 PregenManager 注入 StreamHandler，於 `src/handler/stream.go` 與 `src/main.go`

**Checkpoint**: 順序播放時大部分分段由預生成提供，切換無中斷

---

## Phase 4: User Story 2 — 跳轉後音訊準確對齊 (Priority: P2)

**Goal**: 按需生成的分段時間戳與 playlist 對齊，跳轉後音訊準確

**Independent Test**: 跳轉到多個時間點，用 ffprobe 驗證每個分段的 PTS 與 playlist 一致

### Tests for User Story 2 ⚠️

> **NOTE: 先寫這些測試，確認測試失敗後才開始實作**

- [x] T013 [P] [US2] 單元測試：BuildSegmentArgs 產生的 FFmpeg 參數包含正確的 -output_ts_offset 於 `tests/unit/composer_test.go`
- [x] T014 [P] [US2] 整合測試：按需生成分段 0 和分段 1 後，用 ffprobe 驗證 PTS 連續（分段 1 的起始 PTS ≈ 6 秒）於 `tests/integration/seek_test.go`
- [x] T015 [P] [US2] 整合測試：跳轉到中間分段後音訊內容正確（非從頭開始）於 `tests/integration/seek_test.go`

### Implementation for User Story 2

- [x] T016 [US2] 修正 BuildSegmentArgs：加入 -output_ts_offset 參數，設定為 segIndex * segDuration 於 `src/composer/ffmpeg.go`
- [x] T017 [US2] 修正 BuildSegmentArgs 的 -ss 行為：確保音訊輸入的 seek 位置精確於 `src/composer/ffmpeg.go`

**Checkpoint**: 跳轉到任意位置後，音訊與 playlist 時間戳對齊

---

## Phase 5: User Story 3 — 背景預生成不影響服務可用性 (Priority: P3)

**Goal**: 多素材並發預生成不拖垮服務，失敗時自動回退

**Independent Test**: 同時請求 3 個不同素材，驗證全部能正常播放

### Tests for User Story 3 ⚠️

> **NOTE: 先寫這些測試，確認測試失敗後才開始實作**

- [x] T018 [P] [US3] 整合測試：同時請求 3 個不同素材的 playlist，全部成功回應於 `tests/integration/pregen_test.go`
- [x] T019 [P] [US3] 整合測試：預生成失敗時（模擬 FFmpeg 錯誤），按需生成仍可正常運作於 `tests/integration/pregen_test.go`

### Implementation for User Story 3

- [x] T020 [US3] PregenManager 加入錯誤處理：FFmpeg 失敗時標記 status=failed，後續請求回退按需生成於 `src/composer/pregen.go`
- [x] T021 [US3] 新增預生成相關設定（最大並發預生成數）於 `src/config/config.go`
- [x] T022 [US3] PregenManager 加入並發限制：使用 semaphore 限制同時進行的預生成任務數量於 `src/composer/pregen.go`

**Checkpoint**: 多素材並發預生成穩定，失敗時自動回退

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: 跨 user story 的改善

- [x] T023 [P] 更新 Dockerfile 加入 golang.org/x/sync 相依於 `Dockerfile`（go mod download 自動處理）
- [x] T024 [P] 更新既有整合測試適配新的 StreamHandler 建構方式於 `tests/integration/stream_test.go` 與 `tests/integration/api_test.go`
- [x] T025 程式碼清理與重構
- [x] T026 執行 quickstart.md 驗證流程

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: 無相依——可立即開始
- **Foundational (Phase 2)**: 相依於 Setup 完成——阻塞所有 user stories
- **User Story 1 (Phase 3)**: 相依於 Foundational 完成
- **User Story 2 (Phase 4)**: 相依於 Foundational 完成（可與 US1 平行）
- **User Story 3 (Phase 5)**: 相依於 US1 完成（需要可運行的預生成來測試並發）
- **Polish (Phase 6)**: 相依於所有 user story 完成

### Within Each User Story

- 測試 MUST 先寫並確認失敗，然後才開始實作
- 資料結構 → 核心邏輯 → Handler 整合

### Parallel Opportunities

- T007-T009 所有 US1 測試可平行
- T013-T015 所有 US2 測試可平行
- T018-T019 所有 US3 測試可平行
- T023-T024 所有 Polish 任務可平行

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. 完成 Phase 1: Setup
2. 完成 Phase 2: Foundational（PregenManager 核心）
3. 完成 Phase 3: User Story 1（背景預生成 + 優先回傳）
4. **停下驗證**：播放長音檔，確認順序播放無中斷
5. 可先行部署/展示

### Incremental Delivery

1. Setup + Foundational → PregenManager 就緒
2. User Story 1 → 順序播放流暢（MVP!）
3. User Story 2 → 跳轉音訊對齊
4. User Story 3 → 並發穩定性
5. Polish → 清理與驗證

---

## Notes

- [P] 任務 = 不同檔案、無相依
- [Story] 標籤對應 spec.md 中的 user story
- 每個 user story 可獨立完成與測試
- 驗證測試失敗後再開始實作
- 每完成一個任務或邏輯群組後 commit
