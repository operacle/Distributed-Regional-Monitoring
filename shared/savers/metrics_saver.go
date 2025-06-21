package savers

import (
	"time"

	"service-operation/pocketbase"
	"service-operation/types"
)

type MetricsSaver struct {
	pbClient    *pocketbase.PocketBaseClient
	regionName  string
	agentID     string
}

func NewMetricsSaver(pbClient *pocketbase.PocketBaseClient) *MetricsSaver {
	return &MetricsSaver{
		pbClient:   pbClient,
		regionName: "default", // Default fallback
		agentID:    "1",       // Default fallback
	}
}

func NewMetricsSaverWithRegion(pbClient *pocketbase.PocketBaseClient, regionName, agentID string) *MetricsSaver {
	return &MetricsSaver{
		pbClient:   pbClient,
		regionName: regionName,
		agentID:    agentID,
	}
}

func (ms *MetricsSaver) SaveMetricsToPocketBase(result *types.OperationResult, serviceID string) {
	// Save general metrics using the new structure
	metrics := pocketbase.MetricsRecord{
		ServiceName:  result.Host,
		Host:         result.Host,
		Uptime:       0, // This would need to be calculated based on your requirements
		ResponseTime: result.ResponseTime.Milliseconds(),
		LastChecked:  time.Now().Format(time.RFC3339),
		Port:         result.Port,
		ServiceType:  string(result.Type),
		Status:       GetStatusString(result.Success),
		ErrorMessage: result.Error,
		Details:      FormatResultDetails(result),
		CheckedAt:    time.Now().Format(time.RFC3339),
	}

	if err := ms.pbClient.SaveMetrics(metrics); err != nil {
		// Log error but don't fail the operation
		println("Failed to save metrics to PocketBase:", err.Error())
	}

	// Save detailed data based on operation type - only once per check
	if serviceID != "" {
		switch result.Type {
		case types.OperationPing:
			ms.SavePingDataToPocketBase(result, serviceID)
		case types.OperationHTTP:
			ms.SaveUptimeDataToPocketBase(result, serviceID)
		case types.OperationDNS:
			ms.SaveDNSDataToPocketBase(result, serviceID)
		case types.OperationTCP:
			ms.SaveTCPDataToPocketBase(result, serviceID)
		}
	}
}

// Primary method for monitoring service usage - this prevents duplicates
func (ms *MetricsSaver) SaveMetricsForService(service pocketbase.Service, result *types.OperationResult) {
	// Save general metrics first - reduced logging
	metrics := pocketbase.MetricsRecord{
		ServiceName:  service.Name,
		Host:         service.Host,
		Uptime:       0, // This would need to be calculated based on your requirements
		ResponseTime: result.ResponseTime.Milliseconds(),
		LastChecked:  time.Now().Format(time.RFC3339),
		Port:         service.Port,
		ServiceType:  service.ServiceType,
		Status:       GetStatusString(result.Success),
		ErrorMessage: result.Error,
		Details:      FormatResultDetails(result),
		CheckedAt:    time.Now().Format(time.RFC3339),
	}

	if err := ms.pbClient.SaveMetrics(metrics); err != nil {
		// Silent error - no logging to reduce output
		return
	}

	// Save detailed data based on service type - only once per service with minimal logging
	switch service.ServiceType {
	case "ping", "icmp":
		ms.SavePingDataToPocketBase(result, service.ID)
	case "dns":
		ms.SaveDNSDataToPocketBase(result, service.ID)
	case "http", "https":
		ms.SaveUptimeDataToPocketBase(result, service.ID)
	case "tcp":
		ms.SaveTCPDataToPocketBase(result, service.ID)
	}
}