# logsmith-connector

Open-source (Apache-2.0) outbound connector that lets Logsmith reach resources in
your private subnet — Kubernetes, telemetry, databases — **without inbound
firewall changes**. It dials out to Logsmith over TLS and forwards TCP streams to
private targets you choose. It is a dumb L4 byte-forwarder: it never inspects
your traffic, and end-to-end TLS terminates at your real target, not at Logsmith.

## Run

```bash
docker run -d logsmith/connector --server tunnel.logsmith.ai:443 --token <enrollment-token>
```

## Layout

- `proto/` — wire protocol (tokens, frames). Single source of truth.
- `connector/` — the agent: yamux client session + stream forwarder.
- `refserver/` — reference tunnel-server for local dev and audits.
- `cmd/logsmith-connector/` — the CLI binary.
- `docs/PROTOCOL.md` — the wire-protocol spec.

## Build & test

```bash
go build ./...
go test ./...
```