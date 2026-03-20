package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	chkr "github.com/tweithoener/checker"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// PingArgs defines the arguments for a Ping check.
type PingArgs struct {
	Address    string
	WarnMillis int
	FailMillis int
}

type pingMaker struct{}

var pingMkr = pingMaker{}

func (pingMaker) Maker() string {
	return "Ping"
}

func (pingMaker) UnmarshalArgs(j json.RawMessage) (any, error) {
	args := PingArgs{}
	if err := json.Unmarshal(j, &args); err != nil {
		return args, fmt.Errorf("can't unmarshal Ping arguments: %v", err)
	}
	return args, nil
}

func (pingMaker) FromConfig(c chkr.CheckConfig) (chkr.Check, error) {
	args, ok := c.Args.(PingArgs)
	if !ok {
		return nil, fmt.Errorf("configured arguments are not Ping arguments")
	}
	return Ping(args.Address, args.WarnMillis, args.FailMillis), nil
}

// Ping returns a check that verifies connectivity via ICMP echo requests.
func Ping(address string, warnMillis, failMillis int) chkr.Check {
	return func(ctx context.Context, h chkr.History) (chkr.State, string) {
		dest, err := net.ResolveIPAddr("ip4", address)
		if err != nil {
			return chkr.Fail, fmt.Sprintf("Failed to resolve address: %v", err)
		}

		c, err := icmp.ListenPacket("udp4", "0.0.0.0")
		if err != nil {
			c, err = icmp.ListenPacket("ip4:icmp", "0.0.0.0")
			if err != nil {
				return chkr.Fail, fmt.Sprintf("Failed to listen for ICMP: %v (Note: Ping might require root privileges)", err)
			}
		}
		defer c.Close()

		id := os.Getpid() & 0xffff
		isUdp := c.LocalAddr().Network() == "udp" || c.LocalAddr().Network() == "udp4"
		if isUdp {
			if udpAddr, ok := c.LocalAddr().(*net.UDPAddr); ok {
				id = udpAddr.Port
			}
		}

		msg := icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 0,
			Body: &icmp.Echo{
				ID:   id,
				Seq:  1,
				Data: []byte("GOMOD-CHECKER-PING"),
			},
		}

		b, err := msg.Marshal(nil)
		if err != nil {
			return chkr.Fail, fmt.Sprintf("Failed to marshal ICMP message: %v", err)
		}

		start := time.Now()
		failDuration := time.Duration(failMillis) * time.Millisecond
		warnDuration := time.Duration(warnMillis) * time.Millisecond

		deadline := start.Add(failDuration)
		if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
			deadline = d
		}
		c.SetDeadline(deadline)

		var target net.Addr = dest // Standardmäßig *net.IPAddr (für Raw Sockets)
		if isUdp {
			target = &net.UDPAddr{IP: dest.IP}
		}

		if _, err := c.WriteTo(b, target); err != nil {
			return chkr.Fail, fmt.Sprintf("Failed to send ICMP: %v", err)
		}

		reply := make([]byte, 1500)
		for {
			n, _, err := c.ReadFrom(reply)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					return chkr.Fail, fmt.Sprintf("Ping timeout after %dms", failMillis)
				}
				return chkr.Fail, fmt.Sprintf("Failed to receive ICMP reply: %v", err)
			}

			duration := time.Since(start)
			rm, err := icmp.ParseMessage(1, reply[:n])
			if err != nil {
				continue
			}

			pkt, ok := rm.Body.(*icmp.Echo)
			if !ok {
				continue
			}
			if pkt.ID != id {
				continue // Not our Ping
			}

			msg := fmt.Sprintf("Ping latency: %v", duration.Truncate(time.Microsecond))
			if duration > failDuration {
				return chkr.Fail, msg
			}
			if duration > warnDuration {
				return chkr.Warn, msg
			}
			return chkr.OK, msg
		}
	}
}
