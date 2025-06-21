package operations

import (
	"fmt"
	"net"
	"strings"
	"time"

	"service-operation/types"
)

type DNSOperation struct {
	timeout time.Duration
}

func NewDNSOperation(timeout time.Duration) *DNSOperation {
	return &DNSOperation{timeout: timeout}
}

func (d *DNSOperation) Execute(host, query string) (*types.OperationResult, error) {
	// Validate inputs
	if host == "" {
		return nil, fmt.Errorf("host cannot be empty")
	}
	
	if query == "" {
		query = "A" // Default to A record
	}

	result := &types.OperationResult{
		Type:      types.OperationDNS,
		Host:      host,
		DNSType:   query,
		StartTime: time.Now(),
	}

	start := time.Now()

	// Resolve the host first to get detailed info
	var resolvedIPs []string
	var err error

	switch strings.ToUpper(query) {
	case "A":
		resolvedIPs, err = d.performARecordLookup(host)
		
	case "AAAA":
		resolvedIPs, err = d.performAAAARecordLookup(host)
		
	case "MX":
		resolvedIPs, err = d.performMXRecordLookup(host)
		
	case "TXT":
		resolvedIPs, err = d.performTXTRecordLookup(host)
		
	case "CNAME":
		resolvedIPs, err = d.performCNAMERecordLookup(host)
		
	case "NS":
		resolvedIPs, err = d.performNSRecordLookup(host)
		
	default:
		// Default to A record lookup for unknown types
		resolvedIPs, err = d.performARecordLookup(host)
	}

	result.ResponseTime = time.Since(start)
	result.EndTime = time.Now()

	if err != nil {
		result.Error = err.Error()
		result.Success = false
		result.Details = d.createDetailedErrorMessage(err.Error(), host, query)
	} else {
		result.DNSRecords = resolvedIPs
		result.Success = len(resolvedIPs) > 0
		
		if result.Success {
			result.Details = d.createDetailedSuccessMessage(result, host, query, resolvedIPs)
		} else {
			result.Error = "No DNS records found"
			result.Details = d.createDetailedErrorMessage("no records found", host, query)
		}
	}

	return result, nil
}

func (d *DNSOperation) performARecordLookup(host string) ([]string, error) {
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}
	
	var ipv4Records []string
	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			ipv4Records = append(ipv4Records, ipv4.String())
		}
	}
	
	return ipv4Records, nil
}

func (d *DNSOperation) performAAAARecordLookup(host string) ([]string, error) {
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}
	
	var ipv6Records []string
	for _, ip := range ips {
		if ipv6 := ip.To16(); ipv6 != nil && ip.To4() == nil {
			ipv6Records = append(ipv6Records, ipv6.String())
		}
	}
	
	return ipv6Records, nil
}

func (d *DNSOperation) performMXRecordLookup(host string) ([]string, error) {
	mxRecords, err := net.LookupMX(host)
	if err != nil {
		return nil, err
	}
	
	var records []string
	for _, mx := range mxRecords {
		records = append(records, fmt.Sprintf("%s (priority: %d)", mx.Host, mx.Pref))
	}
	
	return records, nil
}

func (d *DNSOperation) performTXTRecordLookup(host string) ([]string, error) {
	txtRecords, err := net.LookupTXT(host)
	if err != nil {
		return nil, err
	}
	
	return txtRecords, nil
}

func (d *DNSOperation) performCNAMERecordLookup(host string) ([]string, error) {
	cname, err := net.LookupCNAME(host)
	if err != nil {
		return nil, err
	}
	
	return []string{cname}, nil
}

func (d *DNSOperation) performNSRecordLookup(host string) ([]string, error) {
	nsRecords, err := net.LookupNS(host)
	if err != nil {
		return nil, err
	}
	
	var records []string
	for _, ns := range nsRecords {
		records = append(records, ns.Host)
	}
	
	return records, nil
}

func (d *DNSOperation) createDetailedSuccessMessage(result *types.OperationResult, host, queryType string, records []string) string {
	var details strings.Builder
	
	// Success indicator with basic info
	details.WriteString(fmt.Sprintf("🟢 DNS SUCCESS - %s query for %s", 
		strings.ToUpper(queryType), host))
	
	// Response time
	details.WriteString(fmt.Sprintf(" | Response time: %.2fms", 
		float64(result.ResponseTime.Nanoseconds())/1000000))
	
	// Record count
	details.WriteString(fmt.Sprintf(" | Records found: %d", len(records)))
	
	// Show first few records for context
	if len(records) > 0 {
		details.WriteString(" | ")
		if len(records) <= 3 {
			details.WriteString(fmt.Sprintf("Results: %s", strings.Join(records, ", ")))
		} else {
			details.WriteString(fmt.Sprintf("Results: %s... (+%d more)", 
				strings.Join(records[:3], ", "), len(records)-3))
		}
	}
	
	return details.String()
}

func (d *DNSOperation) createDetailedErrorMessage(errorMsg, host, queryType string) string {
	var details strings.Builder
	
	errorLower := strings.ToLower(errorMsg)
	
	if strings.Contains(errorLower, "timeout") {
		details.WriteString("⏱️ DNS TIMEOUT - Query timed out")
	} else if strings.Contains(errorLower, "no such host") || strings.Contains(errorLower, "host not found") {
		details.WriteString("🔍 HOST NOT FOUND - DNS resolution failed")
	} else if strings.Contains(errorLower, "no answer") || strings.Contains(errorLower, "no records found") {
		details.WriteString("📝 NO RECORDS - No DNS records of requested type")
	} else if strings.Contains(errorLower, "server failure") {
		details.WriteString("🔧 SERVER FAILURE - DNS server error")
	} else if strings.Contains(errorLower, "refused") {
		details.WriteString("🚫 QUERY REFUSED - DNS server refused query")
	} else {
		details.WriteString("❌ DNS FAILED - Query error")
	}
	
	// Add query details
	details.WriteString(fmt.Sprintf(" | Query: %s %s", 
		strings.ToUpper(queryType), host))
	
	// Add specific error details
	if errorMsg != "" {
		details.WriteString(fmt.Sprintf(" | Error: %s", d.getShortErrorMessage(errorMsg)))
	}
	
	return details.String()
}

func (d *DNSOperation) getShortErrorMessage(errorMessage string) string {
	if errorMessage == "" {
		return "Unknown error"
	}
	
	errorLower := strings.ToLower(errorMessage)
	
	if strings.Contains(errorLower, "timeout") {
		return "Query timeout"
	} else if strings.Contains(errorLower, "no such host") {
		return "Host not found"
	} else if strings.Contains(errorLower, "no answer") {
		return "No records found"
	} else if strings.Contains(errorLower, "server failure") {
		return "DNS server failure"
	} else if strings.Contains(errorLower, "refused") {
		return "Query refused"
	} else if strings.Contains(errorLower, "network unreachable") {
		return "Network unreachable"
	}
	
	// For other errors, take first 50 characters and clean it up
	shortMsg := errorMessage
	if len(shortMsg) > 50 {
		shortMsg = shortMsg[:50] + "..."
	}
	
	return shortMsg
}
