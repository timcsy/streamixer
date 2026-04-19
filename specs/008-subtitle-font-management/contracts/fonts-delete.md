# Contract: DELETE /fonts/{id} + PUT /fonts/default

## DELETE /fonts/{id}

### 目的

刪除一個使用者上傳字體。

### 請求

- Method：`DELETE`
- Path：`/fonts/{id}`
- 認證：API Key

### 成功回應

- Status：`204 No Content`
- Body：空

### 錯誤回應

| Status | 條件 |
|--------|------|
| 404 | id 不存在 |
| 403 | 嘗試刪除系統字體 |
| 500 | 檔案系統錯誤 |

### 行為契約

- MUST 移除實體檔、metadata JSON、symlink
- MUST NOT 檢查「字體是否被素材引用」——此責任在 WP 端，後端只負責儲存
- 刪除後若 `default.txt` 指向該 family name，系統 MUST 清空 default（寫入空字串）
- 既有快取分段不受影響；後續合成若仍指定該 family name，由 libass fallback

---

## PUT /fonts/default

### 目的

設定全站預設字體的 family name。

### 請求

- Method：`PUT`
- Path：`/fonts/default`
- 認證：API Key
- Content-Type：`application/json`
- Body：`{"family_name": "Noto Sans CJK TC"}` 或 `{"family_name": ""}`（清除預設）

### 成功回應

- Status：`200 OK`
- Body：`{"default_family": "..."}`

### 錯誤回應

| Status | 條件 |
|--------|------|
| 400 | family_name 非空但不在目前字體列表中 |
| 500 | 寫入 default.txt 失敗 |

### 行為契約

- 空字串為合法輸入（代表清除預設，回到 libass 原生 fallback）
- MUST 在寫入 `/fonts/default.txt` 前檢查 family name 確實存在於 `List()` 結果中
