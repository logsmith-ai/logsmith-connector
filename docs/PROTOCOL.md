# logsmith-connector wire protocol

This is the single source of truth for the connector ↔ tunnel-server wire
format. The connector and reference server in this repo, the production
tunnel-server, and gotham's token signer all conform to it.

## Transport

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

- `claimsJSON` = `{"connectorId":string,"workspaceId":string,"exp":int64}`
  (`exp` = unix seconds), serialized with no extra whitespace.
- Signature is Ed25519 over the **raw `claimsJSON` bytes**.
- base64url uses the URL alphabet **without padding**.
- The tunnel-server rejects when `now >= exp` or the signature fails.

## Per-stream target header (first bytes of every opened stream)

`[2-byte big-endian length N][N bytes UTF-8 "host:port"]`, `N <= 255`.

After the header, the stream carries **opaque application bytes** in both
directions. The connector dials `host:port` and copies bytes verbatim — it never
parses payload. This is an auditability guarantee: end-to-end TLS (e.g.
Kubernetes/Postgres/HTTPS) rides inside the stream and terminates at the real
target, never at the tunnel.
