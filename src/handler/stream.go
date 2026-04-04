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
		loader: media.NewLoader(cfg.MediaDir),
		pregen: composer.NewPregenManager(
			cfg.TmpDir, cfg.SegmentDuration,
			cfg.OutputWidth, cfg.OutputHeight,
			cfg.MaxPregenConcurrent,
			cache,
		),
		cache: cache,
	}
}

// SetupRouter 建立包含所有路由的 chi router
func SetupRouter(h *StreamHandler, uh *UploadHandler, sh *SampleHandler) http.Handler {
	r := chi.NewRouter()
	r.Get("/health", HealthHandler)
	r.Post("/upload/{id}", uh.Upload)
	r.Get("/compositions", uh.ListCompositions)
	r.Post("/sample", sh.GenerateSample)
	r.Get("/stream/{id}/index.m3u8", h.PlaylistHandler)
	r.Get("/stream/{id}/init.mp4", h.InitHandler)
	r.Get("/stream/{id}/{segment}", h.SegmentHandler)

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
