
package pocketbase

import (
	"fmt"
	"net/http"
	"time"
)

func (c *PocketBaseClient) GetRegionalServices() ([]RegionalService, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/api/collections/regional_service/records", c.baseURL))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		// Collection might require authentication or doesn't exist
		fmt.Printf("Warning: Cannot access regional_service collection (403). Regional monitoring will be disabled.\n")
		return nil, fmt.Errorf("access denied to regional_service collection")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get regional services: status %d", resp.StatusCode)
	}

	var response RegionalServicesResponse
	if err := c.parseResponse(resp, &response); err != nil {
		return nil, err
	}

	return response.Items, nil
}

func (c *PocketBaseClient) GetDefaultRegionalService() (*RegionalService, error) {
	services, err := c.GetRegionalServices()
	if err != nil {
		// Try to create a default regional service if we can't get existing ones
		return c.CreateDefaultRegionalService()
	}

	// Look for a service with region_name "Default"
	for _, service := range services {
		if service.RegionName == "Default" {
			return &service, nil
		}
	}

	// If no default service found, try to create one
	return c.CreateDefaultRegionalService()
}

func (c *PocketBaseClient) CreateDefaultRegionalService() (*RegionalService, error) {
	defaultService := map[string]interface{}{
		"region_name":       "Default",
		"status":           "active",
		"agent_id":         "1",
		"agent_ip_address": "127.0.0.1",
		"connection":       "offline",
		"token":            fmt.Sprintf("default-%d", time.Now().Unix()),
	}

	err := c.createRecord("regional_service", defaultService)
	if err != nil {
		fmt.Printf("Warning: Could not create default regional service: %v\n", err)
		// Return a fallback service for local operations
		return &RegionalService{
			ID:              "default",
			RegionName:      "Default",
			Status:          "active",
			AgentID:         "1",
			AgentIPAddress:  "127.0.0.1",
			Connection:      "offline",
			Token:           "default-token",
		}, nil
	}

	// Try to fetch the created service
	services, err := c.GetRegionalServices()
	if err == nil {
		for _, service := range services {
			if service.RegionName == "Default" {
				return &service, nil
			}
		}
	}

	// Return fallback if still can't get the created service
	return &RegionalService{
		ID:              "default",
		RegionName:      "Default",
		Status:          "active",
		AgentID:         "1",
		AgentIPAddress:  "127.0.0.1",
		Connection:      "offline",
		Token:           "default-token",
	}, nil
}

func (c *PocketBaseClient) UpdateRegionalServiceConnection(serviceID, connection string) error {
	// Determine status based on connection
	status := "inactive"
	if connection == "online" {
		status = "active"
	}

	data := map[string]interface{}{
		"connection": connection,
		"status":     status,
	}

	return c.updateRecord("regional_service", serviceID, data)
}