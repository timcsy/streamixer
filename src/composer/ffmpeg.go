package composer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/timcsy/streamixer/src/media"
)

// BuildFFmpegArgs 組裝完整 HLS 合成的 FFmpeg 參數（fMP4 格式）
func BuildFFmpegArgs(comp *media.MediaComposition, outDir string, segDuration, width, height int) []string {
	args := []string{
		"-y",
		"-loop", "1",
		"-i", comp.Background.Path,
		"-i", comp.Audio.Path,
	}

	vf := fmt.Sprintf("scale=%d:%d", width, height)
	if comp.Subtitle != nil {
		// MarginV=30 讓字幕貼近底部但避開控制列，Fontsize=28 放大字體
		vf = fmt.Sprintf("subtitles=%s:force_style='MarginV=30,Fontsize=28',scale=%d:%d", comp.Subtitle.Path, width, height)
	}

	args = append(args,
		"-vf", vf,
		"-c:v", "libx264",
		"-tune", "stillimage",
		"-preset", "ultrafast",
		"-pix_fmt", "yuv420p",
		// 強制每 segDuration 秒一個 keyframe，確保分段切割精確
		"-force_key_frames", fmt.Sprintf("expr:gte(t,n_forced*%d)", segDuration),
		"-c:a", "aac",
		"-b:a", "128k",
		"-shortest",
		"-f", "hls",
		"-hls_segment_type", "fmp4",
		"-hls_fmp4_init_filename", "init.mp4",
		"-hls_time", fmt.Sprintf("%d", segDuration),
		"-hls_list_size", "0",
		"-hls_segment_filename", filepath.Join(outDir, "seg_%03d.m4s"),
		filepath.Join(outDir, "index.m3u8"),
	)

	return args
}

// BuildSegmentArgs 組裝單一 fMP4 分段合成的 FFmpeg 參數（用於按需生成）
// 使用 -output_ts_offset 確保輸出 PTS 與 HLS playlist 對齊
func BuildSegmentArgs(comp *media.MediaComposition, outPath string, segIndex, segDuration, width, height int) []string {
	startTime := segIndex * segDuration

	args := []string{
		"-y",
		"-loop", "1",
		"-i", comp.Background.Path,
		"-ss", fmt.Sprintf("%d", startTime),
		"-i", comp.Audio.Path,
		"-t", fmt.Sprintf("%d", segDuration),
	}

	vf := fmt.Sprintf("scale=%d:%d", width, height)
	if comp.Subtitle != nil {
		// MarginV=30 讓字幕貼近底部但避開控制列，Fontsize=28 放大字體
		vf = fmt.Sprintf("subtitles=%s:force_style='MarginV=30,Fontsize=28',scale=%d:%d", comp.Subtitle.Path, width, height)
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
		"-f", "mp4",
		"-movflags", "+frag_keyframe+empty_moov+default_base_moof+cmaf",
		outPath,
	)

	return args
}

// GenerateSegment 合成單一 fMP4 分段
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

// GenerateInitSegment 獨立產生 init.mp4（用於按需生成前 init 不存在的情況）
func GenerateInitSegment(comp *media.MediaComposition, outDir string, segDuration, width, height int) error {
	os.MkdirAll(outDir, 0755)

	initPath := filepath.Join(outDir, "init.mp4")

	// 產生一個極短的 fMP4 HLS，只為了取得 init.mp4
	args := []string{
		"-y",
		"-loop", "1",
		"-i", comp.Background.Path,
		"-i", comp.Audio.Path,
		"-t", "0.1",
		"-vf", fmt.Sprintf("scale=%d:%d", width, height),
		"-c:v", "libx264",
		"-tune", "stillimage",
		"-preset", "ultrafast",
		"-pix_fmt", "yuv420p",
		"-c:a", "aac",
		"-b:a", "128k",
		"-shortest",
		"-f", "hls",
		"-hls_segment_type", "fmp4",
		"-hls_fmp4_init_filename", "init.mp4",
		"-hls_time", fmt.Sprintf("%d", segDuration),
		"-hls_list_size", "0",
		"-hls_segment_filename", filepath.Join(outDir, "tmp_init_%03d.m4s"),
		filepath.Join(outDir, "tmp_init.m3u8"),
	}

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("產生 init segment 失敗：%w", err)
	}

	// 清理臨時檔案，保留 init.mp4
	os.Remove(filepath.Join(outDir, "tmp_init.m3u8"))
	matches, _ := filepath.Glob(filepath.Join(outDir, "tmp_init_*.m4s"))
	for _, f := range matches {
		os.Remove(f)
	}

	if _, err := os.Stat(initPath); err != nil {
		return fmt.Errorf("init.mp4 未產生")
	}

	return nil
}
