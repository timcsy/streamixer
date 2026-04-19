# Quickstart: 字幕字體管理

## 管理員：上傳全站字體

1. 後台 → 設定 → Streamixer → 拉到「字體管理」區塊
2. 按「選擇檔案」挑選 `Church.ttf`（≤ 10 MB）
3. 按「上傳字體」→ 數秒後列表出現新字體，family name 自動解析
4. 「全站預設字體」下拉選擇剛上傳的字體 → 儲存
5. 重新合成任一素材（可透過清除快取觸發），字幕即使用新字體

## 管理員：為單一素材指定不同字體

1. 編輯素材組合 → 新增的「字體」下拉
2. 選擇非「使用全站預設」的其他字體 → 更新
3. 清除該素材的快取（若有），重新播放觀察字幕變化

## 管理員：刪除字體

1. 字體管理區塊的列表，使用者上傳字體有「刪除」按鈕
2. 按刪除 → WP 先查有哪些素材用了此字體，列在 confirm dialog
3. 確認後字體從清單消失；指定該字體的素材下次合成自動 fallback 到全站預設

## 開發者：本機驗證端點

```bash
# 列出字體
curl http://localhost:8080/fonts | jq

# 上傳（需 API Key）
curl -X POST http://localhost:8080/fonts \
  -H "X-API-Key: $API_KEY" \
  -F "font=@/path/to/font.ttf" | jq

# 設為預設
curl -X PUT http://localhost:8080/fonts/default \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"family_name":"Church Font"}'

# 刪除
curl -X DELETE http://localhost:8080/fonts/<id> \
  -H "X-API-Key: $API_KEY"
```

## 開發者：整合測試

```bash
go test ./tests/integration/... -run TestFonts -v
```

## 失敗排查

| 現象 | 可能原因 | 建議 |
|------|----------|------|
| 上傳 400「檔頭驗證失敗」 | 檔案非真正字體 | 確認是 ttf/otf/ttc，非 .txt 改名 |
| 合成後字幕仍是舊字體 | 既有分段在快取中 | 縮短 cache TTL 或手動清除 `/dev/shm/streamixer/<id>` |
| 字體列表沒有新上傳的字體 | fontconfig cache 未更新 | 檢查 symlink `/usr/share/fonts/user/<id>.ttf` 是否存在；手動執行 `fc-cache -f` |
| 指定字體合成失敗 | family name 含特殊字元 | 字元集限制 `[A-Za-z0-9 \u4e00-\u9fff-_]` |
