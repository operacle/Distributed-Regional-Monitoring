
package monitoring

import (
	"log"
	"time"

	"service-operation/pocketbase"
)

// Simplified metrics saver for basic service status updates only
func (ms *MonitoringService) saveMetrics(service pocketbase.Service, status string, responseTime int64, errorMessage string) {
	// Save general metrics using the new structure
	metrics := pocketbase.MetricsRecord{
		ServiceName:  service.Name,
		Host:         service.Host,
		Uptime:       0, // This would need to be calculated based on your requirements
		ResponseTime: responseTime,
		LastChecked:  time.Now().Format(time.RFC3339),
		Port:         service.Port,
		ServiceType:  service.ServiceType,
		Status:       status,
		ErrorMessage: errorMessage,
		CheckedAt:    time.Now().Format(time.RFC3339),
	}

	if err := ms.pbClient.SaveMetrics(metrics); err != nil {
		log.Printf("Failed to save metrics for %s: %v", service.Name, err)
	}
}

// This method is now removed to prevent duplicate saves
// The shared savers handle everything in SaveMetricsForService
