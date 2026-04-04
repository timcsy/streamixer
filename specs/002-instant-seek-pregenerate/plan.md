# Implementation Plan: 即時跳轉與背景預生成

**Branch**: `002-instant-seek-pregenerate` | **Date**: 2026-04-04 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/002-instant-seek-pregenerate/spec.md`

## Summary

以混合策略改進 HLS 串流播放體驗：首次請求時背景啟動 FFmpeg 預生成所有分段，同時保留按需生成作為 fallback。解決目前按需生成造成的音訊時間戳偏移問題，確保長音檔（1 小時以上）流暢播放與精確跳轉。

## Technical Context

**Language/Version**: Go 1.25（既有）
**Primary Dependencies**: FFmpeg 6.x（既有）、go-chi/chi v5（既有）、golang.org/x/sync/singleflight（新增）
**Storage**: tmpfs /dev/shm（既有）
**Testing**: go test + ffprobe（既有）
**Target Platform**: Linux server / Docker（既有）
**Project Type**: web-service（既有）
**Performance Goals**: playlist 回應 < 5 秒、跳轉後 < 2 秒播放、音訊偏移 < 500ms
**Constraints**: 背景預生成不阻塞請求、同一素材不重複預生成
**Scale/Scope**: 單一實例，支援多個並發素材組合的預生成

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| 原則 | 狀態 | 說明 |
|------|------|------|
| I. 測試驅動開發 | ✅ 通過 | 音訊對齊需有自動化測試驗證 |
| II. 模組化架構 | ✅ 通過 | 預生成管理器為獨立模組，不影響既有 handler/composer |
| III. 即時性優先 | ✅ 通過 | 背景預生成 + 按需 fallback 確保低延遲 |
| IV. 簡約原則 | ✅ 通過 | 使用 singleflight 而非自建複雜鎖機制 |
| 轉換成本守恆 | ✅ 通過 | 預生成避免重複合成，降低總運算成本 |

## Project Structure

### Documentation (this feature)

```text
specs/002-instant-seek-pregenerate/
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
│   ├── ffmpeg.go        # 既有：修正 BuildSegmentArgs 時間戳對齊
│   ├── playlist.go      # 既有：不變
│   ├── probe.go         # 既有：不變
│   └── pregen.go        # 新增：預生成管理器（PregenManager）
├── handler/
│   ├── stream.go        # 既有：修改 SegmentHandler 加入預生成邏輯
│   └── ...              # 其他既有 handler 不變
├── media/               # 既有：不變
└── config/
    └── config.go        # 既有：新增預生成相關設定

tests/
├── unit/
│   └── pregen_test.go   # 新增：預生成管理器單元測試
└── integration/
    ├── seek_test.go     # 新增：跳轉音訊對齊整合測試
    └── pregen_test.go   # 新增：預生成流程整合測試
```

**Structure Decision**: 沿用既有的單一專案結構。新增 `src/composer/pregen.go` 作為預生成管理器，修改既有的 `stream.go` 與 `ffmpeg.go`。

## Complexity Tracking

> 無違反事項，無需記錄。
