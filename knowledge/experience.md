# 經驗

## 教訓

### FFmpeg 字幕濾鏡的中文字體 fallback

- **理論說**：FFmpeg 的 `subtitles=` 濾鏡會自動找到系統字體來渲染��幕文字。
- **實際發生**：在 macOS 上，FFmpeg 嘗試載入 PingFangUI.ttc 時出現多次 `Error opening font` 錯誤，最終 fallback 到 PingFangTC-Regular 才成功渲染中文字幕。雖然最終結果正確，但錯誤訊息令人困惑。
- **解決方式**：macOS 上自動 fallback 解決了問題，但 Docker Alpine 容器中預設沒有中文字體，部署時需額外安裝（例如 `apk add font-noto-cjk`）。
- **教訓**：涉及非拉丁文字的字幕渲染時，MUST 確保目標環境有對應字體。Docker 映像建置時應明確安裝所需字體。
- **來源**：階段 1 整合測試 TestSubtitleComposition 的 FFmpeg 輸出

### MPEG-TS 分段切換時的音訊不連續（破音）

- **理論說**：FFmpeg 的 HLS muxer 將影片切成 .ts 分段後，播放器依序播放應該是無縫銜接的。
- **實際發生**：分段切換時出現可感知的小破音。原因是每個 .ts 分段都需要獨立的 AAC 解碼器初始化（PES header + ADTS），切換時解碼器狀態不連續，造成 priming samples 引入的微小音訊間隙。不論是預生成（同一 FFmpeg 進程）還是按需生成（不同進程），都會出現。
- **解決方式**：決定將分段格式從 MPEG-TS(.ts) 改為 fragmented MP4（fMP4 / CMAF）。fMP4 使用共享的初始化段（init segment），解碼器狀態在分段之間連續，音訊銜接無縫。hls.js 和 Safari 都支援。
- **教訓**：對於音訊品質要求高的場景，MUST 使用 fMP4（CMAF）而非 .ts 作為 HLS 分段格式。.ts 格式在影片為主的場景下破音不明顯，但在靜態圖片+音訊的場景下，分段邊界的音訊不連續會被放大感知。
- **來源**：階段 2 實作後的使用者測試回報
