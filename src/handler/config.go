package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/timcsy/streamixer/src/composer"
)

// ConfigHandler 提供執行階段的快取設定讀寫
type ConfigHandler struct {
	cache   *composer.CacheManager
	sweeper *composer.Sweeper
}

// NewConfigHandler 建立設定 handler
func NewConfigHandler(cache *composer.CacheManager, sweeper *composer.Sweeper) *ConfigHandler {
	return &ConfigHandler{cache: cache, sweeper: sweeper}
}

type configResponse struct {
	CacheTTL           string `json:"cache_ttl"`            // Go time.Duration 字串，例如 "30m"
	CacheMaxSize       int64  `json:"cache_max_size"`       // bytes，0 = 不限制
	CacheSweepInterval string `json:"cache_sweep_interval"` // 例如 "5m"
	CacheUsageBytes    int64  `json:"cache_usage_bytes"`    // 目前使用量
}

type configRequest struct {
	CacheTTL           *string `json:"cache_ttl,omitempty"`
	CacheMaxSize       *int64  `json:"cache_max_size,omitempty"`
	CacheSweepInterval *string `json:"cache_sweep_interval,omitempty"`
}

// Get 回傳目前快取設定
func (h *ConfigHandler) Get(w http.ResponseWriter, r *http.Request) {
	resp := configResponse{
		CacheTTL:           h.cache.GetTTL().String(),
		CacheMaxSize:       h.cache.GetMaxSize(),
		CacheSweepInterval: h.sweeper.GetInterval().String(),
		CacheUsageBytes:    h.cache.CalculateUsage(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Put 更新快取設定（任何欄位可省略，僅更新有傳的部分）
func (h *ConfigHandler) Put(w http.ResponseWriter, r *http.Request) {
	var req configRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON 解析失敗："+err.Error())
		return
	}

	if req.CacheTTL != nil {
		d, err := time.ParseDuration(*req.CacheTTL)
		if err != nil || d < time.Second {
			writeError(w, http.StatusBadRequest, "cache_ttl 格式錯誤（例：30m、1h，至少 1s）")
			return
		}
		h.cache.SetTTL(d)
	}
	if req.CacheMaxSize != nil {
		if *req.CacheMaxSize < 0 {
			writeError(w, http.StatusBadRequest, "cache_max_size 不可為負")
			return
		}
		h.cache.SetMaxSize(*req.CacheMaxSize)
	}
	if req.CacheSweepInterval != nil {
		d, err := time.ParseDuration(*req.CacheSweepInterval)
		if err != nil || d < 10*time.Second {
			writeError(w, http.StatusBadRequest, "cache_sweep_interval 格式錯誤（至少 10s）")
			return
		}
		h.sweeper.SetInterval(d)
	}

	// 回傳最新設定
	h.Get(w, r)
}
