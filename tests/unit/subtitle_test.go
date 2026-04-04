package unit

import (
	"strings"
	"testing"

	"github.com/timcsy/streamixer/src/composer"
	"github.com/timcsy/streamixer/src/media"
)

func TestBuildFFmpegArgs_WithSubtitle(t *testing.T) {
	comp := &media.MediaComposition{
		ID:         "test-001",
		Audio:      media.Audio{Path: "/media/test/audio.mp3"},
		Background: media.Background{Path: "/media/test/background.jpg"},
		Subtitle:   &media.Subtitle{Path: "/media/test/subtitle.srt", Format: "srt"},
	}

	args := composer.BuildFFmpegArgs(comp, "/tmp/out", 6, 1280, 720)
	argsStr := strings.Join(args, " ")

	if !strings.Contains(argsStr, "subtitles=/media/test/subtitle.srt") {
		t.Error("含字幕時 MUST 包含 subtitles 濾鏡")
	}
}

func TestBuildSegmentArgs_WithSubtitle(t *testing.T) {
	comp := &media.MediaComposition{
		ID:         "test-001",
		Audio:      media.Audio{Path: "/media/test/audio.mp3"},
		Background: media.Background{Path: "/media/test/background.jpg"},
		Subtitle:   &media.Subtitle{Path: "/media/test/subtitle.srt", Format: "srt"},
	}

	args := composer.BuildSegmentArgs(comp, "/tmp/out/seg_000.m4s", 0, 6, 1280, 720)
	argsStr := strings.Join(args, " ")

	if !strings.Contains(argsStr, "subtitles=/media/test/subtitle.srt") {
		t.Error("含字幕時分段合成 MUST 包含 subtitles 濾鏡")
	}
}
