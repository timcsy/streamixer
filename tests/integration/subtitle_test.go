package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/timcsy/streamixer/src/composer"
	"github.com/timcsy/streamixer/src/media"
)

func setupTestMediaWithSubtitle(t *testing.T) (string, *media.MediaComposition) {
	t.Helper()

	tmpDir := t.TempDir()
	mediaDir := filepath.Join(tmpDir, "media", "test-sub")
	os.MkdirAll(mediaDir, 0755)

	// 產生 10 秒測試音檔
	audioPath := filepath.Join(mediaDir, "audio.mp3")
	cmd := exec.Command("ffmpeg", "-y", "-f", "lavfi", "-i",
		"sine=frequency=440:duration=10", "-q:a", "9", audioPath)
	if err := cmd.Run(); err != nil {
		t.Skipf("需要 ffmpeg：%v", err)
	}

	// 產生測試圖片
	bgPath := filepath.Join(mediaDir, "background.jpg")
	cmd = exec.Command("ffmpeg", "-y", "-f", "lavfi", "-i",
		"color=c=blue:s=320x240:d=1", "-frames:v", "1", bgPath)
	if err := cmd.Run(); err != nil {
		t.Skipf("需要 ffmpeg：%v", err)
	}

	// 建立 SRT 字幕檔
	srtPath := filepath.Join(mediaDir, "subtitle.srt")
	srtContent := `1
00:00:01,000 --> 00:00:05,000
測試字幕第一行

2
00:00:05,000 --> 00:00:09,000
測試字幕第二行
`
	os.WriteFile(srtPath, []byte(srtContent), 0644)

	comp := &media.MediaComposition{
		ID:         "test-sub",
		Audio:      media.Audio{Path: audioPath, Format: "mp3"},
		Background: media.Background{Path: bgPath, Format: "jpg"},
		Subtitle:   &media.Subtitle{Path: srtPath, Format: "srt"},
	}

	return tmpDir, comp
}

func TestSubtitleComposition(t *testing.T) {
	tmpDir, comp := setupTestMediaWithSubtitle(t)

	outDir := filepath.Join(tmpDir, "output")
	segPath := filepath.Join(outDir, "seg_000.m4s")
	os.MkdirAll(outDir, 0755)

	err := composer.GenerateSegment(comp, segPath, 0, 6, 320, 240)
	if err != nil {
		t.Fatalf("含字幕合成失敗：%v", err)
	}

	// 驗證分段存在且有內容
	info, err := os.Stat(segPath)
	if err != nil {
		t.Fatalf("分段檔案不存在：%v", err)
	}
	if info.Size() == 0 {
		t.Error("分段檔案不應為空")
	}

	// 用 ffprobe 驗證輸出含有影片串流
	cmd := exec.Command("ffprobe", "-v", "error",
		"-show_entries", "stream=codec_type",
		"-of", "csv=p=0", segPath)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("ffprobe 驗證失敗：%v", err)
	}
	outStr := string(out)
	if !strings.Contains(outStr, "video") {
		t.Error("輸出分段 MUST 含有影片串流")
	}
	if !strings.Contains(outStr, "audio") {
		t.Error("輸出分段 MUST 含有音訊串流")
	}
}

func TestNoSubtitle_StillWorks(t *testing.T) {
	tmpDir, comp := setupTestMediaWithSubtitle(t)
	comp.Subtitle = nil // 移除字幕

	outDir := filepath.Join(tmpDir, "output")
	segPath := filepath.Join(outDir, "seg_000.m4s")
	os.MkdirAll(outDir, 0755)

	err := composer.GenerateSegment(comp, segPath, 0, 6, 320, 240)
	if err != nil {
		t.Fatalf("無字幕合成應正常運作，但失敗：%v", err)
	}

	info, err := os.Stat(segPath)
	if err != nil || info.Size() == 0 {
		t.Error("無字幕合成的分段不應為空")
	}
}
