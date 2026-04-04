package composer

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

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

// IsSegmentReady 檢查分段是否已預生成
func (m *PregenManager) IsSegmentReady(compositionID string, segIndex int) bool {
	outDir := filepath.Join(m.tmpDir, compositionID)
	segPath := filepath.Join(outDir, fmt.Sprintf("seg_%03d.ts", segIndex))

	info, err := os.Stat(segPath)
	if err != nil {
		return false
	}
	return info.Size() > 0
}

// GetSegmentPath 取得預生成分段的路徑
func (m *PregenManager) GetSegmentPath(compositionID string, segIndex int) string {
	return filepath.Join(m.tmpDir, compositionID, fmt.Sprintf("seg_%03d.ts", segIndex))
}

// GetStatus 取得預生成任務狀態
func (m *PregenManager) GetStatus(compositionID string) *PregenTask {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tasks[compositionID]
}
