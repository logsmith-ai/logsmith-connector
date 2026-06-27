# logsmith-connector wire protocol

This is the single source of truth for the connector ↔ tunnel-server wire
format. The connector and reference server in this repo, the production
tunnel-server, and gotham's token signer all conform to it.

## Transport

0. The connector sends a single **protocol-version byte** (`0x01`) as the very
   first byte on the raw connection. The tunnel-server reads it and rejects the
   connection if it is not a version it supports. All later wire changes bump
   this byte and negotiate against it — v1 is the first version.
1. The connector dials **outbound** TCP+TLS to the tunnel-server. Nothing dials
   into the connector.
2. On the raw connection, the connector sends one **auth frame** (below).
3. The tunnel-server verifies it, then **both sides start yamux** on the same
   connection: connector = `yamux.Client`, tunnel-server = `yamux.Server`.
4. yamux is bidirectional. The **tunnel-server opens** streams toward the
   connector to establish tunnels; the connector **accepts** them.

## Auth frame (once, before yamux)

`[4-byte big-endian length N][N bytes UTF-8 token]`, `N <= 4096`.

## Token

`base64url(claimsJSON) + "." + base64url(ed25519_signature)`

- Signature is Ed25519 over the **raw `claimsJSON` bytes**.
- base64url uses the URL alphabet **without padding**.
- The tunnel-server rejects when `now >= exp` or the signature fails.

The signed `claimsJSON` is **canonical** and a cross-language signer MUST
reproduce it byte-for-byte (the signature is over these raw bytes):

- Key order: `connectorId`, then `workspaceId`, then `exp`.
- `exp` is a JSON **number** (not a string).
- UTF-8; non-ASCII is NOT `\uXXXX`-escaped.
- HTML escaping is **disabled**: `<`, `>`, `&` stay literal (matches a naive
  `JSON.stringify`; note Go's `json.Marshal` escapes these by default — the
  reference Go signer disables it via `Encoder.SetEscapeHTML(false)`).
- No insignificant whitespace.

### Reference vector

- Ed25519 seed (32 bytes): `0x01 0x02 0x03 … 0x20` (byte `i+1` for `i` in `0..31`).
- Claims: `connectorId="conn_ref"`, `workspaceId="ws_ref"`, `exp=1700000000`.
- Canonical claimsJSON: `{"connectorId":"conn_ref","workspaceId":"ws_ref","exp":1700000000}`
- Token (Ed25519 is deterministic, so this is stable): `eyJjb25uZWN0b3JJZCI6ImNvbm5fcmVmIiwid29ya3NwYWNlSWQiOiJ3c19yZWYiLCJleHAiOjE3MDAwMDAwMDB9.0QPtJHdbJiqICxwyIJ1OE0WJWOoFvAaFifVg45219LLCBezHt_Hzbq5Sjt3DtOuCsRUpq1j1c6dVjQwpaLYwCQ`

Any implementation that produces this exact token from this seed and claims is
wire-compatible.

## yamux configuration

Both sides use **library-default** yamux config (default window size, keepalive,
accept backlog). A conforming implementation MUST use yamux defaults — no custom
tuning — so both sides agree on flow-control parameters without negotiation.

## Half-close behavior (v1)

When one direction of a tunnel stream closes, the connector currently tears down
both directions rather than performing a graceful half-close, so bytes still
buffered in the opposite direction may be truncated. Full request/response cycles
complete; streaming protocols that rely on independent half-close should be
aware. (A future version may switch to `CloseWrite` + drain.)

## Per-stream target header (first bytes of every opened stream)

`[2-byte big-endian length N][N bytes UTF-8 "host:port"]`, `N <= 255`.

After the header, the stream carries **opaque application bytes** in both
directions. The connector dials `host:port` and copies bytes verbatim — it never
parses payload. This is an auditability guarantee: end-to-end TLS (e.g.
Kubernetes/Postgres/HTTPS) rides inside the stream and terminates at the real
target, never at the tunnel.
