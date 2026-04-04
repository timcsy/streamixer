package config

import (
	"os"
)

// Config 應用程式設定
type Config struct {
	MediaDir        string // 素材檔案根目錄
	TmpDir          string // 合成分段暫存目錄（建議使用 tmpfs）
	Port            string // HTTP 伺服器 port
	SegmentDuration int    // 每個 HLS 分段秒數
	OutputWidth     int    // 輸出影片寬度
	OutputHeight    int    // 輸出影片高度
}

// Load 從環境變數載入設定，使用合理預設值
func Load() Config {
	return Config{
		MediaDir:        getEnv("MEDIA_DIR", "/media"),
		TmpDir:          getEnv("TMP_DIR", "/dev/shm/streamixer"),
		Port:            getEnv("PORT", "8080"),
		SegmentDuration: 6,
		OutputWidth:     1280,
		OutputHeight:    720,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
