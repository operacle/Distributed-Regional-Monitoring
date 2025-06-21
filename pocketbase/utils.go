
package pocketbase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *PocketBaseClient) CreateRecord(collection string, data interface{}) error {
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
		return fmt.Errorf("failed to create record, status: %d", resp.StatusCode)
	}

	return nil
}