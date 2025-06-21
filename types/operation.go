
package types

import "time"

type OperationType string

const (
	OperationPing OperationType = "ping"
	OperationDNS  OperationType = "dns"
	OperationTCP  OperationType = "tcp"
	OperationHTTP OperationType = "http"
)

type OperationRequest struct {
	Type      OperationType `json:"type"`
	Host      string        `json:"host"`
	Port      int           `json:"port,omitempty"`    // For TCP
	Count     int           `json:"count,omitempty"`   // For ping
	Timeout   int           `json:"timeout,omitempty"` // In seconds
	Query     string        `json:"query,omitempty"`   // For DNS
	URL       string        `json:"url,omitempty"`     // For HTTP
	Method    string        `json:"method,omitempty"`  // For HTTP (GET, POST, etc.)
	ServiceID string        `json:"service_id,omitempty"` // For linking to specific service
}

type OperationResult struct {
	Type        OperationType   `json:"type"`
	Host        string          `json:"host"`
	Port        int             `json:"port,omitempty"`
	Success     bool            `json:"success"`
	ResponseTime time.Duration  `json:"response_time"`
	Error       string          `json:"error,omitempty"`
	Details     string          `json:"details,omitempty"`
	
	// Ping specific fields
	PacketsSent int             `json:"packets_sent,omitempty"`
	PacketsRecv int             `json:"packets_recv,omitempty"`
	PacketLoss  float64         `json:"packet_loss,omitempty"`
	MinRTT      time.Duration   `json:"min_rtt,omitempty"`
	MaxRTT      time.Duration   `json:"max_rtt,omitempty"`
	AvgRTT      time.Duration   `json:"avg_rtt,omitempty"`
	RTTs        []time.Duration `json:"rtts,omitempty"`
	
	// DNS specific fields
	DNSRecords  []string        `json:"dns_records,omitempty"`
	DNSType     string          `json:"dns_type,omitempty"`
	
	// TCP specific fields
	TCPConnected bool           `json:"tcp_connected,omitempty"`
	
	// HTTP specific fields
	HTTPStatusCode int          `json:"http_status_code,omitempty"`
	HTTPMethod     string       `json:"http_method,omitempty"`
	HTTPHeaders    map[string]string `json:"http_headers,omitempty"`
	ContentLength  int64        `json:"content_length,omitempty"`
	ResponseBody   string       `json:"response_body,omitempty"`
	
	StartTime   time.Time       `json:"start_time"`
	EndTime     time.Time       `json:"end_time"`
}
