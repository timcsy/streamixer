# 研究筆記：即時跳轉與背景預生成

## 決策 1：並發控制機制

**選擇**：golang.org/x/sync/singleflight

**理由**：
- 確保同一素材組合只有一個背景預生成在執行
- 確保同一分段的按需生成不會重複執行
- Go 官方擴展套件，API 簡潔，無需自建鎖機制

**考慮過的替代方案**：
- sync.Mutex + map：可行但需手動管理鎖粒度
- channel-based coordination：較複雜，此場景不需要

## 決策 2：預生成策略

**選擇**：FFmpeg 完整 HLS 合成（BuildFFmpegArgs）而非逐段生成

**理由**：
- FFmpeg 的 HLS muxer 一次產生所有分段，時間戳自動連續
- 避免逐段生成時的時間戳對齊問題（這正是目前 seek 偏移的根因）
- 對靜態圖片+音檔場景，完整合成的 CPU 效率高於多次啟動 FFmpeg

**考慮過的替代方案**：
- 逐段預生成（按順序呼叫 BuildSegmentArgs）：時間戳仍可能不連續
- 使用 FFmpeg segment muxer 而非 HLS muxer：更低階，需要自行管理 playlist

## 決策 3：按需生成 fallback 的時間戳修正

**選擇**：按需生成時使用 FFmpeg 的 `-output_ts_offset` 參數對齊時間戳

**理由**：
- 當預生成尚未完成時，按需生成的分段 MUST 與 playlist 宣告的時間戳一致
- `-output_ts_offset` 可設定輸出的起始 PTS，確保與 HLS playlist 對齊
- 搭配 `-ss`（輸入 seek）+ `-t`（分段長度）+ `-output_ts_offset`（輸出時間戳）三者組合

**考慮過的替代方案**：
- 不修正時間戳：會導致音訊偏移（目前的問題）
- 使用 `-copyts`：保留原始時間戳，但與 HLS muxer 的行為不同，可能造成其他問題
