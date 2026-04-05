# Tasks: 影片下載功能

**Input**: Design documents from `/specs/006-video-download/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, contracts/

**Tests**: 包含測試任務（Constitution 要求 TDD）

## Format: `[ID] [P?] [Story] Description`

## Phase 1: Setup

**Purpose**: 無新增相依

- [x] T001 確認既有預生成分段結構（init.mp4 + seg_*.m4s）可用於串接

---

## Phase 2: Foundational

**Purpose**: 下載核心邏輯

- [x] T002 實作 DownloadHandler：讀取 init.mp4 + 所有 .m4s 分段，串流寫入 HTTP response，設定 Content-Type/Content-Disposition/Content-Length 於 `src/handler/download.go`
- [x] T003 在 SetupRouter 中註冊 `GET /download/{id}` 路由於 `src/handler/stream.go`

**Checkpoint**: `/download/{id}` 端點可回傳完整 MP4

---

## Phase 3: User Story 1 — 訪客下載合成影片 (Priority: P1) 🎯 MVP

**Goal**: 訪客在播放頁面可下載完整 MP4

**Independent Test**: 請求 `/download/{id}`，用 ffprobe 驗證回傳的 MP4 格式正確

### Tests for User Story 1 ⚠️

- [x] T004 [P] [US1] 整合測試：GET /download/{id} 回傳有效 MP4（Content-Type、Content-Length、ffprobe 驗證）於 `tests/integration/download_test.go`
- [x] T005 [P] [US1] 整合測試：不存在的 ID 回傳 404 於 `tests/integration/download_test.go`

### Implementation for User Story 1

- [x] T006 [US1] WordPress 播放器 HTML 加入「下載影片」按鈕（連結到 Streamixer download 端點）於 `wordpress/streamixer/includes/class-frontend.php`
- [x] T007 [US1] 下載按鈕的 CSS 樣式於 `wordpress/streamixer/assets/css/player.css`

**Checkpoint**: 訪客可在前台點擊下載並取得可播放的 MP4

---

## Phase 4: User Story 2 — 管理員逐個匯出 (Priority: P2)

**Goal**: 後台編輯頁可匯出影片

### Implementation for User Story 2

- [x] T008 [US2] 在素材編輯頁的 meta box 新增「匯出影片」按鈕（已同步時顯示，連結到 download 端點）於 `wordpress/streamixer/includes/class-cpt.php`

**Checkpoint**: 管理員可在編輯頁一鍵匯出

---

## Phase 5: User Story 3 — 管理員批次匯出 (Priority: P3)

**Goal**: 列表頁可選取多個素材批次下載

### Implementation for User Story 3

- [x] T009 [US3] 註冊批次操作「匯出影片」（WordPress bulk_actions filter）於 `wordpress/streamixer/streamixer.php`
- [x] T010 [US3] 處理批次操作：前端 JS 逐個開啟下載 URL 於 `wordpress/streamixer/streamixer.php`

**Checkpoint**: 管理員可批次匯出多個影片

---

## Phase 6: Polish

- [x] T011 更新 README.md 加入下載端點說明於 `README.md`
- [x] T012 重建 Docker 映像並驗證
- [x] T013 執行 quickstart.md 驗證流程

---

## Dependencies & Execution Order

- **Setup (Phase 1)**: 無相依
- **Foundational (Phase 2)**: 相依於 Setup
- **US1 (Phase 3)**: 相依於 Foundational
- **US2 (Phase 4)**: 相依於 Foundational
- **US3 (Phase 5)**: 相依於 US2
- **Polish (Phase 6)**: 相依於全部完成

---

## Implementation Strategy

### MVP First (User Story 1)

1. Foundational → DownloadHandler + 路由
2. User Story 1 → 前台下載按鈕
3. **停下驗證**：下載 MP4 並用 ffprobe 和播放器確認

### Incremental Delivery

1. US1 → 訪客可下載（MVP!）
2. US2 → 後台逐個匯出
3. US3 → 批次匯出
