
package pocketbase

import (
	"fmt"
	"net/http"
)

// GetService retrieves a specific service by ID
func (c *PocketBaseClient) GetService(serviceID string) (*Service, error) {
	url := fmt.Sprintf("%s/api/collections/services/records/%s", c.baseURL, serviceID)
	
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get service %s: status %d", serviceID, resp.StatusCode)
	}

	var service Service
	if err := c.parseResponse(resp, &service); err != nil {
		return nil, err
	}

	return &service, nil
}

// UpdateServiceStatus updates the status of a service
func (c *PocketBaseClient) UpdateServiceStatus(serviceID, status string, responseTime int64, errorMessage string) error {
	data := map[string]interface{}{
		"status":        status,
		"response_time": responseTime,
		"last_checked":  "now",
	}
	
	if errorMessage != "" {
		data["error_message"] = errorMessage
	}

	return c.updateRecord("services", serviceID, data)
}