# 快速驗證：影片下載功能

## 驗證下載端點

```bash
# 確認素材已預生成
curl -s http://localhost:8081/stream/sample/index.m3u8 > /dev/null
sleep 10

# 下載 MP4
curl -o sample.mp4 http://localhost:8081/download/sample

# 驗證 MP4 可播放
ffprobe -v error -show_format sample.mp4
# 應顯示 format_name=mov,mp4,...

# 用 VLC 播放
open sample.mp4  # macOS
```

## 驗證 WordPress 下載按鈕

1. 前往素材播放頁面
2. 確認播放器下方有「下載影片」按鈕
3. 點擊按鈕，確認瀏覽器開始下載 MP4

## 驗證批次匯出

1. WordPress 後台 → 素材組合列表
2. 勾選多個素材
3. 批次操作 → 匯出影片 → 套用
4. 確認每個素材的 MP4 依序下載
