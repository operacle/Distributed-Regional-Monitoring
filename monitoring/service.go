
package monitoring

import (
	"log"
	"sync"
	"time"

	"service-operation/pocketbase"
)

type MonitoringService struct {
	pbClient        *pocketbase.PocketBaseClient
	activeServices  map[string]*ServiceMonitor
	regionalMonitor *RegionalMonitor
	mu              sync.RWMutex
	stopChan        chan bool
	isRunning       bool
	regionName      string
	agentID         string
}

func NewMonitoringService(pbClient *pocketbase.PocketBaseClient) *MonitoringService {
	return &MonitoringService{
		pbClient:        pbClient,
		activeServices:  make(map[string]*ServiceMonitor),
		regionalMonitor: NewRegionalMonitor(pbClient),
		stopChan:        make(chan bool),
		isRunning:       false,
	}
}

func NewMonitoringServiceWithRegional(pbClient *pocketbase.PocketBaseClient, regionalService *pocketbase.RegionalService) *MonitoringService {
	return &MonitoringService{
		pbClient:        pbClient,
		activeServices:  make(map[string]*ServiceMonitor),
		regionalMonitor: NewRegionalMonitorWithService(pbClient, regionalService),
		stopChan:        make(chan bool),
		isRunning:       false,
		regionName:      regionalService.RegionName,
		agentID:         regionalService.AgentID,
	}
}

func (ms *MonitoringService) Start() {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.isRunning {
		log.Println("⚠️  Monitoring service is already running")
		return
	}

	// Critical validation: Ensure we have valid regional configuration
	if ms.regionName == "" || ms.agentID == "" {
		log.Printf("❌ Cannot start monitoring: Invalid regional configuration")
		log.Printf("   Region Name: '%s'", ms.regionName)
		log.Printf("   Agent ID: '%s'", ms.agentID)
		log.Printf("   Both values must be non-empty to start monitoring")
		return
	}

	ms.isRunning = true
	//log.Printf("🚀 Starting regional monitoring service")
	//log.Printf("   Assigned Region: %s", ms.regionName)
	//log.Printf("   Assigned Agent ID: %s", ms.agentID)
	//log.Printf("   Filter Mode: Only services with matching region_name AND agent_id will be monitored")

	// Start regional monitoring (connection status tracking)
	ms.regionalMonitor.Start()

	// Start the main monitoring loop
	go ms.monitoringLoop()
}

func (ms *MonitoringService) Stop() {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if !ms.isRunning {
		return
	}

	log.Println("🛑 Stopping monitoring service...")
	ms.isRunning = false
	
	// Stop regional monitoring
	ms.regionalMonitor.Stop()
	
	// Stop all active monitors
	for serviceID, monitor := range ms.activeServices {
		ms.stopMonitor(serviceID, monitor)
	}

	ms.stopChan <- true
}

func (ms *MonitoringService) GetRegionalInfo() (string, string) {
	return ms.regionalMonitor.GetRegionalInfo()
}

func (ms *MonitoringService) monitoringLoop() {
	ticker := time.NewTicker(30 * time.Second) // Check for assigned services every 30 seconds
	defer ticker.Stop()

	// Initial load of assigned services
	ms.loadAndStartAssignedServices()

	for {
		select {
		case <-ticker.C:
			ms.loadAndStartAssignedServices()
		case <-ms.stopChan:
			return
		}
	}
}

func (ms *MonitoringService) loadAndStartAssignedServices() {
	// Critical: Only get services specifically assigned to this agent's region and ID
	services, err := ms.pbClient.GetAssignedServices(ms.regionName, ms.agentID)
	if err != nil {
		log.Printf("❌ Failed to load assigned services for region='%s', agent='%s': %v", 
			ms.regionName, ms.agentID, err)
		return
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Track which services are currently assigned to this specific agent
	assignedServiceIDs := make(map[string]bool)
	newServicesCount := 0
	
	for _, service := range services {
		// Double-check: Ensure service really matches our configuration
		if service.RegionName != ms.regionName || service.AgentID != ms.agentID {
			log.Printf("⚠️  Skipping service %s: region/agent mismatch (service: %s/%s, agent: %s/%s)", 
				service.Name, service.RegionName, service.AgentID, ms.regionName, ms.agentID)
			continue
		}
		
		assignedServiceIDs[service.ID] = true
		
		// Start monitoring if not already active
		if _, exists := ms.activeServices[service.ID]; !exists {
			log.Printf("🎯 Starting monitoring: %s (%s) - Region: %s, Agent: %s", 
				service.Name, service.ServiceType, ms.regionName, ms.agentID)
			ms.startMonitor(service)
			newServicesCount++
		}
	}

	// Stop monitoring for services no longer assigned to this agent
	stoppedServicesCount := 0
	for serviceID, monitor := range ms.activeServices {
		if !assignedServiceIDs[serviceID] {
			log.Printf("🛑 Stopping monitoring: service %s (no longer assigned to region=%s, agent=%s)", 
				serviceID, ms.regionName, ms.agentID)
			ms.stopMonitor(serviceID, monitor)
			stoppedServicesCount++
		}
	}

	// Status summary
// 	totalAssigned := len(services)
// 	totalActive := len(ms.activeServices)
	
// 	if totalAssigned == 0 {
// 		log.Printf("📋 No services assigned to region='%s', agent='%s'", ms.regionName, ms.agentID)
		//log.Printf("💡 Assign services to this agent in PocketBase to start monitoring")
// 	} else {
// 		log.Printf("📊 Monitoring Status: %d services assigned, %d actively monitored", 
// 			totalAssigned, totalActive)
// 		if newServicesCount > 0 {
// 		//	log.Printf("   ✅ Started monitoring %d new services", newServicesCount)
// 		}
// 		if stoppedServicesCount > 0 {
// 			log.Printf("   🛑 Stopped monitoring %d unassigned services", stoppedServicesCount)
// 		}
// 	}
}