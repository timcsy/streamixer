package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/timcsy/streamixer/src/composer"
	"github.com/timcsy/streamixer/src/config"
	"github.com/timcsy/streamixer/src/media"
)

// StreamHandler 處理串流相關的 HTTP 請求
type StreamHandler struct {
	cfg    config.Config
	loader *media.Loader
	pregen *composer.PregenManager
	cache  *composer.CacheManager
}

// NewStreamHandler 建立新的串流 handler
func NewStreamHandler(cfg config.Config, cache *composer.CacheManager) *StreamHandler {
	return &StreamHandler{
		cfg:    cfg,
		loader: media.NewLoaderWithDefaultFont(cfg.MediaDir, defaultFontFile(cfg)),
		pregen: composer.NewPregenManager(
			cfg.TmpDir, cfg.SegmentDuration,
			cfg.OutputWidth, cfg.OutputHeight,
			cfg.MaxPregenConcurrent,
			cache,
		),
		cache: cache,
	}
}

func defaultFontFile(cfg config.Config) string {
	if cfg.FontDir == "" {
		return ""
	}
	return cfg.FontDir + "/default.txt"
}

// corsMiddleware 設定 CORS 允許來源
func corsMiddleware(origins string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", origins)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Range, X-API-Key")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// apiKeyMiddleware 驗證 API Key（僅保護寫入端點）
func apiKeyMiddleware(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if apiKey != "" {
				key := r.Header.Get("X-API-Key")
				if key == "" {
					key = r.URL.Query().Get("api_key")
				}
				if key != apiKey {
					writeError(w, http.StatusUnauthorized, "無效的 API Key")
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// SetupRouter 建立包含所有路由的 chi router
func SetupRouter(h *StreamHandler, uh *UploadHandler, sh *SampleHandler, cfg config.Config) http.Handler {
	return SetupRouterWithSweeper(h, uh, sh, cfg, nil)
}

// SetupRouterWithSweeper 同 SetupRouter，但接受 sweeper 以啟用 /config 端點
func SetupRouterWithSweeper(h *StreamHandler, uh *UploadHandler, sh *SampleHandler, cfg config.Config, sweeper *composer.Sweeper) http.Handler {
	return SetupRouterFull(h, uh, sh, cfg, sweeper, nil)
}

// SetupRouterFull 加上 FontHandler
func SetupRouterFull(h *StreamHandler, uh *UploadHandler, sh *SampleHandler, cfg config.Config, sweeper *composer.Sweeper, fh *FontHandler) http.Handler {
	r := chi.NewRouter()
	r.Use(corsMiddleware(cfg.CORSOrigins))
	r.Get("/health", HealthHandler)

	// 寫入端點需要 API Key
	authMw := apiKeyMiddleware(cfg.APIKey)
	r.With(authMw).Post("/upload/{id}", uh.Upload)
	r.With(authMw).Post("/sample", sh.GenerateSample)

	// 快取設定端點（GET 公開、PUT 需認證）
	if sweeper != nil {
		ch := NewConfigHandler(h.cache, sweeper)
		r.Get("/config", ch.Get)
		r.With(authMw).Put("/config", ch.Put)
	}

	// 字體管理端點
	if fh != nil {
		r.Group(func(r chi.Router) {
			r.Get("/fonts", fh.List)
			r.Group(func(r chi.Router) {
				r.Use(authMw)
				r.Post("/fonts", fh.Upload)
				r.Put("/fonts/default", fh.SetDefault)
				r.Delete("/fonts/{id}", fh.Delete)
			})
		})
	}

	r.Get("/compositions", uh.ListCompositions)
	r.Get("/stream/{id}/index.m3u8", h.PlaylistHandler)
	r.Get("/stream/{id}/init.mp4", h.InitHandler)
	r.Get("/stream/{id}/{segment}", h.SegmentHandler)

	// 下載端點
	dh := NewDownloadHandler(cfg, h.pregen)
	r.Get("/download/{id}", dh.Download)
	r.Get("/progress/{id}", dh.Progress)

	// 原始媒體下載端點（音檔、逐字稿）
	mdh := NewMediaDownloadHandler(cfg)
	r.Get("/audio/{id}", mdh.Audio)
	r.Get("/transcript/{id}", mdh.Transcript)

	// 靜態檔案（前端）
	staticDir := http.Dir("static")
	r.Get("/*", http.FileServer(staticDir).ServeHTTP)

	return r
}

// PlaylistHandler 處理 HLS playlist 請求
func (h *StreamHandler) PlaylistHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	comp, err := h.loader.Load(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	duration, err := composer.ProbeDuration(comp.Audio.Path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("無法讀取音檔資訊：%v", err))
		return
	}

	// 背景啟動預生成
	h.pregen.StartPregen(comp, duration)
	h.cache.Touch(id)

	// 立即回傳手動計算的 playlist（顯示完整時長，不等 FFmpeg）
	// 分段由預生成在背景產生，因為 -force_key_frames 確保切割點精確匹配
	playlist := composer.GeneratePlaylist(duration, h.cfg.SegmentDuration)

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Cache-Control", "no-cache")
	w.Write([]byte(playlist))
}

// InitHandler 處理 init.mp4 請��
func (h *StreamHandler) InitHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	comp, err := h.loader.Load(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	// 確保預生成已啟動
	duration, err := composer.ProbeDuration(comp.Audio.Path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("無法讀取音檔資訊：%v", err))
		return
	}
	h.pregen.StartPregen(comp, duration)

	// 等待 init.mp4 就緒
	if err := h.pregen.WaitForInit(id, 30); err != nil {
		writeError(w, http.StatusServiceUnavailable, fmt.Sprintf("init segment 尚未就緒：%v", err))
		return
	}

	h.cache.Touch(id)
	initPath := h.pregen.GetInitPath(id)
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	http.ServeFile(w, r, initPath)
}

// SegmentHandler 處理 HLS fMP4 分段請求
func (h *StreamHandler) SegmentHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	segment := chi.URLParam(r, "segment")

	if !strings.HasPrefix(segment, "seg_") || !strings.HasSuffix(segment, ".m4s") {
		writeError(w, http.StatusBadRequest, "無效的分段名稱")
		return
	}

	numStr := strings.TrimPrefix(segment, "seg_")
	numStr = strings.TrimSuffix(numStr, ".m4s")
	segIndex, err := strconv.Atoi(numStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "無效的分段編號")
		return
	}

	comp, err := h.loader.Load(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	duration, err := composer.ProbeDuration(comp.Audio.Path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("無法讀取音檔資訊：%v", err))
		return
	}

	totalSegments := composer.SegmentCount(duration, h.cfg.SegmentDuration)
	if segIndex < 0 || segIndex >= totalSegments {
		writeError(w, http.StatusNotFound, fmt.Sprintf("分段 %d 不存在��共 %d 個分段）", segIndex, totalSegments))
		return
	}

	h.cache.Touch(id)

	// 等待預生成產出該分段
	if err := h.pregen.WaitForSegment(id, segIndex, 30); err != nil {
		writeError(w, http.StatusServiceUnavailable, fmt.Sprintf("分段 %d 尚未就緒：%v", segIndex, err))
		return
	}

	segPath := h.pregen.GetSegmentPath(id, segIndex)
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	http.ServeFile(w, r, segPath)
}
