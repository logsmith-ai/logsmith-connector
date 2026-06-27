package proto

import (
	"crypto/ed25519"
	"testing"
)

func mustKeys(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("genkey: %v", err)
	}
	return pub, priv
}

func TestSignVerifyRoundTrip(t *testing.T) {
	pub, priv := mustKeys(t)
	in := Claims{ConnectorID: "conn_42", WorkspaceID: "ws_1", Exp: 1000}
	tok, err := SignToken(priv, in)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	out, err := VerifyToken(pub, tok, 900)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if out != in {
		t.Fatalf("claims mismatch: got %+v want %+v", out, in)
	}
}

func TestVerifyExpired(t *testing.T) {
	pub, priv := mustKeys(t)
	tok, _ := SignToken(priv, Claims{ConnectorID: "c", WorkspaceID: "w", Exp: 1000})
	if _, err := VerifyToken(pub, tok, 1001); err == nil {
		t.Fatal("expected expiry error, got nil")
	}
}

func TestVerifyBadSignature(t *testing.T) {
	pub, priv := mustKeys(t)
	tok, _ := SignToken(priv, Claims{ConnectorID: "c", WorkspaceID: "w", Exp: 1000})
	otherPub, _ := mustKeys(t)
	_ = priv
	if _, err := VerifyToken(otherPub, tok, 900); err == nil {
		t.Fatal("expected signature error, got nil")
	}
	_ = pub
}

func TestVerifyMalformed(t *testing.T) {
	pub, _ := mustKeys(t)
	if _, err := VerifyToken(pub, "not-a-token", 0); err == nil {
		t.Fatal("expected malformed error, got nil")
	}
}
