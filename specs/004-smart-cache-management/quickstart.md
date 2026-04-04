# 快速驗證：智慧快取管理

## 驗證 TTL 清除

```bash
# 設定 TTL 為 1 分鐘以方便測試
docker run -d --name streamixer-cache-test \
  -p 8081:8080 \
  -e CACHE_TTL=1m \
  -e CACHE_SWEEP_INTERVAL=30s \
  --tmpfs /dev/shm/streamixer:size=512M \
  streamixer

# 產生範例並播放
curl -X POST http://localhost:8081/sample

# 請求 playlist 觸發預生成
curl -s http://localhost:8081/stream/sample/index.m3u8 > /dev/null

# 等待 2 分鐘（TTL 過期 + 清掃執行）
sleep 120

# 檢查容器日誌確認清掃執行
docker logs streamixer-cache-test 2>&1 | grep -i "清除\|sweep\|cache"
```

## 驗證容量保護

```bash
# 設定較小的容量上限
docker run -d --name streamixer-cap-test \
  -p 8082:8080 \
  -e CACHE_MAX_SIZE=10000000 \
  -e CACHE_SWEEP_INTERVAL=10s \
  --tmpfs /dev/shm/streamixer:size=512M \
  streamixer

# 上傳多個不同素材，觀察舊的是否被淘汰
```

## 驗證重新觸發

```bash
# 在 TTL 過期後重新請求，確認能正常播放
curl -s http://localhost:8081/stream/sample/index.m3u8
# 應能正常回傳 playlist，背景重新預生成
```
