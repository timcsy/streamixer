# 願景

## 問題陳述

教會的 WordPress 網站存放大量講道影片，但這些影片的內容本質上只是錄音配上背景圖片與字幕。以完整影片格式儲存造成高昂的儲存成本，而這些成本是不必要的——因為影片只是底層素材的一種展示投影。

## 核心想法

Streamixer 是一個即時合成串流服務：將分開儲存的音檔、背景影像與字幕，在使用者觀看時即時合成為 HLS 影片串流。儲存的是素材，播放的是影片。

## 現狀

階段 1（最小可用串流）已完成。系統可接受音檔（MP3/WAV）、背景圖片（JPG/PNG）與字幕（SRT/VTT），即時合成為 HLS 影片串流。共 16 個自動化測試（8 單元 + 8 整合）全部通過，Docker 映像可正常建置與啟動。

## 架構

- **語言**：Go 1.25
- **HTTP 路由**：go-chi/chi v5
- **媒體合成**：FFmpeg 子程序，以 `-preset ultrafast -tune stillimage` 合成
- **字幕處理**：燒入（burned-in），透過 FFmpeg `-vf subtitles=` 濾鏡
- **暫存**：tmpfs（/dev/shm），避免磁碟 I/O
- **串流協定**：HLS（HTTP Live Streaming），相容性廣，適合 WordPress 嵌入播放
- **部署方式**：Docker 容器化，多階段建置，Alpine 基礎映像

**程式碼結構**：
- `src/handler/` — HTTP handlers（串流、上傳、範例產生、健康檢查、錯誤回應）
- `src/composer/` — FFmpeg 合成引擎、playlist 產生、音檔探測
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
