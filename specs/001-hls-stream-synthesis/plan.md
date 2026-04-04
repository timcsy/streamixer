# Implementation Plan: HLS 即時合成串流

**Branch**: `001-hls-stream-synthesis` | **Date**: 2026-04-04 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-hls-stream-synthesis/spec.md`

## Summary

將音檔、背景圖片與字幕即時合成為 HLS 影片串流。使用 Go 建構 HTTP 服務，透過 FFmpeg 子程序進行媒體合成，輸出至 tmpfs 以降低 I/O 延遲，並以 Docker 容器化部署。

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: FFmpeg 6.x（子程序）、go-chi/chi v5（HTTP 路由）
**Storage**: 檔案系統（素材）+ tmpfs /dev/shm（合成分段暫存）
**Testing**: go test + ffprobe 驗證
**Target Platform**: Linux server（Docker 容器）
**Project Type**: web-service
**Performance Goals**: 首個分段 5 秒內回應
**Constraints**: Docker 容器執行，tmpfs 記憶體佔用需控制
**Scale/Scope**: 單一實例，支援多個並發請求

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| 原則 | 狀態 | 說明 |
|------|------|------|
| I. 測試驅動開發 | ✅ 通過 | 採用 go test + ffprobe，TDD 流程可執行 |
| II. 模組化架構 | ✅ 通過 | 音檔、影像、字幕處理分離，合成引擎透過介面組合 |
| III. 即時性優先 | ✅ 通過 | FFmpeg 即時合成 + tmpfs 降低延遲 |
| IV. 簡約原則 | ✅ 通過 | 最少相依、標準函式庫為主、燒入字幕避免複雜度 |
| 轉換成本守恆 | ✅ 通過 | FFmpeg ultrafast preset + tmpfs，運算成本低於儲存影片成本 |

## Project Structure

### Documentation (this feature)

```text
specs/001-hls-stream-synthesis/
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
├── handler/        # HTTP handlers（playlist、segment）
├── composer/       # 合成引擎（FFmpeg 呼叫、分段管理）
├── media/          # 媒體處理（音檔、圖片、字幕的驗證與解析）
└── config/         # 設定管理

tests/
├── integration/    # 整合測試（完整合成流程、ffprobe 驗證）
└── unit/           # 單元測試（handler、composer 邏輯）

Dockerfile
go.mod
go.sum
```

**Structure Decision**: 採用單一專案結構。此服務為獨立的後端服務，無前端，以 `src/` 組織核心程式碼，`tests/` 分離測試。handler、composer、media 三個套件對應 HTTP 層、合成層、媒體層的職責分離。

## Complexity Tracking

> 無違反事項，無需記錄。
