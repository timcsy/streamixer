# 研究筆記：HLS 即時合成串流

## 決策 1：程式語言

**選擇**：Go

**理由**：
- FFmpeg 整合：透過 `os/exec` 呼叫 FFmpeg 子程序，最可靠且無需維護繁瑣的 binding
- HTTP 效能：Go 標準函式庫 `net/http` 搭配 goroutine 可處理大量並發連線
- Docker 映像大小：靜態編譯的 Go 二進位檔 + Alpine + FFmpeg 約 80MB，遠小於 Python（400MB+）或 Node.js（300MB+）
- 並發控制：goroutine 與 channel 適合管理每個請求對應的 FFmpeg 子程序生命週期

**考慮過的替代方案**：
- Python：開發速度快，但並發模型較弱，Docker 映像較大
- Node.js��原型開發方便，但對長時間執行的 FFmpeg 子程序管理較差
- Rust：效能最佳，但開發成本高，對此專案規模而言過度

## 決策 2：FFmpeg 合成模式

**選擇**：FFmpeg 子程序，輸出至 tmpfs（/dev/shm）

**理由**：
- FFmpeg 的 HLS muxer（`-f hls`）必須寫入磁碟，無法 pipe 至 stdout（因為需同時寫入 .m3u8 與多個 .ts）
- 使用 Docker 容器內的 tmpfs（/dev/shm）可達到近乎零 I/O 延遲
- 合成指令模式：`ffmpeg -loop 1 -i bg.jpg -i audio.mp3 -vf "subtitles=subs.srt" -c:v libx264 -tune stillimage -preset ultrafast -c:a aac -shortest -f hls ...`

**隨機存取（Seek）策略**：
- 可預先計算完整 .m3u8 playlist（根據音檔長度推算分段時間）
- 使用者 seek 時，以 FFmpeg `-ss` 與 `-t` 參數按需生成對應的 .ts 分段

**考慮過���替代方案**：
- 串流 pipe 至 stdout：FFmpeg HLS muxer 不支援
- 預先生成所有分段：違反即時合成原則，佔用記憶體

## 決策 3：HTTP 框架

**��擇**：Go 標準函式庫 `net/http` + `github.com/go-chi/chi/v5` 路由

**理由**：
- chi 輕量、與標準函式庫相容，無過多抽象
- 只需兩個端點：playlist 與分段
- 符合簡約原則

**考慮過的替代方案**：
- Gin、Echo：功能較多但此專案不需要
- 純 `net/http`：可行但路由較不便

## 決策 4：字幕處理方式

**選擇**：燒入（Burned-in）字幕

**理由**：
- 使用 FFmpeg 的 `-vf "subtitles=file.srt"` 將字幕渲染至畫面中
- 最簡單，所有播放器都支援，無需額外的字幕 playlist
- 適合教會場景：字幕通常預設開啟，不需要切換功能

**考���過的替代方案**：
- WebVTT sidecar 字幕軌：需要額外的 subtitle playlist，增加 HLS manifest 複雜度，但可讓使用��切換字幕開關。留作未來擴展。

## 決策 5：測試策略

**選擇**：`go test` + `ffprobe` 驗證

**理由**：
- 單元測試：`go test` 搭配 `net/http/httptest` 測試 HTTP handler
- 整合測試：用 `ffprobe` 驗證產生的 .m3u8 與 .ts 分段格式、時長、編碼
- 端對端測��：用 `ffprobe -i index.m3u8` 驗證完整串流可播放性

**驗證工具**：
- `ffprobe`：驗證分段時長、編碼、串流資訊
- `go test` + `httptest`：HTTP handler 單元測試
