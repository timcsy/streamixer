package integration

import (
	"bytes"
	"crypto/sha256"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/timcsy/streamixer/src/composer"
	"github.com/timcsy/streamixer/src/config"
	"github.com/timcsy/streamixer/src/handler"
)

func setupMediaDownloadServer(t *testing.T) (*httptest.Server, config.Config, string) {
	t.Helper()
	tmpDir := t.TempDir()
	cfg := config.Config{
		MediaDir:            filepath.Join(tmpDir, "media"),
		TmpDir:              filepath.Join(tmpDir, "output"),
		SegmentDuration:     6,
		OutputWidth:         320,
		OutputHeight:        240,
		MaxPregenConcurrent: 3,
		CORSOrigins:         "*",
	}
	os.MkdirAll(cfg.MediaDir, 0755)

	cache := composer.NewCacheManager(cfg.TmpDir, 10*time.Minute, 0)
	h := handler.NewStreamHandler(cfg, cache)
	uh := handler.NewUploadHandler(cfg)
	sh := handler.NewSampleHandler(cfg)
	router := handler.SetupRouter(h, uh, sh, cfg)

	return httptest.NewServer(router), cfg, tmpDir
}

func writeFixture(t *testing.T, dir, name string, content []byte) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("寫入 fixture 失敗：%v", err)
	}
	return path
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return string([]byte{
		hex(sum[0] >> 4), hex(sum[0] & 0xf),
		hex(sum[1] >> 4), hex(sum[1] & 0xf),
		hex(sum[2] >> 4), hex(sum[2] & 0xf),
		hex(sum[3] >> 4), hex(sum[3] & 0xf),
	})
}

func hex(b byte) byte {
	if b < 10 {
		return '0' + b
	}
	return 'a' + (b - 10)
}

// placeMediaFiles 直接把檔案放到 composition dir（跳過 multipart 上傳簡化測試）
func placeMediaFiles(t *testing.T, cfg config.Config, id string, files map[string][]byte) {
	t.Helper()
	dir := filepath.Join(cfg.MediaDir, id)
	os.MkdirAll(dir, 0755)
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), content, 0644); err != nil {
			t.Fatalf("寫入 %s 失敗：%v", name, err)
		}
	}
}

func TestAudioDownload_Basic(t *testing.T) {
	srv, cfg, _ := setupMediaDownloadServer(t)
	defer srv.Close()

	content := []byte("FAKE MP3 BINARY CONTENT")
	placeMediaFiles(t, cfg, "abc", map[string][]byte{
		"audio.mp3":      content,
		"background.jpg": []byte("fakejpg"),
	})

	resp, err := http.Get(srv.URL + "/audio/abc")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Content-Type"); got != "audio/mpeg" {
		t.Errorf("Content-Type = %q, want audio/mpeg", got)
	}
	disp := resp.Header.Get("Content-Disposition")
	if !strings.Contains(disp, "attachment") || !strings.Contains(disp, "abc.mp3") {
		t.Errorf("Content-Disposition = %q", disp)
	}

	body, _ := io.ReadAll(resp.Body)
	if sha256Hex(body)[:8] != sha256Hex(content)[:8] {
		t.Errorf("下載內容與原檔不一致")
	}
}

func TestAudioDownload_404_NoAudio(t *testing.T) {
	srv, cfg, _ := setupMediaDownloadServer(t)
	defer srv.Close()

	// 組合存在但沒音檔（也沒背景會在 loader 層先失敗，先塞背景）
	placeMediaFiles(t, cfg, "noaudio", map[string][]byte{
		"background.jpg": []byte("fakejpg"),
	})

	resp, err := http.Get(srv.URL + "/audio/noaudio")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 404 {
		t.Errorf("want 404, got %d", resp.StatusCode)
	}
}

func TestAudioDownload_404_MissingComposition(t *testing.T) {
	srv, _, _ := setupMediaDownloadServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/audio/does-not-exist")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 404 {
		t.Errorf("want 404, got %d", resp.StatusCode)
	}
}

func TestAudioDownload_ChineseSlug(t *testing.T) {
	srv, cfg, _ := setupMediaDownloadServer(t)
	defer srv.Close()

	id := "測試"
	placeMediaFiles(t, cfg, id, map[string][]byte{
		"audio.mp3":      []byte("x"),
		"background.jpg": []byte("y"),
	})

	resp, err := http.Get(srv.URL + "/audio/" + url.PathEscape(id))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}
	disp := resp.Header.Get("Content-Disposition")
	if !strings.Contains(disp, "filename*=UTF-8''") {
		t.Errorf("應包含 RFC 5987 filename*，實際 %q", disp)
	}
}

func TestTranscriptDownload_PDF(t *testing.T) {
	srv, cfg, _ := setupMediaDownloadServer(t)
	defer srv.Close()

	content := []byte("%PDF-1.4 fake pdf")
	placeMediaFiles(t, cfg, "pdf-test", map[string][]byte{
		"audio.mp3":      []byte("a"),
		"background.jpg": []byte("b"),
		"transcript.pdf": content,
	})

	resp, err := http.Get(srv.URL + "/transcript/pdf-test")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Content-Type"); got != "application/pdf" {
		t.Errorf("Content-Type = %q", got)
	}
	if got := resp.Header.Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q", got)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != string(content) {
		t.Error("內容不一致")
	}
}

func TestTranscriptDownload_Markdown(t *testing.T) {
	srv, cfg, _ := setupMediaDownloadServer(t)
	defer srv.Close()

	placeMediaFiles(t, cfg, "md-test", map[string][]byte{
		"audio.mp3":      []byte("a"),
		"background.jpg": []byte("b"),
		"transcript.md":  []byte("# hello"),
	})

	resp, err := http.Get(srv.URL + "/transcript/md-test")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/markdown") {
		t.Errorf("Content-Type = %q", resp.Header.Get("Content-Type"))
	}
}

func TestTranscriptDownload_404_NoTranscript(t *testing.T) {
	srv, cfg, _ := setupMediaDownloadServer(t)
	defer srv.Close()

	placeMediaFiles(t, cfg, "no-tx", map[string][]byte{
		"audio.mp3":      []byte("a"),
		"background.jpg": []byte("b"),
	})

	resp, err := http.Get(srv.URL + "/transcript/no-tx")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 404 {
		t.Errorf("want 404, got %d", resp.StatusCode)
	}
}

// TestTranscriptUpload_ReplaceExtension 驗證副檔名替換時舊檔案被清除
func TestTranscriptUpload_ReplaceExtension(t *testing.T) {
	srv, cfg, _ := setupMediaDownloadServer(t)
	defer srv.Close()

	id := "replace-ext"
	// 先建立 composition（音檔 + 背景）
	placeMediaFiles(t, cfg, id, map[string][]byte{
		"audio.mp3":      []byte("a"),
		"background.jpg": []byte("b"),
	})

	uploadTranscript(t, srv.URL, id, "sermon.pdf", []byte("pdf1"))
	if _, err := os.Stat(filepath.Join(cfg.MediaDir, id, "transcript.pdf")); err != nil {
		t.Fatalf("上傳 pdf 後應存在：%v", err)
	}

	uploadTranscript(t, srv.URL, id, "sermon.md", []byte("md1"))
	if _, err := os.Stat(filepath.Join(cfg.MediaDir, id, "transcript.md")); err != nil {
		t.Fatalf("上傳 md 後應存在：%v", err)
	}
	if _, err := os.Stat(filepath.Join(cfg.MediaDir, id, "transcript.pdf")); !os.IsNotExist(err) {
		t.Errorf("替換副檔名後舊 pdf 應被清除，實際 err=%v", err)
	}
}

func TestTranscriptUpload_Delete(t *testing.T) {
	srv, cfg, _ := setupMediaDownloadServer(t)
	defer srv.Close()

	id := "tx-delete"
	placeMediaFiles(t, cfg, id, map[string][]byte{
		"audio.mp3":      []byte("a"),
		"background.jpg": []byte("b"),
		"transcript.pdf": []byte("pdfdata"),
	})

	// POST transcript_delete=1
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	w.WriteField("transcript_delete", "1")
	w.Close()

	req, _ := http.NewRequest("POST", srv.URL+"/upload/"+id, &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}

	if _, err := os.Stat(filepath.Join(cfg.MediaDir, id, "transcript.pdf")); !os.IsNotExist(err) {
		t.Errorf("delete 後 pdf 仍存在")
	}
}

func uploadTranscript(t *testing.T, base, id, filename string, content []byte) {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	fw, _ := w.CreateFormFile("transcript", filename)
	fw.Write(content)
	w.Close()

	req, _ := http.NewRequest("POST", base+"/upload/"+id, &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		io.Copy(io.Discard, resp.Body)
		t.Fatalf("上傳失敗 status=%d", resp.StatusCode)
	}
}
