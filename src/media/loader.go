package media

import (
	"fmt"
	"os"
	"path/filepath"
)

// Loader 負責從檔案系統載入與驗證素材
type Loader struct {
	mediaDir string
}

// NewLoader 建立新的素材載入器
func NewLoader(mediaDir string) *Loader {
	return &Loader{mediaDir: mediaDir}
}

// Load 根據 ID 載入素材組合
func (l *Loader) Load(id string) (*MediaComposition, error) {
	dir := filepath.Join(l.mediaDir, id)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("素材組合 %q 不存在", id)
	}

	audio, err := l.findAudio(dir)
	if err != nil {
		return nil, err
	}

	bg, err := l.findBackground(dir)
	if err != nil {
		return nil, err
	}

	subtitle := l.findSubtitle(dir)
	transcript := l.findTranscript(dir)

	return &MediaComposition{
		ID:         id,
		Audio:      *audio,
		Background: *bg,
		Subtitle:   subtitle,
		Transcript: transcript,
	}, nil
}

func (l *Loader) findAudio(dir string) (*Audio, error) {
	for _, ext := range SupportedAudioFormats() {
		path := filepath.Join(dir, "audio"+ext)
		if _, err := os.Stat(path); err == nil {
			return &Audio{Path: path, Format: ext[1:]}, nil
		}
	}
	return nil, fmt.Errorf("在 %q 中找不到音檔（支援格式：MP3、WAV）", dir)
}

func (l *Loader) findBackground(dir string) (*Background, error) {
	for _, ext := range SupportedImageFormats() {
		path := filepath.Join(dir, "background"+ext)
		if _, err := os.Stat(path); err == nil {
			return &Background{Path: path, Format: ext[1:]}, nil
		}
	}
	return nil, fmt.Errorf("在 %q 中找不到背景圖片（支援格式：JPG、PNG）", dir)
}

func (l *Loader) findSubtitle(dir string) *Subtitle {
	for _, ext := range SupportedSubtitleFormats() {
		path := filepath.Join(dir, "subtitle"+ext)
		if _, err := os.Stat(path); err == nil {
			return &Subtitle{Path: path, Format: ext[1:]}
		}
	}
	return nil
}

func (l *Loader) findTranscript(dir string) *Transcript {
	for _, ext := range SupportedTranscriptFormats() {
		path := filepath.Join(dir, "transcript"+ext)
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			return &Transcript{Path: path, Format: ext[1:], Size: info.Size()}
		}
	}
	return nil
}
