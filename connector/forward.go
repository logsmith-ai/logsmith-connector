package connector

import (
	"io"
	"net"

	"github.com/logsmith-ai/logsmith-connector/proto"
)

// HandleStream reads the target header from an accepted tunnel stream, dials the
// target, and pumps bytes both directions until either side closes. It never
// inspects payload bytes. dial is injected (net.Dial in production).
func HandleStream(stream net.Conn, dial func(network, addr string) (net.Conn, error)) {
	defer stream.Close()

	addr, err := proto.ReadTargetHeader(stream)
	if err != nil {
		return
	}
	target, err := dial("tcp", addr)
	if err != nil {
		return
	}
	defer target.Close()

	done := make(chan struct{}, 2)
	go func() { _, _ = io.Copy(target, stream); done <- struct{}{} }()
	go func() { _, _ = io.Copy(stream, target); done <- struct{}{} }()
	<-done // first half-close ends the stream; deferred Close unblocks the other
}
