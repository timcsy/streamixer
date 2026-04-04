package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/timcsy/streamixer/src/config"
	"github.com/timcsy/streamixer/src/handler"
)

func TestHealthEndpoint(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	handler.HealthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("預期 200，實際 %d", w.Code)
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Errorf("預期 status=ok，實際 %s", resp["status"])
	}
}

func TestPlaylistEndpoint_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.Config{
		MediaDir:        tmpDir,
		TmpDir:          t.TempDir(),
		SegmentDuration: 6,
		OutputWidth:     320,
		OutputHeight:    240,
	}

	h := handler.NewStreamHandler(cfg)
	uh := handler.NewUploadHandler(cfg)
	sh := handler.NewSampleHandler(cfg)
	router := handler.SetupRouter(h, uh, sh)

	req := httptest.NewRequest("GET", "/stream/nonexistent/index.m3u8", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("預期 404，實際 %d", w.Code)
	}
}

func TestSegmentEndpoint_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.Config{
		MediaDir:        tmpDir,
		TmpDir:          t.TempDir(),
		SegmentDuration: 6,
		OutputWidth:     320,
		OutputHeight:    240,
	}

	h := handler.NewStreamHandler(cfg)
	uh := handler.NewUploadHandler(cfg)
	sh := handler.NewSampleHandler(cfg)
	router := handler.SetupRouter(h, uh, sh)

	req := httptest.NewRequest("GET", "/stream/nonexistent/seg_000.m4s", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("預期 404，實際 %d", w.Code)
	}
}
