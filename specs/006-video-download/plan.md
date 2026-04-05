# Implementation Plan: 影片下載功能

**Branch**: `006-video-download` | **Date**: 2026-04-05 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/006-video-download/spec.md`

## Summary

新增下載端點，將已預生成的 fMP4 分段（init.mp4 + .m4s）直接串接寫入 HTTP response，產生可播放的 MP4 檔案。Go 中用 `io.Copy` 串流回傳，不需 FFmpeg 重新編碼。WordPress 外掛在播放器旁加下載按鈕，後台加批次匯出。

## Technical Context

**Language/Version**: Go 1.25（既有）+ PHP 8.0+（既有 WordPress 外掛）
**Primary Dependencies**: 無新增（使用標準函式庫 io、os）
**Storage**: tmpfs（既有，讀取預生成的分段）
**Testing**: go test（既有）
**Target Platform**: Linux server / Docker（既有）+ WordPress（既有）
**Project Type**: web-service + WordPress 外掛
**Performance Goals**: 串流回傳，不需全部讀入記憶體
**Constraints**: 需等待預生成完成才能開始下載
**Scale/Scope**: Go 新增 1 個 handler，WordPress 修改 3 個檔案

## Constitution Check

| 原則 | 狀態 | 說明 |
|------|------|------|
| I. 測試驅動開發 | ✅ 通過 | 下載端點需有整合測試 |
| II. 模組化架構 | ✅ 通過 | DownloadHandler 為獨立 handler |
| III. 即時性優先 | ✅ 通過 | 串流回傳，不需全部緩衝 |
| IV. 簡約原則 | ✅ 通過 | 直接串接二進位，不需 FFmpeg |
| 轉換成本守恆 | ✅ 通過 | 使用已預生成的分段，零額外合成成本 |

## Project Structure

### Source Code

```text
src/
├── handler/
│   ├── stream.go        # 既有：新增路由
│   └── download.go      # 新增：DownloadHandler

wordpress/streamixer/
├── includes/
│   ├── class-frontend.php   # 修改：播放器加下載按鈕
│   └── class-cpt.php        # 修改：編輯頁加匯出按鈕
├── assets/js/player.js      # 修改：下載按鈕事件
└── streamixer.php           # 修改：批次操作註冊
```

**Structure Decision**: Go 新增 `download.go`，WordPress 修改既有檔案。

## Complexity Tracking

> 無違反事項。
