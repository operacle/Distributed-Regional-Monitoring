
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"service-operation/types"
)

func (h *OperationHandler) HandleQuickOperation(w http.ResponseWriter, r *http.Request) {
	opType := r.URL.Query().Get("type")
	host := r.URL.Query().Get("host")
	
	if host == "" || opType == "" {
		http.Error(w, "Type and host parameters are required", http.StatusBadRequest)
		return
	}

	req := types.OperationRequest{
		Type: types.OperationType(opType),
		Host: host,
	}

	// Parse optional parameters
	if countStr := r.URL.Query().Get("count"); countStr != "" {
		if c, err := strconv.Atoi(countStr); err == nil && c > 0 && c <= h.config.MaxCount {
			req.Count = c
		}
	}

	if portStr := r.URL.Query().Get("port"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil && p > 0 && p <= 65535 {
			req.Port = p
		}
	}

	if query := r.URL.Query().Get("query"); query != "" {
		req.Query = query
	}

	if url := r.URL.Query().Get("url"); url != "" {
		req.URL = url
	}

	if method := r.URL.Query().Get("method"); method != "" {
		req.Method = method
	}

	if serviceID := r.URL.Query().Get("service_id"); serviceID != "" {
		req.ServiceID = serviceID
	}

	// Forward to main handler
	reqBody, _ := json.Marshal(req)
	r.Body = http.NoBody
	r.Method = http.MethodPost
	r.Header.Set("Content-Type", "application/json")

	// Create a new request with the body
	newReq := r.Clone(r.Context())
	newReq.Body = http.NoBody
	newReq.ContentLength = int64(len(reqBody))

	h.HandleOperation(w, newReq)
}
