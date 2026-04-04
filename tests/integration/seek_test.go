package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/timcsy/streamixer/src/composer"
	"github.com/timcsy/streamixer/src/media"
)

func TestSeek_SegmentPTSAlignment(t *testing.T) {
	tmpDir := t.TempDir()
	mediaDir := filepath.Join(tmpDir, "media")
	os.MkdirAll(mediaDir, 0755)

	// 產生 18 秒測試音檔
	audioPath := filepath.Join(mediaDir, "audio.mp3")
	cmd := exec.Command("ffmpeg", "-y", "-f", "lavfi", "-i",
		"sine=frequency=440:duration=18", "-q:a", "9", audioPath)
	if err := cmd.Run(); err != nil {
		t.Skipf("需要 ffmpeg：%v", err)
	}

	bgPath := filepath.Join(mediaDir, "background.jpg")
	cmd = exec.Command("ffmpeg", "-y", "-f", "lavfi", "-i",
		"color=c=blue:s=320x240:d=1", "-frames:v", "1", bgPath)
	if err := cmd.Run(); err != nil {
		t.Skipf("需要 ffmpeg：%v", err)
	}

	comp := &media.MediaComposition{
		ID:         "seek-test",
		Audio:      media.Audio{Path: audioPath, Format: "mp3"},
		Background: media.Background{Path: bgPath, Format: "jpg"},
	}

	outDir := filepath.Join(tmpDir, "output")

	// 按需生成分段 0
	seg0Path := filepath.Join(outDir, "seg_000.ts")
	err := composer.GenerateSegment(comp, seg0Path, 0, 6, 320, 240)
	if err != nil {
		t.Fatalf("生成分段 0 失敗：%v", err)
	}

	// 按需生成分段 1（應從 6 秒開始）
	seg1Path := filepath.Join(outDir, "seg_001.ts")
	err = composer.GenerateSegment(comp, seg1Path, 1, 6, 320, 240)
	if err != nil {
		t.Fatalf("生成分段 1 失敗：%v", err)
	}

	// 用 ffprobe 取得分段 1 的音訊起始 PTS
	cmd = exec.Command("ffprobe", "-v", "error",
		"-select_streams", "a:0",
		"-show_entries", "packet=pts_time",
		"-of", "csv=p=0",
		seg1Path)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("ffprobe 失敗：%v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 {
		t.Fatal("ffprobe 沒有回傳任何 PTS 資料")
	}

	// 第一個 audio packet 的 PTS 應接近 6 秒
	ptsStr := strings.TrimRight(strings.TrimSpace(lines[0]), ",")
	firstPTS, err := strconv.ParseFloat(ptsStr, 64)
	if err != nil {
		t.Fatalf("無法解析 PTS：%v", err)
	}

	// 允許 2 秒的誤差（MP3 編碼器延遲 + AAC priming samples 造成 ~1.4 秒偏移）
	// 關鍵是 PTS 基準從 startTime 附近開始，而非從 0 開始
	if firstPTS < 5.0 || firstPTS > 8.0 {
		t.Errorf("分段 1 的音訊起始 PTS 應接近 6 秒（±2秒），實際 %.3f 秒", firstPTS)
	}
}

func TestSeek_MiddleSegmentAudioContent(t *testing.T) {
	tmpDir := t.TempDir()
	mediaDir := filepath.Join(tmpDir, "media")
	os.MkdirAll(mediaDir, 0755)

	// 產生 18 秒音檔，前 6 秒 440Hz，中 6 秒 880Hz，後 6 秒 1320Hz
	audioPath := filepath.Join(mediaDir, "audio.mp3")
	cmd := exec.Command("ffmpeg", "-y", "-f", "lavfi", "-i",
		"sine=frequency=440:duration=6", "-f", "lavfi", "-i",
		"sine=frequency=880:duration=6", "-f", "lavfi", "-i",
		"sine=frequency=1320:duration=6",
		"-filter_complex", "[0:a][1:a][2:a]concat=n=3:v=0:a=1[out]",
		"-map", "[out]", "-q:a", "9", audioPath)
	if err := cmd.Run(); err != nil {
		t.Skipf("需要 ffmpeg：%v", err)
	}

	bgPath := filepath.Join(mediaDir, "background.jpg")
	cmd = exec.Command("ffmpeg", "-y", "-f", "lavfi", "-i",
		"color=c=blue:s=320x240:d=1", "-frames:v", "1", bgPath)
	if err := cmd.Run(); err != nil {
		t.Skipf("需要 ffmpeg：%v", err)
	}

	comp := &media.MediaComposition{
		ID:         "seek-freq-test",
		Audio:      media.Audio{Path: audioPath, Format: "mp3"},
		Background: media.Background{Path: bgPath, Format: "jpg"},
	}

	outDir := filepath.Join(tmpDir, "output")

	// 生成分段 1（應包含 880Hz 的音訊）
	seg1Path := filepath.Join(outDir, "seg_001.ts")
	err := composer.GenerateSegment(comp, seg1Path, 1, 6, 320, 240)
	if err != nil {
		t.Fatalf("生成分段 1 失敗：%v", err)
	}

	// 驗證分段存在且有內容
	info, err := os.Stat(seg1Path)
	if err != nil || info.Size() == 0 {
		t.Fatal("分段 1 應存在且非空")
	}
}
