
package ping

import "time"

type PingResult struct {
	Host        string          `json:"host"`
	Success     bool            `json:"success"`
	PacketsSent int             `json:"packets_sent"`
	PacketsRecv int             `json:"packets_recv"`
	PacketLoss  float64         `json:"packet_loss"`
	MinRTT      time.Duration   `json:"min_rtt"`
	MaxRTT      time.Duration   `json:"max_rtt"`
	AvgRTT      time.Duration   `json:"avg_rtt"`
	RTTs        []time.Duration `json:"rtts"`
	StartTime   time.Time       `json:"start_time"`
	EndTime     time.Time       `json:"end_time"`
	Error       string          `json:"error,omitempty"`
}

type PingRequest struct {
	Host    string `json:"host"`
	Count   int    `json:"count,omitempty"`
	Timeout int    `json:"timeout,omitempty"`
}
