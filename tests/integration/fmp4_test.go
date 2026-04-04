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
	"github.com/timcsy/streamixer/src/media"
)

func setupFmp4TestMedia(t *testing.T) (string, *media.MediaComposition) {
	t.Helper()

	tmpDir := t.TempDir()
	mediaDir := filepath.Join(tmpDir, "media", "fmp4-test")
	os.MkdirAll(mediaDir, 0755)

	audioPath := filepath.Join(mediaDir, "audio.mp3")
	cmd := exec.Command("ffmpeg", "-y", "-f", "lavfi", "-i",
		"sine=frequency=440:duration=18", "-q:a", "9", audioPath)
	if err := cmd.Run(); err != nil {
		t.Skipf("需要 ffmpeg：%v", err)
	}

	bgPath := filepath.Join(mediaDir, "background.jpg")
	cmd = exec.Command("ffmpeg", "-y", "-f", "lavfi", "-i",
		"color=c=blue:s=320x240:d=1", "-frames:v", "1", bgPath)
	if err := cmd.Run(); err != nil {
		t.Skipf("需要 ffmpeg：%v", err)
	}

	comp := &media.MediaComposition{
		ID:         "fmp4-test",
		Audio:      media.Audio{Path: audioPath, Format: "mp3"},
		Background: media.Background{Path: bgPath, Format: "jpg"},
	}

	return tmpDir, comp
}

func TestFmp4_PregenProducesInitAndM4S(t *testing.T) {
	tmpDir, comp := setupFmp4TestMedia(t)
	outDir := filepath.Join(tmpDir, "output")

	mgr := composer.NewPregenManager(outDir, 6, 320, 240, 3)

	duration, err := composer.ProbeDuration(comp.Audio.Path)
	if err != nil {
		t.Fatalf("取得音檔長度失敗：%v", err)
	}

	mgr.StartPregen(comp, duration)

	// 等待預生成完成
	deadline := time.After(30 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("預生成超時")
		default:
			status := mgr.GetStatus(comp.ID)
			if status != nil && status.Status == composer.PregenCompleted {
				// 驗證 init.mp4 存在
				if !mgr.IsInitReady(comp.ID) {
					t.Error("init.mp4 應存在")
				}

				// 驗證 .m4s 分段存在
				if !mgr.IsSegmentReady(comp.ID, 0) {
					t.Error("seg_000.m4s 應存在")
				}

				// 驗證用 ffprobe 可辨識 init.mp4
				initPath := mgr.GetInitPath(comp.ID)
				cmd := exec.Command("ffprobe", "-v", "error", "-show_format", initPath)
				out, err := cmd.Output()
				if err != nil {
					t.Fatalf("ffprobe init.mp4 失敗：%v", err)
				}
				if !strings.Contains(string(out), "mp4") {
					t.Error("init.mp4 應為有效的 MP4 格式")
				}

				// 驗證 .m4s 分段存在且有內容
				segPath := mgr.GetSegmentPath(comp.ID, 0)
				segInfo, segErr := os.Stat(segPath)
				if segErr != nil {
					t.Fatalf("seg_000.m4s 不存在：%v", segErr)
				}
				if segInfo.Size() == 0 {
					t.Error("seg_000.m4s 不應為空")
				}

				return
			}
			if status != nil && status.Status == composer.PregenFailed {
				t.Fatalf("預生成失敗：%v", status.Error)
			}
			time.Sleep(200 * time.Millisecond)
		}
	}
}

func TestFmp4_PlaylistContainsExtXMap(t *testing.T) {
	tmpDir, _ := setupFmp4TestMedia(t)

	cfg := config.Config{
		MediaDir:            filepath.Join(tmpDir, "media"),
		TmpDir:              filepath.Join(tmpDir, "output"),
		SegmentDuration:     6,
		OutputWidth:         320,
		OutputHeight:        240,
		MaxPregenConcurrent: 3,
	}

	h := handler.NewStreamHandler(cfg)
	uh := handler.NewUploadHandler(cfg)
	sh := handler.NewSampleHandler(cfg)
	router := handler.SetupRouter(h, uh, sh)

	req := httptest.NewRequest("GET", "/stream/fmp4-test/index.m3u8", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("預期 200，實際 %d：%s", w.Code, w.Body.String())
	}

	body := w.Body.String()
	if !strings.Contains(body, "#EXT-X-VERSION:7") {
		t.Error("playlist 應包含 #EXT-X-VERSION:7")
	}
	if !strings.Contains(body, "#EXT-X-MAP:URI=\"init.mp4\"") {
		t.Error("playlist 應包含 #EXT-X-MAP:URI=\"init.mp4\"")
	}
	if !strings.Contains(body, ".m4s") {
		t.Error("playlist 分段應為 .m4s")
	}
	if strings.Contains(body, ".ts") {
		t.Error("playlist 不應包含 .ts 分段")
	}
}

func TestFmp4_InitEndpoint(t *testing.T) {
	tmpDir, _ := setupFmp4TestMedia(t)

	cfg := config.Config{
		MediaDir:            filepath.Join(tmpDir, "media"),
		TmpDir:              filepath.Join(tmpDir, "output"),
		SegmentDuration:     6,
		OutputWidth:         320,
		OutputHeight:        240,
		MaxPregenConcurrent: 3,
	}

	h := handler.NewStreamHandler(cfg)
	uh := handler.NewUploadHandler(cfg)
	sh := handler.NewSampleHandler(cfg)
	router := handler.SetupRouter(h, uh, sh)

	// 先請求 playlist 觸發預生成
	req := httptest.NewRequest("GET", "/stream/fmp4-test/index.m3u8", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 等待預生成產生 init.mp4
	time.Sleep(3 * time.Second)

	// 請求 init.mp4
	req = httptest.NewRequest("GET", "/stream/fmp4-test/init.mp4", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("預期 200，實際 %d", w.Code)
	}
	if w.Header().Get("Content-Type") != "video/mp4" {
		t.Errorf("Content-Type 應為 video/mp4，實際 %s", w.Header().Get("Content-Type"))
	}
	if w.Body.Len() == 0 {
		t.Error("init.mp4 不應為空")
	}
}

func TestFmp4_SegmentContentType(t *testing.T) {
	tmpDir, _ := setupFmp4TestMedia(t)

	cfg := config.Config{
		MediaDir:            filepath.Join(tmpDir, "media"),
		TmpDir:              filepath.Join(tmpDir, "output"),
		SegmentDuration:     6,
		OutputWidth:         320,
		OutputHeight:        240,
		MaxPregenConcurrent: 3,
	}

	h := handler.NewStreamHandler(cfg)
	uh := handler.NewUploadHandler(cfg)
	sh := handler.NewSampleHandler(cfg)
	router := handler.SetupRouter(h, uh, sh)

	// 先請求 playlist 觸發預生成
	req := httptest.NewRequest("GET", "/stream/fmp4-test/index.m3u8", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 請求 .m4s 分段
	req = httptest.NewRequest("GET", "/stream/fmp4-test/seg_000.m4s", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("預期 200，實際 %d：%s", w.Code, w.Body.String())
	}
	if w.Header().Get("Content-Type") != "video/mp4" {
		t.Errorf("Content-Type 應為 video/mp4，實際 %s", w.Header().Get("Content-Type"))
	}
}
