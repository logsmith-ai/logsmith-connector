package proto

import (
	"bytes"
	"testing"
)

func TestVersionRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteVersion(&buf); err != nil {
		t.Fatalf("write: %v", err)
	}
	v, err := ReadVersion(&buf)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if v != ProtocolVersion {
		t.Fatalf("got %d want %d", v, ProtocolVersion)
	}
}
