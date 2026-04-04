package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
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
}

// NewStreamHandler 建立新的串流 handler
func NewStreamHandler(cfg config.Config) *StreamHandler {
	return &StreamHandler{
		cfg:    cfg,
		loader: media.NewLoader(cfg.MediaDir),
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

	playlist := composer.GeneratePlaylist(duration, h.cfg.SegmentDuration)

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Cache-Control", "no-cache")
	w.Write([]byte(playlist))
}

// SegmentHandler 處理 HLS 分段請求
func (h *StreamHandler) SegmentHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	segment := chi.URLParam(r, "segment")

	if !strings.HasPrefix(segment, "seg_") || !strings.HasSuffix(segment, ".ts") {
		writeError(w, http.StatusBadRequest, "無效的分段名稱")
		return
	}

	// 解析分段編號
	numStr := strings.TrimPrefix(segment, "seg_")
	numStr = strings.TrimSuffix(numStr, ".ts")
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

	// 檢查分段編號是否在範圍內
	duration, err := composer.ProbeDuration(comp.Audio.Path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("無法讀取音檔資訊：%v", err))
		return
	}

	totalSegments := composer.SegmentCount(duration, h.cfg.SegmentDuration)
	if segIndex < 0 || segIndex >= totalSegments {
		writeError(w, http.StatusNotFound, fmt.Sprintf("分段 %d 不存在（共 %d 個分段）", segIndex, totalSegments))
		return
	}

	// 合成分段
	outDir := filepath.Join(h.cfg.TmpDir, id)
	segPath := filepath.Join(outDir, segment)

	// 如果分段已存在，直接回傳
	if _, err := os.Stat(segPath); err != nil {
		err = composer.GenerateSegment(comp, segPath, segIndex, h.cfg.SegmentDuration, h.cfg.OutputWidth, h.cfg.OutputHeight)
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("合成分段失敗：%v", err))
			return
		}
	}

	w.Header().Set("Content-Type", "video/mp2t")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	http.ServeFile(w, r, segPath)
}
