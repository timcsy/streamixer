# 資料模型：即時跳轉與背景預生成

## 實體

### PregenTask（預生成任務）

管理一個素材組合的背景預生成狀態。

| 欄位 | 說明 |
|------|------|
| compositionID | 對應的素材組合 ID |
| status | 任務狀態：pending、running、completed、failed |
| totalSegments | 總分段數 |
| completedSegments | 已完成的分段數 |
| outputDir | 分段輸出目錄 |
| error | 失敗時的錯誤訊息 |

**狀態轉換**：
```
pending → running → completed
                  → failed
```

- pending：已建立但尚未開始
- running：FFmpeg 正在背景執行
- completed：所有分段已生成
- failed：FFmpeg 執行失敗，回退為按需生成

### SegmentStatus（分段狀態）

每個分段的生成狀態，由檔案系統判定。

| 狀態 | 判定方式 |
|------|----------|
| 未生成 | 分段檔案不存在 |
| 已完成 | 分段檔案存在且大小 > 0 |

**備註**：不需要「生成中」狀態——預生成由 FFmpeg HLS muxer 一次產生所有分段，寫入為原子操作。按需生成使用 singleflight 確保不重複。

## 與既有實體的關係

```
MediaComposition 1──0..1 PregenTask
PregenTask 1──N SegmentStatus（透過檔案系統）
```

一個 MediaComposition 最多有一個 PregenTask。PregenTask 的分段狀態由 tmpfs 中的檔案是否存在決定。
