
package savers

import (
	"fmt"
	"strings"

	"service-operation/types"
)

// Helper function to format bytes in a readable way
func FormatBytes(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	} else {
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	}
}

// Helper function to create short error messages
func GetShortErrorMessage(errorMessage string) string {
	if errorMessage == "" {
		return "Unknown error"
	}
	
	errorLower := strings.ToLower(errorMessage)
	
	if strings.Contains(errorLower, "timeout") {
		return "Request timeout"
	} else if strings.Contains(errorLower, "connection refused") {
		return "Connection refused"
	} else if strings.Contains(errorLower, "dns") || strings.Contains(errorLower, "no such host") {
		return "DNS resolution failed"
	} else if strings.Contains(errorLower, "certificate") || strings.Contains(errorLower, "ssl") || strings.Contains(errorLower, "tls") {
		return "SSL certificate error"
	} else if strings.Contains(errorLower, "server error") || strings.Contains(errorLower, "internal server error") {
		return "Internal server error"
	} else if strings.Contains(errorLower, "not found") {
		return "Page not found"
	} else if strings.Contains(errorLower, "unauthorized") {
		return "Unauthorized access"
	} else if strings.Contains(errorLower, "forbidden") {
		return "Access forbidden"
	}
	
	// For other errors, take first 50 characters and clean it up
	shortMsg := errorMessage
	if len(shortMsg) > 50 {
		shortMsg = shortMsg[:50] + "..."
	}
	
	return shortMsg
}

func GetStatusString(success bool) string {
	if success {
		return "up"
	}
	return "down"
}

func FormatResultDetails(result *types.OperationResult) string {
	// This can be expanded based on operation type
	if result.Details != "" {
		return result.Details
	}
	
	switch result.Type {
	case types.OperationPing:
		if result.Success {
			return fmt.Sprintf("Ping successful - %d packets sent, %d received", result.PacketsSent, result.PacketsRecv)
		}
		return fmt.Sprintf("Ping failed - %s", result.Error)
	case types.OperationHTTP:
		if result.Success {
			return fmt.Sprintf("HTTP %d - Response time: %.2fms", result.HTTPStatusCode, float64(result.ResponseTime.Nanoseconds())/1000000)
		}
		return fmt.Sprintf("HTTP failed - %s", result.Error)
	case types.OperationDNS:
		if result.Success {
			return fmt.Sprintf("DNS %s query successful - %d records found", result.DNSType, len(result.DNSRecords))
		}
		return fmt.Sprintf("DNS query failed - %s", result.Error)
	default:
		return "Operation completed"
	}
}
