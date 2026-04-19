package main

import (
	"log"
	"net/http"
	"os"

	"github.com/timcsy/streamixer/src/composer"
	"github.com/timcsy/streamixer/src/config"
	"github.com/timcsy/streamixer/src/fonts"
	"github.com/timcsy/streamixer/src/handler"
)

func main() {
	cfg := config.Load()

	// 確保目錄存在
	os.MkdirAll(cfg.TmpDir, 0755)
	os.MkdirAll(cfg.MediaDir, 0755)

	// 初始化快取管理
	cache := composer.NewCacheManager(cfg.TmpDir, cfg.CacheTTL, cfg.CacheMaxSize)
	sweeper := composer.NewSweeper(cache, cfg.CacheSweepInterval)
	sweeper.Start()

	// 字體管理
	fontMgr, err := fonts.NewManager(fonts.Config{
		FontDir:    cfg.FontDir,
		SymlinkDir: cfg.FontSymlinkDir,
		SystemDirs: cfg.SystemFontDirs,
		MaxSize:    cfg.MaxFontSize,
		MaxCount:   cfg.MaxFontCount,
	})
	if err != nil {
		log.Fatalf("字體管理初始化失敗：%v", err)
	}

	streamH := handler.NewStreamHandler(cfg, cache)
	uploadH := handler.NewUploadHandler(cfg)
	sampleH := handler.NewSampleHandler(cfg)
	fontH := handler.NewFontHandler(fontMgr)
	router := handler.SetupRouterFull(streamH, uploadH, sampleH, cfg, sweeper, fontH)

	log.Printf("Streamixer 啟動中，port %s，素材目錄 %s", cfg.Port, cfg.MediaDir)
	log.Printf("快取設定：TTL %v，容量上限 %d bytes，清掃頻率 %v", cfg.CacheTTL, cfg.CacheMaxSize, cfg.CacheSweepInterval)
	log.Printf("開啟瀏覽器前往 http://localhost:%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatalf("伺服器啟動失敗：%v", err)
	}
}
