
package pocketbase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *PocketBaseClient) createRecord(collection string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Post(
		fmt.Sprintf("%s/api/collections/%s/records", c.baseURL, collection),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create record in %s, status: %d", collection, resp.StatusCode)
	}

	return nil
}

func (c *PocketBaseClient) updateRecord(collection, recordID string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", 
		fmt.Sprintf("%s/api/collections/%s/records/%s", c.baseURL, collection, recordID),
		bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update record %s in %s, status: %d", recordID, collection, resp.StatusCode)
	}

	return nil
}