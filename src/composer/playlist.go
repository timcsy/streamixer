package composer

import (
	"fmt"
	"math"
	"strings"
)

// GeneratePlaylist 根據音檔總長度與分段秒數產生 HLS playlist 內容
func GeneratePlaylist(totalDuration float64, segmentDuration int) string {
	var b strings.Builder

	b.WriteString("#EXTM3U\n")
	b.WriteString("#EXT-X-VERSION:3\n")
	b.WriteString(fmt.Sprintf("#EXT-X-TARGETDURATION:%d\n", segmentDuration))
	b.WriteString("#EXT-X-MEDIA-SEQUENCE:0\n")

	segDur := float64(segmentDuration)
	numSegments := int(math.Ceil(totalDuration / segDur))

	for i := 0; i < numSegments; i++ {
		remaining := totalDuration - float64(i)*segDur
		dur := segDur
		if remaining < segDur {
			dur = remaining
		}
		b.WriteString(fmt.Sprintf("#EXTINF:%f,\n", dur))
		b.WriteString(fmt.Sprintf("seg_%03d.ts\n", i))
	}

	b.WriteString("#EXT-X-ENDLIST\n")
	return b.String()
}

// SegmentCount 計算分段數
func SegmentCount(totalDuration float64, segmentDuration int) int {
	return int(math.Ceil(totalDuration / float64(segmentDuration)))
}
