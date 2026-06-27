package refserver

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/hashicorp/yamux"
	"github.com/logsmith-ai/logsmith-connector/proto"
)

// Server is an open, minimal tunnel-server for local dev, audits, and tests.
// The production tunnel-server lives in a separate repo but mirrors this shape.
type Server struct {
	pub   ed25519.PublicKey
	nowFn func() int64

	mu       sync.RWMutex
	sessions map[string]*yamux.Session // connectorID -> live session
}

// New builds a Server that verifies tokens against pub. nowFn supplies current
// unix seconds (injectable for tests).
func New(pub ed25519.PublicKey, nowFn func() int64) *Server {
	return &Server{pub: pub, nowFn: nowFn, sessions: map[string]*yamux.Session{}}
}

// Handle authenticates one connector connection and registers its session.
// It returns when the session ends or auth fails.
func (s *Server) Handle(conn net.Conn) error {
	ver, err := proto.ReadVersion(conn)
	if err != nil {
		conn.Close()
		return err
	}
	if ver != proto.ProtocolVersion {
		conn.Close()
		return fmt.Errorf("refserver: unsupported protocol version %d", ver)
	}
	token, err := proto.ReadAuthFrame(conn)
	if err != nil {
		conn.Close()
		return err
	}
	claims, err := proto.VerifyToken(s.pub, token, s.nowFn())
	if err != nil {
		conn.Close()
		return err
	}
	sess, err := yamux.Server(conn, nil)
	if err != nil {
		conn.Close()
		return err
	}
	s.register(claims.ConnectorID, sess)
	defer s.unregister(claims.ConnectorID, sess)

	<-sess.CloseChan() // block until the session dies
	return nil
}

// Dial opens a stream to a registered connector and writes the target header,
// returning the stream ready for opaque byte exchange.
func (s *Server) Dial(connectorID, target string) (net.Conn, error) {
	s.mu.RLock()
	sess := s.sessions[connectorID]
	s.mu.RUnlock()
	if sess == nil {
		return nil, errors.New("refserver: connector not connected")
	}
	stream, err := sess.OpenStream()
	if err != nil {
		return nil, err
	}
	if err := proto.WriteTargetHeader(stream, target); err != nil {
		stream.Close()
		return nil, err
	}
	return stream, nil
}

// Has reports whether a connector is currently registered.
func (s *Server) Has(connectorID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[connectorID] != nil
}

func (s *Server) register(id string, sess *yamux.Session) {
	s.mu.Lock()
	s.sessions[id] = sess
	s.mu.Unlock()
}

func (s *Server) unregister(id string, sess *yamux.Session) {
	s.mu.Lock()
	if s.sessions[id] == sess {
		delete(s.sessions, id)
	}
	s.mu.Unlock()
}
