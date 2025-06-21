package monitoring

import (
	"log"
	"time"

	"service-operation/pocketbase"
)

type RegionalMonitor struct {
	pbClient        *pocketbase.PocketBaseClient
	regionalService *pocketbase.RegionalService
	isOnline        bool
	ticker          *time.Ticker
	stopChan        chan bool
}

func NewRegionalMonitor(pbClient *pocketbase.PocketBaseClient) *RegionalMonitor {
	return &RegionalMonitor{
		pbClient: pbClient,
		isOnline: false,
		stopChan: make(chan bool),
	}
}

func NewRegionalMonitorWithService(pbClient *pocketbase.PocketBaseClient, regionalService *pocketbase.RegionalService) *RegionalMonitor {
	return &RegionalMonitor{
		pbClient:        pbClient,
		regionalService: regionalService,
		isOnline:        false,
		stopChan:        make(chan bool),
	}
}

func (rm *RegionalMonitor) Start() {
	// Use the pre-configured regional service
	if rm.regionalService == nil {
		// Fallback to the original behavior if no service provided
		service, err := rm.pbClient.GetDefaultRegionalService()
		if err != nil {
			log.Printf("Warning: Could not get default regional service: %v", err)
			log.Printf("Regional monitoring will continue with fallback values")
			// Continue with fallback service
			service = &pocketbase.RegionalService{
				ID:              "default",
				RegionName:      "Default",
				Status:          "active",
				AgentID:         "1",
				AgentIPAddress:  "127.0.0.1",
				Connection:      "offline",
				Token:           "default-token",
			}
		}
		rm.regionalService = service
	}

	rm.ticker = time.NewTicker(30 * time.Second) // Check connection every 30 seconds

	// Initial connection status update
	rm.updateConnectionStatus("online")
	rm.isOnline = true

	//log.Printf("Regional monitor started for region: %s (Agent ID: %s)", 
		//rm.regionalService.RegionName, rm.regionalService.AgentID)

	go func() {
		for {
			select {
			case <-rm.ticker.C:
				rm.checkConnection()
			case <-rm.stopChan:
				rm.ticker.Stop()
				return
			}
		}
	}()
}

func (rm *RegionalMonitor) Stop() {
	if rm.regionalService != nil {
		rm.updateConnectionStatus("offline")
		log.Printf("Regional monitor stopped for region: %s", rm.regionalService.RegionName)
	}
	
	if rm.ticker != nil {
		rm.ticker.Stop()
	}
	
	rm.stopChan <- true
}

func (rm *RegionalMonitor) checkConnection() {
	// Test PocketBase connection to determine if we're online
	err := rm.pbClient.TestConnection()
	
	if err != nil && rm.isOnline {
		// We were online but now we're offline
		rm.updateConnectionStatus("offline")
		rm.isOnline = false
		log.Printf("Regional agent went offline: %v", err)
	} else if err == nil && !rm.isOnline {
		// We were offline but now we're online
		rm.updateConnectionStatus("online")
		rm.isOnline = true
		log.Printf("Regional agent back online")
	}
}

func (rm *RegionalMonitor) updateConnectionStatus(status string) {
	if rm.regionalService == nil {
		return
	}
	
	if err := rm.pbClient.UpdateRegionalServiceConnection(rm.regionalService.ID, status); err != nil {
		// Don't log errors for update failures - might be due to collection access issues
		// log.Printf("Failed to update regional service connection status: %v", err)
	}
}

func (rm *RegionalMonitor) GetRegionalInfo() (string, string) {
	if rm.regionalService == nil {
		return "Default", "1" // Fallback values
	}
	return rm.regionalService.RegionName, rm.regionalService.AgentID
}