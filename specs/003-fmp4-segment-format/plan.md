# Implementation Plan: fMP4 分段格式（消除破音）

**Branch**: `003-fmp4-segment-format` | **Date**: 2026-04-04 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/003-fmp4-segment-format/spec.md`

## Summary

將 HLS 分段格式從 MPEG-TS(.ts) 改為 fMP4/CMAF(.m4s + init.mp4)，消除分段切換時的 AAC 解碼器重新初始化造成的音訊破音。預生成使用 FFmpeg `-hls_segment_type fmp4`，按需生成使用 `-movflags +frag_keyframe+empty_moov+default_base_moof+cmaf`。Init segment 生成一次後重複使用。

## Technical Context

**Language/Version**: Go 1.25（既有）
**Primary Dependencies**: FFmpeg 6.x+（既有，需 `-hls_segment_type fmp4` 支援，FFmpeg 3.3+ 可用）
**Storage**: tmpfs /dev/shm（既有）
**Testing**: go test + ffprobe（既有）
**Target Platform**: Linux server / Docker（既有）
**Project Type**: web-service（既有）
**Performance Goals**: 分段切換無破音，首個分段回應時間不劣化
**Constraints**: init segment 必須跨分段共享，按需生成的分段必須與預生成分段相容
**Scale/Scope**: 修改既有的合成邏輯，不新增外部相依

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| 原則 | 狀態 | 說明 |
|------|------|------|
| I. 測試驅動開發 | ✅ 通過 | 破音消除需有自動化測試驗證 |
| II. 模組化架構 | ✅ 通過 | 修改集中在 composer 套件，handler 僅調整 Content-Type 與路徑 |
| III. 即時性優先 | ✅ 通過 | fMP4 編碼效能與 .ts 相當 |
| IV. 簡約原則 | ✅ 通過 | 使用 FFmpeg 內建 fMP4 支援，無需新增外部相依 |
| 轉換成本守恆 | ✅ 通過 | fMP4 的編碼成本與 .ts 相當，不增加運算成本 |

## Project Structure

### Documentation (this feature)

```text
specs/003-fmp4-segment-format/
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
│   ├── ffmpeg.go        # 修改：BuildFFmpegArgs 改用 -hls_segment_type fmp4
│   │                    # 修改：BuildSegmentArgs 改用 -f mp4 -movflags fmp4 旗標
│   ├── playlist.go      # 修改：GeneratePlaylist 加入 #EXT-X-MAP、版本升至 7
│   ├── pregen.go        # 修改：IsSegmentReady/GetSegmentPath 改用 .m4s 副檔名
│   └── probe.go         # 不變
├── handler/
│   ├── stream.go        # 修改：SegmentHandler 支援 init.mp4 與 .m4s
│   └── ...              # 其他不變
├── media/               # 不變
└── config/              # 不變

tests/
├── unit/
│   ├── playlist_test.go  # 修改：驗證 #EXT-X-MAP 與 .m4s 分段名稱
│   └── composer_test.go  # 修改：驗證 fMP4 FFmpeg 參數
└── integration/
    ├── stream_test.go    # 修改：驗證 fMP4 分段的 Content-Type
    └── fmp4_test.go      # 新增：驗證分段切換無音訊不連續
```

**Structure Decision**: 沿用既有結構。主要修改 composer 套件中的 FFmpeg 參數，不新增套件或檔案（除測試）。

## Complexity Tracking

> 無違反事項，無需記錄。
