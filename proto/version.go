package proto

import "io"

// ProtocolVersion is the current wire-protocol version, sent as the first byte
// of every connector connection before the auth frame. The tunnel-server rejects
// connections whose first byte is not a version it supports. This gives the
// protocol a negotiation point so future wire changes are not flag-day breaks.
const ProtocolVersion byte = 1

// WriteVersion writes the current protocol-version byte. Called by the connector
// as the very first byte on the raw connection, before the auth frame.
func WriteVersion(w io.Writer) error {
	_, err := w.Write([]byte{ProtocolVersion})
	return err
}

// ReadVersion reads and returns the single protocol-version byte.
func ReadVersion(r io.Reader) (byte, error) {
	var b [1]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return 0, err
	}
	return b[0], nil
}
