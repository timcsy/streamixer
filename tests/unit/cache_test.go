package unit

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/timcsy/streamixer/src/composer"
)

func TestCacheManager_Touch(t *testing.T) {
	tmpDir := t.TempDir()
	cm := composer.NewCacheManager(tmpDir, 10*time.Minute, 0)

	cm.Touch("test-1")

	if !cm.HasCache("test-1") {
		t.Error("Touch 後應有快取條目")
	}
	if !cm.HasCache("test-1") {
		t.Error("HasCache 應回傳 true")
	}
}

func TestCacheManager_SweepExpired(t *testing.T) {
	tmpDir := t.TempDir()
	cm := composer.NewCacheManager(tmpDir, 100*time.Millisecond, 0)

	// 建立測試目錄
	os.MkdirAll(filepath.Join(tmpDir, "old"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "old", "seg_000.m4s"), []byte("data"), 0644)

	cm.Touch("old")

	// 等待 TTL 過期
	time.Sleep(200 * time.Millisecond)

	removed := cm.SweepExpired()
	if removed != 1 {
		t.Errorf("預期清除 1 個，實際 %d 個", removed)
	}

	// 目錄應已刪除
	if _, err := os.Stat(filepath.Join(tmpDir, "old")); !os.IsNotExist(err) {
		t.Error("過期素材的目錄應已刪除")
	}

	if cm.HasCache("old") {
		t.Error("過期素材的快取條目應已移除")
	}
}

func TestCacheManager_SweepExpired_SkipsActive(t *testing.T) {
	tmpDir := t.TempDir()
	cm := composer.NewCacheManager(tmpDir, 100*time.Millisecond, 0)

	os.MkdirAll(filepath.Join(tmpDir, "active"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "active", "seg_000.m4s"), []byte("data"), 0644)

	cm.Touch("active")
	cm.SetActive("active", true)

	time.Sleep(200 * time.Millisecond)

	removed := cm.SweepExpired()
	if removed != 0 {
		t.Errorf("活躍素材不應被清除，但清除了 %d 個", removed)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "active")); os.IsNotExist(err) {
		t.Error("活躍素材的目錄不應被刪除")
	}
}

func TestCacheManager_SweepByCapacity(t *testing.T) {
	tmpDir := t.TempDir()
	// 設定很小的容量上限
	cm := composer.NewCacheManager(tmpDir, 10*time.Minute, 100)

	// 建立兩個素材，每個約 50 bytes → 總共 100 bytes = 100% 容量
	os.MkdirAll(filepath.Join(tmpDir, "old"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "old", "data"), make([]byte, 50), 0644)
	cm.Touch("old")

	time.Sleep(10 * time.Millisecond)

	os.MkdirAll(filepath.Join(tmpDir, "new"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "new", "data"), make([]byte, 50), 0644)
	cm.Touch("new")

	removed := cm.SweepByCapacity()
	if removed < 1 {
		t.Errorf("應至少淘汰 1 個素材，實際 %d 個", removed)
	}

	// old 應被淘汰（較舊），new 應保留
	if _, err := os.Stat(filepath.Join(tmpDir, "old")); !os.IsNotExist(err) {
		t.Error("最舊的素材應被淘汰")
	}
}

func TestCacheManager_SweepByCapacity_SkipsActive(t *testing.T) {
	tmpDir := t.TempDir()
	cm := composer.NewCacheManager(tmpDir, 10*time.Minute, 100)

	os.MkdirAll(filepath.Join(tmpDir, "active"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "active", "data"), make([]byte, 95), 0644)
	cm.Touch("active")
	cm.SetActive("active", true)

	removed := cm.SweepByCapacity()
	if removed != 0 {
		t.Errorf("活躍素材不應被淘汰，但淘汰了 %d 個", removed)
	}
}
