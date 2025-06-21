
package ping

import (
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type ICMPPinger struct {
	timeout time.Duration
}

func NewICMPPinger(timeout time.Duration) *ICMPPinger {
	return &ICMPPinger{timeout: timeout}
}

func (p *ICMPPinger) Ping(host string, count int) (*PingResult, error) {
	dst, err := net.ResolveIPAddr("ip4", host)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve host %s: %v", host, err)
	}

	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return nil, fmt.Errorf("failed to create ICMP connection: %v", err)
	}
	defer conn.Close()

	result := &PingResult{
		Host:        host,
		PacketsSent: count,
		StartTime:   time.Now(),
	}

	var totalRTT time.Duration
	var minRTT, maxRTT time.Duration

	for i := 0; i < count; i++ {
		message := &icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 0,
			Body: &icmp.Echo{
				ID:   os.Getpid() & 0xffff,
				Seq:  i + 1,
				Data: []byte("Hello, World!"),
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

		// Parse the reply
		rm, err := icmp.ParseMessage(ipv4.ICMPTypeEchoReply, reply)
		if err != nil {
			continue
		}

		switch rm.Type {
		case ipv4.ICMPTypeEchoReply:
			if peer.String() == dst.String() {
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
		result.Success = true
	}

	return result, nil
}
