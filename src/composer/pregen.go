package composer

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/timcsy/streamixer/src/media"
	"golang.org/x/sync/singleflight"
)

// PregenStatus 預生成任務狀態
type PregenStatus int

const (
	PregenPending   PregenStatus = iota
	PregenRunning
	PregenCompleted
	PregenFailed
)

// PregenTask 預生成任務
type PregenTask struct {
	CompositionID string
	Status        PregenStatus
	TotalSegments int
	OutputDir     string
	Error         error
}

// PregenManager 管理背景預生成任務
type PregenManager struct {
	tmpDir          string
	segmentDuration int
	outputWidth     int
	outputHeight    int

	mu    sync.RWMutex
	tasks map[string]*PregenTask
	group singleflight.Group

	// 並發限制
	sem chan struct{}
}

// NewPregenManager 建立預生成管理器
func NewPregenManager(tmpDir string, segDuration, width, height, maxConcurrent int) *PregenManager {
	if maxConcurrent <= 0 {
		maxConcurrent = 3
	}
	return &PregenManager{
		tmpDir:          tmpDir,
		segmentDuration: segDuration,
		outputWidth:     width,
		outputHeight:    height,
		tasks:           make(map[string]*PregenTask),
		sem:             make(chan struct{}, maxConcurrent),
	}
}

// StartPregen 啟動背景預生成（singleflight 防止重複）
func (m *PregenManager) StartPregen(comp *media.MediaComposition, duration float64) {
	m.group.Do(comp.ID, func() (interface{}, error) {
		totalSegs := SegmentCount(duration, m.segmentDuration)
		outDir := filepath.Join(m.tmpDir, comp.ID)

		task := &PregenTask{
			CompositionID: comp.ID,
			Status:        PregenPending,
			TotalSegments: totalSegs,
			OutputDir:     outDir,
		}

		m.mu.Lock()
		// 如果已經完成或正在執行，不重複啟動
		if existing, ok := m.tasks[comp.ID]; ok {
			if existing.Status == PregenCompleted || existing.Status == PregenRunning {
				m.mu.Unlock()
				return nil, nil
			}
		}
		m.tasks[comp.ID] = task
		m.mu.Unlock()

		go m.runPregen(comp, task)
		return nil, nil
	})
}

func (m *PregenManager) runPregen(comp *media.MediaComposition, task *PregenTask) {
	// 取得 semaphore
	m.sem <- struct{}{}
	defer func() { <-m.sem }()

	m.mu.Lock()
	task.Status = PregenRunning
	m.mu.Unlock()

	os.MkdirAll(task.OutputDir, 0755)

	args := BuildFFmpegArgs(comp, task.OutputDir, m.segmentDuration, m.outputWidth, m.outputHeight)
	cmd := exec.Command("ffmpeg", args...)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		m.mu.Lock()
		task.Status = PregenFailed
		task.Error = fmt.Errorf("預生成失敗：%w", err)
		m.mu.Unlock()
		log.Printf("預生成 %s 失敗：%v", comp.ID, err)
		return
	}

	m.mu.Lock()
	task.Status = PregenCompleted
	m.mu.Unlock()
	log.Printf("預生成 %s 完成（%d 分段）", comp.ID, task.TotalSegments)
}

// IsSegmentReady 檢查 fMP4 分段是否已預生成
func (m *PregenManager) IsSegmentReady(compositionID string, segIndex int) bool {
	outDir := filepath.Join(m.tmpDir, compositionID)
	segPath := filepath.Join(outDir, fmt.Sprintf("seg_%03d.m4s", segIndex))

	info, err := os.Stat(segPath)
	if err != nil {
		return false
	}
	return info.Size() > 0
}

// GetSegmentPath 取得預生成 fMP4 分段的路徑
func (m *PregenManager) GetSegmentPath(compositionID string, segIndex int) string {
	return filepath.Join(m.tmpDir, compositionID, fmt.Sprintf("seg_%03d.m4s", segIndex))
}

// IsInitReady 檢查 init.mp4 是否已存在
func (m *PregenManager) IsInitReady(compositionID string) bool {
	initPath := filepath.Join(m.tmpDir, compositionID, "init.mp4")
	info, err := os.Stat(initPath)
	if err != nil {
		return false
	}
	return info.Size() > 0
}

// GetInitPath 取得 init.mp4 的路徑
func (m *PregenManager) GetInitPath(compositionID string) string {
	return filepath.Join(m.tmpDir, compositionID, "init.mp4")
}

// GetPlaylistPath 取得 FFmpeg 產生的 playlist 路徑
func (m *PregenManager) GetPlaylistPath(compositionID string) string {
	return filepath.Join(m.tmpDir, compositionID, "index.m3u8")
}

// WaitForPlaylist 等待預生成產生 playlist（至少有一個分段），最多等待 timeoutSec 秒
func (m *PregenManager) WaitForPlaylist(compositionID string, timeoutSec int) error {
	playlistPath := m.GetPlaylistPath(compositionID)

	for i := 0; i < timeoutSec*10; i++ {
		data, err := os.ReadFile(playlistPath)
		if err == nil {
			content := string(data)
			// EVENT 模式下，FFmpeg 邊合成邊寫入 playlist
			// 只要有至少一個 .m4s 分段就可以開始播放
			if strings.Contains(content, ".m4s") {
				return nil
			}
		}

		// 檢查是否已失敗
		m.mu.RLock()
		task := m.tasks[compositionID]
		m.mu.RUnlock()
		if task != nil && task.Status == PregenFailed {
			return task.Error
		}

		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("等待 playlist 逾時（%d 秒）", timeoutSec)
}

// WaitForInit 等待 init.mp4 就緒，最多等待 timeoutSec 秒
func (m *PregenManager) WaitForInit(compositionID string, timeoutSec int) error {
	for i := 0; i < timeoutSec*10; i++ {
		if m.IsInitReady(compositionID) {
			return nil
		}

		m.mu.RLock()
		task := m.tasks[compositionID]
		m.mu.RUnlock()
		if task != nil && task.Status == PregenFailed {
			return task.Error
		}

		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("等待 init.mp4 逾時（%d 秒）", timeoutSec)
}

// WaitForSegment 等待指定分段就緒，最多等待 timeoutSec 秒
func (m *PregenManager) WaitForSegment(compositionID string, segIndex int, timeoutSec int) error {
	for i := 0; i < timeoutSec*10; i++ {
		if m.IsSegmentReady(compositionID, segIndex) {
			return nil
		}

		m.mu.RLock()
		task := m.tasks[compositionID]
		m.mu.RUnlock()
		if task != nil && task.Status == PregenFailed {
			return task.Error
		}

		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("等待分段 %d 逾時（%d 秒）", segIndex, timeoutSec)
}

// IsPlaylistComplete 檢查 playlist 是否已完成（包含 EXT-X-ENDLIST）
func (m *PregenManager) IsPlaylistComplete(compositionID string) bool {
	data, err := os.ReadFile(m.GetPlaylistPath(compositionID))
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "#EXT-X-ENDLIST")
}

// GetStatus 取得預生成任務狀態
func (m *PregenManager) GetStatus(compositionID string) *PregenTask {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tasks[compositionID]
}
