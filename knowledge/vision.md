# 願景

## 問題陳述

教會的 WordPress 網站存放大量講道影片，但這些影片的內容本質上只是錄音配上背景圖片與字幕。以完整影片格式儲存造成高昂的儲存成本，而這些成本是不必要的——因為影片只是底層素材的一種展示投影。

## 核心想法

Streamixer 是一個即時合成串流服務：將分開儲存的音檔、背景影像與字幕，在使用者觀看時即時合成為 HLS 影片串流。儲存的是素材，播放的是影片。

## 現狀

階段 1 已完成，階段 2 與 2.5 的程式碼已實作。系統使用 fMP4（CMAF）分段格式，所有分段由同一個 FFmpeg 預生成進程產出（無按需生成），透過 `-force_key_frames` 強制精確的分段切割點。分段請求以 WaitForSegment 等待預生成完成。待使用者確認播放體驗後可勾選階段 2 與 2.5 完成。

## 架構

- **語言**：Go 1.25
- **HTTP 路由**：go-chi/chi v5
- **媒體合成**：FFmpeg 子程序，以 `-preset ultrafast -tune stillimage -force_key_frames` 合成
- **分段格式**：fMP4 / CMAF（`-hls_segment_type fmp4`），init.mp4 + .m4s 分段
- **字幕處理**：燒入（burned-in），透過 FFmpeg `-vf subtitles=` 濾鏡
- **暫存**：tmpfs（/dev/shm），避免磁碟 I/O
- **串流協定**：HLS（HTTP Live Streaming），`#EXT-X-VERSION:7` + `#EXT-X-MAP`
- **部署方式**：Docker 容器化，多階段建置，Alpine 基礎映像 + font-noto-cjk

**程式碼結構**：
- `src/handler/` — HTTP handlers（串流、init.mp4、上傳、範例產生、健康檢查）
- `src/composer/` — FFmpeg 合成引擎、playlist 產生、音檔探測、PregenManager（背景預生成 + WaitForSegment/WaitForInit + singleflight）
- `src/media/` — 素材載入、驗證、資料結構
- `src/config/` — 環境變數設定管理
- `static/` — 測試用前端（上傳素材、HLS 播放器）

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

- [ ] 完成

交付：背景預生成所有分段，分段請求等待預生成完成，確保長音檔流暢播放。
前置條件：階段 1

**策略**：
- 首次請求 playlist 時，立即回傳手動計算的 playlist（顯示完整時長），同時背景啟動 FFmpeg 預生成
- 使用者請求分段時：已生成 → 直接回傳；尚未生成 → 等待預生成完成（WaitForSegment）
- 順序播放時預生成跑在使用者前面，等待時間極短
- 使用 singleflight 防止同一素材重複預生成，semaphore 限制並發數

**成功標準：**
- [ ] 順序播放時分段切換無中斷
- [ ] 跳轉後音訊位置正確
- [ ] 1 小時音檔的首次 playlist 請求在 5 秒內回應
- [ ] 跳轉到任意位置後 2 秒內開始播放

### 階段 2.5：fMP4 分段格式（消除破音）

- [ ] 完成

交付：使用 fMP4（CMAF）分段格式，消除分段切換時的音訊不連續。
前置條件：階段 2

**策略**：
- FFmpeg 使用 `-hls_segment_type fmp4` 產出 init.mp4 + .m4s 分段
- `-force_key_frames expr:gte(t,n_forced*N)` 強制精確的 keyframe 間距，確保分段時長與 playlist 一致
- 所有分段來自同一個 FFmpeg 進程（不使用按需生成），確保 init segment 與分段格式完全相容
- playlist 使用 `#EXT-X-VERSION:7` + `#EXT-X-MAP:URI="init.mp4"`

**成功標準：**
- [ ] 順序播放時分段切換無可感知的破音或中斷
- [ ] 跳轉後音訊銜接平順
- [ ] 既有的 HLS 播放器（hls.js、Safari）可正常播放 fMP4 分段

### 階段 3：智慧快取管理

- [ ] 完成

交付：自動管理 tmpfs 中合成分段的生命週期——熱門內容保留快取，冷門內容自動清除，避免記憶體耗盡。
前置條件：階段 2.5

**策略**：
- 以 LRU + TTL 管理每組素材的合成分段
- 記錄每組素材最後被存取的時間，超過 TTL（可設定）未被存取則自動清除
- 背景 goroutine 定期掃描，淘汰過期分段目錄
- tmpfs 容量接近上限時，主動淘汰最久未存取的內容
- 被清除的素材再次被請求時，重新觸發合成（對使用者透明）

**成功標準：**
- [ ] tmpfs 使用量不超過設定的容量上限
- [ ] 無人存取的素材在 TTL 過期後自動清除
- [ ] 清除後重新請求時，使用者體驗與首次請求一致
- [ ] 多人同時觀看同一內容時共享同一份快取，不重複合成
