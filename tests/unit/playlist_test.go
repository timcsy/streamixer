package unit

import (
	"strings"
	"testing"

	"github.com/timcsy/streamixer/src/composer"
)

func TestGeneratePlaylist_BasicDuration(t *testing.T) {
	playlist := composer.GeneratePlaylist(18.0, 6)

	if !strings.Contains(playlist, "#EXTM3U") {
		t.Error("playlist MUST 包含 #EXTM3U 標頭")
	}
	if !strings.Contains(playlist, "#EXT-X-VERSION:7") {
		t.Error("playlist MUST 包含 #EXT-X-VERSION:7")
	}
	if !strings.Contains(playlist, "#EXT-X-TARGETDURATION:6") {
		t.Error("playlist MUST 包含正確的 target duration")
	}
	if !strings.Contains(playlist, "#EXT-X-MAP:URI=\"init.mp4\"") {
		t.Error("playlist MUST 包含 #EXT-X-MAP:URI=\"init.mp4\"")
	}
	if !strings.Contains(playlist, "#EXT-X-ENDLIST") {
		t.Error("playlist MUST 包含 #EXT-X-ENDLIST")
	}
	count := strings.Count(playlist, ".m4s")
	if count != 3 {
		t.Errorf("預期 3 個 .m4s 分段，實際 %d 個", count)
	}
	if strings.Contains(playlist, ".ts") {
		t.Error("playlist 不應包含 .ts 分段")
	}
}

func TestGeneratePlaylist_NonEvenDuration(t *testing.T) {
	playlist := composer.GeneratePlaylist(16.0, 6)

	count := strings.Count(playlist, ".m4s")
	if count != 3 {
		t.Errorf("預期 3 個分段，實際 %d 個", count)
	}
	if !strings.Contains(playlist, "#EXTINF:4.000000,") {
		t.Error("最後一個分段時長應為 4 秒")
	}
}

func TestGeneratePlaylist_ShortDuration(t *testing.T) {
	playlist := composer.GeneratePlaylist(3.0, 6)

	count := strings.Count(playlist, ".m4s")
	if count != 1 {
		t.Errorf("預期 1 個分段，實際 %d 個", count)
	}
}
