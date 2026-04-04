package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/timcsy/streamixer/src/config"
)

// SampleHandler 產生範例素材
type SampleHandler struct {
	cfg config.Config
}

// NewSampleHandler 建立範例 handler
func NewSampleHandler(cfg config.Config) *SampleHandler {
	return &SampleHandler{cfg: cfg}
}

// GenerateSample 產生範例素材（10 秒正弦波音檔 + 藍色背景 + 範例字幕）
func (h *SampleHandler) GenerateSample(w http.ResponseWriter, r *http.Request) {
	id := "sample"
	dir := filepath.Join(h.cfg.MediaDir, id)
	os.MkdirAll(dir, 0755)

	audioPath := filepath.Join(dir, "audio.mp3")
	bgPath := filepath.Join(dir, "background.jpg")
	subPath := filepath.Join(dir, "subtitle.srt")

	// 產生 10 秒測試音檔
	cmd := exec.Command("ffmpeg", "-y", "-f", "lavfi", "-i",
		"sine=frequency=440:duration=10", "-q:a", "2", audioPath)
	if err := cmd.Run(); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("產生音檔失敗：%v", err))
		return
	}

	// 產生藍色背景圖片 1280x720
	cmd = exec.Command("ffmpeg", "-y", "-f", "lavfi", "-i",
		"color=c=#1a237e:s=1280x720:d=1",
		"-frames:v", "1", bgPath)
	if err := cmd.Run(); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("產生背景圖片失敗：%v", err))
		return
	}

	// 產生範例字幕
	srtContent := `1
00:00:00,500 --> 00:00:03,000
Streamixer 範例字幕

2
00:00:03,500 --> 00:00:06,000
即時合成串流測試中...

3
00:00:06,500 --> 00:00:09,500
音檔 + 背景 + 字幕 = 影片串流
`
	if err := os.WriteFile(subPath, []byte(srtContent), 0644); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("產生字幕失敗：%v", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":     id,
		"files":  []string{"audio.mp3", "background.jpg", "subtitle.srt"},
		"stream": fmt.Sprintf("/stream/%s/index.m3u8", id),
	})
}
