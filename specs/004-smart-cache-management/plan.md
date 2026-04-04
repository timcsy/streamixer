# Implementation Plan: 智慧快取管理

**Branch**: `004-smart-cache-management` | **Date**: 2026-04-05 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/004-smart-cache-management/spec.md`

## Summary

在 PregenManager 上層新增快取管理層（CacheManager），以 LRU + TTL 策略管理 tmpfs 中的合成分段。背景 Sweeper goroutine 定期掃描，清除過期或超量的快取。分段被存取時更新最後存取時間，確保熱門內容保留、冷門內容淘汰。

## Technical Context

**Language/Version**: Go 1.25（既有）
**Primary Dependencies**: 無新增外部相依（使用標準函式庫 sync、time、os）
**Storage**: tmpfs /dev/shm（既有）
**Testing**: go test（既有）
**Target Platform**: Linux server / Docker（既有）
**Project Type**: web-service（既有）
**Performance Goals**: 清掃不影響正常請求回應、暫存不超過容量上限
**Constraints**: 清掃必須安全（不清除正在使用的分段）、整組清除（不留殘片）
**Scale/Scope**: 新增 CacheManager + Sweeper，修改 handler 更新存取時間

## Constitution Check

| 原則 | 狀態 | 說明 |
|------|------|------|
| I. 測試驅動開發 | ✅ 通過 | TTL 清掃與 LRU 淘汰需有自動化測試 |
| II. 模組化架構 | ✅ 通過 | CacheManager 為獨立模組，不侵入 PregenManager 核心邏輯 |
| III. 即時性優先 | ✅ 通過 | 清掃在背景 goroutine 執行，不阻塞請求 |
| IV. 簡約原則 | ✅ 通過 | 使用標準函式庫，不引入外部快取框架 |
| 轉換成本守恆 | ✅ 通過 | 快取減少重複合成，降低總運算成本 |

## Project Structure

### Documentation (this feature)

```text
specs/004-smart-cache-management/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
└── tasks.md
```

### Source Code (repository root)

```text
src/
├── composer/
│   ├── cache.go         # 新增：CacheManager（LRU + TTL、存取時間追蹤、容量計算）
│   ├── sweeper.go       # 新增：Sweeper（背景清掃 goroutine）
│   ├── pregen.go        # 既有：修改——清除快取時同步清除預生成任務狀態
│   └── ...              # 其他不變
├── handler/
│   └── stream.go        # 既有：修改——分段存取時更新 CacheManager 時間
├── config/
│   └── config.go        # 既有：修改——新增 TTL、容量上限、清掃頻率設定
└── ...

tests/
├── unit/
│   ├── cache_test.go    # 新增：CacheManager 單元測試
│   └── sweeper_test.go  # 新增：Sweeper 單元測試
└── integration/
    └── cache_test.go    # 新增：快取清掃整合測試
```

**Structure Decision**: 沿用既有結構。新增 `cache.go`（快取管理）和 `sweeper.go`（清掃排程）到 `src/composer/`。

## Complexity Tracking

> 無違反事項。
