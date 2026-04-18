# Contract: GET /transcript/{id}

## 目的

回傳指定素材組合的逐字稿全文檔，供訪客下載。

## 請求

- Method：`GET`
- Path：`/transcript/{id}`
- 認證：無（公開端點）

## 成功回應

- Status：`200 OK`
- Headers：
  - `Content-Type`：依副檔名（`.txt` → `text/plain; charset=utf-8`、`.md` → `text/markdown; charset=utf-8`、`.pdf` → `application/pdf`、`.doc` → `application/msword`、`.docx` → `application/vnd.openxmlformats-officedocument.wordprocessingml.document`）
  - `Content-Length`
  - `Content-Disposition`：`attachment; filename="<id>.<ext>"; filename*=UTF-8''<percent-encoded-utf8>` — 注意即使為 txt/md 也 MUST 是 `attachment`，避免瀏覽器直接顯示
  - `X-Content-Type-Options`：`nosniff`（防止瀏覽器把 txt/md 當作 HTML 解析）
- Body：逐字稿原始檔案內容

## 錯誤回應

| Status | 條件 | Body |
|--------|------|------|
| 404 | 素材組合不存在 | `{"error":"..."}` |
| 404 | 逐字稿未上傳 | `{"error":"..."}` |
| 500 | 其他伺服器錯誤 | `{"error":"..."}` |

## 行為契約

- MUST 以串流方式回應，不讀整個檔案進記憶體
- MUST 以 `attachment` disposition 強制下載（不論格式）
- MUST 保留上傳時的副檔名於回應檔名中
- MUST 支援中文 slug 與中文原始檔名的 RFC 5987 下載

## 驗收測試

1. 上傳 `逐字稿.pdf` 到 composition `sermon-2026-01` → `GET /transcript/sermon-2026-01` 回傳 PDF 位元相符。
2. 上傳 `notes.md` → `GET` 回應 `Content-Type: text/markdown; charset=utf-8` 且有 `X-Content-Type-Options: nosniff`。
3. 無逐字稿的素材 → `GET /transcript/<id>` 回 404。
4. 中文 slug `測試` 含 `/` encoded URL → 成功下載，檔名正確。
