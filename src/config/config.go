package config

import (
	"os"
	"strconv"
	"time"
)

// Config 應用程式設定
type Config struct {
	MediaDir            string        // 素材檔案根目錄
	TmpDir              string        // 合成分段暫存目錄（建議使用 tmpfs）
	Port                string        // HTTP 伺服器 port
	SegmentDuration     int           // 每個 HLS 分段秒數
	OutputWidth         int           // 輸出影片寬度
	OutputHeight        int           // 輸出影片高度
	MaxPregenConcurrent int           // 最大並發預生成數
	CacheTTL            time.Duration // 快取過期時間
	CacheMaxSize        int64         // 快取容量上限（bytes，0 = tmpfs 的 80%）
	CacheSweepInterval  time.Duration // 清掃頻率
	APIKey              string        // Upload API 認證金鑰（空字串 = 不認證）
	CORSOrigins         string        // CORS 允許來源（逗號分隔，* = 全部）
	FontDir             string        // 使用者字體儲存根目錄（含 user/、default.txt）
	FontSymlinkDir      string        // 掛進 fontconfig 可見路徑的 symlink 目錄
	SystemFontDirs      []string      // 系統字體掃描路徑
	MaxFontSize         int64         // 單一字體檔大小上限（bytes）
	MaxFontCount        int           // 使用者字體總數上限
}

// Load 從環境變數載入設定，使用合理預設值
func Load() Config {
	return Config{
		MediaDir:            getEnv("MEDIA_DIR", "/media"),
		TmpDir:              getEnv("TMP_DIR", "/dev/shm/streamixer"),
		Port:                getEnv("PORT", "8080"),
		SegmentDuration:     6,
		OutputWidth:         1280,
		OutputHeight:        720,
		MaxPregenConcurrent: 3,
		CacheTTL:            getDuration("CACHE_TTL", 30*time.Minute),
		CacheMaxSize:        getInt64("CACHE_MAX_SIZE", 0),
		CacheSweepInterval:  getDuration("CACHE_SWEEP_INTERVAL", 5*time.Minute),
		APIKey:              getEnv("API_KEY", ""),
		CORSOrigins:         getEnv("CORS_ORIGINS", "*"),
		FontDir:             getEnv("FONT_DIR", "/fonts"),
		FontSymlinkDir:      getEnv("FONT_SYMLINK_DIR", "/usr/share/fonts/user"),
		SystemFontDirs:      []string{"/usr/share/fonts"},
		MaxFontSize:         getInt64("MAX_FONT_SIZE", 50*1024*1024),
		MaxFontCount:        int(getInt64("MAX_FONT_COUNT", 50)),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

func getInt64(key string, fallback int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return fallback
}
