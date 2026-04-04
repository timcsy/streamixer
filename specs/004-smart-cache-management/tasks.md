# Tasks: 智慧快取管理

**Input**: Design documents from `/specs/004-smart-cache-management/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: 包含測試任務（Constitution 要求 TDD）

**Organization**: 任務按 user story 分組。

## Format: `[ID] [P?] [Story] Description`

- **[P]**: 可平行執行（不同檔案、無相依）
- **[Story]**: 所屬 user story（US1、US2、US3）
- 包含確切的檔案路徑

## Phase 1: Setup

**Purpose**: 新增快取相關設定

- [x] T001 新增快取設定：CacheTTL、CacheMaxSize、CacheSweepInterval 於 `src/config/config.go`

---

## Phase 2: Foundational（阻塞性前置作業）

**Purpose**: CacheManager 核心資料結構

**⚠️ CRITICAL**: 所有 user story 必須等此階段完成才能開始

- [x] T002 定義 CacheEntry 資料結構（compositionID、lastAccessed、size、active）於 `src/composer/cache.go`
- [x] T003 實作 CacheManager（管理 CacheEntry map、RWMutex 保護、Touch 更新存取時間）於 `src/composer/cache.go`
- [x] T004 實作 CacheManager.CalculateUsage()：掃描 tmpfs 目錄計算各素材佔用空間於 `src/composer/cache.go`

**Checkpoint**: CacheManager 核心邏輯完成

---

## Phase 3: User Story 1 — 冷門內容自動清除 (Priority: P1) 🎯 MVP

**Goal**: 超過 TTL 未被存取的素材自動從 tmpfs 清除

**Independent Test**: 播放素材，等待 TTL 過期，確認分段被清除，重新播放正常

### Tests for User Story 1 ⚠️

> **NOTE: 先寫這些測試，確認測試失敗後才開始實作**

- [x] T005 [P] [US1] 單元測試：CacheManager.Touch 更新 lastAccessed 時間於 `tests/unit/cache_test.go`
- [x] T006 [P] [US1] 單元測試：CacheManager.SweepExpired 清除超過 TTL 的條目於 `tests/unit/cache_test.go`
- [x] T007 [P] [US1] 單元測試：SweepExpired 不清除 active=true（正在預生成）的條目於 `tests/unit/cache_test.go`
- [x] T008 [P] [US1] 整合測試：分段檔案在 TTL 過期後被實際刪除，重新請求觸發預生成於 `tests/integration/cache_test.go`

### Implementation for User Story 1

- [x] T009 [US1] 實作 CacheManager.SweepExpired()：清除 active=false 且 lastAccessed + TTL < now 的條目，刪除對應目錄於 `src/composer/cache.go`
- [x] T010 [US1] 實作 Sweeper：背景 goroutine 定期呼叫 SweepExpired 於 `src/composer/sweeper.go`
- [x] T011 [US1] 修改 PregenManager：預生成啟動時標記 active=true，完成/失敗時標記 active=false，清除時同步清除任務狀態於 `src/composer/pregen.go`
- [x] T012 [US1] 修改 SegmentHandler 和 InitHandler：每次存取呼叫 CacheManager.Touch 更新時間於 `src/handler/stream.go`
- [x] T013 [US1] 在 main.go 初始化 CacheManager 和 Sweeper，注入 StreamHandler 於 `src/main.go`

**Checkpoint**: 冷門素材在 TTL 後自動清除，重新請求正常

---

## Phase 4: User Story 2 — 暫存容量保護 (Priority: P2)

**Goal**: 暫存接近上限時主動淘汰最久未存取的內容

**Independent Test**: 設定小容量上限，連續播放多個素材，確認舊內容被淘汰

### Tests for User Story 2 ⚠️

- [x] T014 [P] [US2] 單元測試：CacheManager.SweepByCapacity 在超過 90% 時淘汰最久的條目於 `tests/unit/cache_test.go`
- [x] T015 [P] [US2] 單元測試：SweepByCapacity 不淘汰 active=true 的條目於 `tests/unit/cache_test.go`

### Implementation for User Story 2

- [x] T016 [US2] 實作 CacheManager.SweepByCapacity()：按 lastAccessed 升序淘汰直到低於 90% 於 `src/composer/cache.go`
- [x] T017 [US2] 修改 Sweeper：在 SweepExpired 後呼叫 SweepByCapacity 於 `src/composer/sweeper.go`

**Checkpoint**: 暫存不超過容量上限

---

## Phase 5: User Story 3 — 熱門內容共享快取 (Priority: P3)

**Goal**: 多人存取同一素材共享快取，存取延長壽命

**Independent Test**: 兩人先後請求同一素材，確認不重複預生成且存取時間更新

### Tests for User Story 3 ⚠️

- [x] T018 [P] [US3] 整合測試：素材已預生成完成後再次請求 playlist，不觸發新的預生成於 `tests/integration/cache_test.go`
- [x] T019 [P] [US3] 整合測試：素材被第二人存取後 lastAccessed 更新，TTL 重新計時於 `tests/integration/cache_test.go`

### Implementation for User Story 3

- [x] T020 [US3] 修改 PlaylistHandler：已有快取且預生成完成時不重新啟動預生成於 `src/handler/stream.go`

**Checkpoint**: 熱門內容共享快取，存取延長壽命

---

## Phase 6: Polish & Cross-Cutting Concerns

- [x] T021 [P] 新增環境變數解析（CACHE_TTL、CACHE_MAX_SIZE、CACHE_SWEEP_INTERVAL 支援時間/大小格式）於 `src/config/config.go`
- [x] T022 程式碼清理與重構
- [x] T023 重建 Docker 映像並驗證於 `Dockerfile`
- [x] T024 執行 quickstart.md 驗證流程

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: 無相依
- **Foundational (Phase 2)**: 相依於 Setup
- **User Story 1 (Phase 3)**: 相依於 Foundational
- **User Story 2 (Phase 4)**: 相依於 US1（SweepByCapacity 在 SweepExpired 之後呼叫）
- **User Story 3 (Phase 5)**: 相依於 US1（需要 Touch 機制）
- **Polish (Phase 6)**: 相依於所有 user story 完成

### Within Each User Story

- 測試 MUST 先寫並確認失敗，然後才開始實作

### Parallel Opportunities

- T005-T008 所有 US1 測試可平行
- T014-T015 所有 US2 測試可平行
- T018-T019 所有 US3 測試可平行

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. 完成 Phase 1: Setup
2. 完成 Phase 2: Foundational
3. 完成 Phase 3: User Story 1（TTL 清掃）
4. **停下驗證**：設定短 TTL 觀察清掃行為

### Incremental Delivery

1. Setup + Foundational → CacheManager 就緒
2. User Story 1 → TTL 自動清除（MVP!）
3. User Story 2 → 容量保護
4. User Story 3 → 共享快取
5. Polish → 清理與驗證

---

## Notes

- 清掃以素材組合為單位（整組清除）
- CacheManager 用 RWMutex 保護並發安全
- 正在預生成的素材不可被清掃
