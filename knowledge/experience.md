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

### WordPress Gutenberg Block JS 用 editor_script 不一定載入

- **理論說**：在 `register_block_type` 中指定 `editor_script` 會自動在 Gutenberg 編輯器載入該 JS 檔案。
- **實際發生**：純 JS（非 `@wordpress/scripts` 建置）的 block，`editor_script` 指定的 script 不會被載入到 Gutenberg 編輯器。搜尋 block 時顯示「找不到符合條件的搜尋結果」，但手動在 console 載入同一 JS 檔後 block 成功註冊。
- **解決方式**：改用 `admin_enqueue_scripts` hook，在 `post.php` 和 `post-new.php` 頁面直接 enqueue block JS。
- **教訓**：不使用 `@wordpress/scripts` 建置的 Gutenberg Block，MUST 使用 `admin_enqueue_scripts` 而非 `editor_script` 載入 JS。`editor_script` 的行為假設 JS 已由 build 系統處理。
- **來源**：階段 4 WordPress 外掛開發中的 block 搜尋除錯

### WordPress esc_url 對已編碼的 URL 雙重編碼

- **理論說**：`esc_url()` 是 WordPress 標準的 URL 跳脫函式，用於 HTML 屬性中的 URL 應該是安全的。
- **實際發生**：中文 slug 經過 `rawurlencode()` 後（例如 `%e6%b8%ac`），再經過 `esc_url()` 會把 `%` 編碼成 `%25`，變成 `%25e6%25b8%25ac`。瀏覽器無法解析雙重編碼的 URL，HLS 請求失敗（ERR_NAME_NOT_RESOLVED 或 404）。
- **解決方式**：在 HTML 的 `data-*` 屬性中使用 `esc_attr()` 取代 `esc_url()`，避免二次編碼。
- **教訓**：WordPress 中含有 percent-encoded 字元的 URL，在 HTML 屬性中 MUST 使用 `esc_attr()` 而非 `esc_url()`。`esc_url()` 會重新編碼 `%` 字元，破壞已編碼的 URL。
- **來源**：階段 4 WordPress 外掛中中文素材 slug 的播放 URL 錯誤

### HLS 預生成下載 MUST 等 EXT-X-ENDLIST 而非首段

- **理論說**：下載端點只要等 playlist 就緒（`WaitForPlaylist`，出現第一個 `.m4s` 就回傳）就能把所有分段串接為完整 MP4。
- **實際發生**：`WaitForPlaylist` 在首段產出時即返回，下載 handler 接著 `ReadDir` 收集當下已存在的分段，對長音檔（FFmpeg 尚未寫完）只拿到前幾段，使用者得到被截斷的影片檔。
- **解決方式**：新增 `WaitForComplete`，要求任務狀態為 `PregenCompleted` 且 playlist 含 `#EXT-X-ENDLIST` 才放行；下載前再用 `SegmentCount(duration)` 驗證實際分段數不少於預期。同時加上 `/progress/{id}` 端點讓前端輪詢進度。
- **教訓**：預生成是異步的，「playlist 存在」不等於「全部分段就緒」。任何需要整份媒體的端點（下載、打包、封存）MUST 等 `EXT-X-ENDLIST` 並比對分段數，不可只等首段。
- **來源**：使用者回報「下載影片檔不完整」，追查發現 `src/handler/download.go` 的等待條件錯誤。

### Gutenberg 編輯器不觸發傳統 meta box 的 $_POST save

- **理論說**：WordPress 的 `save_post` hook 會在文章儲存時觸發，傳統 meta box 裡的 `<input name="...">` 會透過 `$_POST` 傳進來，不論編輯器是 Gutenberg 或 Classic。
- **實際發生**：streamixer CPT 的素材選擇（音檔、背景、字幕）透過傳統 meta box + hidden input 送出。Gutenberg 編輯器改用 REST API 儲存文章，`$_POST` 裡沒有 meta box 欄位，`save_meta` 收到的全是 0，素材 id 被清空，導致 `/media/{id}/` 目錄空白，串流回 404。
- **解決方式**：對 streamixer CPT 強制使用傳統編輯器（`use_block_editor_for_post_type` filter 回傳 false）。
- **教訓**：依賴傳統 meta box 的 CPT，MUST 強制使用傳統編輯器，或改寫成 Gutenberg sidebar panel + `register_post_meta` + REST。混用 Gutenberg 與傳統 meta box 的 `$_POST` 流程會靜默丟資料。
- **來源**：使用者回報「測試影片讀不到」，追查發現 Gutenberg 儲存路徑未帶 meta box 欄位。

### 已清除本地檔的 WordPress 素材，新增選填欄位 MUST 走增量同步

- **理論說**：`sync_composition` 每次都重新上傳全套檔案（audio+background+可選 subtitle/transcript），即使管理員只動了其中一個選填欄位也無妨。
- **實際發生**：階段 6 的「同步後自動清除本地檔案」會把 `_streamixer_audio_id` / `_streamixer_background_id` 歸零並刪除 WP attachment 實體檔。之後管理員只想新加逐字稿，`sync_composition` 的前置檢查「若 audio_id 或 background_id 為 0 則回傳 pending」會直接 bail out，逐字稿根本沒送到後端，播放頁點下載得到 404「沒有逐字稿」。
- **解決方式**：以 `_streamixer_files_cleaned === '1'` 判斷「曾經同步成功」，若為真則進入增量模式：跳過 audio/background 的必填檢查、只把實際有變動的欄位（subtitle/transcript）送到後端；Go `/upload/{id}` 本來就接受任意欄位子集，直接寫入對應檔即可。
- **教訓**：帶有「同步後清除本地檔」機制的寫入管線 MUST 支援「增量同步」——用可靠的歷史旗標（如 `_streamixer_files_cleaned`）判斷曾經成功，而不是用當下的 sync_status（pending/error 會誤判）。`sync_status` 只記錄最近一次嘗試的結果，不是歷史事實。
- **來源**：使用者回報「明明上傳了逐字稿卻顯示沒有」，追查發現 `class-api.php` 的 bail-out 條件在本地檔已清除的情境下仍觸發。

### Go `time.NewTicker` 無法動態換頻率，改用 timer + kick channel

- **理論說**：背景清掃用 `time.NewTicker(interval)`，若需要讓使用者在執行期調整頻率，改一下變數下次 tick 就會生效。
- **實際發生**：`ticker.C` 的發射週期在 `NewTicker` 時就鎖定，後續改 interval 變數對 ticker 毫無影響；若硬要換頻率得 `ticker.Stop()` 再 `NewTicker`，但這需要跳脫現有的 `for range ticker.C` 迴圈結構。加上「WP 設定頁可即時調整清掃頻率」需求時，原本的 Sweeper 無法滿足。
- **解決方式**：改寫為 `for { timer := time.NewTimer(getInterval()); select { <-timer.C: ...; <-kick: ...; <-stop: ... } }`；`SetInterval` 更新欄位後往 `kick` 丟一個訊號，讓迴圈立即重建 timer 以新頻率起算。
- **教訓**：需要在執行期動態調整頻率的排程迴圈 MUST 使用每輪重建的 `time.NewTimer` + kick channel，不得依賴 `time.NewTicker`。同理適用於所有「初始化時鎖定、之後不可調整」的 Go 標準函式庫原語。
- **來源**：階段 8.5 把快取清掃頻率搬到 WP 設定頁時，發現既有 `Sweeper` 架構無法支援執行期調整。

### WordPress post_name 已是 URL 編碼，不需再 encode

- **理論說**：組合 URL 時應該用 `rawurlencode()` 或 `urlencode()` 編碼路徑段，確保特殊字元被正確處理。
- **實際發生**：WordPress 的 `post_name`（slug）在資料庫中已經是 URL 編碼形式（中文字 → `%e6%b8%ac%e8%a9%a6`）。再用 `rawurlencode()` 會把 `%` 編碼成 `%25`，產生 `%25e6%25b8%25ac` 的雙重編碼 URL。同樣的錯誤在串流 URL 和下載 URL 中各出現一次。
- **解決方式**：直接拼接 `post_name` 到 URL 中，不再 encode。
- **教訓**：WordPress 的 `post_name` MUST 直接用於 URL 拼接，不需要 `rawurlencode()` 或 `urlencode()`。在使用任何 encode 函式前，先確認資料來源是否已經是編碼後的形式。
- **來源**：階段 7 下載功能的 404 錯誤（同一問題第二次出現）
