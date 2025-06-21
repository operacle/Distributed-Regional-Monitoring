
package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

func (h *OperationHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"service":   "service-operation",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
		"operations": []string{"ping", "dns", "tcp", "http"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}
