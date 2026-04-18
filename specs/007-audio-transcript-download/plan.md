# Implementation Plan: 音檔與逐字稿下載

**Branch**: `007-audio-transcript-download` | **Date**: 2026-04-18 | **Spec**: [spec.md](./spec.md)
**Input**: `specs/007-audio-transcript-download/spec.md`

## Summary

新增訪客可從播放頁下載原始音檔與管理員上傳逐字稿全文的能力。技術上沿用既有 Streamixer HTTP 服務（Go 1.25 + chi）與 WordPress 外掛（PHP 8.0+）架構：在 Streamixer 後端新增兩個唯讀串流端點（`/audio/{id}`、`/transcript/{id}`），在 WordPress CPT 新增逐字稿 meta 欄位並擴充 `sync_composition` multipart 上傳內容，同時維持階段 6、7.5 的「同步後自動清除本地檔案」語意一致。

## Technical Context

**Language/Version**: Go 1.25（Streamixer 服務）、PHP 8.0+（WordPress 外掛）、原生 JavaScript（前端 toolbar）
**Primary Dependencies**: go-chi/chi v5（既有路由）、WordPress 6.0+ Media/Post Meta API（既有）；無新增套件
**Storage**: tmpfs（`/dev/shm`）／`/media/{id}/` 目錄；逐字稿檔案以 `transcript.<ext>` 存於素材目錄
**Testing**: Go 單元測試 + 整合測試（httptest）；WP 外掛手動驗收（沿用既有模式）
**Target Platform**: Linux Docker 容器（既有）；前端支援 Chrome、Safari、Firefox
**Project Type**: web-service + WordPress plugin（雙元件）
**Performance Goals**: 音檔／逐字稿下載採串流輸出；下載啟動時間 ≤ 3s（對已存在檔案）
**Constraints**: 下載端點記憶體用量不隨檔案大小線性成長；`Content-Disposition` 使用 RFC 5987 格式
**Scale/Scope**: 既有負載（教會網站層級，單素材音檔 ≤ 100MB、逐字稿 ≤ 50MB）

## Constitution Check

| 原則 | 狀態 | 說明 |
|------|------|------|
| I. TDD | 🟡 部分適用 | Go 端新增端點 MUST 有 `httptest` 整合測試（檔案完整性、404、RFC 5987 檔名）；WordPress 側 UI 與同步流程沿用手動驗收（與近期實作一致，已於知識庫原則 4 修訂為「資訊等價性 MUST 可驗證」）。 |
| II. 模組化架構 | 🟢 通過 | 新端點 handler 獨立檔案；逐字稿 loader 與音檔 loader 共用既有 `media.Loader` 檔案探測模式。 |
| III. 即時性優先 | 🟢 通過 | 兩個新端點皆用 `http.ServeFile` 或等同串流方式，不將檔案讀入記憶體。 |
| IV. 簡約原則 | 🟢 通過 | 不引入新套件、不新增抽象層；沿用既有 `MediaComposition`、`Loader`、`upload.Upload` 結構。逐字稿單檔、不建多檔抽象。 |

**Gate 結論**：無違反，可進入 Phase 0。

## Project Structure

### Documentation (this feature)

```text
specs/007-audio-transcript-download/
├── plan.md              # 本檔
├── research.md          # Phase 0 產出
├── data-model.md        # Phase 1 產出
├── quickstart.md        # Phase 1 產出
├── contracts/           # Phase 1 產出（HTTP 端點 schema）
│   ├── audio.md
│   └── transcript.md
├── checklists/
│   └── requirements.md  # spec 階段產出
└── tasks.md             # /speckit.tasks 產出（本指令不建立）
```

### Source Code (repository root)

```text
src/                                # Go 服務（既有）
├── handler/
│   ├── download.go                 # 既有：影片下載 + 進度
│   ├── media_download.go           # 新增：/audio/{id}、/transcript/{id}
│   └── stream.go                   # 路由註冊（修改）
├── media/
│   ├── loader.go                   # 既有；新增 Transcript 探測
│   └── composition.go              # 既有 MediaComposition 新增 Transcript 欄位
├── handler/upload.go               # 既有；multipart 解析新增 transcript
└── ...

wordpress/streamixer/               # WP 外掛（既有）
├── includes/
│   ├── class-cpt.php               # 素材編輯頁新增逐字稿欄位（修改）
│   ├── class-api.php               # sync_composition 多帶 transcript；
│   │                                # cleanup_local_files 納入 transcript；
│   │                                # 新增 get_audio_url / get_transcript_url
│   └── class-frontend.php          # 播放頁 toolbar 新增兩顆下載按鈕
├── assets/js/player.js             # 既有下載進度邏輯保留，音檔/逐字稿為直接連結
└── streamixer.php                  # 批次操作下拉新增兩選項；版本號

tests/                              # 新增
└── integration/
    ├── audio_download_test.go
    └── transcript_download_test.go
```

**Structure Decision**: 沿用既有 Streamixer（Go 單一服務）+ WordPress 外掛（PHP）雙元件結構，不新增模組或子專案。測試置於既有 `tests/integration/`。

## Complexity Tracking

無違反，毋須填寫。
