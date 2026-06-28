package proto

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
)

// Claims is the payload of a connector enrollment token. Signed by gotham,
// verified by the tunnel-server. Kept tiny on purpose.
type Claims struct {
	ConnectorID string `json:"connectorId"`
	WorkspaceID string `json:"workspaceId"`
	Exp         int64  `json:"exp"` // unix seconds
}

var b64 = base64.RawURLEncoding

// marshalClaims produces the canonical JSON that is signed and that any
// cross-language signer (e.g. gotham's TypeScript) MUST reproduce byte-for-byte:
// struct field order (connectorId, workspaceId, exp), exp as a JSON number,
// UTF-8, HTML escaping DISABLED so '<' '>' '&' stay literal (matching a naive
// JSON.stringify), and no insignificant whitespace.
func marshalClaims(c Claims) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(c); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil // Encode appends a newline
}

// SignToken serializes claims to JSON and signs the raw JSON bytes with Ed25519.
// Output: base64url(claimsJSON) + "." + base64url(signature).
func SignToken(priv ed25519.PrivateKey, c Claims) (string, error) {
	payload, err := marshalClaims(c)
	if err != nil {
		return "", err
	}
	sig := ed25519.Sign(priv, payload)
	return b64.EncodeToString(payload) + "." + b64.EncodeToString(sig), nil
}

// VerifyToken checks the signature and expiry. nowUnix is the current unix time;
// the token is rejected when nowUnix >= Exp.
func VerifyToken(pub ed25519.PublicKey, token string, nowUnix int64) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return Claims{}, errors.New("proto: malformed token")
	}
	payload, err := b64.DecodeString(parts[0])
	if err != nil {
		return Claims{}, errors.New("proto: bad payload encoding")
	}
	sig, err := b64.DecodeString(parts[1])
	if err != nil {
		return Claims{}, errors.New("proto: bad signature encoding")
	}
	if !ed25519.Verify(pub, payload, sig) {
		return Claims{}, errors.New("proto: signature verification failed")
	}
	var c Claims
	if err := json.Unmarshal(payload, &c); err != nil {
		return Claims{}, errors.New("proto: bad claims json")
	}
	if nowUnix >= c.Exp {
		return Claims{}, errors.New("proto: token expired")
	}
	return c, nil
}
