# Phase 1 Data Model: 音檔與逐字稿下載

## 後端（Go）

### TranscriptFile（新增）

位置：`src/media/types.go`

| 欄位 | 型別 | 說明 |
|------|------|------|
| `Path` | `string` | 絕對路徑，例如 `/data/media/<id>/transcript.pdf` |
| `Ext`  | `string` | 副檔名（不含點），例如 `pdf` / `txt` / `md` / `doc` / `docx` |
| `Size` | `int64`  | 檔案大小（bytes） |

驗證規則：
- `Path` 存在且為一般檔案
- `Ext` 屬於白名單 {`txt`, `pdf`, `doc`, `docx`, `md`}

### MediaComposition（修改）

位置：`src/media/composition.go`

| 欄位 | 型別 | 備註 |
|------|------|------|
| `ID` | `string` | 既有 |
| `Audio` | `*AudioFile` | 既有 |
| `Background` | `*BackgroundFile` | 既有 |
| `Subtitle` | `*SubtitleFile` | 既有，選填 |
| **`Transcript`** | `*TranscriptFile` | **新增，選填**。僅當 `/media/{id}/transcript.*` 存在時 non-nil。 |

### Loader（修改）

位置：`src/media/loader.go`

新增方法 / 擴充：
- `loadTranscript(compDir string) (*TranscriptFile, error)`：以 `filepath.Glob(filepath.Join(compDir, "transcript.*"))` 搜尋，取第一個 match；若無則回傳 `nil, nil`（非錯誤）。
- 於 `Load(id)` 結尾呼叫 `loadTranscript` 並賦值到 `comp.Transcript`。

## 前端 WordPress（PHP）

### Post Meta（新增 / 擴充）

| Meta key | 型別 | 說明 |
|----------|------|------|
| `_streamixer_transcript_id` | int | WP attachment ID（與音檔、背景、字幕同模式；同步後清除會歸零） |
| `_streamixer_transcript_id_filename` | string | 原檔名（含副檔名），同步後清除時寫入，用於編輯頁與按鈕顯示判斷 |

其餘 meta（`_streamixer_sync_status`、`_streamixer_files_cleaned` 等）沿用既有語意。

### CPT 欄位狀態機

```
未選擇 ──(上傳)──▶ 已選擇未同步 ──(儲存/同步成功)──▶ 已同步有本地檔
                         │                                  │
                         │                            (自動清除)
                         │                                  ▼
                         │                        已同步本地清除（保留檔名）
                         ▼
                    (選擇新檔)
                         │
                         ▼
                    已選擇未同步（覆蓋）
```

「下載逐字稿」按鈕顯示條件：`_streamixer_sync_status === 'synced'` 且（`_streamixer_transcript_id` 存在 或 `_streamixer_transcript_id_filename` 非空）。

## 檔案系統佈局

```
/data/media/<composition_id>/
├── audio.<ext>          # 既有
├── background.<ext>     # 既有
├── subtitle.<ext>       # 既有，選填
└── transcript.<ext>     # 新增，選填
```

## 無狀態性

所有新增端點為唯讀，不影響現有狀態模型。逐字稿上傳由既有 `/upload/{id}` 端點擴充 multipart 解析處理，不引入新的寫入路徑。
