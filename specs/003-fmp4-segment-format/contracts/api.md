# API 契約：fMP4 分段格式

## 端點變更

### GET /stream/{id}/index.m3u8

**變更**：playlist 內容改變。
- 新增 `#EXT-X-VERSION:7`（從 3 升級）
- 新增 `#EXT-X-MAP:URI="init.mp4"` 指向 init segment
- 分段檔名從 `seg_000.ts` 改為 `seg_000.m4s`

**回應範例**：
```
#EXTM3U
#EXT-X-VERSION:7
#EXT-X-TARGETDURATION:6
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-MAP:URI="init.mp4"
#EXTINF:6.000000,
seg_000.m4s
#EXTINF:6.000000,
seg_001.m4s
#EXT-X-ENDLIST
```

### GET /stream/{id}/init.mp4（新增）

取得初始化段。

**成功回應（200）**：
- Content-Type: `video/mp4`
- Body: fMP4 init segment 二進位資料

**錯誤回應**：
- 404：素材組合不存在

### GET /stream/{id}/seg_{n}.m4s（修改副檔名）

**變更**：從 `.ts` 改為 `.m4s`。
- Content-Type 從 `video/mp2t` 改為 `video/mp4`
- 其餘行為不變

### GET /health

不變。
