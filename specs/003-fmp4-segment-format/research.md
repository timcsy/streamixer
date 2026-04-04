# 研究筆記：fMP4 分段格式

## 決策 1：預生成的 FFmpeg fMP4 參數

**選擇**：`-hls_segment_type fmp4 -hls_fmp4_init_filename init.mp4 -hls_segment_filename seg_%03d.m4s`

**理由**：
- FFmpeg 內建支援 fMP4 HLS 輸出，只需加一個旗標
- 自動產生 init.mp4（包含 moov atom）和 .m4s 分段（包含 moof + mdat）
- init segment 在所有分段之間共享解碼器配置，消除 .ts 的重新初始化問題
- 需搭配 `-hls_playlist_type vod` 和 `#EXT-X-VERSION:7`

**完整指令**：
```
ffmpeg -loop 1 -i bg.jpg -i audio.mp3 \
  -vf "scale=W:H" -c:v libx264 -tune stillimage -preset ultrafast \
  -pix_fmt yuv420p -c:a aac -b:a 128k -shortest \
  -f hls -hls_segment_type fmp4 \
  -hls_fmp4_init_filename init.mp4 \
  -hls_segment_filename seg_%03d.m4s \
  -hls_time 6 -hls_list_size 0 \
  outdir/index.m3u8
```

**考慮過的替代方案**：
- DASH 格式：相容性不如 HLS，Safari 不原生支援
- .ts + 音訊 padding：治標不治本

## 決策 2：按需生成的 fMP4 分段

**選擇**：`-f mp4 -movflags +frag_keyframe+empty_moov+default_base_moof+cmaf`

**理由**：
- 產生獨立的 fMP4 分段，與預生成的 init segment 相容
- `empty_moov` 讓分段不包含完整的 moov（靠 init segment 提供）
- `default_base_moof` 是 CMAF/HLS 相容性所需
- `cmaf` 旗標確保嚴格的 CMAF 合規

**關鍵限制**：init segment 跨 FFmpeg 執行不保證 byte-identical，但在相同編碼參數下功能相容。策略：預生成時產生 init segment，按需生成的分段使用同一個 init segment。

**考慮過的替代方案**：
- 每次按需生成都重新產生 init segment：浪費且可能造成不一致
- 使用 `-copyts` 保留原始時間戳：與 fMP4 muxer 行為不同

## 決策 3：Init segment 管理策略

**選擇**：預生成時產生一次，按需生成前確保存在

**理由**：
- Init segment 非常小（< 1KB），產生成本極低
- 按需生成分段前，先檢查 init segment 是否存在
- 若不存在（預生成尚未開始），獨立產生 init segment（用極短的 FFmpeg 執行）
- 或直接觸發預生成（會自動產生 init segment）

## 決策 4：Content-Type

**選擇**：
- init.mp4：`video/mp4`
- .m4s 分段：`video/mp4`

**理由**：雖然 RFC 規定 .m4s 的 MIME 為 `video/iso.segment`，但實務上所有瀏覽器和 CDN 都使用 `video/mp4`，hls.js 和 Safari 均接受。
