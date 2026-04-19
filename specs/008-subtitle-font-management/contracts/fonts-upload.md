# Contract: POST /fonts

## 目的

上傳一個使用者字體檔。

## 請求

- Method：`POST`
- Path：`/fonts`
- 認證：API Key（沿用既有 `apiKeyMiddleware`）
- Content-Type：`multipart/form-data`
- Form field：`font`（檔案，必填）

## 成功回應

- Status：`201 Created`
- Content-Type：`application/json`
- Body：

```json
{
  "id": "8b7c1f2e-...",
  "family_name": "Church Font",
  "source": "user",
  "ext": "ttf",
  "size": 1234567,
  "uploaded_at": "2026-04-19T10:30:00Z"
}
```

## 錯誤回應

| Status | 條件 | Body |
|--------|------|------|
| 400 | 欄位缺失、副檔名不符、檔頭驗證失敗、family name 不合法字元 | `{"error":"..."}` |
| 413 | 單檔大小超過 10 MB | `{"error":"..."}` |
| 409 | 字體總數已達 50，或 family name 與既有 user 字體重複 | `{"error":"..."}` |
| 500 | 寫入失敗 | `{"error":"..."}` |

## 行為契約

- MUST 以檔頭 magic number 驗證檔案格式（`\x00\x01\x00\x00` / `OTTO` / `ttcf`）
- MUST 解析字體 name table 取 family name；無法解析時拒絕
- MUST 建立 `/usr/share/fonts/user/<id>.<ext>` symlink 指向 `/fonts/user/<id>.<ext>`，讓 fontconfig 可發現
- MUST 寫入 metadata JSON 供後續 list 使用，無需重新 parse 字體
- MUST 在驗證失敗時清理已寫入的臨時檔，不留垃圾
- MAY 在上傳完成後呼叫 `fc-cache -f /usr/share/fonts/user` 作為保險

## 驗收測試

1. 上傳合法 `.ttf` → 201，列表 GET 多一筆
2. 上傳 `.txt` 改名 `.ttf` → 400
3. 上傳 15MB `.ttf` → 413
4. 上傳第 51 個字體 → 409
5. 上傳時不帶 X-API-Key → 401
