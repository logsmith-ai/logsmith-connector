package proto

import (
	"bytes"
	"strings"
	"testing"
)

func TestAuthFrameRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteAuthFrame(&buf, "tok-abc"); err != nil {
		t.Fatalf("write: %v", err)
	}
	got, err := ReadAuthFrame(&buf)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if got != "tok-abc" {
		t.Fatalf("got %q want %q", got, "tok-abc")
	}
}

func TestAuthFrameTooLong(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteAuthFrame(&buf, strings.Repeat("x", 4097)); err == nil {
		t.Fatal("expected error for oversized token")
	}
}

func TestTargetHeaderRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteTargetHeader(&buf, "10.0.1.5:9090"); err != nil {
		t.Fatalf("write: %v", err)
	}
	got, err := ReadTargetHeader(&buf)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if got != "10.0.1.5:9090" {
		t.Fatalf("got %q want %q", got, "10.0.1.5:9090")
	}
}

func TestTargetHeaderTooLong(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteTargetHeader(&buf, strings.Repeat("h", 256)); err == nil {
		t.Fatal("expected error for oversized target")
	}
}

func TestReadTargetHeaderRejectsOversized(t *testing.T) {
	var buf bytes.Buffer
	buf.Write([]byte{0x01, 0x00}) // big-endian 256, exceeds maxTargetLen (255)
	buf.Write(make([]byte, 256))
	if _, err := ReadTargetHeader(&buf); err == nil {
		t.Fatal("expected error for over-limit declared length")
	}
}
