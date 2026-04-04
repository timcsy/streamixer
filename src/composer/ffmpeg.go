package composer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/timcsy/streamixer/src/media"
)

// BuildFFmpegArgs 組裝完整 HLS 合成的 FFmpeg 參數
func BuildFFmpegArgs(comp *media.MediaComposition, outDir string, segDuration, width, height int) []string {
	args := []string{
		"-y",
		"-loop", "1",
		"-i", comp.Background.Path,
		"-i", comp.Audio.Path,
	}

	vf := fmt.Sprintf("scale=%d:%d", width, height)
	if comp.Subtitle != nil {
		vf = fmt.Sprintf("subtitles=%s,scale=%d:%d", comp.Subtitle.Path, width, height)
	}

	args = append(args,
		"-vf", vf,
		"-c:v", "libx264",
		"-tune", "stillimage",
		"-preset", "ultrafast",
		"-pix_fmt", "yuv420p",
		"-c:a", "aac",
		"-b:a", "128k",
		"-shortest",
		"-f", "hls",
		"-hls_time", fmt.Sprintf("%d", segDuration),
		"-hls_list_size", "0",
		"-hls_segment_filename", filepath.Join(outDir, "seg_%03d.ts"),
		filepath.Join(outDir, "index.m3u8"),
	)

	return args
}

// BuildSegmentArgs 組裝單一分段合成的 FFmpeg 參數（用於 seek/按需生成）
func BuildSegmentArgs(comp *media.MediaComposition, outPath string, segIndex, segDuration, width, height int) []string {
	startTime := segIndex * segDuration

	args := []string{
		"-y",
		"-ss", fmt.Sprintf("%d", startTime),
		"-loop", "1",
		"-i", comp.Background.Path,
		"-ss", fmt.Sprintf("%d", startTime),
		"-i", comp.Audio.Path,
		"-t", fmt.Sprintf("%d", segDuration),
	}

	vf := fmt.Sprintf("scale=%d:%d", width, height)
	if comp.Subtitle != nil {
		vf = fmt.Sprintf("subtitles=%s,scale=%d:%d", comp.Subtitle.Path, width, height)
	}

	args = append(args,
		"-vf", vf,
		"-c:v", "libx264",
		"-tune", "stillimage",
		"-preset", "ultrafast",
		"-pix_fmt", "yuv420p",
		"-c:a", "aac",
		"-b:a", "128k",
		"-shortest",
		"-f", "mpegts",
		outPath,
	)

	return args
}

// GenerateSegment 合成單一 .ts 分段
func GenerateSegment(comp *media.MediaComposition, outPath string, segIndex, segDuration, width, height int) error {
	os.MkdirAll(filepath.Dir(outPath), 0755)

	args := BuildSegmentArgs(comp, outPath, segIndex, segDuration, width, height)
	cmd := exec.Command("ffmpeg", args...)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("FFmpeg 合成分段 %d 失敗：%w", segIndex, err)
	}

	return nil
}
