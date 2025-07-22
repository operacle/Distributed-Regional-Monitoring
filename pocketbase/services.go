package pocketbase

import (
	"fmt"
	"net/http"
	"strings"
)

// Helper function to split comma-separated values and trim whitespace
func SplitCommaValues(value string) []string {
	if value == "" {
		return []string{}
	}
	
	parts := strings.Split(value, ",")
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// Helper function to check if a value exists in comma-separated string
func ContainsValue(commaString, targetValue string) bool {
	values := SplitCommaValues(commaString)
	for _, value := range values {
		if value == targetValue {
			return true
		}
	}
	return false
}

// Helper function to check if service is assigned to specific region and agent
func IsAssignedToRegionAndAgent(service Service, regionName, agentID string) bool {
	// Check if the service's region_name contains our region
	regionMatch := ContainsValue(service.RegionName, regionName)
	
	// Check if the service's agent_id contains our agent ID
	agentMatch := ContainsValue(service.AgentID, agentID)
	
	return regionMatch && agentMatch
}

// GetActiveServices retrieves all active services from PocketBase
func (c *PocketBaseClient) GetActiveServices() ([]Service, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/api/collections/services/records", c.baseURL))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get services: status %d", resp.StatusCode)
	}

	var response ServicesResponse
	if err := c.parseResponse(resp, &response); err != nil {
		return nil, err
	}

	return response.Items, nil
}

// GetAssignedServices retrieves all services assigned to specific region and agent
// Now supports comma-separated values in region_name and agent_id fields
func (c *PocketBaseClient) GetAssignedServices(regionName, agentID string) ([]Service, error) {
	var allServices []Service
	var assignedServices []Service
	page := 1
	perPage := 30 // Use default pagination size

	// Get all services first (we'll filter on the client side for comma-separated values)
	for {
		// Fetch services page by page without filtering (we'll filter locally)
		requestURL := fmt.Sprintf("%s/api/collections/services/records?page=%d&perPage=%d&filter=status!='paused'", 
			c.baseURL, page, perPage)
		
		resp, err := c.httpClient.Get(requestURL)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to get assigned services: status %d", resp.StatusCode)
		}

		var response ServicesResponse
		if err := c.parseResponse(resp, &response); err != nil {
			return nil, err
		}

		// Add current page items to the result
		allServices = append(allServices, response.Items...)

		// Check if we've fetched all pages
		if page >= response.TotalPages || len(response.Items) == 0 {
			break
		}

		page++
	}

	// Filter services locally to support comma-separated values
	for _, service := range allServices {
		if IsAssignedToRegionAndAgent(service, regionName, agentID) {
			assignedServices = append(assignedServices, service)
		}
	}

	return assignedServices, nil
}