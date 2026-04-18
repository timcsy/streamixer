package composer

import (
	"log"
	"sync"
	"time"
)

// Sweeper 背景清掃排程
type Sweeper struct {
	cache    *CacheManager
	mu       sync.RWMutex
	interval time.Duration
	stop     chan struct{}
	kick     chan struct{}
}

// NewSweeper 建立清掃排程
func NewSweeper(cache *CacheManager, interval time.Duration) *Sweeper {
	return &Sweeper{
		cache:    cache,
		interval: interval,
		stop:     make(chan struct{}),
		kick:     make(chan struct{}, 1),
	}
}

// GetInterval 回傳目前清掃頻率
func (s *Sweeper) GetInterval() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.interval
}

// SetInterval 動態更新清掃頻率，下次排程立即採用新值
func (s *Sweeper) SetInterval(d time.Duration) {
	s.mu.Lock()
	s.interval = d
	s.mu.Unlock()
	select {
	case s.kick <- struct{}{}:
	default:
	}
}

// Start 啟動背景清掃 goroutine
func (s *Sweeper) Start() {
	go func() {
		log.Printf("快取清掃排程啟動，頻率 %v", s.GetInterval())
		for {
			timer := time.NewTimer(s.GetInterval())
			select {
			case <-timer.C:
				s.cache.SweepExpired()
				s.cache.SweepByCapacity()
			case <-s.kick:
				// 設定已更新，重新以新頻率起算
				if !timer.Stop() {
					<-timer.C
				}
			case <-s.stop:
				timer.Stop()
				return
			}
		}
	}()
}

// Stop 停止清掃排程
func (s *Sweeper) Stop() {
	close(s.stop)
}
