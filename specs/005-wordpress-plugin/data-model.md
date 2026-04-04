# 資料模型：WordPress 外掛

## 實體

### Composition（素材組合）— Custom Post Type

WordPress 的 `streamixer` 文章類型。

| 欄位 | 來源 | 說明 |
|------|------|------|
| post_title | WP 內建 | 標題 |
| post_content | WP 內建 | 描述 |
| post_name (slug) | WP 內建 | 用作 Streamixer composition ID |
| _streamixer_audio_id | post meta | 音檔 attachment ID |
| _streamixer_background_id | post meta | 背景圖片 attachment ID |
| _streamixer_subtitle_id | post meta | 字幕檔 attachment ID（可為空） |
| _streamixer_composition_id | post meta | Streamixer 服務中的 ID |
| _streamixer_sync_status | post meta | 同步狀態 |

### Category（素材分類）— Custom Taxonomy

階層式分類法 `streamixer_category`。

### Tag（素材標籤）— Custom Taxonomy

非階層式分類法 `streamixer_tag`。

### Settings（外掛設定）— WordPress Options

| 鍵 | 說明 |
|----|----|
| streamixer_service_url | Streamixer 服務 URL |
| streamixer_default_background | 預設背景圖片 attachment ID |

## 實體關係

```
Composition N──N Category
Composition N──N Tag
Composition 1──0..1 Audio Attachment
Composition 1──0..1 Background Attachment
Composition 1──0..1 Subtitle Attachment
```
