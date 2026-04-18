package composer

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// CacheEntry 快取條目
type CacheEntry struct {
	CompositionID string
	LastAccessed  time.Time
	Size          int64
	Active        bool // 正在預生成中
}

// CacheManager 管理 tmpfs 中合成分段的快取
type CacheManager struct {
	tmpDir  string
	ttl     time.Duration
	maxSize int64
	mu      sync.RWMutex
	entries map[string]*CacheEntry
}

// GetTTL 取得目前 TTL 設定
func (cm *CacheManager) GetTTL() time.Duration {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.ttl
}

// SetTTL 動態更新 TTL
func (cm *CacheManager) SetTTL(d time.Duration) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.ttl = d
}

// GetMaxSize 取得目前容量上限
func (cm *CacheManager) GetMaxSize() int64 {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.maxSize
}

// SetMaxSize 動態更新容量上限（bytes，0 = 不限制）
func (cm *CacheManager) SetMaxSize(n int64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.maxSize = n
}

// NewCacheManager 建立快取管理器
func NewCacheManager(tmpDir string, ttl time.Duration, maxSize int64) *CacheManager {
	return &CacheManager{
		tmpDir:  tmpDir,
		ttl:     ttl,
		maxSize: maxSize,
		entries: make(map[string]*CacheEntry),
	}
}

// Touch 更新素材的最後存取時間
func (cm *CacheManager) Touch(compositionID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	entry, ok := cm.entries[compositionID]
	if !ok {
		entry = &CacheEntry{CompositionID: compositionID}
		cm.entries[compositionID] = entry
	}
	entry.LastAccessed = time.Now()
}

// SetActive 標記素材為活躍（正在預生成）
func (cm *CacheManager) SetActive(compositionID string, active bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	entry, ok := cm.entries[compositionID]
	if !ok {
		entry = &CacheEntry{CompositionID: compositionID, LastAccessed: time.Now()}
		cm.entries[compositionID] = entry
	}
	entry.Active = active
}

// IsActive 檢查素材是否活躍
func (cm *CacheManager) IsActive(compositionID string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	entry, ok := cm.entries[compositionID]
	return ok && entry.Active
}

// HasCache 檢查素材是否有快取條目
func (cm *CacheManager) HasCache(compositionID string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	_, ok := cm.entries[compositionID]
	return ok
}

// SweepExpired 清除超過 TTL 的非活躍快取
func (cm *CacheManager) SweepExpired() int {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	now := time.Now()
	removed := 0

	for id, entry := range cm.entries {
		if entry.Active {
			continue
		}
		if now.Sub(entry.LastAccessed) > cm.ttl {
			cm.removeEntry(id)
			removed++
		}
	}

	if removed > 0 {
		log.Printf("快取清掃：清除 %d 個過期素材", removed)
	}
	return removed
}

// SweepByCapacity 當容量超過上限的 90% 時，淘汰最久未存取的素材
func (cm *CacheManager) SweepByCapacity() int {
	if cm.maxSize <= 0 {
		return 0
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	totalSize := cm.calculateTotalSize()
	threshold := int64(float64(cm.maxSize) * 0.9)

	if totalSize <= threshold {
		return 0
	}

	// 按 lastAccessed 排序（最舊的優先）
	type entryWithID struct {
		id    string
		entry *CacheEntry
	}
	var candidates []entryWithID
	for id, entry := range cm.entries {
		if !entry.Active {
			candidates = append(candidates, entryWithID{id, entry})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].entry.LastAccessed.Before(candidates[j].entry.LastAccessed)
	})

	removed := 0
	for _, c := range candidates {
		if totalSize <= threshold {
			break
		}
		entrySize := cm.dirSize(c.id)
		cm.removeEntry(c.id)
		totalSize -= entrySize
		removed++
	}

	if removed > 0 {
		log.Printf("快取容量控制：淘汰 %d 個素材，目前使用 %d / %d bytes", removed, totalSize, cm.maxSize)
	}
	return removed
}

// CalculateUsage 計算所有快取的總使用量
func (cm *CacheManager) CalculateUsage() int64 {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.calculateTotalSize()
}

func (cm *CacheManager) calculateTotalSize() int64 {
	var total int64
	entries, err := os.ReadDir(cm.tmpDir)
	if err != nil {
		return 0
	}
	for _, e := range entries {
		if e.IsDir() {
			total += cm.dirSize(e.Name())
		}
	}
	return total
}

func (cm *CacheManager) dirSize(compositionID string) int64 {
	var size int64
	dir := filepath.Join(cm.tmpDir, compositionID)
	filepath.Walk(dir, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

func (cm *CacheManager) removeEntry(compositionID string) {
	dir := filepath.Join(cm.tmpDir, compositionID)
	os.RemoveAll(dir)
	delete(cm.entries, compositionID)
}
