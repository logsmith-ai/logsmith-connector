// Package proto defines the logsmith-connector wire protocol: the auth frame,
// the per-stream target header, and the signed-token format. It is the single
// source of truth shared by the connector agent, the reference tunnel-server,
// and the production tunnel-server.
package proto
