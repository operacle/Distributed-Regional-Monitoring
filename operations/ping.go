package operations

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"service-operation/types"
)

type PingOperation struct {
	timeout time.Duration
}

func NewPingOperation(timeout time.Duration) *PingOperation {
	return &PingOperation{timeout: timeout}
}

func (p *PingOperation) Execute(host string, count int) (*types.OperationResult, error) {
	// Validate host/IP
	if host == "" {
		return nil, fmt.Errorf("host cannot be empty")
	}

	// Always try system ping first for better reliability
	result, err := p.executeSystemPing(host, count)
	if err == nil && result.PacketsRecv > 0 {
		return result, nil
	}
	
	// If system ping fails, try raw ICMP as fallback
	fmt.Printf("System ping failed (%v), trying raw ICMP\n", err)
	rawResult, rawErr := p.executeRawICMP(host, count)
	if rawErr == nil && rawResult.PacketsRecv > 0 {
		return rawResult, nil
	}
	
	// If both fail, return the system ping result with error info
	if result != nil {
		return result, nil
	}
	
	return nil, fmt.Errorf("both system ping and raw ICMP failed: system_err=%v, raw_err=%v", err, rawErr)
}

func (p *PingOperation) executeSystemPing(host string, count int) (*types.OperationResult, error) {
	result := &types.OperationResult{
		Type:        types.OperationPing,
		Host:        host,
		PacketsSent: count,
		StartTime:   time.Now(),
	}

	// Resolve host to get IP address for better details
	ips, err := net.LookupIP(host)
	var resolvedIP string
	if err == nil && len(ips) > 0 {
		resolvedIP = ips[0].String()
	}

	// Build ping command based on OS
	var cmd *exec.Cmd
	timeoutSeconds := int(p.timeout.Seconds())
	if timeoutSeconds < 1 {
		timeoutSeconds = 10 // Minimum 10 seconds timeout
	}

	switch runtime.GOOS {
	case "linux":
		// Linux ping: -c count -W timeout_in_seconds
		cmd = exec.Command("ping", "-c", fmt.Sprintf("%d", count), "-W", fmt.Sprintf("%d", timeoutSeconds), host)
	case "darwin":
		// macOS ping: -c count -W timeout_in_milliseconds
		cmd = exec.Command("ping", "-c", fmt.Sprintf("%d", count), "-W", fmt.Sprintf("%d", timeoutSeconds*1000), host)
	case "windows":
		// Windows ping: -n count -w timeout_in_milliseconds
		cmd = exec.Command("ping", "-n", fmt.Sprintf("%d", count), "-w", fmt.Sprintf("%d", timeoutSeconds*1000), host)
	default:
		// Default to Linux-style
		cmd = exec.Command("ping", "-c", fmt.Sprintf("%d", count), "-W", fmt.Sprintf("%d", timeoutSeconds), host)
	}

	// Set command timeout slightly longer than ping timeout
	cmdTimeout := time.Duration(timeoutSeconds+5) * time.Second
	done := make(chan error, 1)
	var output []byte
	var cmdErr error

	go func() {
		output, cmdErr = cmd.Output()
		done <- cmdErr
	}()

	select {
	case err := <-done:
		result.EndTime = time.Now()
		
		if err != nil {
			// Even if command fails, try to parse partial output
			result.PacketLoss = 100.0
			if len(output) > 0 {
				p.parseSystemPingOutput(result, string(output), count)
			}
			// Set detailed error information
			result.Error = p.createDetailedErrorMessage(err.Error(), host, resolvedIP)
			return result, nil // Don't return error, return result with failure info
		}
		
		// Parse successful output
		if len(output) > 0 {
			p.parseSystemPingOutput(result, string(output), count)
		}

		// Create detailed success message
		if result.Success {
			result.Details = p.createDetailedSuccessMessage(result, host, resolvedIP)
		}
		
		return result, nil
		
	case <-time.After(cmdTimeout):
		// Command timed out
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		result.EndTime = time.Now()
		result.PacketLoss = 100.0
		result.Error = fmt.Sprintf("Ping request timed out after %v", cmdTimeout)
		result.Details = p.createDetailedErrorMessage("timeout", host, resolvedIP)
		return result, fmt.Errorf("ping command timed out after %v", cmdTimeout)
	}
}

func (p *PingOperation) createDetailedSuccessMessage(result *types.OperationResult, host, resolvedIP string) string {
	var details strings.Builder
	
	// Success indicator with basic info
	details.WriteString(fmt.Sprintf("🟢 PING SUCCESS - %d/%d packets received", 
		result.PacketsRecv, result.PacketsSent))
	
	// Host and IP information
	if resolvedIP != "" && resolvedIP != host {
		details.WriteString(fmt.Sprintf(" | Host: %s (%s)", host, resolvedIP))
	} else {
		details.WriteString(fmt.Sprintf(" | Host: %s", host))
	}
	
	// Timing information
	if result.AvgRTT > 0 {
		details.WriteString(fmt.Sprintf(" | Avg RTT: %.2fms", 
			float64(result.AvgRTT.Nanoseconds())/1000000))
	}
	
	if result.MinRTT > 0 && result.MaxRTT > 0 {
		details.WriteString(fmt.Sprintf(" | Range: %.2f-%.2fms", 
			float64(result.MinRTT.Nanoseconds())/1000000,
			float64(result.MaxRTT.Nanoseconds())/1000000))
	}
	
	// Packet loss information
	if result.PacketLoss > 0 {
		details.WriteString(fmt.Sprintf(" | Packet Loss: %.1f%%", result.PacketLoss))
	} else {
		details.WriteString(" | No packet loss")
	}
	
	return details.String()
}

func (p *PingOperation) createDetailedErrorMessage(errorMsg, host, resolvedIP string) string {
	var details strings.Builder
	
	errorLower := strings.ToLower(errorMsg)
	
	if strings.Contains(errorLower, "timeout") {
		details.WriteString("⏱️ PING TIMEOUT - Host did not respond within timeout period")
	} else if strings.Contains(errorLower, "unreachable") {
		details.WriteString("🚫 HOST UNREACHABLE - Network path to host is blocked")
	} else if strings.Contains(errorLower, "no route") {
		details.WriteString("🛤️ NO ROUTE - No network route to destination")
	} else if strings.Contains(errorLower, "name resolution") || strings.Contains(errorLower, "unknown host") {
		details.WriteString("🔍 DNS RESOLUTION FAILED - Unable to resolve hostname")
	} else if strings.Contains(errorLower, "permission denied") {
		details.WriteString("🔐 PERMISSION DENIED - Insufficient privileges for ICMP")
	} else {
		details.WriteString("❌ PING FAILED - Connection error")
	}
	
	// Add host information
	if resolvedIP != "" && resolvedIP != host {
		details.WriteString(fmt.Sprintf(" | Target: %s (%s)", host, resolvedIP))
	} else {
		details.WriteString(fmt.Sprintf(" | Target: %s", host))
	}
	
	// Add specific error details
	if errorMsg != "" {
		details.WriteString(fmt.Sprintf(" | Error: %s", p.getShortErrorMessage(errorMsg)))
	}
	
	return details.String()
}

func (p *PingOperation) getShortErrorMessage(errorMessage string) string {
	if errorMessage == "" {
		return "Unknown error"
	}
	
	errorLower := strings.ToLower(errorMessage)
	
	if strings.Contains(errorLower, "timeout") {
		return "Request timeout"
	} else if strings.Contains(errorLower, "unreachable") {
		return "Host unreachable"
	} else if strings.Contains(errorLower, "no route") {
		return "No route to host"
	} else if strings.Contains(errorLower, "name resolution") || strings.Contains(errorLower, "unknown host") {
		return "DNS resolution failed"
	} else if strings.Contains(errorLower, "permission denied") {
		return "Permission denied"
	} else if strings.Contains(errorLower, "network is down") {
		return "Network is down"
	}
	
	// For other errors, take first 50 characters and clean it up
	shortMsg := errorMessage
	if len(shortMsg) > 50 {
		shortMsg = shortMsg[:50] + "..."
	}
	
	return shortMsg
}

func (p *PingOperation) parseSystemPingOutput(result *types.OperationResult, output string, expectedCount int) {
	// Extract packet loss
	lossRegex := regexp.MustCompile(`(\d+(?:\.\d+)?)%.*packet.*loss`)
	if matches := lossRegex.FindStringSubmatch(output); len(matches) > 1 {
		if loss, err := strconv.ParseFloat(matches[1], 64); err == nil {
			result.PacketLoss = loss
		}
	}

	// Extract timing information (Linux/macOS format)
	timingRegex := regexp.MustCompile(`rtt min/avg/max/(?:mdev|stddev) = ([\d.]+)/([\d.]+)/([\d.]+)/([\d.]+) ms`)
	if matches := timingRegex.FindStringSubmatch(output); len(matches) > 4 {
		if min, err := strconv.ParseFloat(matches[1], 64); err == nil {
			result.MinRTT = time.Duration(min * float64(time.Millisecond))
		}
		if avg, err := strconv.ParseFloat(matches[2], 64); err == nil {
			result.AvgRTT = time.Duration(avg * float64(time.Millisecond))
			result.ResponseTime = result.AvgRTT
		}
		if max, err := strconv.ParseFloat(matches[3], 64); err == nil {
			result.MaxRTT = time.Duration(max * float64(time.Millisecond))
		}
	}

	// Extract individual ping times and count successful pings
	timeRegex := regexp.MustCompile(`time[<=]([\d.]+) ?ms`)
	timeMatches := timeRegex.FindAllStringSubmatch(output, -1)
	result.PacketsRecv = len(timeMatches)
	
	for _, match := range timeMatches {
		if len(match) > 1 {
			if t, err := strconv.ParseFloat(match[1], 64); err == nil {
				rtt := time.Duration(t * float64(time.Millisecond))
				result.RTTs = append(result.RTTs, rtt)
				
				// Update min/max if not set by summary stats
				if result.MinRTT == 0 || rtt < result.MinRTT {
					result.MinRTT = rtt
				}
				if result.MaxRTT == 0 || rtt > result.MaxRTT {
					result.MaxRTT = rtt
				}
			}
		}
	}

	// Calculate average RTT if not set by summary stats
	if result.AvgRTT == 0 && len(result.RTTs) > 0 {
		var total time.Duration
		for _, rtt := range result.RTTs {
			total += rtt
		}
		result.AvgRTT = total / time.Duration(len(result.RTTs))
		result.ResponseTime = result.AvgRTT
	}

	// Recalculate packet loss if we have individual pings
	if expectedCount > 0 {
		result.PacketLoss = float64(expectedCount-result.PacketsRecv) / float64(expectedCount) * 100
	}

	// Mark as successful if we got any responses
	if result.PacketsRecv > 0 {
		result.Success = true
	}

	// Handle Windows output format if Linux/macOS parsing didn't work
	if result.PacketsRecv == 0 && strings.Contains(strings.ToLower(output), "reply from") {
		p.parseWindowsPingOutput(result, output, expectedCount)
	}
}

func (p *PingOperation) parseWindowsPingOutput(result *types.OperationResult, output string, expectedCount int) {
	// Windows: "Reply from 1.1.1.1: bytes=32 time=14ms TTL=56"
	timeRegex := regexp.MustCompile(`time[<=]([\d.]+)ms`)
	timeMatches := timeRegex.FindAllStringSubmatch(output, -1)
	result.PacketsRecv = len(timeMatches)
	
	for _, match := range timeMatches {
		if len(match) > 1 {
			if t, err := strconv.ParseFloat(match[1], 64); err == nil {
				rtt := time.Duration(t * float64(time.Millisecond))
				result.RTTs = append(result.RTTs, rtt)
				
				if result.MinRTT == 0 || rtt < result.MinRTT {
					result.MinRTT = rtt
				}
				if result.MaxRTT == 0 || rtt > result.MaxRTT {
					result.MaxRTT = rtt
				}
			}
		}
	}

	if len(result.RTTs) > 0 {
		var total time.Duration
		for _, rtt := range result.RTTs {
			total += rtt
		}
		result.AvgRTT = total / time.Duration(len(result.RTTs))
		result.ResponseTime = result.AvgRTT
		result.Success = true
	}

	if expectedCount > 0 {
		result.PacketLoss = float64(expectedCount-result.PacketsRecv) / float64(expectedCount) * 100
	}
}

func (p *PingOperation) executeRawICMP(host string, count int) (*types.OperationResult, error) {
	dst, err := net.ResolveIPAddr("ip4", host)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve host %s: %v", host, err)
	}

	// Try to create ICMP connection with better error handling
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return nil, fmt.Errorf("failed to create ICMP connection (requires root/admin privileges): %v", err)
	}
	defer conn.Close()

	result := &types.OperationResult{
		Type:        types.OperationPing,
		Host:        host,
		PacketsSent: count,
		StartTime:   time.Now(),
	}

	var totalRTT time.Duration
	var minRTT, maxRTT time.Duration
	pid := os.Getpid() & 0xffff

	for i := 0; i < count; i++ {
		message := &icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 0,
			Body: &icmp.Echo{
				ID:   pid,
				Seq:  i + 1,
				Data: []byte(fmt.Sprintf("Hello, World! %d", i)),
			},
		}

		data, err := message.Marshal(nil)
		if err != nil {
			continue
		}

		start := time.Now()
		_, err = conn.WriteTo(data, dst)
		if err != nil {
			continue
		}

		// Set read deadline
		err = conn.SetReadDeadline(time.Now().Add(p.timeout))
		if err != nil {
			continue
		}

		reply := make([]byte, 1500)
		_, peer, err := conn.ReadFrom(reply)
		if err != nil {
			continue
		}

		rtt := time.Since(start)

		// Parse the reply with better error handling - fix the type conversion
		rm, err := icmp.ParseMessage(int(ipv4.ICMPTypeEchoReply), reply)
		if err != nil {
			continue
		}

		switch rm.Type {
		case ipv4.ICMPTypeEchoReply:
			if peer.String() == dst.String() {
				echoReply, ok := rm.Body.(*icmp.Echo)
				if ok && echoReply.ID == pid {
					result.PacketsRecv++
					totalRTT += rtt

					if result.PacketsRecv == 1 || rtt < minRTT {
						minRTT = rtt
					}
					if result.PacketsRecv == 1 || rtt > maxRTT {
						maxRTT = rtt
					}

					result.RTTs = append(result.RTTs, rtt)
				}
			}
		}

		// Sleep between pings except for the last one
		if i < count-1 {
			time.Sleep(1 * time.Second)
		}
	}

	result.EndTime = time.Now()
	result.PacketLoss = float64(count-result.PacketsRecv) / float64(count) * 100

	if result.PacketsRecv > 0 {
		result.MinRTT = minRTT
		result.MaxRTT = maxRTT
		result.AvgRTT = totalRTT / time.Duration(result.PacketsRecv)
		result.ResponseTime = result.AvgRTT
		result.Success = true
		// Create detailed success message for raw ICMP
		result.Details = p.createDetailedSuccessMessage(result, host, dst.String())
	} else {
		result.Error = "No ICMP echo replies received"
		result.Details = p.createDetailedErrorMessage("no replies", host, dst.String())
	}

	return result, nil
}
