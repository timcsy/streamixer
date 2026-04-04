# 介面契約：WordPress 外掛

## WordPress Hooks（外掛與 WordPress 的契約）

### Custom Post Type

- **名稱**：`streamixer`
- **標籤**：「素材組合」
- **支援**：title、editor（描述）、thumbnail（縮圖）
- **公開**：是（有前台 archive 和 single 頁面）
- **自訂分類法**：`streamixer_category`（階層式）、`streamixer_tag`（非階層式）

### Post Meta（素材組合的 metadata）

| meta_key | 說明 | 型態 |
|----------|------|------|
| `_streamixer_audio_id` | 音檔 attachment ID | int |
| `_streamixer_background_id` | 背景圖片 attachment ID | int |
| `_streamixer_subtitle_id` | 字幕檔 attachment ID | int（可為空） |
| `_streamixer_composition_id` | Streamixer 服務中的素材 ID | string |
| `_streamixer_sync_status` | 同步狀態（synced / pending / error） | string |
| `_streamixer_sync_error` | 同步錯誤訊息 | string |

### Shortcode

- **名稱**：`streamixer`
- **屬性**：`id`（素材組合的 post slug 或 Streamixer ID）
- **用法**：`[streamixer id="my-audio"]`
- **輸出**：播放器 HTML（`<div>` + `<video>` + hls.js 初始化）

### Gutenberg Block

- **名稱**：`streamixer/player`
- **屬性**：`compositionId`（選取的素材組合 post ID）
- **渲染**：動態（`render.php`），輸出與 Shortcode 相同的播放器 HTML

### 設定（Options）

| option_key | 說明 | 預設值 |
|------------|------|--------|
| `streamixer_service_url` | Streamixer 服務 URL | `http://localhost:8080` |
| `streamixer_default_background` | 預設背景圖片 attachment ID | 空 |

## 外掛與 Streamixer 服務的通訊

使用既有的 Streamixer HTTP API：
- `POST /upload/{id}` — 同步素材（multipart form: audio + background + subtitle）
- `GET /stream/{id}/index.m3u8` — 播放器連接的串流 URL
