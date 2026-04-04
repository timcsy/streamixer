package unit

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/timcsy/streamixer/src/composer"
	"github.com/timcsy/streamixer/src/media"
)

func TestPregenManager_StartAndStatus(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := composer.NewPregenManager(tmpDir, 6, 320, 240, 3)

	status := mgr.GetStatus("nonexistent")
	if status != nil {
		t.Error("不存在的素材應回傳 nil 狀態")
	}
}

func TestPregenManager_IsSegmentReady(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := composer.NewPregenManager(tmpDir, 6, 320, 240, 3)

	if mgr.IsSegmentReady("test", 0) {
		t.Error("未生成的分段應回傳 false")
	}

	// 模擬預生成完成
	segDir := filepath.Join(tmpDir, "test")
	os.MkdirAll(segDir, 0755)
	os.WriteFile(filepath.Join(segDir, "seg_000.ts"), []byte("fake segment data"), 0644)

	if !mgr.IsSegmentReady("test", 0) {
		t.Error("已存在的分段應回傳 true")
	}

	// 空檔案應回傳 false
	os.WriteFile(filepath.Join(segDir, "seg_001.ts"), []byte(""), 0644)
	if mgr.IsSegmentReady("test", 1) {
		t.Error("空檔案應回傳 false")
	}
}

func TestPregenManager_Singleflight(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := composer.NewPregenManager(tmpDir, 6, 320, 240, 3)

	comp := &media.MediaComposition{
		ID:         "singleflight-test",
		Audio:      media.Audio{Path: "/nonexistent/audio.mp3"},
		Background: media.Background{Path: "/nonexistent/bg.jpg"},
	}

	// 呼叫多次不應 panic
	for i := 0; i < 5; i++ {
		mgr.StartPregen(comp, 10.0)
	}

	// 等待背景任務啟動
	time.Sleep(200 * time.Millisecond)

	// 應有任務記錄（可能是 running 或 failed，因為檔案不存在）
	status := mgr.GetStatus("singleflight-test")
	if status == nil {
		t.Error("應有任務記錄")
	}
}
