package main

import (
	"crypto/tls"
	"net"
	"testing"
	"time"
)

// TestDialTLSConnects starts a throwaway TLS listener and checks dialTLS reaches it.
func TestDialTLSConnects(t *testing.T) {
	cert := selfSignedCert(t)
	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{Certificates: []tls.Certificate{cert}})
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	go func() {
		c, err := ln.Accept()
		if err == nil {
			_ = c.(*tls.Conn).Handshake()
			time.Sleep(50 * time.Millisecond)
			_ = c.Close()
		}
	}()

	conn, err := dialTLS(ln.Addr().String(), true)
	if err != nil {
		t.Fatalf("dialTLS: %v", err)
	}
	_ = conn.Close()
	var _ net.Conn = conn
}
