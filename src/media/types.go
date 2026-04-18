package media

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Audio 代表音檔素材
type Audio struct {
	Path     string
	Format   string
	Duration float64 // 秒
}

// Background 代表背景影像素材
type Background struct {
	Path   string
	Format string
	Width  int
	Height int
}

// Subtitle 代表字幕素材
type Subtitle struct {
	Path   string
	Format string
}

// Transcript 代表逐字稿全文檔案（選填）
type Transcript struct {
	Path   string // 絕對路徑，例如 /data/media/<id>/transcript.pdf
	Format string // 副檔名（不含點），例如 pdf、txt、md、doc、docx
	Size   int64  // 檔案大小（bytes）
}

var supportedAudioFormats = map[string]bool{
	".mp3": true,
	".wav": true,
}

var supportedImageFormats = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
}

var supportedSubtitleFormats = map[string]bool{
	".srt": true,
	".vtt": true,
}

var supportedTranscriptFormats = map[string]bool{
	".txt":  true,
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".md":   true,
}

// ValidateAudioFormat 檢查音檔格式是否支援
func ValidateAudioFormat(path string) error {
	ext := strings.ToLower(filepath.Ext(path))
	if !supportedAudioFormats[ext] {
		return fmt.Errorf("不支援的音檔格式 %q，支援的格式：MP3、WAV", ext)
	}
	return nil
}

// ValidateImageFormat 檢查圖片格式是否支援
func ValidateImageFormat(path string) error {
	ext := strings.ToLower(filepath.Ext(path))
	if !supportedImageFormats[ext] {
		return fmt.Errorf("不支援的圖片格式 %q，支援的格式：JPG、PNG", ext)
	}
	return nil
}

// ValidateSubtitleFormat 檢查字幕格式是否支援
func ValidateSubtitleFormat(path string) error {
	ext := strings.ToLower(filepath.Ext(path))
	if !supportedSubtitleFormats[ext] {
		return fmt.Errorf("不支援的字幕格式 %q，支援的格式：SRT、VTT", ext)
	}
	return nil
}

// SupportedAudioFormats 回傳支援的音檔格式清單
func SupportedAudioFormats() []string {
	return []string{".mp3", ".wav"}
}

// SupportedImageFormats 回傳支援的圖片格式清單
func SupportedImageFormats() []string {
	return []string{".jpg", ".jpeg", ".png"}
}

// SupportedSubtitleFormats 回傳支援的字幕格式清單
func SupportedSubtitleFormats() []string {
	return []string{".srt", ".vtt"}
}

// ValidateTranscriptFormat 檢查逐字稿格式是否支援
func ValidateTranscriptFormat(path string) error {
	ext := strings.ToLower(filepath.Ext(path))
	if !supportedTranscriptFormats[ext] {
		return fmt.Errorf("不支援的逐字稿格式 %q，支援的格式：TXT、PDF、DOC、DOCX、MD", ext)
	}
	return nil
}

// SupportedTranscriptFormats 回傳支援的逐字稿格式清單
func SupportedTranscriptFormats() []string {
	return []string{".txt", ".pdf", ".doc", ".docx", ".md"}
}
