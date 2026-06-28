# logsmith-connector

Open-source (Apache-2.0) outbound connector that lets Logsmith reach resources in
your private subnet — Kubernetes, telemetry, databases — **without inbound
firewall changes**. It dials out to Logsmith over TLS and forwards TCP streams to
private targets you choose. It is a dumb L4 byte-forwarder: it never inspects
your traffic, and end-to-end TLS terminates at your real target, not at Logsmith.

## Install & run

You need a `--server` (the Logsmith tunnel-server, `host:port`) and a `--token`
(your enrollment token).

### Docker

Images are published to the GitHub Container Registry for `linux/amd64` and
`linux/arm64`:

```bash
docker run -d --name logsmith-connector \
  ghcr.io/logsmith-ai/logsmith-connector:latest \
  --server tunnel.logsmith.ai:443 --token <enrollment-token>
```

### Binary

Download the archive for your OS/arch from the
[Releases](https://github.com/logsmith-ai/logsmith-connector/releases) page
(`linux` and `darwin`, `amd64` and `arm64`), then:

```bash
tar -xzf logsmith-connector_*_linux_amd64.tar.gz
./logsmith-connector --server tunnel.logsmith.ai:443 --token <enrollment-token>
```

Verify the download against `checksums.txt` from the same release.

### Helm (Kubernetes)

The chart is published as an OCI artifact on ghcr.io:

```bash
helm install logsmith-connector \
  oci://ghcr.io/logsmith-ai/charts/logsmith-connector \
  --set server=tunnel.logsmith.ai:443 \
  --set token=<enrollment-token>
```

For production, store the token in your own Secret and reference it instead of
passing it on the command line:

```bash
kubectl create secret generic logsmith-token --from-literal=token=<enrollment-token>

helm install logsmith-connector \
  oci://ghcr.io/logsmith-ai/charts/logsmith-connector \
  --set server=tunnel.logsmith.ai:443 \
  --set existingSecret=logsmith-token
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

## Releasing (maintainers)

CI runs on every push and PR (`go vet`, `go test -race`, `golangci-lint`,
`helm lint`). Releases are fully automated: push a semver tag and the
`Release` workflow builds and publishes everything.

```bash
git tag v0.1.0
git push origin v0.1.0
```

That produces, with no external secrets (uses the built-in `GITHUB_TOKEN`):

- `linux`/`darwin` × `amd64`/`arm64` binaries + `checksums.txt` on a GitHub Release
- a multi-arch Docker image at `ghcr.io/logsmith-ai/logsmith-connector`
- the Helm chart at `oci://ghcr.io/logsmith-ai/charts/logsmith-connector`

Release tooling lives in `.goreleaser.yaml` and `.github/workflows/`.