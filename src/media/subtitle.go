package media

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidateSubtitleFile 驗證字幕檔存在且格式正確
func ValidateSubtitleFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("字幕檔不存在：%s", path)
	}

	ext := strings.ToLower(filepath.Ext(path))
	if !supportedSubtitleFormats[ext] {
		return fmt.Errorf("不支援的字幕格式 %q，支援的格式：SRT、VTT", ext)
	}

	return nil
}
