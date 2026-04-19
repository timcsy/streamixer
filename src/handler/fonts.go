package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/timcsy/streamixer/src/fonts"
)

// FontHandler 處理字體相關 HTTP 請求
type FontHandler struct {
	manager *fonts.Manager
}

// NewFontHandler 建立 font handler
func NewFontHandler(manager *fonts.Manager) *FontHandler {
	return &FontHandler{manager: manager}
}

// List GET /fonts
func (h *FontHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.manager.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := fonts.ListResponse{
		DefaultFamily: h.manager.DefaultFamily(),
		Fonts:         list,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Upload POST /fonts
func (h *FontHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(60 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "multipart 解析失敗："+err.Error())
		return
	}
	file, header, err := r.FormFile("font")
	if err != nil {
		writeError(w, http.StatusBadRequest, "缺少 font 欄位")
		return
	}
	defer file.Close()

	font, err := h.manager.Upload(file, header.Size)
	if err != nil {
		status := http.StatusBadRequest
		switch {
		case errors.Is(err, fonts.ErrFontLimit), errors.Is(err, fonts.ErrFamilyTaken):
			status = http.StatusConflict
		case errors.Is(err, fonts.ErrInvalidFontHeader), errors.Is(err, fonts.ErrUnparsableFont):
			status = http.StatusBadRequest
		}
		writeError(w, status, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(font)
}

// Delete DELETE /fonts/{id}
func (h *FontHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.manager.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, fonts.ErrSystemFont):
			writeError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, fonts.ErrFontNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type setDefaultReq struct {
	FamilyName string `json:"family_name"`
}

// SetDefault PUT /fonts/default
func (h *FontHandler) SetDefault(w http.ResponseWriter, r *http.Request) {
	var req setDefaultReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON 解析失敗："+err.Error())
		return
	}
	if err := h.manager.SetDefaultFamily(req.FamilyName); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"default_family": h.manager.DefaultFamily()})
}
