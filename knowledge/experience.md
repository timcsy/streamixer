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

### fMP4 按需生成的分段與預生成的 init segment 不相容

- **理論說**：只要 FFmpeg 的編碼參數（codec、解析度、sample rate）相同，不同 FFmpeg 進程產生的 fMP4 init segment 應該功能相容，按需生成的 .m4s 分段可以搭配預生成的 init.mp4 使用。
- **實際發生**：按需生成（`-f mp4 -movflags +empty_moov+cmaf`）產出的 .m4s 分段搭配預生成的 init.mp4 後，播放器播完第一段就不斷重整。不同 FFmpeg 進程的 moov atom 在 metadata、時間基底等方面有微妙差異。
- **解決方式**：完全移除按需生成，所有分段（包含 init.mp4）由同一個 FFmpeg 預生成進程產出。分段請求改為等待預生成完成（WaitForSegment）。
- **教訓**：fMP4 的 init segment 與 media segment MUST 來自同一個 FFmpeg 進程。不同進程即使參數相同，也不保證相容。需要混合來源時，應使用 MPEG-TS（每段獨立）而非 fMP4。
- **來源**：階段 2.5 實作過程中的播放器相容性測試

### FFmpeg -tune stillimage 導致 GOP 遠超 -hls_time 目標

- **理論說**：FFmpeg 的 `-hls_time 6` 應該產生約 6 秒一段的 HLS 分段。
- **實際發生**：搭配 `-tune stillimage` 時，libx264 將 GOP 設為很大的值（因為靜態圖片幾乎不變），實際分段變成 10 秒而非 6 秒。手動計算的 playlist 宣告 6 秒分段，但 FFmpeg 產出的分段是 10 秒，導致播放時間軸錯亂（亂跳）。
- **解決方式**：加入 `-force_key_frames "expr:gte(t,n_forced*6)"` 強制每 6 秒一個 keyframe。FFmpeg 的 HLS muxer 只在 keyframe 處切割分段，有了強制 keyframe，分段時長就精確匹配。
- **教訓**：使用 `-tune stillimage` 時 MUST 搭配 `-force_key_frames` 明確指定 keyframe 間距。不能依賴 `-hls_time` 的「目標」行為——它只是建議值，實際切割取決於 keyframe 位置。
- **來源**：階段 2.5 實作過程中比對 FFmpeg 產出的 playlist 與手動計算的 playlist

### FFmpeg VOD 模式 playlist 要等全部合成完才寫出

- **理論說**：請求 playlist 時可以立即回傳，因為 FFmpeg 邊合成邊寫入 playlist。
- **實際發生**：FFmpeg 的 `-hls_list_size 0`（VOD 模式）在合成全部完成後才寫出 `index.m3u8`。等待 playlist 產生的 HTTP 請求會 block 直到 FFmpeg 跑完，長音檔（如 1 小時）可能需要 100 秒以上，導致 hls.js 的 manifest 請求 timeout。
- **解決方式**：playlist 改由程式碼手動計算並立即回傳（根據音檔 duration 計算分段數與時長），不依賴 FFmpeg 產生的 playlist。搭配 `-force_key_frames` 確保 FFmpeg 實際的分段時長與手動計算一致。
- **教訓**：對於即時回應的場景，不應依賴 FFmpeg 的 HLS playlist 輸出。手動計算 playlist 並搭配 `-force_key_frames` 確保一致性，是更可控的做法。
- **來源**：階段 2.5 實作過程中的 hls.js manifest timeout 錯誤
