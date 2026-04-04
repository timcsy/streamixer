# streamixer Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-04-05

## Active Technologies
- Go 1.25（既有） + FFmpeg 6.x（既有）、go-chi/chi v5（既有）、golang.org/x/sync/singleflight（新增） (002-instant-seek-pregenerate)
- mpfs /dev/shm（既有） (002-instant-seek-pregenerate)
- Go 1.25（既有） + FFmpeg 6.x+（既有，需 `-hls_segment_type fmp4` 支援，FFmpeg 3.3+ 可用） (003-fmp4-segment-format)
- Go 1.25（既有） + 無新增外部相依（使用標準函式庫 sync、time、os） (004-smart-cache-management)

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
- 004-smart-cache-management: Added Go 1.25（既有） + 無新增外部相依（使用標準函式庫 sync、time、os）
- 003-fmp4-segment-format: Added Go 1.25（既有） + FFmpeg 6.x+（既有，需 `-hls_segment_type fmp4` 支援，FFmpeg 3.3+ 可用）
- 002-instant-seek-pregenerate: Added Go 1.25（既有） + FFmpeg 6.x（既有）、go-chi/chi v5（既有）、golang.org/x/sync/singleflight（新增）


<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
