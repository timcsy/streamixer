package composer

import (
	"log"
	"time"
)

// Sweeper 背景清掃排程
type Sweeper struct {
	cache    *CacheManager
	interval time.Duration
	stop     chan struct{}
}

// NewSweeper 建立清掃排程
func NewSweeper(cache *CacheManager, interval time.Duration) *Sweeper {
	return &Sweeper{
		cache:    cache,
		interval: interval,
		stop:     make(chan struct{}),
	}
}

// Start 啟動背景清掃 goroutine
func (s *Sweeper) Start() {
	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		log.Printf("快取清掃排程啟動，頻率 %v", s.interval)

		for {
			select {
			case <-ticker.C:
				s.cache.SweepExpired()
				s.cache.SweepByCapacity()
			case <-s.stop:
				return
			}
		}
	}()
}

// Stop 停止清掃排程
func (s *Sweeper) Stop() {
	close(s.stop)
}
