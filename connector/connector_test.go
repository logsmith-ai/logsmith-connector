package connector

import (
	"net"
	"testing"
	"time"

	"github.com/hashicorp/yamux"
	"github.com/logsmith-ai/logsmith-connector/proto"
)

// TestRunAcceptsAndForwards drives Run over an in-memory transport. The "server"
// side reads the auth frame, starts a yamux server, opens a stream to a fake
// target, and checks the round trip.
func TestRunAcceptsAndForwards(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	targetClient, targetServer := net.Pipe()

	dial := func(network, addr string) (net.Conn, error) { return targetClient, nil }

	go func() {
		_ = Run(clientConn, Options{Token: "tok-1", Dial: dial})
	}()

	// Server side: consume version byte, then auth frame, then yamux server.
	ver, err := proto.ReadVersion(serverConn)
	if err != nil {
		t.Fatalf("read version: %v", err)
	}
	if ver != proto.ProtocolVersion {
		t.Fatalf("version got %d want %d", ver, proto.ProtocolVersion)
	}
	tok, err := proto.ReadAuthFrame(serverConn)
	if err != nil {
		t.Fatalf("read auth: %v", err)
	}
	if tok != "tok-1" {
		t.Fatalf("token got %q want tok-1", tok)
	}
	sess, err := yamux.Server(serverConn, nil)
	if err != nil {
		t.Fatalf("yamux server: %v", err)
	}
	stream, err := sess.OpenStream()
	if err != nil {
		t.Fatalf("open stream: %v", err)
	}
	if err := proto.WriteTargetHeader(stream, "x:1"); err != nil {
		t.Fatalf("write header: %v", err)
	}
	go func() { _, _ = stream.Write([]byte("ping")) }()

	got := make([]byte, 4)
	_ = targetServer.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, err := readFull(targetServer, got); err != nil {
		t.Fatalf("target read: %v", err)
	}
	if string(got) != "ping" {
		t.Fatalf("target got %q want ping", got)
	}
}
