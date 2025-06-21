
package savers

import (
	"fmt"
	"strconv"
	"time"

	"service-operation/pocketbase"
	"service-operation/types"
)

func (ms *MetricsSaver) SaveTCPDataToPocketBase(result *types.OperationResult, serviceID string) {
	// Create a short, professional status message
	var details string
	
	if result.Success && result.TCPConnected {
		// Success message with connection info
		details = fmt.Sprintf("✅ TCP Connection OK - Port %d accessible", result.Port)
		
		// Add response time
		details += fmt.Sprintf(" | Connection time: %.2fms", 
			float64(result.ResponseTime.Nanoseconds())/1000000)
	} else {
		// Error message with port info
		details = fmt.Sprintf("❌ TCP Connection Failed - Port %d unreachable", result.Port)
		
		if result.Error != "" {
			details += fmt.Sprintf(" (%s)", GetShortErrorMessage(result.Error))
		}
		
		// Add response time if available
		if result.ResponseTime > 0 {
			details += fmt.Sprintf(" | Timeout: %.2fms", 
				float64(result.ResponseTime.Nanoseconds())/1000000)
		}
	}

	connectionStatus := "disconnected"
	if result.TCPConnected {
		connectionStatus = "connected"
	}

	tcpData := pocketbase.TCPDataRecord{
		ServiceID:    serviceID,
		Timestamp:    time.Now(),
		ResponseTime: result.ResponseTime.Milliseconds(),
		Status:       GetStatusString(result.Success),
		Connection:   connectionStatus,
		Latency:      fmt.Sprintf("%.2fms", float64(result.ResponseTime.Nanoseconds())/1000000),
		Port:         strconv.Itoa(result.Port),
		ErrorMessage: result.Error,
		Details:      details,
		RegionName:   ms.regionName, // Use actual regional info
		AgentID:      ms.agentID,    // Use actual agent ID
	}

	if err := ms.pbClient.SaveTCPData(tcpData); err != nil {
		fmt.Printf("Failed to save TCP data to PocketBase: %v\n", err)
	}
}

// Method for monitoring service usage
func (ms *MetricsSaver) SaveTCPDataForService(service pocketbase.Service, result *types.OperationResult) {
	ms.SaveTCPDataToPocketBase(result, service.ID)
}