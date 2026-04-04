# 快速驗證：fMP4 分段格式

## 驗證 playlist 格式

```bash
curl -s http://localhost:8081/stream/sample/index.m3u8
# 應包含 #EXT-X-VERSION:7 和 #EXT-X-MAP:URI="init.mp4"
# 分段應為 .m4s 而非 .ts
```

## 驗證 init segment

```bash
curl -s http://localhost:8081/stream/sample/init.mp4 -o /tmp/init.mp4
ffprobe -v error -show_format /tmp/init.mp4
# 應為有效的 MP4 檔案，包含 moov atom
```

## 驗證分段格式

```bash
curl -s http://localhost:8081/stream/sample/seg_000.m4s -o /tmp/seg_000.m4s
ffprobe -v error -show_format /tmp/seg_000.m4s
# 應為有效的 fMP4 分段
```

## 驗證播放（無破音）

在瀏覽器開啟 http://localhost:8081，播放 sample：
1. 仔細聆聽分段切換點，確認無破音
2. 跳轉到不同位置，確認音訊平順
3. 在 Chrome 和 Safari 上分別測試
