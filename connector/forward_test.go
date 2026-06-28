package connector

import (
	"net"
	"testing"

	"github.com/logsmith-ai/logsmith-connector/proto"
)

// TestHandleStreamForwardsBytes wires a fake "yamux stream" (net.Pipe) to a
// fake target (another net.Pipe via the dial func) and checks bytes flow both
// ways after the target header is consumed.
func TestHandleStreamForwardsBytes(t *testing.T) {
	serverSide, connectorSide := net.Pipe() // serverSide simulates the tunnel-server end of the stream
	targetClient, targetServer := net.Pipe() // targetServer simulates the private resource

	dial := func(network, addr string) (net.Conn, error) {
		if addr != "resource.internal:9090" {
			t.Errorf("dialed wrong addr: %q", addr)
		}
		return targetClient, nil
	}

	go HandleStream(connectorSide, dial)

	// Server writes the target header, then a request payload.
	go func() {
		_ = proto.WriteTargetHeader(serverSide, "resource.internal:9090")
		_, _ = serverSide.Write([]byte("ping"))
	}()

	// The target should receive "ping".
	got := make([]byte, 4)
	if _, err := readFull(targetServer, got); err != nil {
		t.Fatalf("target read: %v", err)
	}
	if string(got) != "ping" {
		t.Fatalf("target got %q want ping", got)
	}

	// Target replies "pong"; server side should receive it.
	go func() { _, _ = targetServer.Write([]byte("pong")) }()
	back := make([]byte, 4)
	if _, err := readFull(serverSide, back); err != nil {
		t.Fatalf("server read: %v", err)
	}
	if string(back) != "pong" {
		t.Fatalf("server got %q want pong", back)
	}
}

func readFull(c net.Conn, b []byte) (int, error) {
	total := 0
	for total < len(b) {
		n, err := c.Read(b[total:])
		total += n
		if err != nil {
			return total, err
		}
	}
	return total, nil
}
