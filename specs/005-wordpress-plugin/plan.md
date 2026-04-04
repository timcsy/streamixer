# Implementation Plan: WordPress 外掛

**Branch**: `005-wordpress-plugin` | **Date**: 2026-04-05 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/005-wordpress-plugin/spec.md`

## Summary

WordPress 外掛，使用 Custom Post Type 管理素材組合，透過 WordPress Media Library 處理檔案上傳，以 `wp_remote_post()` 同步至 Streamixer 服務。提供 Shortcode 與 Gutenberg Block 兩種嵌入方式，前端使用 hls.js 播放合成的 HLS 串流。

## Technical Context

**Language/Version**: PHP 8.0+、JavaScript（React/JSX for Gutenberg Block）
**Primary Dependencies**: WordPress 6.0+、hls.js（CDN）、@wordpress/scripts（Block 建置工具）
**Storage**: WordPress 資料庫（post meta）+ WordPress 媒體庫（檔案）
**Testing**: WordPress Plugin Unit Tests（PHPUnit + WP_UnitTestCase）
**Target Platform**: WordPress 網站（PHP server）
**Project Type**: WordPress 外掛
**Performance Goals**: 播放器載入 < 2 秒、素材同步 < 5 秒
**Constraints**: 不修改 Streamixer Go 服務、使用 WordPress 標準 API
**Scale/Scope**: 單一外掛，約 15 個檔案

## Constitution Check

| 原則 | 狀態 | 說明 |
|------|------|------|
| I. 測試驅動開發 | ✅ 通過 | 使用 WordPress PHPUnit 測試框架 |
| II. 模組化架構 | ✅ 通過 | CPT、Shortcode、Block、Settings 各為獨立 class |
| III. 即時性優先 | ✅ 通過 | 播放器直接連接 Streamixer 串流，無額外轉接層 |
| IV. 簡約原則 | ✅ 通過 | 使用 WordPress 標準 API（CPT、Settings API、Media Library），不自建框架 |
| 轉換成本守恆 | ✅ 通過 | 外掛只是管理與嵌入層，不增加合成成本 |

## Project Structure

### Documentation (this feature)

```text
specs/005-wordpress-plugin/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
└── tasks.md
```

### Source Code (repository root)

```text
wordpress/streamixer/
├── streamixer.php              # 主外掛檔案（header、bootstrap）
├── includes/
│   ├── class-cpt.php           # Custom Post Type 註冊 + meta box
│   ├── class-settings.php      # 設定頁（Settings API）
│   ├── class-api.php           # Streamixer HTTP 通訊（wp_remote_post）
│   ├── class-shortcode.php     # Shortcode 註冊與渲染
│   └── class-frontend.php      # 前端 asset enqueue + 播放器渲染
├── src/                        # Gutenberg Block 原始碼（React/JSX）
│   ├── block.json
│   ├── index.js
│   ├── edit.js
│   └── save.js
├── build/                      # wp-scripts 建置輸出
├── assets/
│   ├── css/
│   │   └── player.css          # 播放器與列表頁樣式
│   └── js/
│       └── player.js           # HLS 播放器初始化
├── templates/
│   ├── single-streamixer.php   # 單則播放頁模板
│   └── archive-streamixer.php  # 列表頁模板
└── package.json                # @wordpress/scripts 設定
```

**Structure Decision**: WordPress 外掛獨立存放在 `wordpress/streamixer/` 目錄，與 Go 服務分開。遵循 WordPress 外掛標準結構。

## Complexity Tracking

> 無違反事項。
