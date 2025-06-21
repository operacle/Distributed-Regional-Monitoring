
package savers

import (
	"fmt"
	"time"

	"service-operation/pocketbase"
	"service-operation/types"
)

func (ms *MetricsSaver) SavePingDataToPocketBase(result *types.OperationResult, serviceID string) {
	// Create a short, professional status message
	var details string
	
	if result.Success {
		// Success message with packet stats
		details = fmt.Sprintf("✅ Ping OK - %d/%d packets received", 
			result.PacketsRecv, result.PacketsSent)
		
		// Add loss percentage if there was any loss
		if result.PacketLoss > 0 {
			details += fmt.Sprintf(" (%.1f%% loss)", result.PacketLoss)
		}
		
		// Add response time info
		details += fmt.Sprintf(" | Avg: %.2fms", 
			float64(result.AvgRTT.Nanoseconds())/1000000)
		
		// Add min/max if different from average (significant variance)
		if result.MinRTT != result.MaxRTT {
			details += fmt.Sprintf(", Min: %.2fms, Max: %.2fms", 
				float64(result.MinRTT.Nanoseconds())/1000000,
				float64(result.MaxRTT.Nanoseconds())/1000000)
		}
	} else {
		// Error message
		if result.PacketLoss >= 100 {
			details = fmt.Sprintf("❌ Ping Failed - 100%% packet loss (%s)", 
				GetShortErrorMessage(result.Error))
		} else {
			details = fmt.Sprintf("⚠️ Ping Partial - %.1f%% packet loss (%s)", 
				result.PacketLoss, GetShortErrorMessage(result.Error))
		}
	}

	pingData := pocketbase.PingDataRecord{
		ServiceID:    serviceID,
		Timestamp:    time.Now(),
		ResponseTime: result.ResponseTime.Milliseconds(),
		Status:       GetStatusString(result.Success),
		PacketsSent:  fmt.Sprintf("%d", result.PacketsSent),
		PacketsRecv:  fmt.Sprintf("%d", result.PacketsRecv),
		PacketLoss:   fmt.Sprintf("%.1f%%", result.PacketLoss),
		MinRTT:       fmt.Sprintf("%.2fms", float64(result.MinRTT.Nanoseconds())/1000000),
		MaxRTT:       fmt.Sprintf("%.2fms", float64(result.MaxRTT.Nanoseconds())/1000000),
		AvgRTT:       fmt.Sprintf("%.2fms", float64(result.AvgRTT.Nanoseconds())/1000000),
		RTTs:         "", // Not currently tracked
		Latency:      fmt.Sprintf("%.2fms", float64(result.AvgRTT.Nanoseconds())/1000000),
		ErrorMessage: result.Error,
		Details:      details, // Short, clean message
		RegionName:   ms.regionName, // Use actual regional info
		AgentID:      ms.agentID,    // Use actual agent ID
	}

	if err := ms.pbClient.SavePingData(pingData); err != nil {
		println("Failed to save ping data to PocketBase:", err.Error())
	}
}

// Method for monitoring service usage
func (ms *MetricsSaver) SavePingDataForService(service pocketbase.Service, result *types.OperationResult) {
	ms.SavePingDataToPocketBase(result, service.ID)
}