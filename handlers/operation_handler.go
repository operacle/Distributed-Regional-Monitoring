
package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"service-operation/operations"
	"service-operation/types"
)

func (h *OperationHandler) HandleOperation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req types.OperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if req.Host == "" && req.URL == "" {
		http.Error(w, "Host or URL is required", http.StatusBadRequest)
		return
	}

	// Set defaults
	if req.Count <= 0 {
		req.Count = h.config.DefaultCount
	}
	if req.Count > h.config.MaxCount {
		req.Count = h.config.MaxCount
	}

	if req.Timeout <= 0 {
		req.Timeout = int(h.config.DefaultTimeout.Seconds())
	}
	if time.Duration(req.Timeout)*time.Second > h.config.MaxTimeout {
		req.Timeout = int(h.config.MaxTimeout.Seconds())
	}

	timeout := time.Duration(req.Timeout) * time.Second
	var result *types.OperationResult
	var err error

	switch req.Type {
	case types.OperationPing:
		pingOp := operations.NewPingOperation(timeout)
		result, err = pingOp.Execute(req.Host, req.Count)
		
	case types.OperationDNS:
		dnsOp := operations.NewDNSOperation(timeout)
		query := req.Query
		if query == "" {
			query = "A"
		}
		result, err = dnsOp.Execute(req.Host, query)
		
	case types.OperationTCP:
		if req.Port <= 0 {
			http.Error(w, "Port is required for TCP operations", http.StatusBadRequest)
			return
		}
		tcpOp := operations.NewTCPOperation(timeout)
		result, err = tcpOp.Execute(req.Host, req.Port)
		
	case types.OperationHTTP:
		httpOp := operations.NewHTTPOperation(timeout)
		url := req.URL
		if url == "" {
			url = req.Host
		}
		method := req.Method
		if method == "" {
			method = "GET"
		}
		result, err = httpOp.Execute(url, method)
		
	default:
		http.Error(w, "Invalid operation type", http.StatusBadRequest)
		return
	}

	if err != nil {
		result = &types.OperationResult{
			Type:    req.Type,
			Host:    req.Host,
			Port:    req.Port,
			Success: false,
			Error:   err.Error(),
		}
	}

	// Save metrics to PocketBase if available
	if h.pbClient != nil && h.pbClient.IsAuthenticated() {
		go h.saveMetricsToPocketBase(result, req.ServiceID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
