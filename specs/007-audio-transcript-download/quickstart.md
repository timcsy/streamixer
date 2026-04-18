# Quickstart: 音檔與逐字稿下載

> 本文件描述本功能合併後，使用者與開發者如何快速驗證其正常運作。

## 管理員：上傳並同步逐字稿

1. 後台 → Streamixer → 新增或編輯素材組合
2. 填寫標題（例如 `2026-01-07-主日證道`）
3. 選擇「音檔」「背景圖片」（必填）、字幕（選填）
4. **新增步驟**：在「逐字稿」欄位選擇一份檔案（txt / pdf / doc / docx / md，≤ 50MB）
5. 發布／更新文章
6. 下方同步狀態顯示「✓ 已同步至 Streamixer」
7. 若開啟「同步後自動清除本地檔案」：媒體庫中本素材相關的附件（含逐字稿）被移除，編輯頁仍顯示檔名與「（已同步至 Streamixer，本地已清除）」

## 訪客：從播放頁下載

1. 進入文章頁 `/streamixer/<slug>/`
2. 播放器下方 toolbar 看到三個按鈕：
   - ⬇ 下載影片（MP4，等待合成進度）
   - ⬇ 下載音檔（原始 MP3/WAV/…）
   - ⬇ 下載逐字稿（原始檔；若無逐字稿則此按鈕不顯示）
3. 點擊音檔或逐字稿 → 瀏覽器直接下載

## 開發者：本機驗證端點

```bash
# 假設 Streamixer 服務跑在 http://localhost:8080
curl -I http://localhost:8080/audio/test-composition
# 期望：HTTP 200，Content-Type: audio/mpeg，Content-Disposition: attachment

curl -I http://localhost:8080/transcript/test-composition
# 期望：HTTP 200，Content-Type: application/pdf，Content-Disposition: attachment

# 中文 slug 測試（請先用 %e6... 形式的 URL encoded path）
curl -I "http://localhost:8080/audio/%e6%b8%ac%e8%a9%a6"
```

## 開發者：跑整合測試

```bash
go test ./tests/integration/... -run TestAudioDownload -v
go test ./tests/integration/... -run TestTranscriptDownload -v
```

## 失敗排查

| 現象 | 可能原因 | 建議 |
|------|----------|------|
| 下載音檔 404 | 素材未同步或音檔缺失 | 確認 `_streamixer_sync_status === 'synced'`；重新儲存素材強制同步 |
| 下載逐字稿 404 | 未上傳逐字稿 | 編輯素材上傳逐字稿並儲存 |
| 檔名亂碼 | 瀏覽器不支援 RFC 5987 | 新版 Chrome/Safari/Firefox 皆支援；若仍亂碼檢查 `Content-Disposition` 格式 |
| 下載被當作網頁顯示 | `Content-Disposition` 未設 `attachment` | 檢查回應標頭 |
| 中文 slug 404 | URL 雙重編碼 | 檢查 `class-api.php` 的 `get_audio_url` / `get_transcript_url` 是否直接拼 `post_name`（不再 `rawurlencode`） |
