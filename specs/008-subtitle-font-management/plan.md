# Implementation Plan: 字幕字體管理

**Branch**: `008-subtitle-font-management` | **Date**: 2026-04-19 | **Spec**: [spec.md](./spec.md)
**Input**: `specs/008-subtitle-font-management/spec.md`

## Summary

在 Streamixer Go 服務新增字體管理能力（列表、上傳、刪除），在 WordPress 外掛新增字體管理 UI 與素材組合的字體選項。系統內建字體直接從 Docker 映像掃描，使用者上傳字體存於獨立 Docker named volume 保證重啟持久。FFmpeg 合成時以 `fontsdir=` 合併掃描路徑並以 `FontName=...` 指定 family name；字體不存在時 libass 自動 fallback，不使合成失敗。

## Technical Context

**Language/Version**: Go 1.25（既有）、PHP 8.0+（既有）、原生 JavaScript（WP UI）
**Primary Dependencies**: go-chi/chi v5（既有路由）、FFmpeg + libass（既有合成管線）、fontconfig（Alpine）
**Storage**：
- 系統字體：映像內 `/usr/share/fonts/`（既有）
- 使用者上傳字體：Docker named volume 掛在 `/fonts`
**Testing**: Go `httptest` 整合測試；字體驗證以小型真實 ttf fixture
**Target Platform**: Linux Docker 容器（Alpine base + font-noto-cjk）
**Project Type**: web-service + WordPress plugin（雙元件）
**Performance Goals**: 字體列表 API < 200ms；上傳驗證 < 1s（10MB 檔）；合成效能不受字體數量影響
**Constraints**: 單檔 ≤ 10MB；字體總數 ≤ 50；上傳只支援 .ttf/.otf/.ttc
**Scale/Scope**: 教會網站層級；全站字體數量現實上 < 10 個

## Constitution Check

| 原則 | 狀態 | 說明 |
|------|------|------|
| I. TDD | 🟡 部分適用 | Go 端字體驗證、列表合併邏輯 MUST 有單元測試；字體 HTTP 端點 MUST 有整合測試；WP UI 沿用手動驗收。 |
| II. 模組化架構 | 🟢 通過 | 新增 `src/fonts/` 套件封裝掃描、驗證、儲存；`FontHandler` 與既有 `MediaDownloadHandler` 結構一致。 |
| III. 即時性優先 | 🟢 通過 | 列表與上傳為 I/O 綁定操作，皆以串流寫入；不影響合成管線的延遲特性。 |
| IV. 簡約原則 | 🟢 通過 | 無新外部套件；字體驗證以標準函式庫讀檔頭；刪除保護以檔案系統檢查 + WP 端查詢處理，不新增抽象層。 |

**Gate 結論**：無違反，進入 Phase 0。

## Project Structure

### Documentation (this feature)

```text
specs/008-subtitle-font-management/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── fonts-list.md
│   ├── fonts-upload.md
│   └── fonts-delete.md
├── checklists/
│   └── requirements.md
└── tasks.md            # /speckit.tasks 產出
```

### Source Code (repository root)

```text
src/
├── fonts/                              # 新增套件
│   ├── manager.go                      # FontManager：scan + list + upload + delete
│   ├── validate.go                     # sniff ttf/otf/ttc 檔頭、parse family name
│   └── types.go                        # Font struct、Source 列舉
├── handler/
│   ├── fonts.go                        # 新增：/fonts GET/POST/DELETE
│   ├── stream.go                       # 路由註冊（修改）
│   └── ...
├── composer/
│   └── ffmpeg.go                       # subtitles filter 加 fontsdir + FontName（修改）
├── config/
│   └── config.go                       # 新增 FontDir 與 SystemFontDirs 欄位
└── ...

wordpress/streamixer/
├── includes/
│   ├── class-settings.php              # 「字體管理」區塊 + 全站預設下拉
│   ├── class-cpt.php                   # 素材編輯頁新增字體下拉
│   ├── class-api.php                   # sync_composition 帶 font 欄位 + fetch_fonts helper
│   └── class-fonts.php                 # 新增：thin client for /fonts
└── ...

Dockerfile                              # 預裝 Noto Sans/Serif TC（若 base 未內含）
docker-compose.yml                      # 新增 streamixer-fonts volume

tests/
└── integration/
    └── fonts_test.go                   # 字體端點與驗證的整合測試
```

**Structure Decision**：新增 `src/fonts/` 封裝字體領域邏輯，與現有 `src/media/`、`src/composer/` 同層。WP 新增 `class-fonts.php` 為呼叫後端 `/fonts` API 的 thin client，讓 settings 與 CPT 共用。

## Complexity Tracking

無違反，毋須填寫。
