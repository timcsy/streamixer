# 研究筆記：影片下載功能

## 決策 1：fMP4 分段合併方式

**選擇**：Go 中直接串接二進位（`io.Copy` init.mp4 + 所有 .m4s）

**理由**：
- fMP4 的設計本身就支援串接：init.mp4 含 `ftyp` + `moov`，每個 .m4s 含 `moof` + `mdat`
- 直接串接產生的是合法的 fragmented MP4，VLC、系統播放器都能播放
- 不需要 FFmpeg 重新編碼或重新 mux，零額外 CPU 成本
- 可串流回傳（不需全部讀入記憶體），直接寫入 http.ResponseWriter

**考慮過的替代方案**：
- FFmpeg `ffmpeg -i index.m3u8 -c copy output.mp4`：產生更標準的 MP4（含完整 moov），但需要啟動 FFmpeg 子程序，耗時且佔用資源
- 只提供 HLS 串流下載（m3u8 + 分段）：使用者需要特殊工具才能離線播放

## 決策 2：Content-Length 計算

**選擇**：先 `os.Stat` 所有分段檔案加總大小，再設定 Content-Length

**理由**：
- `os.Stat` 成本極低（只讀 inode，不讀檔案內容）
- 有 Content-Length 讓瀏覽器能顯示下載進度
- 沒有 Content-Length 的話瀏覽器只能顯示「下載中」，UX 較差

## 決策 3：下載存取控制

**選擇**：與串流端點一致——預設公開，可透過 API Key 限制

**理由**：
- 如果使用者能看串流，邏輯上也應該能下載
- 不需要額外的認證機制
