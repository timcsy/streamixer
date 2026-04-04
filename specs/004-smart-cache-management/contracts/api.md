# API 契約：智慧快取管理

## 端點變更

本功能不新增或修改任何外部 API 端點。快取管理完全在內部運作，對使用者透明。

既有端點行為不變：
- `GET /stream/{id}/index.m3u8` — 不變（內部：存取時更新快取時間）
- `GET /stream/{id}/init.mp4` — 不變（內部：存取時更新快取時間）
- `GET /stream/{id}/seg_{n}.m4s` — 不變（內部：存取時更新快取時間）
- `GET /health` — 不變

## 新增環境變數

| 變數 | 預設值 | 說明 |
|------|--------|------|
| CACHE_TTL | 30m | 快取過期時間（未被存取後多久清除） |
| CACHE_MAX_SIZE | 0 | 快取容量上限（bytes，0 表示使用 tmpfs 的 80%） |
| CACHE_SWEEP_INTERVAL | 5m | 清掃頻率 |
