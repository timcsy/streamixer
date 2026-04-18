# Contract: GET /audio/{id}

## 目的

回傳指定素材組合的原始音檔，供訪客下載。

## 請求

- Method：`GET`
- Path：`/audio/{id}`（`id` 為 composition_id，可能含 URL 編碼的中文）
- 認證：無（公開端點，與 `/stream/*`、`/download/*` 同類）
- Headers：無必要欄位；CORS 沿用 `corsMiddleware`

## 成功回應

- Status：`200 OK`
- Headers：
  - `Content-Type`：依副檔名設定（`.mp3` → `audio/mpeg`、`.wav` → `audio/wav`、`.m4a` → `audio/mp4`、其他 → `application/octet-stream`）
  - `Content-Length`：檔案大小
  - `Content-Disposition`：`attachment; filename="<id>.<ext>"; filename*=UTF-8''<percent-encoded-utf8>`
  - `Accept-Ranges`：`bytes`（沿用 `http.ServeFile` 行為）
- Body：原始音檔內容（可支援 Range request）

## 錯誤回應

| Status | 條件 | Body |
|--------|------|------|
| 404 | 素材組合不存在 | `{"error":"..."}` |
| 404 | 素材組合存在但音檔缺失 | `{"error":"..."}` |
| 500 | 其他伺服器錯誤 | `{"error":"..."}` |

## 行為契約

- MUST 以串流方式回應（`http.ServeFile`），不將檔案整個讀入記憶體
- MUST 支援 HTTP Range，利於大檔續傳
- MUST 在檔名含非 ASCII 時同時提供 `filename*=UTF-8''...` 並維持 `filename=` fallback
- MUST NOT 修改或重新編碼音檔內容

## 驗收測試

1. 上傳 `test.mp3` 到 composition `abc` → `GET /audio/abc` 回傳與上傳相同位元，`Content-Type: audio/mpeg`。
2. 上傳 `範例.wav` 到中文 slug `測試` → `GET /audio/<url-encoded>` 成功，`Content-Disposition` 正確解析。
3. 刪除 `audio.*` 後 → `GET /audio/abc` 回 404 JSON 錯誤。
