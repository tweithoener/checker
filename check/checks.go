package check

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	chkr "github.com/tweithoener/checker"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

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

func HTTPCheck(method, url string, expected int) chkr.Check {
	return func(ctx context.Context, h chkr.History) (chkr.State, string) {
		req, err := http.NewRequestWithContext(ctx, method, url, nil)
		if err != nil {
			return chkr.Fail, fmt.Sprintf("Failed to create request: %v", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return chkr.Fail, fmt.Sprintf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != expected {
			return chkr.Fail, fmt.Sprintf("Unexpected status code: %d (expected %d)", resp.StatusCode, expected)
		}

		return chkr.OK, ""
	}
}

func HTTPProxy(method, request, proxy string, expected int) chkr.Check {
	return func(ctx context.Context, h chkr.History) (chkr.State, string) {
		req, err := http.NewRequestWithContext(ctx, method, request, nil)
		if err != nil {
			return chkr.Fail, fmt.Sprintf("Failed to create request: %v", err)
		}

		proxyUrl, err := url.Parse(proxy)
		if err != nil {
			return chkr.Fail, fmt.Sprintf("Failed to parse proxy URL: %v", err)
		}
		cl := http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
		resp, err := cl.Do(req)
		if err != nil {
			return chkr.Fail, fmt.Sprintf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != expected {
			return chkr.Fail, fmt.Sprintf("Unexpected status code: %d (expected %d)", resp.StatusCode, expected)
		}

		return chkr.OK, ""
	}
}

func Fail(chk chkr.Check) chkr.Check {
	return func(ctx context.Context, h chkr.History) (chkr.State, string) {
		s, msg := chk(ctx, h)
		if s != chkr.Fail {
			return chkr.Fail, fmt.Sprintf("Check was supposed to fail but did not: %s %s", s, msg)
		}

		return chkr.OK, fmt.Sprintf("Check failed as expected: %s %s", s, msg)
	}
}
