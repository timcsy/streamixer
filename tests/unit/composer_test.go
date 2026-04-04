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
		t.Error("MUST 包含 -loop 1 讓圖片循環")
	}
	if !strings.Contains(argsStr, "-i /media/test/background.jpg") {
		t.Error("MUST 包含背景圖片輸入")
	}
	if !strings.Contains(argsStr, "-i /media/test/audio.mp3") {
		t.Error("MUST 包含音檔輸入")
	}
	if !strings.Contains(argsStr, "-f hls") {
		t.Error("MUST 輸出 HLS 格式")
	}
	if !strings.Contains(argsStr, "-shortest") {
		t.Error("MUST 包含 -shortest 讓影片長度與音檔一致")
	}
	if !strings.Contains(argsStr, "scale=1280:720") {
		t.Error("MUST 包含輸出解析度縮放")
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

func TestBuildSegmentArgs_SeekToPosition(t *testing.T) {
	comp := &media.MediaComposition{
		ID:         "test-001",
		Audio:      media.Audio{Path: "/media/test/audio.mp3"},
		Background: media.Background{Path: "/media/test/background.jpg"},
	}

	args := composer.BuildSegmentArgs(comp, "/tmp/out/seg_002.ts", 2, 6, 1280, 720)
	argsStr := strings.Join(args, " ")

	// 分段 2 應從第 12 秒開始
	if !strings.Contains(argsStr, "-ss 12") {
		t.Error("MUST 以 -ss 跳至正確的音檔起始位置（分段 2 = 12 秒）")
	}
	if !strings.Contains(argsStr, "-t 6") {
		t.Error("MUST 以 -t 限制分段長度")
	}
	if !strings.Contains(argsStr, "-output_ts_offset 12") {
		t.Error("MUST 以 -output_ts_offset 對齊輸出時間戳")
	}
	// -ss 應在音檔 -i 之前（input seek），不應在圖片 -i 之前
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
