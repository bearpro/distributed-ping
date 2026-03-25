package node

// ping.go is responsible for executing local ICMP echo requests and
// translating network responses into transport-level ping results.

import (
	"net"
	"os"
	"strings"
	"time"

	"github.com/bearpro/distributed-ping/model"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const icmpProtocolIPv4 = 1

func executePing(identity model.NodeIdentity, req model.PingRequest) model.PingResult {
	result := model.PingResult{
		RequestID:      req.ID,
		Executor:       identity,
		ReporterNodeID: identity.NodeID,
		ObservedAt:     time.Now().UTC(),
	}

	dst, err := net.ResolveIPAddr("ip4", req.Target)
	if err != nil {
		result.Error = summarizeError(err)
		return result
	}

	conn, err := icmp.ListenPacket("udp4", "0.0.0.0")
	if err != nil {
		result.Error = summarizeError(err)
		return result
	}
	defer conn.Close()

	deadline := time.Now().Add(5 * time.Second)
	if err := conn.SetDeadline(deadline); err != nil {
		result.Error = summarizeError(err)
		return result
	}

	echoID := os.Getpid() & 0xffff
	echoSeq := int(time.Now().UnixNano() & 0xffff)
	message := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   echoID,
			Seq:  echoSeq,
			Data: []byte("distributed-ping"),
		},
	}

	payload, err := message.Marshal(nil)
	if err != nil {
		result.Error = summarizeError(err)
		return result
	}

	startedAt := time.Now()
	if _, err := conn.WriteTo(payload, &net.UDPAddr{IP: dst.IP}); err != nil {
		result.Error = summarizeError(err)
		return result
	}

	buffer := make([]byte, 1500)
	for {
		n, _, err := conn.ReadFrom(buffer)
		if err != nil {
			result.Error = summarizeError(err)
			return result
		}

		reply, err := icmp.ParseMessage(icmpProtocolIPv4, buffer[:n])
		if err != nil {
			continue
		}
		if reply.Type != ipv4.ICMPTypeEchoReply {
			continue
		}

		body, ok := reply.Body.(*icmp.Echo)
		if !ok || body.ID != echoID || body.Seq != echoSeq {
			continue
		}

		result.Success = true
		result.LatencyMs = float64(time.Since(startedAt).Microseconds()) / 1000
		return result
	}
}

func summarizeError(err error) string {
	trimmed := strings.TrimSpace(err.Error())
	if len(trimmed) > 240 {
		trimmed = trimmed[:240]
	}
	return trimmed
}
