
package savers

import (
	"fmt"
	"strings"
	"time"

	"service-operation/pocketbase"
	"service-operation/types"
)

func (ms *MetricsSaver) SaveDNSDataToPocketBase(result *types.OperationResult, serviceID string) {
	// Create a short, professional status message
	var details string
	
	if result.Success {
		// Success message with record count and query info
		recordCount := len(result.DNSRecords)
		details = fmt.Sprintf("✅ DNS %s Query OK - %d records found", 
			strings.ToUpper(result.DNSType), recordCount)
		
		// Add response time
		details += fmt.Sprintf(" | Response time: %.2fms", 
			float64(result.ResponseTime.Nanoseconds())/1000000)
		
		// Add first few records for context
		if recordCount > 0 {
			if recordCount <= 2 {
				details += fmt.Sprintf(" | Records: %s", strings.Join(result.DNSRecords, ", "))
			} else {
				details += fmt.Sprintf(" | Records: %s... (+%d more)", 
					strings.Join(result.DNSRecords[:2], ", "), recordCount-2)
			}
		}
	} else {
		// Error message with query type
		details = fmt.Sprintf("❌ DNS %s Query Failed - %s", 
			strings.ToUpper(result.DNSType), 
			GetShortErrorMessage(result.Error))
		
		// Add response time if available
		if result.ResponseTime > 0 {
			details += fmt.Sprintf(" | Response time: %.2fms", 
				float64(result.ResponseTime.Nanoseconds())/1000000)
		}
	}

	dnsData := pocketbase.DNSDataRecord{
		ServiceID:    serviceID,
		Timestamp:    time.Now(),
		ResponseTime: result.ResponseTime.Milliseconds(),
		Status:       GetStatusString(result.Success),
		QueryType:    result.DNSType,
		ResolveIP:    strings.Join(result.DNSRecords, ","),
		MsgSize:      fmt.Sprintf("%d", len(result.DNSRecords)),
		Question:     result.Host,
		Answer:       strings.Join(result.DNSRecords, ","),
		Authority:    "", // Not available in current implementation
		ErrorMessage: result.Error,
		Details:      details, // Short, clean message
		RegionName:   ms.regionName, // Use actual regional info
		AgentID:      ms.agentID,    // Use actual agent ID
	}

	if err := ms.pbClient.SaveDNSData(dnsData); err != nil {
		println("Failed to save DNS data to PocketBase:", err.Error())
	}
}

// Method for monitoring service usage
func (ms *MetricsSaver) SaveDNSDataForService(service pocketbase.Service, result *types.OperationResult) {
	ms.SaveDNSDataToPocketBase(result, service.ID)
}