
package config

import (
	"fmt"
	"log"
	"time"

	"service-operation/pocketbase"
)

type RegionalConfigManager struct {
	config   *Config
	pbClient *pocketbase.PocketBaseClient
	service  *pocketbase.RegionalService
}

func NewRegionalConfigManager(config *Config, pbClient *pocketbase.PocketBaseClient) *RegionalConfigManager {
	return &RegionalConfigManager{
		config:   config,
		pbClient: pbClient,
	}
}

func (rcm *RegionalConfigManager) LoadOrCreateRegionalService() (*pocketbase.RegionalService, error) {
	// First try to find existing service by agent_id
	services, err := rcm.pbClient.GetRegionalServices()
	if err != nil {
		log.Printf("Warning: Could not get regional services: %v", err)
		return rcm.createFallbackService(), nil
	}

	// Look for existing service with matching agent_id
	for _, service := range services {
		if service.AgentID == rcm.config.AgentID {
			rcm.service = &service
			rcm.updateConfigFromService(&service)
			log.Printf("Loaded existing regional service configuration: Region=%s, Agent=%s", 
				service.RegionName, service.AgentID)
			return &service, nil
		}
	}

	// If no existing service found, create new one
	return rcm.createNewRegionalService()
}

func (rcm *RegionalConfigManager) createNewRegionalService() (*pocketbase.RegionalService, error) {
	log.Printf("Creating new regional service: Region=%s, Agent=%s", 
		rcm.config.RegionName, rcm.config.AgentID)

	token := rcm.config.Token
	if token == "" {
		token = fmt.Sprintf("agent-%s-%d", rcm.config.AgentID, time.Now().Unix())
	}

	serviceData := map[string]interface{}{
		"region_name":       rcm.config.RegionName,
		"status":           "active",
		"agent_id":         rcm.config.AgentID,
		"agent_ip_address": rcm.config.AgentIPAddress,
		"connection":       "offline",
		"token":            token,
	}

	err := rcm.pbClient.CreateRecord("regional_service", serviceData)
	if err != nil {
		log.Printf("Warning: Could not create regional service: %v", err)
		return rcm.createFallbackService(), nil
	}

	// Fetch the created service
	services, err := rcm.pbClient.GetRegionalServices()
	if err == nil {
		for _, service := range services {
			if service.AgentID == rcm.config.AgentID {
				rcm.service = &service
				rcm.updateConfigFromService(&service)
				return &service, nil
			}
		}
	}

	return rcm.createFallbackService(), nil
}

func (rcm *RegionalConfigManager) createFallbackService() *pocketbase.RegionalService {
	token := rcm.config.Token
	if token == "" {
		token = fmt.Sprintf("fallback-%s-%d", rcm.config.AgentID, time.Now().Unix())
	}

	return &pocketbase.RegionalService{
		ID:              fmt.Sprintf("fallback-%s", rcm.config.AgentID),
		RegionName:      rcm.config.RegionName,
		Status:          "active",
		AgentID:         rcm.config.AgentID,
		AgentIPAddress:  rcm.config.AgentIPAddress,
		Connection:      "offline",
		Token:           token,
	}
}

func (rcm *RegionalConfigManager) updateConfigFromService(service *pocketbase.RegionalService) {
	// Update config with values from PocketBase if they exist
	if service.RegionName != "" {
		rcm.config.RegionName = service.RegionName
	}
	if service.AgentIPAddress != "" {
		rcm.config.AgentIPAddress = service.AgentIPAddress
	}
	if service.Token != "" {
		rcm.config.Token = service.Token
	}
}

func (rcm *RegionalConfigManager) GetRegionalInfo() (string, string) {
	if rcm.service != nil {
		return rcm.service.RegionName, rcm.service.AgentID
	}
	return rcm.config.RegionName, rcm.config.AgentID
}

func (rcm *RegionalConfigManager) GetService() *pocketbase.RegionalService {
	return rcm.service
}