
package handlers

import (
	"service-operation/config"
	"service-operation/pocketbase"
)

type OperationHandler struct {
	config   *config.Config
	pbClient *pocketbase.PocketBaseClient
}

func NewOperationHandler(cfg *config.Config, pbClient *pocketbase.PocketBaseClient) *OperationHandler {
	return &OperationHandler{
		config:   cfg,
		pbClient: pbClient,
	}
}
