# 資料模型：fMP4 分段格式

## 實體變更

### Init Segment（初始化段）— 新增

包含解碼器配置（moov atom）的檔案，每個素材組合一個。

| 屬性 | 說明 |
|------|------|
| compositionID | 對應的素材組合 ID |
| path | init.mp4 檔案路徑 |

**生命週期**：與該素材組合的分段共存亡。預生成時自動產生。

### Media Segment（媒體分段）— 修改

| 變更 | 原本 | 改為 |
|------|------|------|
| 副檔名 | .ts | .m4s |
| 格式 | MPEG-TS | fMP4 (CMAF) |
| 解碼器初始化 | 每段獨立 | 共享 init segment |

### HLSStream — 修改

| 變更 | 原本 | 改為 |
|------|------|------|
| playlistVersion | 3 | 7 |
| segmentExtension | .ts | .m4s |
| initSegment | 無 | init.mp4 路徑 |

## 與既有實體的關係

```
MediaComposition 1──0..1 PregenTask
PregenTask 1──1 InitSegment
PregenTask 1──N MediaSegment(.m4s)
```
