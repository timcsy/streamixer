# Streamixer

將音檔、背景圖片與字幕即時合成為 HLS 影片串流的服務。儲存的是素材，播放的是影片——大幅降低影片儲存成本。

## 適用場景

適合內容本質上是「音檔配圖」的場景：講道錄音、線上課程、podcast 影片化等。不需要預先合成完整影片檔，節省 50% 以上的儲存空間。

## 快速開始

### Docker 部署

```bash
docker run -d \
  --name streamixer \
  -p 8080:8080 \
  -v /path/to/media:/media:ro \
  --tmpfs /dev/shm/streamixer:size=512M \
  ghcr.io/timcsy/streamixer:latest
```

### 準備素材

```
/path/to/media/
└── my-sermon/
    ├── audio.mp3       # 音檔（必要，支援 MP3/WAV）
    ├── background.jpg  # 背景圖片（必要，支援 JPG/PNG）
    └── subtitle.srt    # 字幕（選填，支援 SRT/VTT）
```

### 播放

在任何 HLS 播放器中開啟：

```
http://localhost:8080/stream/my-sermon/index.m3u8
```

或開啟內建測試頁面：`http://localhost:8080`

## 環境變數

| 變數 | 預設值 | 說明 |
|------|--------|------|
| `PORT` | `8080` | HTTP 伺服器 port |
| `MEDIA_DIR` | `/media` | 素材檔案目錄 |
| `TMP_DIR` | `/dev/shm/streamixer` | 合成分段暫存目錄（建議 tmpfs） |
| `API_KEY` | 空（不認證） | Upload API 認證金鑰 |
| `CORS_ORIGINS` | `*` | CORS 允許來源（逗號分隔） |
| `CACHE_TTL` | `30m` | 快取過期時間 |
| `CACHE_MAX_SIZE` | `0`（不限） | 快取容量上限（bytes） |
| `CACHE_SWEEP_INTERVAL` | `5m` | 快取清掃頻率 |

## API

### 串流端點

| 方法 | 路徑 | 說明 |
|------|------|------|
| GET | `/stream/{id}/index.m3u8` | HLS playlist |
| GET | `/stream/{id}/init.mp4` | fMP4 init segment |
| GET | `/stream/{id}/seg_{n}.m4s` | fMP4 media segment |
| GET | `/health` | 健康檢查 |

### 管理端點（需 API Key）

| 方法 | 路徑 | 說明 |
|------|------|------|
| POST | `/upload/{id}` | 上傳素材（multipart: audio, background, subtitle） |
| POST | `/sample` | 產生範例素材 |
| GET | `/compositions` | 列出所有素材組合 |

使用 API Key：

```bash
curl -X POST -H "X-API-Key: your-key" \
  -F "audio=@audio.mp3" \
  -F "background=@bg.jpg" \
  http://localhost:8080/upload/my-sermon
```

## WordPress 外掛

### 安裝

1. 從 [GitHub Releases](https://github.com/timcsy/streamixer/releases) 下載 `streamixer-wordpress.zip`
2. WordPress 後台 → 外掛 → 安裝外掛 → 上傳外掛 → 選擇 zip → 安裝
3. 啟用「Streamixer」外掛

### 設定

前往「設定 > Streamixer」：
- **Streamixer 服務 URL**：Go 服務的內部位址（如 `http://streamixer:8080`）
- **前端播放 URL**：瀏覽器可存取的位址（如 `https://your-domain.com:8080`）
- **API Key**：需與服務端的 `API_KEY` 環境變數一致

### 使用

1. **新增素材**：Streamixer → 新增素材組合 → 上傳音檔、圖片、字幕 → 發佈
2. **嵌入播放器**：
   - **Gutenberg**：點「+」→ 搜尋「streamixer」→ 選取素材
   - **Shortcode**：`[streamixer id="my-sermon"]`
3. **瀏覽**：前台自動有列表頁（`/streamixer/`）和播放頁

## 技術架構

```
瀏覽器 ──HLS──→ Streamixer（Go + FFmpeg）──fMP4──→ tmpfs
                      ↑
WordPress 外掛 ──upload API──→ 素材目錄
```

- **Go + chi**：HTTP 服務
- **FFmpeg**：即時合成（`-hls_segment_type fmp4 -force_key_frames`）
- **fMP4/CMAF**：無破音分段格式
- **背景預生成**：首次播放時自動合成，後續請求直接回傳
- **LRU + TTL 快取**：自動清掃冷門內容

## 開發

```bash
# 執行測試
go test ./tests/...

# 本地開發
MEDIA_DIR=./media TMP_DIR=/tmp/streamixer go run ./src

# Docker 建置
docker build -t streamixer .
```

## 授權

[MIT License](LICENSE)
