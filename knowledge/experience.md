# 經驗

## 教訓

### FFmpeg 字幕濾鏡的中文字體 fallback

- **理論說**：FFmpeg 的 `subtitles=` 濾鏡會自動找到系統字體來渲染��幕文字。
- **實際發生**：在 macOS 上，FFmpeg 嘗試載入 PingFangUI.ttc 時出現多次 `Error opening font` 錯誤，最終 fallback 到 PingFangTC-Regular 才成功渲染中文字幕。雖然最終結果正確，但錯誤訊息令人困惑。
- **解決方式**：macOS 上自動 fallback 解決了問題，但 Docker Alpine 容器中預設沒有中文字體，部署時需額外安裝（例如 `apk add font-noto-cjk`）。
- **教訓**：涉及非拉丁文字的字幕渲染時，MUST 確保目標環境有對應字體。Docker 映像建置時應明確安裝所需字體。
- **來源**：階段 1 整合測試 TestSubtitleComposition 的 FFmpeg 輸出
