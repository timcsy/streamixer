package integration

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/timcsy/streamixer/src/composer"
	"github.com/timcsy/streamixer/src/config"
	"github.com/timcsy/streamixer/src/handler"
	"github.com/timcsy/streamixer/src/media"
)

// setupTestMedia 建立測試用的素材目錄與檔案
func setupTestMedia(t *testing.T) (string, func()) {
	t.Helper()

	tmpDir := t.TempDir()
	mediaDir := filepath.Join(tmpDir, "media", "test-sermon")
	os.MkdirAll(mediaDir, 0755)

	// 用 ffmpeg 產生 10 秒的測試音檔
	audioPath := filepath.Join(mediaDir, "audio.mp3")
	cmd := exec.Command("ffmpeg", "-y", "-f", "lavfi", "-i",
		"sine=frequency=440:duration=10", "-q:a", "9", audioPath)
	if err := cmd.Run(); err != nil {
		t.Skipf("需要 ffmpeg 來執行整合測試：%v", err)
	}

	// 產生測試用背景圖片
	bgPath := filepath.Join(mediaDir, "background.jpg")
	cmd = exec.Command("ffmpeg", "-y", "-f", "lavfi", "-i",
		"color=c=blue:s=320x240:d=1", "-frames:v", "1", bgPath)
	if err := cmd.Run(); err != nil {
		t.Skipf("需要 ffmpeg 來產生測試圖片：%v", err)
	}

	outDir := filepath.Join(tmpDir, "output")
	os.MkdirAll(outDir, 0755)

	return tmpDir, func() {}
}

func TestFullStreamWorkflow(t *testing.T) {
	tmpDir, cleanup := setupTestMedia(t)
	defer cleanup()

	cfg := config.Config{
		MediaDir:        filepath.Join(tmpDir, "media"),
		TmpDir:          filepath.Join(tmpDir, "output"),
		Port:            "0",
		SegmentDuration: 6,
		OutputWidth:     320,
		OutputHeight:    240,
	}

	loader := media.NewLoader(cfg.MediaDir)
	comp, err := loader.Load("test-sermon")
	if err != nil {
		t.Fatalf("載入素材失敗：%v", err)
	}

	// 用 ffprobe 取得音檔長度
	duration, err := composer.ProbeDuration(comp.Audio.Path)
	if err != nil {
		t.Fatalf("取得音檔長度失敗：%v", err)
	}

	if duration < 9.0 || duration > 11.0 {
		t.Errorf("預期音檔長度約 10 秒，實際 %.2f 秒", duration)
	}

	// 產生 playlist
	playlist := composer.GeneratePlaylist(duration, cfg.SegmentDuration)
	if !strings.Contains(playlist, "#EXTM3U") {
		t.Error("playlist 格式不正確")
	}

	// 合成第一個分段
	segPath := filepath.Join(cfg.TmpDir, "test-sermon", "seg_000.ts")
	os.MkdirAll(filepath.Dir(segPath), 0755)
	err = composer.GenerateSegment(comp, segPath, 0, cfg.SegmentDuration, cfg.OutputWidth, cfg.OutputHeight)
	if err != nil {
		t.Fatalf("合成分段失敗：%v", err)
	}

	// 用 ffprobe 驗證分段格式
	cmd := exec.Command("ffprobe", "-v", "error",
		"-show_entries", "format=duration",
		"-of", "csv=p=0", segPath)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("ffprobe 驗證分段失敗：%v", err)
	}
	if len(strings.TrimSpace(string(out))) == 0 {
		t.Error("分段檔案無有效內容")
	}
}

func TestPlaylistHandler_ReturnsValidM3U8(t *testing.T) {
	tmpDir, cleanup := setupTestMedia(t)
	defer cleanup()

	cfg := config.Config{
		MediaDir:        filepath.Join(tmpDir, "media"),
		TmpDir:          filepath.Join(tmpDir, "output"),
		SegmentDuration: 6,
		OutputWidth:     320,
		OutputHeight:    240,
	}

	h := handler.NewStreamHandler(cfg)
	uh := handler.NewUploadHandler(cfg)
	sh := handler.NewSampleHandler(cfg)
	router := handler.SetupRouter(h, uh, sh)

	req := httptest.NewRequest("GET", "/stream/test-sermon/index.m3u8", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("預期 200，實際 %d：%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Header().Get("Content-Type"), "mpegurl") {
		t.Error("Content-Type 應為 application/vnd.apple.mpegurl")
	}
	if !strings.Contains(w.Body.String(), "#EXTM3U") {
		t.Error("回應 body 應為有效的 HLS playlist")
	}
}

func TestSegmentHandler_ReturnsValidTS(t *testing.T) {
	tmpDir, cleanup := setupTestMedia(t)
	defer cleanup()

	cfg := config.Config{
		MediaDir:        filepath.Join(tmpDir, "media"),
		TmpDir:          filepath.Join(tmpDir, "output"),
		SegmentDuration: 6,
		OutputWidth:     320,
		OutputHeight:    240,
	}

	h := handler.NewStreamHandler(cfg)
	uh := handler.NewUploadHandler(cfg)
	sh := handler.NewSampleHandler(cfg)
	router := handler.SetupRouter(h, uh, sh)

	req := httptest.NewRequest("GET", "/stream/test-sermon/seg_000.ts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("預期 200，實際 %d：%s", w.Code, w.Body.String())
	}
	if w.Header().Get("Content-Type") != "video/mp2t" {
		t.Errorf("Content-Type 應為 video/mp2t，實際 %s", w.Header().Get("Content-Type"))
	}
	if w.Body.Len() == 0 {
		t.Error("分段回應不應為空")
	}
}
