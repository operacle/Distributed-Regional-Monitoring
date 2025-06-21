
package handlers

import (
	"service-operation/shared/savers"
	"service-operation/types"
)

func (h *OperationHandler) saveMetricsToPocketBase(result *types.OperationResult, serviceID string) {
	metricsSaver := savers.NewMetricsSaver(h.pbClient)
	metricsSaver.SaveMetricsToPocketBase(result, serviceID)
}
