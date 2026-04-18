package handler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/timcsy/streamixer/src/config"
	"github.com/timcsy/streamixer/src/media"
)

// MediaDownloadHandler 處理原始媒體檔案（音檔、逐字稿）的下載
type MediaDownloadHandler struct {
	cfg    config.Config
	loader *media.Loader
}

// NewMediaDownloadHandler 建立 media download handler
func NewMediaDownloadHandler(cfg config.Config) *MediaDownloadHandler {
	return &MediaDownloadHandler{
		cfg:    cfg,
		loader: media.NewLoader(cfg.MediaDir),
	}
}

// Audio 回傳原始音檔（串流輸出、保留副檔名）
func (h *MediaDownloadHandler) Audio(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	comp, err := h.loader.Load(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	ext := strings.ToLower(filepath.Ext(comp.Audio.Path))
	filename := id + ext

	w.Header().Set("Content-Type", audioMIME(ext))
	w.Header().Set("Content-Disposition", contentDisposition(filename))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeFile(w, r, comp.Audio.Path)
}

// Transcript 回傳逐字稿原檔（強制 attachment 下載、保留副檔名）
func (h *MediaDownloadHandler) Transcript(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	comp, err := h.loader.Load(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if comp.Transcript == nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("素材組合 %q 沒有逐字稿", id))
		return
	}

	ext := "." + comp.Transcript.Format
	filename := id + ext

	w.Header().Set("Content-Type", transcriptMIME(ext))
	w.Header().Set("Content-Disposition", contentDisposition(filename))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeFile(w, r, comp.Transcript.Path)
}

func audioMIME(ext string) string {
	switch strings.ToLower(ext) {
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".m4a", ".mp4":
		return "audio/mp4"
	case ".ogg":
		return "audio/ogg"
	case ".flac":
		return "audio/flac"
	default:
		return "application/octet-stream"
	}
}

func transcriptMIME(ext string) string {
	switch strings.ToLower(ext) {
	case ".txt":
		return "text/plain; charset=utf-8"
	case ".md":
		return "text/markdown; charset=utf-8"
	case ".pdf":
		return "application/pdf"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	default:
		return "application/octet-stream"
	}
}
