# Tasks: WordPress 外掛

**Input**: Design documents from `/specs/005-wordpress-plugin/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: 包含測試任務（Constitution 要求 TDD）

**Organization**: 任務按 user story 分組。外掛原始碼位於 `wordpress/streamixer/`。

## Format: `[ID] [P?] [Story] Description`

- **[P]**: 可平行執行（不同檔案、無相依）
- **[Story]**: 所屬 user story（US1-US5）
- 包含確切的檔案路徑

## Phase 1: Setup

**Purpose**: 建立外掛目錄結構與基礎設定

- [x] T001 建立外掛目錄結構：`wordpress/streamixer/`、`includes/`、`src/`、`assets/css/`、`assets/js/`、`templates/`
- [x] T002 建立主外掛檔案（plugin header、版本、bootstrap）於 `wordpress/streamixer/streamixer.php`
- [x] T003 建立 `package.json`，設定 `@wordpress/scripts` 建置工具於 `wordpress/streamixer/package.json`

---

## Phase 2: Foundational（阻塞性前置作業）

**Purpose**: Custom Post Type、分類法、共用元件

**⚠️ CRITICAL**: 所有 user story 必須等此階段完成才能開始

- [x] T004 實作 Custom Post Type 註冊（`streamixer` CPT，支援 title、editor、thumbnail）於 `wordpress/streamixer/includes/class-cpt.php`
- [x] T005 [P] 實作自訂分類法（`streamixer_category` 階層式、`streamixer_tag` 非階層式）於 `wordpress/streamixer/includes/class-cpt.php`
- [x] T006 [P] 實作 Streamixer HTTP 通訊 class（`wp_remote_post` 上傳素材）於 `wordpress/streamixer/includes/class-api.php`
- [x] T007 [P] 實作前端 asset enqueue class（條件式載入 hls.js 和播放器 CSS/JS）於 `wordpress/streamixer/includes/class-frontend.php`

**Checkpoint**: CPT 可在後台新增，分類法可用，Streamixer 通訊可呼叫

---

## Phase 3: User Story 1 — 後台上傳與管理素材 (Priority: P1) 🎯 MVP

**Goal**: 管理員可在後台新增、編輯、刪除素材組合，檔案自動同步至 Streamixer

**Independent Test**: 在後台新增素材並上傳音檔+圖片，確認儲存成功且同步至 Streamixer

### Tests for User Story 1 ⚠️

- [x] T008 [P] [US1] 單元測試：CPT 註冊成功，post type 存在於 `wordpress/streamixer/tests/test-cpt.php`
- [x] T009 [P] [US1] 單元測試：save_post hook 呼叫 API 同步於 `wordpress/streamixer/tests/test-api.php`

### Implementation for User Story 1

- [x] T010 [US1] 實作 meta box：音檔上傳欄位（使用 `wp_enqueue_media()` + Media Library）於 `wordpress/streamixer/includes/class-cpt.php`
- [x] T011 [US1] 實作 meta box：背景圖片上傳欄位於 `wordpress/streamixer/includes/class-cpt.php`
- [x] T012 [US1] 實作 meta box：字幕檔上傳欄位（選填）於 `wordpress/streamixer/includes/class-cpt.php`
- [x] T013 [US1] 實作 meta box 的 save 邏輯（nonce 驗證、儲存 attachment ID 到 post meta）於 `wordpress/streamixer/includes/class-cpt.php`
- [x] T014 [US1] 實作 save_post_streamixer hook：將素材檔案同步至 Streamixer 服務於 `wordpress/streamixer/includes/class-api.php`
- [x] T015 [US1] 在 meta box 中顯示同步狀態（synced / pending / error）於 `wordpress/streamixer/includes/class-cpt.php`
- [x] T016 [US1] 實作刪除素材時通知 Streamixer 清除對應快取於 `wordpress/streamixer/includes/class-api.php`

**Checkpoint**: 管理員可在後台完整管理素材，自動同步至 Streamixer

---

## Phase 4: User Story 2 — Shortcode 嵌入播放器 (Priority: P2)

**Goal**: 在文章中用 `[streamixer id="..."]` 嵌入 HLS 播放器

**Independent Test**: 在文章中插入 Shortcode，前台確認播放器正常顯示並播放

### Tests for User Story 2 ⚠️

- [x] T017 [P] [US2] 單元測試：Shortcode 註冊成功，回傳包含 video 元素的 HTML 於 `wordpress/streamixer/tests/test-shortcode.php`

### Implementation for User Story 2

- [x] T018 [US2] 實作 Shortcode 註冊（`[streamixer id="..."]`）於 `wordpress/streamixer/includes/class-shortcode.php`
- [x] T019 [US2] 實作 Shortcode 渲染：輸出播放器 HTML（video 元素 + data-hls-url 屬性）於 `wordpress/streamixer/includes/class-shortcode.php`
- [x] T020 [US2] 實作前端播放器 JS：初始化 hls.js、綁定 video 元素、錯誤處理於 `wordpress/streamixer/assets/js/player.js`
- [x] T021 [US2] 實作播放器 CSS：響應式容器、控制列樣式於 `wordpress/streamixer/assets/css/player.css`
- [x] T022 [US2] Shortcode 中條件式 enqueue hls.js 和 player.js/css 於 `wordpress/streamixer/includes/class-shortcode.php`

**Checkpoint**: Shortcode 嵌入的播放器可在前台正常播放

---

## Phase 5: User Story 3 — Gutenberg Block 嵌入 (Priority: P3)

**Goal**: 在 Gutenberg 編輯器中視覺化嵌入播放器

**Independent Test**: 在編輯器中新增 Streamixer Block，選取素材，前台正常播放

### Tests for User Story 3 ⚠️

- [x] T023 [P] [US3] 測試：Block 在前台渲染包含播放器 HTML 於 `wordpress/streamixer/tests/test-block.php`

### Implementation for User Story 3

- [x] T024 [US3] 建立 block.json（名稱、屬性、editorScript、viewScript）於 `wordpress/streamixer/src/block.json`
- [x] T025 [US3] 實作 index.js：Block 註冊入口於 `wordpress/streamixer/src/index.js`
- [x] T026 [US3] 實作 edit.js：編輯器中的素材選擇介面（ComboboxControl 搜尋已建立的素材）於 `wordpress/streamixer/src/edit.js`
- [x] T027 [US3] 實作 render.php：動態前台渲染（與 Shortcode 共用播放器 HTML）於 `wordpress/streamixer/src/render.php`
- [x] T028 [US3] 在主外掛檔案中以 `register_block_type()` 註冊 Block 於 `wordpress/streamixer/streamixer.php`
- [x] T029 [US3] 執行 `npm run build` 產生 build/ 目錄

**Checkpoint**: Gutenberg Block 可在編輯器中選取素材，前台正常播放

---

## Phase 6: User Story 4 — 素材列表頁與播放頁 (Priority: P4)

**Goal**: 前台有獨立的列表頁和播放頁

**Independent Test**: 瀏覽列表頁、篩選分類、點擊進入播放頁，手機與桌面都正常

### Implementation for User Story 4

- [x] T030 [US4] 實作 archive 模板（列表頁：依日期排列、分類篩選、搜尋）於 `wordpress/streamixer/templates/archive-streamixer.php`
- [x] T031 [US4] 實作 single 模板（播放頁：播放器 + 標題 + 描述 + 分類 + 標籤）於 `wordpress/streamixer/templates/single-streamixer.php`
- [x] T032 [US4] 實作模板載入（`single_template` + `archive_template` filter）於 `wordpress/streamixer/includes/class-frontend.php`
- [x] T033 [US4] 實作列表頁與播放頁的響應式 CSS 於 `wordpress/streamixer/assets/css/player.css`

**Checkpoint**: 前台列表頁和播放頁在各裝置上正常顯示

---

## Phase 7: User Story 5 — 外掛設定頁 (Priority: P5)

**Goal**: 管理員可設定 Streamixer URL 和預設背景圖片

**Independent Test**: 變更 Streamixer URL，確認後續操作使用新的位址

### Implementation for User Story 5

- [x] T034 [US5] 實作設定頁（Settings API：`add_options_page`、`register_setting`、`add_settings_field`）於 `wordpress/streamixer/includes/class-settings.php`
- [x] T035 [US5] 實作 Streamixer URL 設定欄位於 `wordpress/streamixer/includes/class-settings.php`
- [x] T036 [US5] 實作預設背景圖片設定欄位（使用 Media Library picker）於 `wordpress/streamixer/includes/class-settings.php`
- [x] T037 [US5] 在主外掛檔案中載入 Settings class 於 `wordpress/streamixer/streamixer.php`

**Checkpoint**: 設定頁可正常儲存與讀取設定

---

## Phase 8: Polish & Cross-Cutting Concerns

- [x] T038 [P] Streamixer 服務不可達時的友善錯誤訊息（播放器和 meta box）於 `wordpress/streamixer/assets/js/player.js` 與 `wordpress/streamixer/includes/class-api.php`
- [x] T039 [P] 外掛停用時的清理（flush rewrite rules）於 `wordpress/streamixer/streamixer.php`
- [x] T040 程式碼清理與 WordPress Coding Standards 檢查
- [x] T041 執行 quickstart.md 驗證流程

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: 無相依
- **Foundational (Phase 2)**: 相依於 Setup
- **US1 後台管理 (Phase 3)**: 相依於 Foundational
- **US2 Shortcode (Phase 4)**: 相依於 US1（需要素材可用）
- **US3 Gutenberg Block (Phase 5)**: 相依於 US2（共用播放器渲染邏輯）
- **US4 列表/播放頁 (Phase 6)**: 相依於 US2（共用播放器 CSS/JS）
- **US5 設定頁 (Phase 7)**: 相依於 Foundational（可與 US1 平行）
- **Polish (Phase 8)**: 相依於所有 user story 完成

### Within Each User Story

- 測試 MUST 先寫並確認失敗，然後才開始實作

### Parallel Opportunities

- T005-T007 Foundational 任務可平行
- T008-T009 US1 測試可平行
- T038-T039 Polish 任務可平行

---

## Implementation Strategy

### MVP First (User Story 1 + 2)

1. Setup + Foundational → 外掛骨架
2. User Story 1 → 後台可管理素材（MVP 基礎!）
3. User Story 2 → Shortcode 可播放（MVP 完成!）
4. **停下驗證**：在 WordPress 中上傳素材並用 Shortcode 播放

### Incremental Delivery

1. US1 + US2 → 管理 + 播放（MVP!）
2. US3 → Gutenberg Block
3. US4 → 列表/播放頁
4. US5 → 設定頁
5. Polish → 錯誤處理、清理

---

## Notes

- 外掛位於 `wordpress/streamixer/`，與 Go 服務分開
- Gutenberg Block 需要 `npm run build` 建置
- 所有 PHP 檔案遵循 WordPress Coding Standards
- 播放器渲染邏輯在 Shortcode 和 Block 之間共用
