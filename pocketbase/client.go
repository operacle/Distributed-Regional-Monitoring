
package pocketbase

import (
	"fmt"
	"net/http"
	"time"
)

type PocketBaseClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewPocketBaseClient(baseURL string) (*PocketBaseClient, error) {
	// Use provided baseURL or default to localhost
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8090"
	}
	
	client := &PocketBaseClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	
	return client, nil
}

func (c *PocketBaseClient) GetBaseURL() string {
	return c.baseURL
}

func (c *PocketBaseClient) TestConnection() error {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/api/health", c.baseURL))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("PocketBase health check failed with status: %d", resp.StatusCode)
	}
	
	return nil
}

func (c *PocketBaseClient) IsAuthenticated() bool {
	// Since we're using public access mode, always return true
	return true
}
