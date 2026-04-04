# Tasks: HLS 即時合成串流

**Input**: Design documents from `/specs/001-hls-stream-synthesis/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: 包含測試任務（Constitution 要求 TDD）

**Organization**: 任務按 user story 分組，確保每個 story 可獨立實作與測試。

## Format: `[ID] [P?] [Story] Description`

- **[P]**: 可平行執行（不同檔案、無相依）
- **[Story]**: 所屬 user story（US1、US2、US3）
- 包含確切的檔案路徑

## Phase 1: Setup（專案初始化）

**Purpose**: 建立專案結構與基礎設定

- [x] T001 初始化 Go module，建立 go.mod（`go.mod`）
- [x] T002 [P] 建立專案目錄結構：`src/handler/`、`src/composer/`、`src/media/`、`src/config/`、`tests/unit/`、`tests/integration/`
- [x] T003 [P] 加入 go-chi/chi v5 相依（`go.mod`）
- [x] T004 [P] 建立 Dockerfile 基礎版本，包含 Go build stage 與 FFmpeg runtime（`Dockerfile`）

---

## Phase 2: Foundational（阻塞性前置作業）

**Purpose**: 所有 user story 共用的核心基礎設施

**⚠️ CRITICAL**: 所有 user story 必須等此階段完成才能開始

- [x] T005 定義 MediaComposition 資料結構（id、audio、background、subtitle）於 `src/media/composition.go`
- [x] T006 [P] 定義 Audio、Background、Subtitle 資料結構與格式驗證於 `src/media/types.go`
- [x] T007 [P] 實作設定管理（媒體目錄路徑、伺服器 port、分段秒數）於 `src/config/config.go`
- [x] T008 實作素材載入與驗證邏輯（檢查檔案存在、格式支援）於 `src/media/loader.go`
- [x] T009 實作錯誤回應格式（統一 JSON 錯誤格式）於 `src/handler/error.go`
- [x] T010 實作 health check handler 於 `src/handler/health.go`
- [x] T011 建立 HTTP server 與路由設定（chi router）於 `src/main.go`

**Checkpoint**: 基礎建設完成，可啟動空服務並回應 /health

---

## Phase 3: User Story 1 — 播放合成影片串流 (Priority: P1) 🎯 MVP

**Goal**: 音檔 + 背景圖片即時合成為可播放的 HLS 串流

**Independent Test**: 提供音檔和背景圖片，請求 HLS playlist 後用 ffprobe 驗證串流可播放

### Tests for User Story 1 ⚠️

> **NOTE: 先寫這些測試，確認測試失敗後才開始實作**

- [x] T012 [P] [US1] 單元測試：playlist 產生邏輯（根據音檔長度計算分段數與時長）於 `tests/unit/playlist_test.go`
- [x] T013 [P] [US1] 單元測試：FFmpeg 指令組裝邏輯（音檔+圖片合成、無字幕情境）於 `tests/unit/composer_test.go`
- [x] T014 [P] [US1] 整合測試：完整合成流程（請求 playlist → 請求分段 → ffprobe 驗證）於 `tests/integration/stream_test.go`
- [x] T015 [P] [US1] 契約測試：GET /stream/{id}/index.m3u8 回傳有效 HLS playlist 於 `tests/integration/api_test.go`
- [x] T016 [P] [US1] 契約測試：GET /stream/{id}/seg_{n}.ts 回傳有效 MPEG-TS 分段於 `tests/integration/api_test.go`

### Implementation for User Story 1

- [x] T017 [US1] 實作 playlist 產生邏輯（根據音檔 duration 計算 .m3u8 內容）於 `src/composer/playlist.go`
- [x] T018 [US1] 實作 FFmpeg 合成引擎（音檔+圖片 → HLS 分段，寫入 tmpfs）於 `src/composer/ffmpeg.go`
- [x] T019 [US1] 實作音檔 duration 讀取（透過 ffprobe）於 `src/media/probe.go`
- [x] T020 [US1] 實作 playlist handler（GET /stream/{id}/index.m3u8）於 `src/handler/stream.go`
- [x] T021 [US1] 實作 segment handler（GET /stream/{id}/seg_{n}.ts）於 `src/handler/stream.go`
- [x] T022 [US1] 實作 seek 支援（根據分段編號以 -ss/-t 參數按需生成對應分段）於 `src/composer/ffmpeg.go`
- [x] T023 [US1] 整合路由：將 playlist 與 segment handler 掛載至 chi router 於 `src/main.go`

**Checkpoint**: 此時 User Story 1 應可完整運作——音檔+圖片可合成為 HLS 串流播放

---

## Phase 4: User Story 2 — 顯示字幕 (Priority: P2)

**Goal**: 將字幕燒入合成影片中，與音訊同步顯示

**Independent Test**: 提供音檔、圖片和字幕檔，用 ffprobe 驗證輸出影片含有字幕畫面

### Tests for User Story 2 ⚠️

> **NOTE: 先寫這些測試，確認測試失敗後才開始實作**

- [x] T024 [P] [US2] 單元測試：FFmpeg 指令組裝含字幕參數（-vf subtitles=...）於 `tests/unit/subtitle_test.go`
- [x] T025 [P] [US2] 整合測試：含字幕合成流程（驗證輸出分段含字幕渲染）於 `tests/integration/subtitle_test.go`
- [x] T026 [P] [US2] 整合測試：無字幕時仍正常播放（回歸測試）於 `tests/integration/subtitle_test.go`

### Implementation for User Story 2

- [x] T027 [US2] 實作字幕檔驗證與格式偵測（SRT、VTT）於 `src/media/subtitle.go`
- [x] T028 [US2] 擴展 FFmpeg 合成引擎，加入字幕燒入邏輯（-vf "subtitles=..."）於 `src/composer/ffmpeg.go`
- [x] T029 [US2] 處理字幕檔不存在的情境（略過字幕、正常合成）於 `src/composer/ffmpeg.go`

**Checkpoint**: User Story 1 和 2 皆可獨立運作——有字幕時顯示，無字幕時正常播放

---

## Phase 5: User Story 3 — 容器化部署 (Priority: P3)

**Goal**: 完整的 Docker 容器化部署，管理員可快速啟動服務

**Independent Test**: 執行 docker build + docker run，驗證服務啟動並可處理請求

### Tests for User Story 3 ⚠️

> **NOTE: 先寫這些測試，確認測試失敗後才開始實作**

- [x] T030 [P] [US3] 整合測試：Docker 容器啟動後 health check 回應 200 於 `tests/integration/docker_test.go`
- [x] T031 [P] [US3] 整合測試：容器掛載外部素材目錄後可正常合成串流於 `tests/integration/docker_test.go`

### Implementation for User Story 3

- [x] T032 [US3] 完善 Dockerfile：多階段建置、安裝 FFmpeg、設定 tmpfs、EXPOSE port 於 `Dockerfile`
- [x] T033 [US3] 建立 docker-compose.yml（含 volume mount 與 tmpfs 設定）於 `docker-compose.yml`
- [x] T034 [US3] 建立 .dockerignore 排除非必要檔案（specs/、tests/、.specify/）於 `.dockerignore`

**Checkpoint**: 所有 user story 皆可完整運作，服務可透過 Docker 部署

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: 跨 user story 的改善

- [x] T035 [P] 素材格式錯誤的錯誤訊息完善（列出支援格式）於 `src/media/loader.go`
- [x] T036 [P] 背景圖片自動縮放至輸出解析度（720p）於 `src/composer/ffmpeg.go`
- [x] T037 [P] 多使用者並發請求的 tmpfs 隔離（每個請求獨立目錄）於 `src/composer/ffmpeg.go`
- [x] T038 程式碼清理與重構
- [x] T039 執行 quickstart.md 驗證流程

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: 無相依——可立即開始
- **Foundational (Phase 2)**: 相依於 Setup 完成——阻塞所有 user stories
- **User Story 1 (Phase 3)**: 相依於 Foundational 完成
- **User Story 2 (Phase 4)**: 相依於 Foundational 完成（可與 US1 平行，但建議先完成 US1）
- **User Story 3 (Phase 5)**: 相依於 US1 完成（需要可運行的服務來測試 Docker）
- **Polish (Phase 6)**: 相依於所有 user story 完成

### Within Each User Story

- 測試 MUST 先寫並確認失敗，然後才開始實作
- Model → Service → Handler 順序
- 核心邏輯優先於整合

### Parallel Opportunities

- T002、T003、T004 可平行（Setup 階段）
- T006、T007 可平行（Foundational 階段）
- T012-T016 所有 US1 測試可平行
- T024-T026 所有 US2 測試可平行
- T030-T031 所有 US3 測試可平行
- T035-T037 所有 Polish 任務可平行

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. 完成 Phase 1: Setup
2. 完成 Phase 2: Foundational
3. 完成 Phase 3: User Story 1（寫測試 → 測試失敗 → 實作 → 測試通過）
4. **停下驗證**：用 ffprobe 與播放器確認 HLS 串流可正常播放
5. 可先行部署/展示

### Incremental Delivery

1. Setup + Foundational → 基礎就緒
2. User Story 1 → 音檔+圖片串流可播放（MVP!）
3. User Story 2 → 加入字幕支援
4. User Story 3 → Docker 容器化部署
5. Polish → 錯誤處理、縮放、並發

---

## Notes

- [P] 任務 = 不同檔案、無相依
- [Story] 標籤對應 spec.md 中的 user story
- 每個 user story 可獨立完成與測試
- 驗證測試失敗後再開始實作
- 每完成一個任務或邏輯群組後 commit
- 在任何 checkpoint 停下來驗證 story 是否獨立可用
