
package handlers

import (
	"encoding/json"

	"service-operation/types"
)

func getStatusString(success bool) string {
	if success {
		return "up"
	}
	return "down"
}

func formatResultDetails(result *types.OperationResult) string {
	details := map[string]interface{}{
		"type":          result.Type,
		"response_time": result.ResponseTime,
		"start_time":    result.StartTime,
		"end_time":      result.EndTime,
	}

	// Add type-specific details
	switch result.Type {
	case types.OperationPing:
		details["packets_sent"] = result.PacketsSent
		details["packets_recv"] = result.PacketsRecv
		details["packet_loss"] = result.PacketLoss
		details["avg_rtt"] = result.AvgRTT
	case types.OperationHTTP:
		details["status_code"] = result.HTTPStatusCode
		details["method"] = result.HTTPMethod
		details["content_length"] = result.ContentLength
	case types.OperationTCP:
		details["tcp_connected"] = result.TCPConnected
	case types.OperationDNS:
		details["dns_records"] = result.DNSRecords
		details["dns_type"] = result.DNSType
	}

	jsonData, _ := json.Marshal(details)
	return string(jsonData)
}
