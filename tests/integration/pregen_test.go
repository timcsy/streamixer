package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/timcsy/streamixer/src/composer"
	"github.com/timcsy/streamixer/src/media"
)

func setupPregenTestMedia(t *testing.T) (string, *media.MediaComposition) {
	t.Helper()

	tmpDir := t.TempDir()
	mediaDir := filepath.Join(tmpDir, "media", "pregen-test")
	os.MkdirAll(mediaDir, 0755)

	// 產生 18 秒測試音檔（3 個 6 秒分段）
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
		ID:         "pregen-test",
		Audio:      media.Audio{Path: audioPath, Format: "mp3"},
		Background: media.Background{Path: bgPath, Format: "jpg"},
	}

	return tmpDir, comp
}

func TestPregenConcurrentMultipleCompositions(t *testing.T) {
	tmpDir := t.TempDir()
	outDir := filepath.Join(tmpDir, "output")

	mgr := composer.NewPregenManager(outDir, 6, 320, 240, 3)

	// 建立 3 個不同素材
	comps := make([]*media.MediaComposition, 3)
	for i := 0; i < 3; i++ {
		mediaDir := filepath.Join(tmpDir, "media", fmt.Sprintf("comp-%d", i))
		os.MkdirAll(mediaDir, 0755)

		audioPath := filepath.Join(mediaDir, "audio.mp3")
		cmd := exec.Command("ffmpeg", "-y", "-f", "lavfi", "-i",
			fmt.Sprintf("sine=frequency=%d:duration=6", 440+i*220), "-q:a", "9", audioPath)
		if err := cmd.Run(); err != nil {
			t.Skipf("需要 ffmpeg：%v", err)
		}

		bgPath := filepath.Join(mediaDir, "background.jpg")
		cmd = exec.Command("ffmpeg", "-y", "-f", "lavfi", "-i",
			"color=c=blue:s=320x240:d=1", "-frames:v", "1", bgPath)
		if err := cmd.Run(); err != nil {
			t.Skipf("需要 ffmpeg：%v", err)
		}

		comps[i] = &media.MediaComposition{
			ID:         fmt.Sprintf("comp-%d", i),
			Audio:      media.Audio{Path: audioPath, Format: "mp3"},
			Background: media.Background{Path: bgPath, Format: "jpg"},
		}
	}

	// 同時啟動 3 個預生成
	for _, comp := range comps {
		mgr.StartPregen(comp, 6.0)
	}

	// 等待全部完成
	deadline := time.After(30 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("並發預生成超時")
		default:
			allDone := true
			for _, comp := range comps {
				status := mgr.GetStatus(comp.ID)
				if status == nil || status.Status == composer.PregenRunning || status.Status == composer.PregenPending {
					allDone = false
					break
				}
			}
			if allDone {
				// 驗證全部有分段
				for _, comp := range comps {
					if !mgr.IsSegmentReady(comp.ID, 0) {
						t.Errorf("%s 的分段 0 應已就緒", comp.ID)
					}
				}
				return
			}
			time.Sleep(200 * time.Millisecond)
		}
	}
}

func TestPregenFailureFallback(t *testing.T) {
	tmpDir := t.TempDir()
	outDir := filepath.Join(tmpDir, "output")

	mgr := composer.NewPregenManager(outDir, 6, 320, 240, 3)

	// 用不存在的檔案觸發預生成失敗
	comp := &media.MediaComposition{
		ID:         "fail-test",
		Audio:      media.Audio{Path: "/nonexistent/audio.mp3"},
		Background: media.Background{Path: "/nonexistent/bg.jpg"},
	}

	mgr.StartPregen(comp, 10.0)
	time.Sleep(500 * time.Millisecond)

	status := mgr.GetStatus("fail-test")
	if status == nil {
		t.Fatal("應有任務記錄")
	}
	if status.Status != composer.PregenFailed {
		t.Errorf("預期狀態 Failed，實際 %d", status.Status)
	}

	// 失敗後 IsSegmentReady 應回傳 false，handler 會回退為按需生成
	if mgr.IsSegmentReady("fail-test", 0) {
		t.Error("失敗的預生成不應有就緒的分段")
	}
}

func TestPregenFullWorkflow(t *testing.T) {
	tmpDir, comp := setupPregenTestMedia(t)
	outDir := filepath.Join(tmpDir, "output")

	mgr := composer.NewPregenManager(outDir, 6, 320, 240, 3)

	duration, err := composer.ProbeDuration(comp.Audio.Path)
	if err != nil {
		t.Fatalf("取得音檔長度失敗：%v", err)
	}

	// 啟動預生成
	mgr.StartPregen(comp, duration)

	// 等待預生成完成（18 秒音檔應在幾秒內完成）
	deadline := time.After(30 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("預生成超時")
		default:
			status := mgr.GetStatus(comp.ID)
			if status != nil && status.Status == composer.PregenCompleted {
				// 至少第一個分段必須就緒
				if !mgr.IsSegmentReady(comp.ID, 0) {
					t.Error("分段 0 應已就緒")
				}
				// 至少應有 2 個分段（18 秒 / 6 秒目標）
				if !mgr.IsSegmentReady(comp.ID, 1) {
					t.Error("分段 1 應已就緒")
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
