# streamixer Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-04-18

## Active Technologies
- Go 1.25（既有） + FFmpeg 6.x（既有）、go-chi/chi v5（既有）、golang.org/x/sync/singleflight（新增） (002-instant-seek-pregenerate)
- mpfs /dev/shm（既有） (002-instant-seek-pregenerate)
- Go 1.25（既有） + FFmpeg 6.x+（既有，需 `-hls_segment_type fmp4` 支援，FFmpeg 3.3+ 可用） (003-fmp4-segment-format)
- Go 1.25（既有） + 無新增外部相依（使用標準函式庫 sync、time、os） (004-smart-cache-management)
- PHP 8.0+、JavaScript（React/JSX for Gutenberg Block） + WordPress 6.0+、hls.js（CDN）、@wordpress/scripts（Block 建置工具） (005-wordpress-plugin)
- WordPress 資料庫（post meta）+ WordPress 媒體庫（檔案） (005-wordpress-plugin)
- Go 1.25（既有）+ PHP 8.0+（既有 WordPress 外掛） + 無新增（使用標準函式庫 io、os） (006-video-download)
- mpfs（既有，讀取預生成的分段） (006-video-download)
- Go 1.25（Streamixer 服務）、PHP 8.0+（WordPress 外掛）、原生 JavaScript（前端 toolbar） + go-chi/chi v5（既有路由）、WordPress 6.0+ Media/Post Meta API（既有）；無新增套件 (007-audio-transcript-download)
- mpfs（`/dev/shm`）／`/media/{id}/` 目錄；逐字稿檔案以 `transcript.<ext>` 存於素材目錄 (007-audio-transcript-download)

- Go 1.22+ + FFmpeg 6.x（子程序）、go-chi/chi v5（HTTP 路由） (001-hls-stream-synthesis)

## Project Structure

```text
backend/
frontend/
tests/
```

## Commands

# Add commands for Go 1.22+

## Code Style

Go 1.22+: Follow standard conventions

## Recent Changes
- 007-audio-transcript-download: Added Go 1.25（Streamixer 服務）、PHP 8.0+（WordPress 外掛）、原生 JavaScript（前端 toolbar） + go-chi/chi v5（既有路由）、WordPress 6.0+ Media/Post Meta API（既有）；無新增套件
- 006-video-download: Added Go 1.25（既有）+ PHP 8.0+（既有 WordPress 外掛） + 無新增（使用標準函式庫 io、os）
- 005-wordpress-plugin: Added PHP 8.0+、JavaScript（React/JSX for Gutenberg Block） + WordPress 6.0+、hls.js（CDN）、@wordpress/scripts（Block 建置工具）


<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
