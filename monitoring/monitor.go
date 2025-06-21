
package monitoring

import (
	"log"
	"time"

	"service-operation/pocketbase"
)

type ServiceMonitor struct {
	service  pocketbase.Service
	ticker   *time.Ticker
	stopChan chan bool
}

func (ms *MonitoringService) startMonitor(service pocketbase.Service) {
	if service.HeartbeatInterval <= 0 {
		service.HeartbeatInterval = 60 // Default to 60 seconds
	}

	monitor := &ServiceMonitor{
		service:  service,
		ticker:   time.NewTicker(time.Duration(service.HeartbeatInterval) * time.Second),
		stopChan: make(chan bool),
	}

	ms.activeServices[service.ID] = monitor

	//log.Printf("Starting monitor for service: %s (%s)", service.Name, service.ServiceType)

	go func() {
		// Perform initial check
		ms.performCheck(service)
		
		for {
			select {
			case <-monitor.ticker.C:
				ms.performCheck(service)
			case <-monitor.stopChan:
				monitor.ticker.Stop()
				return
			}
		}
	}()
}

func (ms *MonitoringService) stopMonitor(serviceID string, monitor *ServiceMonitor) {
	log.Printf("Stopping monitor for service: %s", serviceID)
	monitor.stopChan <- true
	delete(ms.activeServices, serviceID)
}
