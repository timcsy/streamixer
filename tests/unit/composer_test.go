package unit

import (
	"strings"
	"testing"

	"github.com/timcsy/streamixer/src/composer"
	"github.com/timcsy/streamixer/src/media"
)

func TestBuildFFmpegArgs_AudioAndImage(t *testing.T) {
	comp := &media.MediaComposition{
		ID:         "test-001",
		Audio:      media.Audio{Path: "/media/test/audio.mp3"},
		Background: media.Background{Path: "/media/test/background.jpg"},
	}

	args := composer.BuildFFmpegArgs(comp, "/tmp/out", 6, 1280, 720)
	argsStr := strings.Join(args, " ")

	if !strings.Contains(argsStr, "-loop 1") {
		t.Error("MUST 包含 -loop 1")
	}
	if !strings.Contains(argsStr, "-f hls") {
		t.Error("MUST 輸出 HLS 格式")
	}
	if !strings.Contains(argsStr, "-hls_segment_type fmp4") {
		t.Error("MUST 包含 -hls_segment_type fmp4")
	}
	if !strings.Contains(argsStr, "-hls_fmp4_init_filename init.mp4") {
		t.Error("MUST 包含 -hls_fmp4_init_filename init.mp4")
	}
	if !strings.Contains(argsStr, ".m4s") {
		t.Error("分段副檔名 MUST 為 .m4s")
	}
	if !strings.Contains(argsStr, "-shortest") {
		t.Error("MUST 包含 -shortest")
	}
	if !strings.Contains(argsStr, "scale=1280:720") {
		t.Error("MUST 包含解析度縮放")
	}
}

func TestBuildFFmpegArgs_NoSubtitle(t *testing.T) {
	comp := &media.MediaComposition{
		ID:         "test-001",
		Audio:      media.Audio{Path: "/media/test/audio.mp3"},
		Background: media.Background{Path: "/media/test/background.jpg"},
		Subtitle:   nil,
	}

	args := composer.BuildFFmpegArgs(comp, "/tmp/out", 6, 1280, 720)
	argsStr := strings.Join(args, " ")

	if strings.Contains(argsStr, "subtitles=") {
		t.Error("無字幕時不應包含 subtitles 濾鏡")
	}
}

func TestBuildSegmentArgs_Fmp4Format(t *testing.T) {
	comp := &media.MediaComposition{
		ID:         "test-001",
		Audio:      media.Audio{Path: "/media/test/audio.mp3"},
		Background: media.Background{Path: "/media/test/background.jpg"},
	}

	args := composer.BuildSegmentArgs(comp, "/tmp/out/seg_002.m4s", 2, 6, 1280, 720)
	argsStr := strings.Join(args, " ")

	if !strings.Contains(argsStr, "-ss 12") {
		t.Error("MUST 以 -ss 跳至正確位置（分段 2 = 12 秒）")
	}
	if !strings.Contains(argsStr, "-t 6") {
		t.Error("MUST 以 -t 限制分段長度")
	}
	if !strings.Contains(argsStr, "-f mp4") {
		t.Error("MUST 輸出 mp4 格式")
	}
	if !strings.Contains(argsStr, "empty_moov") {
		t.Error("MUST 包含 empty_moov movflag")
	}
	if !strings.Contains(argsStr, "cmaf") {
		t.Error("MUST 包含 cmaf movflag")
	}
}

func TestBuildSegmentArgs_SeekPosition(t *testing.T) {
	comp := &media.MediaComposition{
		ID:         "test-001",
		Audio:      media.Audio{Path: "/media/test/audio.mp3"},
		Background: media.Background{Path: "/media/test/background.jpg"},
	}

	args := composer.BuildSegmentArgs(comp, "/tmp/out/seg_002.m4s", 2, 6, 1280, 720)
	argsStr := strings.Join(args, " ")

	bgIdx := strings.Index(argsStr, comp.Background.Path)
	ssIdx := strings.Index(argsStr, "-ss 12")
	audioIdx := strings.Index(argsStr, comp.Audio.Path)
	if ssIdx < bgIdx {
		t.Error("-ss 不應在背景圖片 -i 之前")
	}
	if ssIdx > audioIdx {
		t.Error("-ss 應在音檔 -i 之前（input seek）")
	}
}
