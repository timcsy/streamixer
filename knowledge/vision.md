# 願景

## 問題陳述

教會的 WordPress 網站存放大量講道影片，但這些影片的內容本質上只是錄音配上背景圖片與字幕。以完整影片格式儲存造成高昂的儲存成本，而這些成本是不必要的——因為影片只是底層素材的一種展示投影。

## 核心想法

Streamixer 是一個即時合成串流服務：將分開儲存的音檔、背景影像與字幕，在使用者觀看時即時合成為 HLS 影片串流。儲存的是素材，播放的是影片。

## 現狀

階段 1 至 4 已全部完成。Go 串流服務使用 fMP4（CMAF）分段格式、背景預生成、LRU+TTL 快取管理。WordPress 外掛提供後台素材管理（CPT + Media Library）、Shortcode 與 Gutenberg Block 嵌入、響應式播放頁面。以 docker-compose 同時運行 Streamixer + WordPress + MySQL。

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
- `src/handler/` — HTTP handlers（串流、init.mp4、上傳、範例產生、健康檢查、CORS middleware）
- `src/composer/` — FFmpeg 合成引擎、playlist 產生、音檔探測、PregenManager、CacheManager、Sweeper
- `src/media/` — 素材載入、驗證、資料結構
- `src/config/` — 環境變數設定管理
- `static/` — 測試用前端

**WordPress 外掛結構**（`wordpress/streamixer/`）：
- `streamixer.php` — 主外掛檔案（bootstrap、Gutenberg Block 註冊、媒體按鈕）
- `includes/` — CPT + 分類法、Settings API、Streamixer HTTP 通訊、Shortcode、前端 asset
- `assets/js/block.js` — Gutenberg Block（純 JS，不需 build）
- `assets/js/player.js` — HLS 播放器（hls.js 整合）
- `templates/` — 列表頁 + 播放頁模板

**部署**：
- Streamixer 以 Docker 獨立部署（docker-compose 或 `docker run`）
- WordPress 外掛複製到既有 WordPress 的 `wp-content/plugins/` 即可使用

## 路線圖

### 階段 1：最小可用串流

- [x] 完成

交付：給定一個音檔、一張背景圖片、一段字幕檔，能透過 HLS 串流播放合成後的影片。
前置條件：無

**成功標準：**
- [x] 發送 HTTP 請求後取得有效的 HLS playlist（.m3u8）
- [x] 播放器能正常播放合成的影片串流，包含音訊、背景畫面與字幕
- [x] Docker 容器可正常啟動服務

### 階段 2：即時跳轉與背景預生成

- [x] 完成

交付：背景預生成所有分段，分段請求等待預生成完成，確保長音檔流暢播放。
前置條件：階段 1

**策略**：
- 首次請求 playlist 時，立即回傳手動計算的 playlist（顯示完整時長），同時背景啟動 FFmpeg 預生成
- 使用者請求分段時：已生成 → 直接回傳；尚未生成 → 等待預生成完成（WaitForSegment）
- 順序播放時預生成跑在使用者前面，等待時間極短
- 使用 singleflight 防止同一素材重複預生成，semaphore 限制並發數

**成功標準：**
- [x] 順序播放時分段切換無中斷
- [x] 跳轉後音訊位置正確
- [x] 1 小時音檔的首次 playlist 請求在 5 秒內回應
- [x] 跳轉到任意位置後 2 秒內開始播放

### 階段 2.5：fMP4 分段格式（消除破音）

- [x] 完成

交付：使用 fMP4（CMAF）分段格式，消除分段切換時的音訊不連續。
前置條件：階段 2

**策略**：
- FFmpeg 使用 `-hls_segment_type fmp4` 產出 init.mp4 + .m4s 分段
- `-force_key_frames expr:gte(t,n_forced*N)` 強制精確的 keyframe 間距，確保分段時長與 playlist 一致
- 所有分段來自同一個 FFmpeg 進程（不使用按需生成），確保 init segment 與分段格式完全相容
- playlist 使用 `#EXT-X-VERSION:7` + `#EXT-X-MAP:URI="init.mp4"`

**成功標準：**
- [x] 順序播放時分段切換無可感知的破音或中斷
- [x] 跳轉後音訊銜接平順
- [x] 既有的 HLS 播放器（hls.js、Safari）可正常播放 fMP4 分段

### 階段 3：智慧快取管理

- [x] 完成

交付：自動管理 tmpfs 中合成分段的生命週期——熱門內容保留快取，冷門內容自動清除，避免記憶體耗盡。
前置條件：階段 2.5

**策略**：
- 以 LRU + TTL 管理每組素材的合成分段
- 記錄每組素材最後被存取的時間，超過 TTL（可設定）未被存取則自動清除
- 背景 goroutine 定期掃描，淘汰過期分段目錄
- tmpfs 容量接近上限時，主動淘汰最久未存取的內容
- 被清除的素材再次被請求時，重新觸發合成（對使用者透明）

**成功標準：**
- [x] tmpfs 使用量不超過設定的容量上限
- [x] 無人存取的素材在 TTL 過期後自動清除
- [x] 清除後重新請求時，使用者體驗與首次請求一致
- [x] 多人同時觀看同一內容時共享同一份快取，不重複合成

### 階段 4：WordPress 外掛

- [x] 完成

交付：通用的 WordPress 外掛，讓使用者在後台管理素材組合（音檔+背景圖片+字幕），並在前台以良好的 UX 播放合成影片。不綁定特定用途，適用於講道、課程、podcast 等任何「音檔配圖」的場景。
前置條件：階段 3

**管理功能（後台）**：
- 新增「Streamixer」選單，以自訂文章類型（Custom Post Type）管理素材組合
- 新增/編輯素材組合：上傳音檔、背景圖片、字幕檔，設定標題、分類、標籤等 metadata
- 素材上傳後自動同步至 Streamixer 服務（透過 upload API）
- 顯示每組素材的串流狀態（預生成進度、快取狀態）
- 設定頁：Streamixer 服務 URL、預設背景圖片、分類管理

**嵌入方式（前台）**：
- Shortcode：`[streamixer id="my-audio"]` 嵌入播放器至任何文章或頁面
- Gutenberg Block：視覺化編輯器中可直接搜尋並選取素材組合嵌入
- 自動嵌入：素材組合的 Custom Post Type 頁面自動顯示播放器

**播放頁面**：
- 列表頁：依日期排列，可依分類、標籤篩選，支援搜尋
- 單則播放頁：含播放器、標題、描述、分類、標籤
- 響應式設計：手機、平板、桌面都有良好的播放體驗
- 播放器支援：播放/暫停、進度條跳轉、音量、全螢幕

**成功標準：**
- [x] 使用者可在 WordPress 後台上傳音檔+圖片+字幕，無需接觸伺服器
- [x] 訪客可在前台播放頁面流暢觀看合成影片
- [x] Shortcode 和 Gutenberg Block 皆可正常嵌入播放器
- [ ] 手機瀏覽器上的播放體驗與桌面一致
- [ ] 列表頁可依分類、標籤篩選與搜尋

### 階段 5：發布準備

- [ ] 完成

交付：將專案整理至可公開發布的狀態，包含文件、CI/CD、安全性強化。
前置條件：階段 4

**文件**：
- README.md：專案介紹、快速開始、架構說明、設定參數
- LICENSE：選擇開源授權條款
- WordPress 外掛 readme.txt（WordPress 外掛目錄標準格式）

**CI/CD**：
- GitHub Actions：推送 tag 時自動建置 Docker Image 並發布至 GitHub Container Registry（ghcr.io）
- WordPress 外掛以 GitHub Release 的 zip 附件發布，使用者下載後解壓到 wp-content/plugins/

**安全性**：
- Upload API 加入 API Key 認證（環境變數設定）
- CORS 改為可設定的允許來源（環境變數，預設限制為設定的 domain）

**分支整理**：
- 合併所有 feature branch 到 main
- 建立 v1.0.0 tag

**成功標準：**
- [ ] README.md 完整且可指引新使用者從零開始部署
- [ ] 推送 git tag 後自動產生 Docker Image 至 ghcr.io
- [ ] WordPress 外掛可從 GitHub Release 下載 zip 安裝
- [ ] Upload API 需要 API Key 才能存取
- [ ] CORS 允許來源可透過環境變數設定
