package refserver

import (
	"crypto/ed25519"
	"net"
	"testing"
	"time"

	"github.com/logsmith-ai/logsmith-connector/connector"
	"github.com/logsmith-ai/logsmith-connector/proto"
)

func TestEndToEndTunnel(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	srv := New(pub, func() int64 { return 500 })

	// A fake private resource: TCP echo server.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		buf := make([]byte, 4)
		if _, err := c.Read(buf); err == nil {
			_, _ = c.Write(buf) // echo
		}
		_ = c.Close()
	}()

	// Connector connects to the reference server over an in-memory transport.
	connSide, srvSide := net.Pipe()
	tok, _ := proto.SignToken(priv, proto.Claims{ConnectorID: "conn_42", WorkspaceID: "ws", Exp: 1000})
	go func() { _ = connector.Run(connSide, connector.Options{Token: tok}) }()
	go func() { _ = srv.Handle(srvSide) }()

	// Wait for registration.
	deadline := time.Now().Add(2 * time.Second)
	for !srv.Has("conn_42") {
		if time.Now().After(deadline) {
			t.Fatal("connector never registered")
		}
		time.Sleep(5 * time.Millisecond)
	}

	// Open a tunnel to the echo server and verify the round trip.
	stream, err := srv.Dial("conn_42", ln.Addr().String())
	if err != nil {
		t.Fatalf("dial through tunnel: %v", err)
	}
	defer stream.Close()
	if _, err := stream.Write([]byte("ping")); err != nil {
		t.Fatalf("write: %v", err)
	}
	got := make([]byte, 4)
	_ = stream.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, err := readFull(stream, got); err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "ping" {
		t.Fatalf("got %q want ping (echoed)", got)
	}
}

func TestHandleRejectsBadToken(t *testing.T) {
	pub, _, _ := ed25519.GenerateKey(nil)
	otherPub, otherPriv, _ := ed25519.GenerateKey(nil)
	_ = otherPub
	srv := New(pub, func() int64 { return 500 })

	connSide, srvSide := net.Pipe()
	tok, _ := proto.SignToken(otherPriv, proto.Claims{ConnectorID: "c", WorkspaceID: "w", Exp: 1000})
	go func() {
		_ = proto.WriteVersion(connSide)
		_ = proto.WriteAuthFrame(connSide, tok)
	}()
	if err := srv.Handle(srvSide); err == nil {
		t.Fatal("expected Handle to reject token signed by unknown key")
	}
}

func TestHandleRejectsBadVersion(t *testing.T) {
	pub, _, _ := ed25519.GenerateKey(nil)
	srv := New(pub, func() int64 { return 500 })
	connSide, srvSide := net.Pipe()
	go func() {
		_, _ = connSide.Write([]byte{0xFF}) // unsupported version
		_ = connSide.Close()
	}()
	if err := srv.Handle(srvSide); err == nil {
		t.Fatal("expected Handle to reject an unsupported protocol version")
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
