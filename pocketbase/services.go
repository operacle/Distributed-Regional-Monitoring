
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

// GetAssignedServices retrieves services assigned to specific region and agent
func (c *PocketBaseClient) GetAssignedServices(regionName, agentID string) ([]Service, error) {
	// Build filter with proper URL encoding and PocketBase filter syntax
	filter := fmt.Sprintf("region_name='%s' && agent_id='%s' && status!='paused'", regionName, agentID)
	encodedFilter := url.QueryEscape(filter)
	url := fmt.Sprintf("%s/api/collections/services/records?filter=%s", c.baseURL, encodedFilter)
	
	resp, err := c.httpClient.Get(url)
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

	return response.Items, nil
}