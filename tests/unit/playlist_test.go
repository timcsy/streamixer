package unit

import (
	"strings"
	"testing"

	"github.com/timcsy/streamixer/src/composer"
)

func TestGeneratePlaylist_BasicDuration(t *testing.T) {
	// 18 秒音檔，6 秒分段 → 3 個分段
	playlist := composer.GeneratePlaylist(18.0, 6)

	if !strings.Contains(playlist, "#EXTM3U") {
		t.Error("playlist MUST 包含 #EXTM3U 標頭")
	}
	if !strings.Contains(playlist, "#EXT-X-TARGETDURATION:6") {
		t.Error("playlist MUST 包含正確的 target duration")
	}
	if !strings.Contains(playlist, "#EXT-X-ENDLIST") {
		t.Error("playlist MUST 包含 #EXT-X-ENDLIST")
	}
	// 應有 3 個分段
	count := strings.Count(playlist, ".ts")
	if count != 3 {
		t.Errorf("預期 3 個分段，實際 %d 個", count)
	}
}

func TestGeneratePlaylist_NonEvenDuration(t *testing.T) {
	// 16 秒音檔，6 秒分段 → 3 個分段（6 + 6 + 4）
	playlist := composer.GeneratePlaylist(16.0, 6)

	count := strings.Count(playlist, ".ts")
	if count != 3 {
		t.Errorf("預期 3 個分段，實際 %d 個", count)
	}
	// 最後一個分段應為 4 秒
	if !strings.Contains(playlist, "#EXTINF:4.000000,") {
		t.Error("最後一個分段時長應為 4 秒")
	}
}

func TestGeneratePlaylist_ShortDuration(t *testing.T) {
	// 3 秒音檔，6 秒分段 → 1 個分段
	playlist := composer.GeneratePlaylist(3.0, 6)

	count := strings.Count(playlist, ".ts")
	if count != 1 {
		t.Errorf("預期 1 個分段，實際 %d 個", count)
	}
}
