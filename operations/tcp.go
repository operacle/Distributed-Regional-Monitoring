
package operations

import (
	"fmt"
	"net"
	"time"

	"service-operation/types"
)

type TCPOperation struct {
	timeout time.Duration
}

func NewTCPOperation(timeout time.Duration) *TCPOperation {
	return &TCPOperation{timeout: timeout}
}

func (t *TCPOperation) Execute(host string, port int) (*types.OperationResult, error) {
	result := &types.OperationResult{
		Type:      types.OperationTCP,
		Host:      host,
		Port:      port,
		StartTime: time.Now(),
	}

	start := time.Now()
	
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", address, t.timeout)
	
	result.ResponseTime = time.Since(start)
	result.EndTime = time.Now()

	if err != nil {
		result.Error = err.Error()
		result.TCPConnected = false
		result.Success = false
		result.Details = fmt.Sprintf("Failed to connect to %s:%d - %s", host, port, err.Error())
	} else {
		conn.Close()
		result.TCPConnected = true
		result.Success = true
		result.Details = fmt.Sprintf("Successfully connected to %s:%d", host, port)
	}

	return result, nil
}
