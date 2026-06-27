package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net"
	"time"

	"github.com/logsmith-ai/logsmith-connector/connector"
)

func dialTLS(server string, insecure bool) (net.Conn, error) {
	return tls.DialWithDialer(
		&net.Dialer{Timeout: 10 * time.Second},
		"tcp", server,
		&tls.Config{InsecureSkipVerify: insecure}, //nolint:gosec // gated behind --insecure dev flag
	)
}

func main() {
	server := flag.String("server", "", "tunnel-server host:port")
	token := flag.String("token", "", "enrollment token")
	insecure := flag.Bool("insecure", false, "skip TLS verification (dev only)")
	flag.Parse()

	if *server == "" || *token == "" {
		log.Fatal("--server and --token are required")
	}

	for { // reconnect loop
		conn, err := dialTLS(*server, *insecure)
		if err != nil {
			log.Printf("dial %s: %v; retrying in 5s", *server, err)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Printf("connected to %s", *server)
		if err := connector.Run(conn, connector.Options{Token: *token}); err != nil {
			log.Printf("session ended: %v; reconnecting in 5s", err)
		}
		time.Sleep(5 * time.Second)
	}
}
