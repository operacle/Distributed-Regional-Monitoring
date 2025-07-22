
package monitoring

import (
	"log"
	"strings"
	"time"

	"service-operation/operations"
	"service-operation/pocketbase"
	"service-operation/shared/savers"
	"service-operation/types"
)

func (ms *MonitoringService) performCheck(service pocketbase.Service) {
	// First, fetch the latest service status from PocketBase to ensure we have current data
	latestService, err := ms.pbClient.GetService(service.ID)
	if err != nil {
		log.Printf("Failed to fetch latest service status for %s: %v", service.Name, err)
		return
	}

	// Respect the service status - don't check paused services
	if latestService.Status == "paused" {
		return // Silently skip paused services
	}

	// Validate that this service is still assigned to our region and agent (supports comma-separated values)
	if !pocketbase.IsAssignedToRegionAndAgent(*latestService, ms.regionName, ms.agentID) {
		log.Printf("⚠️  Skipping check for %s: no longer assigned to region=%s, agent=%s", 
			latestService.Name, ms.regionName, ms.agentID)
		return
	}

	timeout := 10 * time.Second // Default timeout
	var result *types.OperationResult
	
	serviceType := strings.ToLower(latestService.ServiceType)
	
	// Single log message for check start
	//log.Printf("Checking %s (%s)", latestService.Name, serviceType)
	
	switch serviceType {
	case "ping", "icmp":
		pingOp := operations.NewPingOperation(timeout)
		host := latestService.Host
		if host == "" {
			host = latestService.URL
		}
		result, err = pingOp.Execute(host, 1) // Single ping for monitoring
		
	case "dns":
		dnsOp := operations.NewDNSOperation(timeout)
		host := latestService.Host
		if host == "" {
			host = latestService.Domain
		}
		// Default to A record, but could be made configurable
		queryType := "A"
		result, err = dnsOp.Execute(host, queryType)
		
	case "tcp":
		tcpOp := operations.NewTCPOperation(timeout)
		host := latestService.Host
		if host == "" {
			host = latestService.URL
		}
		port := latestService.Port
		if port <= 0 {
			port = 80 // Default port
		}
		result, err = tcpOp.Execute(host, port)
		
	case "http", "https":
		httpOp := operations.NewHTTPOperation(timeout)
		url := latestService.URL
		if url == "" {
			url = latestService.Host
		}
		result, err = httpOp.Execute(url, "GET")
		
	default:
		log.Printf("Unknown service type: %s for service %s", latestService.ServiceType, latestService.Name)
		return
	}

	// Determine status based on result
	status := "down"
	errorMessage := ""
	responseTime := int64(0)
	
	if err != nil {
		errorMessage = err.Error()
		log.Printf("❌ %s failed: %v", latestService.Name, err)
	} else if result != nil {
		responseTime = result.ResponseTime.Milliseconds()
		if result.Success {
			status = "up"
			log.Printf("✅ %s: %.0fms", latestService.Name, float64(responseTime))
		} else {
			status = "down"
			errorMessage = result.Error
			log.Printf("❌ %s failed: %s", latestService.Name, errorMessage)
		}
	}

	// Only update service status if the service is not paused
	// Check one more time before updating to prevent race conditions
	currentService, err := ms.pbClient.GetService(latestService.ID)
	if err != nil {
		log.Printf("Failed to verify service status before update for %s: %v", latestService.Name, err)
		return
	}
	
	if currentService.Status == "paused" {
		return // Silently skip status update for paused services
	}

	// Final assignment validation before updating status
	if !pocketbase.IsAssignedToRegionAndAgent(*currentService, ms.regionName, ms.agentID) {
		log.Printf("⚠️  Skipping status update for %s: assignment changed during check", latestService.Name)
		return
	}

	// Update service status in PocketBase only if not paused and still assigned
	if err := ms.pbClient.UpdateServiceStatus(latestService.ID, status, responseTime, errorMessage); err != nil {
		log.Printf("Failed to update service status for %s: %v", latestService.Name, err)
	}

	// Save metrics data in ONE place to prevent duplicates
	if result != nil {
		// Get regional information from the monitoring service
		regionName, agentID := ms.GetRegionalInfo()
		metricsSaver := savers.NewMetricsSaverWithRegion(ms.pbClient, regionName, agentID)
		metricsSaver.SaveMetricsForService(*latestService, result)
	}
}