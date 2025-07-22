
package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"service-operation/config"
	"service-operation/handlers"
	"service-operation/monitoring"
	"service-operation/pocketbase"
)

func main() {
	cfg := config.Load()
	
	// Initialize PocketBase client (no credentials required)
	var pbClient *pocketbase.PocketBaseClient
	var monitoringService *monitoring.MonitoringService
	var regionalConfig *config.RegionalConfigManager
	
	if cfg.PocketBaseEnabled {
		var err error
		pbClient, err = pocketbase.NewPocketBaseClient(cfg.PocketBaseURL)
		if err != nil {
			log.Printf("Warning: Failed to initialize PocketBase client: %v", err)
		} else {
			if err := pbClient.TestConnection(); err != nil {
				log.Printf("Warning: PocketBase connection test failed: %v", err)
			} else {
				// Initialize regional configuration manager
				regionalConfig = config.NewRegionalConfigManager(cfg, pbClient)
				
				// Validate regional configuration first
				if err := regionalConfig.ValidateRegionalConfig(); err != nil {
					log.Printf("Error: Invalid regional configuration: %v", err)
					log.Printf("Please set REGION_NAME and AGENT_ID environment variables")
					log.Printf("Monitoring will not start without valid regional configuration")
				} else {
					// Load or create regional service configuration
					regionalService, err := regionalConfig.LoadOrCreateRegionalService()
					if err != nil {
						log.Printf("Error: Failed to setup regional configuration: %v", err)
						log.Printf("Monitoring will not start without valid regional configuration")
					} else {
						// Final validation - ensure all required fields are present
						if regionalService.RegionName == "" || regionalService.AgentID == "" {
							log.Printf("Error: Invalid regional service configuration - region_name: '%s', agent_id: '%s'", 
								regionalService.RegionName, regionalService.AgentID)
							log.Printf("Monitoring will not start without valid regional configuration")
						} else {
							// Initialize and start monitoring service with regional support
							monitoringService = monitoring.NewMonitoringServiceWithRegional(pbClient, regionalService)
							go monitoringService.Start()
							//log.Printf("✅ Regional monitoring started successfully with multi-assignment support")
							//log.Printf("   Region: %s", regionalService.RegionName)
							//log.Printf("   Agent ID: %s", regionalService.AgentID)
							//log.Printf("   Agent IP: %s", regionalService.AgentIPAddress)
							//log.Printf("   Multi-Assignment Support: Services can be assigned using comma-separated values")
							//log.Printf("   Example: region_name='us-east,eu-west' agent_id='agent1,agent2'")
							//log.Printf("   This agent will monitor services where its region AND agent ID appear in the assignments")
						}
					}
				}
			}
		}
	}
	
	handler := handlers.NewOperationHandler(cfg, pbClient)

	router := mux.NewRouter()

	// Main operation endpoint
	router.HandleFunc("/operation", handler.HandleOperation).Methods("POST")
	
	// Quick operation endpoint with query parameters
	router.HandleFunc("/operation/quick", handler.HandleQuickOperation).Methods("GET")
	
	// Legacy ping endpoint for backward compatibility
	router.HandleFunc("/ping", handler.HandleOperation).Methods("POST")
	router.HandleFunc("/ping/quick", handler.HandleQuickOperation).Methods("GET")
	
	// Health check
	router.HandleFunc("/health", handler.HandleHealth).Methods("GET")

	log.Printf(" - Regional Check Agent starting on port %s", cfg.Port)
	if pbClient != nil {
		log.Printf(" - Backenbd integration enabled at %s ", pbClient.GetBaseURL())
	}
	if monitoringService != nil && regionalConfig != nil {
		regionName, agentID := regionalConfig.GetRegionalInfo()
		log.Printf("🎯 Regional monitoring active: Region=%s, Agent=%s", regionName, agentID)
		//log.Printf("⚡ Service filtering: Supports comma-separated multi-region/multi-agent assignments")
		//log.Printf("💡 Service Assignment Examples:")
		//log.Printf("   - Single: region_name='%s' agent_id='%s'", regionName, agentID)
		//log.Printf("   - Multi: region_name='us-east,%s,eu-west' agent_id='agent1,%s,agent3'", regionName, agentID)
	} else {
		//log.Printf("❌ Regional monitoring disabled - configuration validation failed")
		//log.Printf("💡 Set REGION_NAME and AGENT_ID environment variables to enable monitoring")
	}
	
	//log.Printf("🔗 Available endpoints:")
	//log.Printf("  POST /operation - Full operation test (ping, dns, tcp, http)")
	//log.Printf("  GET  /operation/quick?type=<type>&host=<host> - Quick operation test")
	//log.Printf("  POST /ping - Legacy ping endpoint")
	//log.Printf("  GET  /ping/quick?host=<host> - Legacy quick ping test")
	//log.Printf("  GET  /health - Health check")
	//log.Printf("📋 Supported operations: ping, dns, tcp, http")

	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-c
		log.Println("🛑 Shutting down monitoring service...")
		if monitoringService != nil {
			monitoringService.Stop()
		}
		log.Println("✅ Regional Check Agent stopped")
		os.Exit(0)
	}()

	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}