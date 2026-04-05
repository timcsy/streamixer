package integration

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/timcsy/streamixer/src/composer"
	"github.com/timcsy/streamixer/src/config"
	"github.com/timcsy/streamixer/src/handler"
)

func setupDownloadTestMedia(t *testing.T) (string, config.Config) {
	t.Helper()

	tmpDir := t.TempDir()
	mediaDir := filepath.Join(tmpDir, "media", "dl-test")
	os.MkdirAll(mediaDir, 0755)

	audioPath := filepath.Join(mediaDir, "audio.mp3")
	cmd := exec.Command("ffmpeg", "-y", "-f", "lavfi", "-i",
		"sine=frequency=440:duration=12", "-q:a", "9", audioPath)
	if err := cmd.Run(); err != nil {
		t.Skipf("需要 ffmpeg：%v", err)
	}

	bgPath := filepath.Join(mediaDir, "background.jpg")
	cmd = exec.Command("ffmpeg", "-y", "-f", "lavfi", "-i",
		"color=c=blue:s=320x240:d=1", "-frames:v", "1", bgPath)
	if err := cmd.Run(); err != nil {
		t.Skipf("需要 ffmpeg：%v", err)
	}

	cfg := config.Config{
		MediaDir:            filepath.Join(tmpDir, "media"),
		TmpDir:              filepath.Join(tmpDir, "output"),
		SegmentDuration:     6,
		OutputWidth:         320,
		OutputHeight:        240,
		MaxPregenConcurrent: 3,
		CORSOrigins:         "*",
	}

	return tmpDir, cfg
}

func TestDownload_ReturnsValidMP4(t *testing.T) {
	_, cfg := setupDownloadTestMedia(t)

	cache := composer.NewCacheManager(cfg.TmpDir, 10*time.Minute, 0)
	h := handler.NewStreamHandler(cfg, cache)
	uh := handler.NewUploadHandler(cfg)
	sh := handler.NewSampleHandler(cfg)
	router := handler.SetupRouter(h, uh, sh, cfg)

	// 先請求 playlist 觸發預生成
	req := httptest.NewRequest("GET", "/stream/dl-test/index.m3u8", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("playlist 請求失敗：%d", w.Code)
	}

	// 等待預生成完成
	time.Sleep(5 * time.Second)

	// 請求下載
	req = httptest.NewRequest("GET", "/download/dl-test", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("預期 200，實際 %d：%s", w.Code, w.Body.String())
	}

	if w.Header().Get("Content-Type") != "video/mp4" {
		t.Errorf("Content-Type 應為 video/mp4，實際 %s", w.Header().Get("Content-Type"))
	}

	if !strings.Contains(w.Header().Get("Content-Disposition"), "attachment") {
		t.Error("應包含 Content-Disposition: attachment")
	}

	if w.Body.Len() == 0 {
		t.Error("回應不應為空")
	}

	// 將回應寫入暫存檔並用 ffprobe 驗證
	tmpFile := filepath.Join(t.TempDir(), "test.mp4")
	os.WriteFile(tmpFile, w.Body.Bytes(), 0644)

	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=format_name", "-of", "csv=p=0", tmpFile)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("ffprobe 驗證失敗：%v", err)
	}
	if !strings.Contains(string(out), "mov") && !strings.Contains(string(out), "mp4") {
		t.Errorf("應為 MP4 格式，實際 %s", strings.TrimSpace(string(out)))
	}
}

func TestDownload_NotFound(t *testing.T) {
	_, cfg := setupDownloadTestMedia(t)

	cache := composer.NewCacheManager(cfg.TmpDir, 10*time.Minute, 0)
	h := handler.NewStreamHandler(cfg, cache)
	uh := handler.NewUploadHandler(cfg)
	sh := handler.NewSampleHandler(cfg)
	router := handler.SetupRouter(h, uh, sh, cfg)

	req := httptest.NewRequest("GET", "/download/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("預期 404，實際 %d", w.Code)
	}
}
