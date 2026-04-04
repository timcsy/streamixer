# 資料模型：HLS 即時合成串流

## 實體

### MediaComposition（素材組合）

代表一組待合成的媒體素材，是系統的核心實體。

| 欄位 | 說明 | 必�� |
|------|------|------|
| id | 唯一識別碼 | 是 |
| audio | 音檔來源路徑 | 是 |
| background | 背景影像路徑 | 是 |
| subtitle | 字幕檔路徑 | 否 |

**驗證規則**：
- audio MUST 為支援的格式（MP3、WAV）
- background MUST 為支援的格式（JPG、PNG）
- subtitle 若提供，MUST 為支援的格式（SRT、VTT）
- 所有路徑指向的檔案 MUST 存在且可讀取

### Audio（音檔）

| 屬性 | 說明 |
|------|------|
| path | 檔案路徑 |
| format | 檔案格式（mp3、wav） |
| duration | 音訊時長（秒） |

### Background（背景影像）

| 屬性 | 說明 |
|------|------|
| path | 檔案路徑 |
| format | 檔案格式（jpg、png） |
| width | 原始寬度（像素） |
| height | 原始高度（像素） |

### Subtitle（字幕）

| 屬性 | 說明 |
|------|------|
| path | 檔案路徑 |
| format | 檔案格式（srt���vtt） |

### HLSStream（HLS 串流）

合成後的輸出，存在於 tmpfs 中。

| 屬性 | 說明 |
|------|------|
| compositionId | 對應的素材組合 ID |
| playlistPath | .m3u8 檔案路徑 |
| segmentDir | .ts 分段所在目錄 |
| segmentDuration | 每個分段的秒數（預設 6 秒） |
| totalSegments | 總分段數（由音檔長度計算） |

## 實體關係

```
MediaComposition 1──1 Audio
MediaComposition 1──1 Background
MediaComposition 1──0..1 Subtitle
MediaComposition 1──1 HLSStream
```

一個 MediaComposition 包含恰好一個 Audio 和��個 Background，可選擇性包含一個 Subtitle。播放請求時產生對應�� HLSStream。
