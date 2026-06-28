# GoReleaser builds the static binary and passes it into this build context,
# so this is a single-stage image that just wraps the prebuilt binary.
FROM gcr.io/distroless/static:nonroot

COPY logsmith-connector /usr/local/bin/logsmith-connector

USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/logsmith-connector"]
