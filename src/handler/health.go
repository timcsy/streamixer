package handler

import (
	"encoding/json"
	"net/http"
)

// HealthHandler 回傳服務健康狀態
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
