# 快速入門：HLS 即時合成串流

## 前置需求

- Docker（已安裝並可執行）

## 啟動服務

```bash
docker build -t streamixer .
docker run -p 8080:8080 -v /path/to/media:/media streamixer
```

其中 `/path/to/media` 為素材檔案的存放目錄。

## 準備素材

在素材目錄中，以素材組合 ID 為名建立子目錄，放入對應檔案：

```
/path/to/media/
└── sermon-001/
    ├── audio.mp3       # 音檔（必要）
    ├── background.jpg  # 背景圖片（必要）
    └── subtitle.srt    # 字幕檔（選擇性）
```

## 播放串流

在瀏覽器或任何 HLS 播放器中開啟：

```
http://localhost:8080/stream/sermon-001/index.m3u8
```

## 驗證服務狀態

```bash
curl http://localhost:8080/health
# 預期回應：{"status": "ok"}
```

## 驗證串流

```bash
# 取得 playlist
curl http://localhost:8080/stream/sermon-001/index.m3u8

# 用 ffprobe 驗證
ffprobe -v error -show_format http://localhost:8080/stream/sermon-001/index.m3u8
```
