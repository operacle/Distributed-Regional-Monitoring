
package savers

import (
	"fmt"
	"time"

	"service-operation/pocketbase"
	"service-operation/types"
)

func (ms *MetricsSaver) SaveUptimeDataToPocketBase(result *types.OperationResult, serviceID string) {
	// Create a short, professional status message
	var details string
	
	if result.Success {
		// Success message with basic info
		details = fmt.Sprintf("✅ HTTP %d OK - Response time: %.2fms", 
			result.HTTPStatusCode, 
			float64(result.ResponseTime.Nanoseconds())/1000000)
		
		// Add content info if available
		if result.ContentLength > 0 {
			details += fmt.Sprintf(" | Content: %s", FormatBytes(result.ContentLength))
		}
		
		// Add server info if available
		if server, exists := result.HTTPHeaders["Server"]; exists {
			details += fmt.Sprintf(" | Server: %s", server)
		}
	} else {
		// Error message with status code if available
		if result.HTTPStatusCode > 0 {
			details = fmt.Sprintf("❌ HTTP %d Error - %s", 
				result.HTTPStatusCode, 
				GetShortErrorMessage(result.Error))
		} else {
			details = fmt.Sprintf("🔌 Connection Error - %s", 
				GetShortErrorMessage(result.Error))
		}
		
		// Add response time if available
		if result.ResponseTime > 0 {
			details += fmt.Sprintf(" | Response time: %.2fms", 
				float64(result.ResponseTime.Nanoseconds())/1000000)
		}
	}

	uptimeData := pocketbase.UptimeDataRecord{
		ServiceID:    serviceID,
		Timestamp:    time.Now(),
		ResponseTime: result.ResponseTime.Milliseconds(),
		Status:       GetStatusString(result.Success),
		Packets:      "N/A", // Not applicable for HTTP
		Latency:      fmt.Sprintf("%.2fms", float64(result.ResponseTime.Nanoseconds())/1000000),
		StatusCodes:  fmt.Sprintf("%d", result.HTTPStatusCode),
		Keyword:      "", // Can be populated later if needed
		ErrorMessage: result.Error,
		Details:      details, // Short, clean message
		Region:       ms.regionName, // Use actual regional info
		RegionID:     ms.agentID,    // Use actual agent ID
		RegionName:   ms.regionName, // Use actual regional info
		AgentID:      ms.agentID,    // Use actual agent ID
	}

	if err := ms.pbClient.SaveUptimeData(uptimeData); err != nil {
		println("Failed to save uptime data to PocketBase:", err.Error())
	}
}

// Method for monitoring service usage
func (ms *MetricsSaver) SaveUptimeDataForService(service pocketbase.Service, result *types.OperationResult) {
	ms.SaveUptimeDataToPocketBase(result, service.ID)
}