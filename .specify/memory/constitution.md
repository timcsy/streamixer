<!-- Sync Impact Report
版本變更：無 → 1.0.0
新增原則：
  - I. 測試驅動開發（TDD）
  - II. 模組化架構
  - III. 即時性優先
  - IV. 簡約原則
新增段落：
  - 開發約束
  - 品質門檻
治理段落已填入
模板同步狀態：
  - .specify/templates/plan-template.md ✅ 無需修改（Constitution Check 段落為動態填入）
  - .specify/templates/spec-template.md ✅ 無需修改（結構與原則相容）
  - .specify/templates/tasks-template.md ✅ 無需修改（已支援測試優先流程）
後續 TODO：無
-->

# Streamixer Constitution

## Core Principles

### I. 測試驅動開發（不可妥協）

所有功能開發 MUST 遵循 TDD 流程：

1. 先撰寫測試，確認測試失敗（Red）
2. 撰寫最小實作使測試通過（Green）
3. 重構程式碼，確保測試持續通過（Refactor）

- 任何新功能或修正 MUST 先有對應的失敗測試，才能開始實作
- 測試 MUST 涵蓋單元測試與整合測試
- 串流處理邏輯 MUST 包含端對端測試，驗證音檔、影像、文字的合成結果

### II. 模組化架構

音檔、影像、文字處理 MUST 為獨立模組，各自可單獨測試與替換。

- 每個媒體類型的處理器 MUST 實作統一的串流介面
- 模組之間 MUST 透過明確定義的介面溝通，禁止直接相依內部實作
- 合成引擎 MUST 與各媒體處理器解耦，透過串流介面組合

### III. 即時性優先

作為即時串流服務，低延遲與穩定性 MUST 為首要考量。

- 串流管線 MUST 支援背壓（backpressure）機制，避免記憶體溢出
- 處理流程 SHOULD 優先使用串流式處理，避免將完整資料載入記憶體
- 效能相關的變更 MUST 附帶基準測試數據佐證

### IV. 簡約原則

避免過度設計，從最簡單可行的方案開始。

- MUST 遵循 YAGNI（You Ain't Gonna Need It）原則，不實作尚未確定的需求
- 新增抽象層 MUST 有至少兩個以上的具體使用場景佐證其必要性
- 優先選擇組合而非繼承

## 開發約束

- 所有規格文件與程式碼註解 MUST 使用繁體中文撰寫
- 每個 commit SHOULD 對應單一邏輯變更
- 相依套件的引入 MUST 評估其維護狀態與授權條款

## 品質門檻

- 所有測試 MUST 在合併前通過
- 新功能 MUST 包含對應的測試，測試覆蓋率不得低於既有水準
- 串流相關功能 MUST 通過壓力測試，確認在持續輸入下無記憶體洩漏

## Governance

本 Constitution 為 Streamixer 專案的最高指導原則，所有開發實踐 MUST 遵循本文件。

- **修訂程序**：任何原則的修改 MUST 記錄變更理由，並更新版本號
- **版本政策**：採用語意化版本（Semantic Versioning）——MAJOR 為不相容的原則移除或重新定義，MINOR 為新增原則或大幅擴充，PATCH 為措辭修正
- **合規審查**：所有 PR 與程式碼審查 MUST 驗證是否符合本 Constitution

**Version**: 1.0.0 | **Ratified**: 2026-04-04 | **Last Amended**: 2026-04-04

<!-- Knowie: Project Knowledge -->
## Project Knowledge

This project maintains structured knowledge in `knowledge/`:

- **Principles** (`knowledge/principles.md`): Core axioms and derived development principles — the project's non-negotiable rules.
- **Vision** (`knowledge/vision.md`): Goals, current state, architecture decisions, and roadmap.
- **Experience** (`knowledge/experience.md`): Distilled lessons from past development — patterns, pitfalls, and takeaways.

Read these files at the start of any task to understand the project's *why* and constraints.
Additional context may be found in `knowledge/research/`, `knowledge/design/`, and `knowledge/history/`.
<!-- /Knowie -->
