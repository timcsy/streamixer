# Contract: GET /fonts

## 目的

回傳 Streamixer 目前可用的所有字體（系統內建 + 使用者上傳）。

## 請求

- Method：`GET`
- Path：`/fonts`
- 認證：無（公開讀取，WP 設定頁需顯示清單）
- Query params：無

## 成功回應

- Status：`200 OK`
- Content-Type：`application/json`
- Body：

```json
{
  "default_family": "Noto Sans CJK TC",
  "fonts": [
    {
      "id": "noto-sans-cjk-tc",
      "family_name": "Noto Sans CJK TC",
      "source": "system",
      "ext": "ttc",
      "size": 17928236,
      "uploaded_at": null
    },
    {
      "id": "8b7c1f2e-...",
      "family_name": "Church Font",
      "source": "user",
      "ext": "ttf",
      "size": 1234567,
      "uploaded_at": "2026-04-19T10:30:00Z"
    }
  ]
}
```

## 錯誤回應

| Status | 條件 | Body |
|--------|------|------|
| 500 | 掃描字體目錄失敗 | `{"error":"..."}` |

## 行為契約

- MUST 同時列出系統與使用者字體
- MUST 以 `source` 欄位區分來源
- MUST 在無任何使用者字體時仍正確回傳系統字體
- MUST 將全站預設 family name 放在 top-level `default_family`；未設定時為空字串
- 排序：系統字體在前、使用者字體在後；同組內依 family name 升序
