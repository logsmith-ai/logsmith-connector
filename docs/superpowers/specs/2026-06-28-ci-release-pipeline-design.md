# CI/CD & Release Pipeline — logsmith-connector

Date: 2026-06-28
Status: Approved

## Goal

Set up GitHub-native CI/CD so that pushing a `vX.Y.Z` tag produces, with no
external secrets:

1. `linux/arm64` binary
2. `linux/amd64` binary (plus darwin amd64/arm64 — free via GoReleaser)
3. multi-arch Docker image on ghcr.io
4. updated README with install/run instructions
5. a Helm chart published as an OCI artifact on ghcr.io

## Decisions

| Area | Choice | Rationale |
|------|--------|-----------|
| Release tooling | GoReleaser | One config → multi-arch binaries, Docker, checksums, GitHub Release |
| Docker registry | `ghcr.io/logsmith-ai/logsmith-connector` | GitHub-only, uses `GITHUB_TOKEN`, no secrets |
| Helm distribution | OCI chart → `oci://ghcr.io/logsmith-ai/charts` | Modern, no `gh-pages` branch |
| Merge strategy | Commit on branch → PR to `main` → merge → tag | Review gate before release |

## Repository additions

```
.github/workflows/ci.yml         # PR/push: vet, test -race, golangci-lint, helm lint
.github/workflows/release.yml     # tag v*: GoReleaser + Helm OCI push
.goreleaser.yaml                  # builds, archives, docker, checksums, release notes
Dockerfile                        # distroless final image (consumed by GoReleaser)
charts/logsmith-connector/        # Helm chart
README.md                         # updated instructions
```

## ci.yml

Triggers: pull_request, push to `main` and feature branches.

Steps:
- `go vet ./...`
- `go test ./... -race`
- `golangci-lint run` (code already uses a `//nolint:gosec` directive)
- `helm lint charts/logsmith-connector` + `helm template` smoke check

This workflow gates the PR into `main`.

## release.yml

Trigger: push of tag matching `v*.*.*`.

Permissions: `contents: write`, `packages: write`.

Jobs:
1. **goreleaser** — runs GoReleaser, which:
   - builds `linux/{amd64,arm64}` and `darwin/{amd64,arm64}` static binaries
   - produces `.tar.gz` archives + `checksums.txt`
   - builds and pushes multi-arch Docker image to
     `ghcr.io/logsmith-ai/logsmith-connector` tagged `:vX.Y.Z`, `:X.Y`, `:latest`
   - creates the GitHub Release with notes + artifacts
2. **helm** — `helm package` the chart at the tag version, then
   `helm push` to `oci://ghcr.io/logsmith-ai/charts`.

Auth: `GITHUB_TOKEN` for both ghcr image and ghcr OCI chart. No external secrets.

## GoReleaser config (`.goreleaser.yaml`)

- `builds`: single binary `logsmith-connector` from `./cmd/logsmith-connector`,
  `CGO_ENABLED=0`, `-trimpath`, ldflags stamping version/commit/date, GOOS
  `linux,darwin`, GOARCH `amd64,arm64`.
- `archives`: tar.gz, name template `{{.ProjectName}}_{{.Os}}_{{.Arch}}`.
- `checksum`: `checksums.txt`.
- `dockers_v2` / `dockers` + `docker_manifests`: multi-arch image to ghcr.io.
- `release`: GitHub, auto-generated changelog.

## Dockerfile

Multi-stage is unnecessary because GoReleaser builds the binary and only uses
the Dockerfile to assemble the image. Final stage = `gcr.io/distroless/static`
(or `scratch`) copying the prebuilt binary; `ENTRYPOINT ["/logsmith-connector"]`.

## Helm chart (`charts/logsmith-connector`)

- `Chart.yaml`: name `logsmith-connector`, version + appVersion set from tag at
  release time.
- `Deployment`: 1 replica (outbound dialer, nothing inbound), runs the binary
  with `--server` and `--token`.
- `--token`: from a Secret. Chart can create one from `values.token`, or use
  `values.existingSecret` (recommended for production).
- `values.yaml`: `replicaCount`, `image.{repository,tag,pullPolicy}`, `server`,
  `token`, `existingSecret`, `resources`.
- No Service/Ingress — consistent with "no inbound firewall changes".

## README updates

Replace the single Docker Hub run line with sections:
- **Docker**: `docker run ghcr.io/logsmith-ai/logsmith-connector:latest --server ... --token ...`
- **Binary**: download per-arch archive from the Releases page.
- **Helm**: `helm install logsmith-connector oci://ghcr.io/logsmith-ai/charts/logsmith-connector --set server=... --set token=...`
- **Releasing** (maintainer note): push a `vX.Y.Z` tag; the pipeline does the rest.

## Rollout

1. Commit all of the above to `feat/connector-data-plane`.
2. Push; open PR to `main`; CI must pass.
3. Merge PR; push tag `v0.1.0` → release pipeline publishes everything.
