
package config

import (
	"fmt"
	"log"
)

// ValidateRegionalConfig validates that required regional configuration is present
func (rcm *RegionalConfigManager) ValidateRegionalConfig() error {
	if rcm.config.RegionName == "" {
		return fmt.Errorf("REGION_NAME is required but not set")
	}
	
	if rcm.config.AgentID == "" {
		return fmt.Errorf("AGENT_ID is required but not set")
	}
	
	if rcm.config.AgentIPAddress == "" {
		log.Printf("Warning: AGENT_IP_ADDRESS not set, using default: 127.0.0.1")
		rcm.config.AgentIPAddress = "127.0.0.1"
	}
	
	return nil
}

// IsConfigurationValid checks if the current configuration is valid for monitoring
func (rcm *RegionalConfigManager) IsConfigurationValid() bool {
	return rcm.config.RegionName != "" && 
		   rcm.config.AgentID != "" && 
		   rcm.config.AgentIPAddress != ""
}

// GetConfigurationSummary returns a summary of the current configuration
func (rcm *RegionalConfigManager) GetConfigurationSummary() string {
	return fmt.Sprintf("Region: %s, Agent: %s, IP: %s", 
		rcm.config.RegionName, 
		rcm.config.AgentID, 
		rcm.config.AgentIPAddress)
}