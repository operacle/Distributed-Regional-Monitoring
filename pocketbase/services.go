
package pocketbase

import (
	"fmt"
	"net/http"
	"net/url"
)

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
func (c *PocketBaseClient) GetAssignedServices(regionName, agentID string) ([]Service, error) {
	var allServices []Service
	page := 1
	perPage := 30 // Use default pagination size

	// Build filter with proper URL encoding and PocketBase filter syntax
	filter := fmt.Sprintf("region_name='%s' && agent_id='%s' && status!='paused'", regionName, agentID)
	encodedFilter := url.QueryEscape(filter)

	for {
		// Fetch services page by page with filter for assigned services
		requestURL := fmt.Sprintf("%s/api/collections/services/records?page=%d&perPage=%d&filter=%s", 
			c.baseURL, page, perPage, encodedFilter)
		
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

	return allServices, nil
}