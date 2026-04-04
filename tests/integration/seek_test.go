package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/timcsy/streamixer/src/composer"
	"github.com/timcsy/streamixer/src/media"
)

func TestSeek_SegmentGeneration(t *testing.T) {
	tmpDir := t.TempDir()
	mediaDir := filepath.Join(tmpDir, "media")
	os.MkdirAll(mediaDir, 0755)

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
	seg0Path := filepath.Join(outDir, "seg_000.m4s")
	err := composer.GenerateSegment(comp, seg0Path, 0, 6, 320, 240)
	if err != nil {
		t.Fatalf("生成分段 0 失敗：%v", err)
	}

	info, err := os.Stat(seg0Path)
	if err != nil || info.Size() == 0 {
		t.Fatal("分段 0 應存在且非空")
	}

	// 按需生成分段 1（應從 6 秒開始）
	seg1Path := filepath.Join(outDir, "seg_001.m4s")
	err = composer.GenerateSegment(comp, seg1Path, 1, 6, 320, 240)
	if err != nil {
		t.Fatalf("生成分段 1 失敗：%v", err)
	}

	info, err = os.Stat(seg1Path)
	if err != nil || info.Size() == 0 {
		t.Fatal("分段 1 應存在且非空")
	}
}

func TestSeek_MiddleSegmentAudioContent(t *testing.T) {
	tmpDir := t.TempDir()
	mediaDir := filepath.Join(tmpDir, "media")
	os.MkdirAll(mediaDir, 0755)

	// 產生 18 秒音檔
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

	seg1Path := filepath.Join(outDir, "seg_001.m4s")
	err := composer.GenerateSegment(comp, seg1Path, 1, 6, 320, 240)
	if err != nil {
		t.Fatalf("生成分段 1 失敗：%v", err)
	}

	info, err := os.Stat(seg1Path)
	if err != nil || info.Size() == 0 {
		t.Fatal("分段 1 應存在且非空")
	}
}
