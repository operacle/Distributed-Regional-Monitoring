
package pocketbase

import (
	"encoding/json"
	"net/http"
)

func (c *PocketBaseClient) parseResponse(resp *http.Response, target interface{}) error {
	return json.NewDecoder(resp.Body).Decode(target)
}