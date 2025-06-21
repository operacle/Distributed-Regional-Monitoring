
package monitoring

// This file previously contained saveDetailedDataByType which was causing duplicate saves
// All detailed data saving is now handled by the shared savers in a single place
// to prevent duplicate records in PocketBase

// The monitoring service now uses only:
// 1. SaveMetricsForService from shared/savers for complete data saving
// 2. Direct service status updates for the services table
