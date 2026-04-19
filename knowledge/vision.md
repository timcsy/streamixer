# 願景

## 問題陳述

教會的 WordPress 網站存放大量講道影片，但這些影片的內容本質上只是錄音配上背景圖片與字幕。以完整影片格式儲存造成高昂的儲存成本，而這些成本是不必要的——因為影片只是底層素材的一種展示投影。

## 核心想法

Streamixer 是一個即時合成串流服務：將分開儲存的音檔、背景影像與字幕，在使用者觀看時即時合成為 HLS 影片串流。儲存的是素材，播放的是影片。

## 現狀

階段 1 至 9 已完成並發布 v1.7.2（v1.7.1 修正字幕樣式：ffmpeg drawbox 繪製半透明黑底條、純白無描邊字；字體上傳上限提至 50MB、啟動時自動重建 fontconfig symlinks。v1.7.2 CPT rewrite slug 預設改為 `media` 並可透過 `streamixer_cpt_slug` 等 filter 覆寫，避免與反向代理 `/streamixer` 路徑衝突）。Go 串流服務使用 fMP4（CMAF）分段格式、背景預生成、LRU+TTL 快取管理（TTL/清掃頻率/容量可由 `/config` 端點動態調整）、API Key 認證、可設定 CORS、影片下載（完整檔案 + 進度端點）、原始音檔與逐字稿下載端點。WordPress 外掛提供後台素材管理（含逐字稿欄位與增量同步）、Shortcode 與 Gutenberg Block 嵌入（嵌入碼使用 post ID 保持簡短）、響應式播放頁面、同步後自動清除本地檔案、下載進度條、影片/音檔/逐字稿三顆下載按鈕、批次匯出、設定頁即時調整後端快取參數；字幕樣式避開播放器控制列；後端目錄採用可讀中文名稱；CPT 強制傳統編輯器避免 Gutenberg 丟失 meta box 欄位；所有下載採 RFC 5987 檔名避免中文亂碼。

## 架構

- **語言**：Go 1.25
- **HTTP 路由**：go-chi/chi v5
- **媒體合成**：FFmpeg 子程序，以 `-preset ultrafast -tune stillimage -force_key_frames` 合成
- **分段格式**：fMP4 / CMAF（`-hls_segment_type fmp4`），init.mp4 + .m4s 分段
- **字幕處理**：燒入（burned-in），透過 FFmpeg `-vf subtitles=` 濾鏡
- **暫存**：tmpfs（/dev/shm），避免磁碟 I/O
- **串流協定**：HLS（HTTP Live Streaming），`#EXT-X-VERSION:7` + `#EXT-X-MAP`
- **部署方式**：Docker 容器化，多階段建置，Alpine 基礎映像 + font-noto-cjk

**Go 服務結構**：
- `src/handler/` — HTTP handlers（串流、下載、上傳、健康檢查、CORS、API Key middleware）
- `src/composer/` — FFmpeg 合成引擎、playlist 產生、PregenManager、CacheManager、Sweeper
- `src/media/` — 素材載入、驗證、資料結構
- `src/config/` — 環境變數設定管理

**WordPress 外掛結構**（`wordpress/streamixer/`）：
- `streamixer.php` — bootstrap、Block 註冊、批次操作、MIME 類型
- `includes/` — CPT、Settings API、Streamixer HTTP 通訊、Shortcode、前端 asset
- `assets/` — Gutenberg Block JS、HLS 播放器 JS、CSS
- `templates/` — 列表頁 + 播放頁模板

**部署**：
- Streamixer 以 Docker 獨立部署（docker-compose 或 `docker run`）
- WordPress 外掛複製到既有 WordPress 的 `wp-content/plugins/` 即可使用

## 路線圖

### 已完成

- [x] **階段 1：最小可用串流** — HLS 即時合成（音檔+圖片+字幕）、Docker 部署
- [x] **階段 2：即時跳轉與背景預生成** — PregenManager、WaitForSegment、singleflight
- [x] **階段 2.5：fMP4 分段格式** — 消除 .ts 破音、`-force_key_frames` 精確切割
- [x] **階段 3：智慧快取管理** — LRU + TTL、Sweeper 背景清掃
- [x] **階段 4：WordPress 外掛** — CPT、Shortcode、Gutenberg Block、響應式播放頁面
- [x] **階段 5：發布準備** — README、LICENSE、GitHub Actions CI/CD、API Key、CORS
- [x] **階段 6：同步後清除 WordPress 端檔案** — 自動刪除本地檔案節省空間
- [x] **階段 7：影片下載功能** — `/download/{id}` 端點、播放器下載按鈕、後台批次匯出
- [x] **階段 7.5：下載體驗與素材管理修正** — 等待 `EXT-X-ENDLIST` 保完整、`/progress/{id}` 進度端點、字幕樣式避開控制列、可讀中文目錄、CPT 強制傳統編輯器
- [x] **階段 8：音檔與逐字稿下載** — `/audio/{id}` 與 `/transcript/{id}` 端點、WP「逐字稿」欄位、三顆下載按鈕、批次匯出音檔/逐字稿、RFC 5987 下載檔名
- [x] **階段 8.5：快取設定動態化與增量同步** — Go `GET/PUT /config` 端點（TTL、清掃頻率、容量上限），WP 設定頁「影片快取設定」區塊即時推送到後端不需重啟；已清除本地檔的素材新增選填欄位走增量同步；shortcode 嵌入碼改用 `post->ID` 避免中文 slug 過長

- [x] **階段 9：字幕字體管理** — `/fonts` GET/POST/DELETE + `/fonts/default` 端點、WP 設定頁「字體管理」區塊、CPT 字體下拉、FFmpeg `force_style` 加 `FontName`、使用者字體存於 named volume、系統字體唯讀

### 待完成

（暫無下一階段）
