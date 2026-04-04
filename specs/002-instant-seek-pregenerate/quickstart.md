# 快速驗證：即時跳轉與背景預生成

## 驗證順序播放

```bash
# 產生 60 秒測試素材
curl -X POST http://localhost:8081/sample

# 播放並觀察分段切換
# 在瀏覽器開啟 http://localhost:8081，播放 sample，確認無中斷
```

## 驗證跳轉對齊

```bash
# 取得 playlist
curl -s http://localhost:8081/stream/sample/index.m3u8

# 取得第 0 個分段並用 ffprobe 驗證時間戳
curl -s http://localhost:8081/stream/sample/seg_000.ts -o /tmp/seg_000.ts
ffprobe -v error -show_entries packet=pts_time -of csv=p=0 /tmp/seg_000.ts | head -5

# 取得第 1 個分段並驗證時間戳從 6 秒開始
curl -s http://localhost:8081/stream/sample/seg_001.ts -o /tmp/seg_001.ts
ffprobe -v error -show_entries packet=pts_time -of csv=p=0 /tmp/seg_001.ts | head -5
```

## 驗證預生成

```bash
# 請求 playlist 後等待幾秒，背景預生成應完成
curl -s http://localhost:8081/stream/sample/index.m3u8 > /dev/null

# 等待預生成完成
sleep 5

# 此時所有分段應已存在，回應應接近瞬間
time curl -s http://localhost:8081/stream/sample/seg_001.ts -o /dev/null
```
