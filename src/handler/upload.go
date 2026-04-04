package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/timcsy/streamixer/src/config"
)

// UploadHandler 處理素材上傳
type UploadHandler struct {
	cfg config.Config
}

// NewUploadHandler 建立上傳 handler
func NewUploadHandler(cfg config.Config) *UploadHandler {
	return &UploadHandler{cfg: cfg}
}

// Upload 處理 POST /upload/{id}，接收 multipart form 的 audio、background、subtitle
func (h *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "缺少素材組合 ID")
		return
	}

	// 限制上傳大小 100MB
	r.ParseMultipartForm(100 << 20)

	dir := filepath.Join(h.cfg.MediaDir, id)
	os.MkdirAll(dir, 0755)

	saved := []string{}

	// 儲存 audio
	if err := saveFormFile(r, "audio", dir); err == nil {
		saved = append(saved, "audio")
	}

	// 儲存 background
	if err := saveFormFile(r, "background", dir); err == nil {
		saved = append(saved, "background")
	}

	// 儲存 subtitle（選擇性）
	if err := saveFormFile(r, "subtitle", dir); err == nil {
		saved = append(saved, "subtitle")
	}

	if len(saved) == 0 {
		writeError(w, http.StatusBadRequest, "未收到任何檔案")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":    id,
		"saved": saved,
		"stream": fmt.Sprintf("/stream/%s/index.m3u8", id),
	})
}

// ListCompositions 列出所有可用的素材組合
func (h *UploadHandler) ListCompositions(w http.ResponseWriter, r *http.Request) {
	entries, err := os.ReadDir(h.cfg.MediaDir)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]string{})
		return
	}

	ids := []string{}
	for _, e := range entries {
		if e.IsDir() {
			ids = append(ids, e.Name())
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ids)
}

func saveFormFile(r *http.Request, fieldName, dir string) error {
	file, header, err := r.FormFile(fieldName)
	if err != nil {
		return err
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	destPath := filepath.Join(dir, fieldName+ext)

	dest, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, file)
	return err
}
