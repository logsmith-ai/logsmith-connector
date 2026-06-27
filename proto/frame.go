package proto

import (
	"encoding/binary"
	"errors"
	"io"
)

const (
	maxAuthLen   = 4096
	maxTargetLen = 255
)

// WriteAuthFrame writes [4-byte BE len][token]. Sent once on the raw conn
// before yamux starts.
func WriteAuthFrame(w io.Writer, token string) error {
	if len(token) > maxAuthLen {
		return errors.New("proto: auth token too long")
	}
	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], uint32(len(token)))
	if _, err := w.Write(hdr[:]); err != nil {
		return err
	}
	_, err := io.WriteString(w, token)
	return err
}

// ReadAuthFrame reads a frame written by WriteAuthFrame.
func ReadAuthFrame(r io.Reader) (string, error) {
	var hdr [4]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return "", err
	}
	n := binary.BigEndian.Uint32(hdr[:])
	if n > maxAuthLen {
		return "", errors.New("proto: auth frame length exceeds limit")
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	return string(buf), nil
}

// WriteTargetHeader writes [2-byte BE len][addr]. First bytes of every tunnel
// stream the server opens.
func WriteTargetHeader(w io.Writer, addr string) error {
	if len(addr) > maxTargetLen {
		return errors.New("proto: target address too long")
	}
	var hdr [2]byte
	binary.BigEndian.PutUint16(hdr[:], uint16(len(addr)))
	if _, err := w.Write(hdr[:]); err != nil {
		return err
	}
	_, err := io.WriteString(w, addr)
	return err
}

// ReadTargetHeader reads a header written by WriteTargetHeader.
func ReadTargetHeader(r io.Reader) (string, error) {
	var hdr [2]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return "", err
	}
	n := binary.BigEndian.Uint16(hdr[:])
	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	return string(buf), nil
}
