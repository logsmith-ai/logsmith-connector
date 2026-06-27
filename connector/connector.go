package connector

import (
	"net"

	"github.com/hashicorp/yamux"
	"github.com/logsmith-ai/logsmith-connector/proto"
)

// Options configures a connector run.
type Options struct {
	Token string
	// Dial opens connections to private targets. Defaults to net.Dial when nil.
	Dial func(network, addr string) (net.Conn, error)
}

// Run authenticates on conn, starts a yamux client session, and serves tunnel
// streams the server opens until the session ends. conn is the established
// transport (TLS in production, a pipe in tests).
func Run(conn net.Conn, opts Options) error {
	if opts.Dial == nil {
		opts.Dial = net.Dial
	}
	if err := proto.WriteAuthFrame(conn, opts.Token); err != nil {
		return err
	}
	sess, err := yamux.Client(conn, nil)
	if err != nil {
		return err
	}
	defer sess.Close()
	for {
		stream, err := sess.AcceptStream()
		if err != nil {
			return err // session closed
		}
		go HandleStream(stream, opts.Dial)
	}
}
