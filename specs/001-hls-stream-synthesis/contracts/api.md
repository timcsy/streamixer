# API 契約：HLS 即時合成串流

## 端點

### GET /stream/{id}/index.m3u8

取得指定素材組合的 HLS playlist。

**路徑參數**：
- `id`：素材組合的唯一識別碼

**成功回應（200）**：
- Content-Type: `application/vnd.apple.mpegurl`
- Body: HLS master playlist 內容

**錯誤回應**：
- 404：指定的素材組合不存在
- 400：素材格式不支援或檔案損毀
- 500：合成過程中發生內部錯誤

**範例回應**：
```
#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:6
#EXT-X-MEDIA-SEQUENCE:0
#EXTINF:6.000000,
seg_000.ts
#EXTINF:6.000000,
seg_001.ts
#EXTINF:4.500000,
seg_002.ts
#EXT-X-ENDLIST
```

---

### GET /stream/{id}/seg_{n}.ts

取得指定分段的影片資料。

**路徑參數**：
- `id`：素材組合的唯一識別碼
- `n`：分段編號（從 000 開始，零填充三位數）

**成功回應（200）**：
- Content-Type: `video/mp2t`
- Body: MPEG-TS 分段二進位資料

**錯誤回應**：
- 404：指定的素材組合不存在，或分段編號超出範圍
- 500：合成過程中發生內部錯誤

---

### GET /health

健康檢查端點。

**成功回應（200）**：
- Content-Type: `application/json`
- Body: `{"status": "ok"}`

## 錯誤回應格式

所有錯誤回應統一格式：

```json
{
  "error": "錯誤描述訊息"
}
```

Content-Type: `application/json`
