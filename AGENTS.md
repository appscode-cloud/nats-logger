# AGENTS.md

This file provides guidance to coding agents (e.g. Claude Code, claude.ai/code) when working with code in this repository.

## Repository purpose

Go module `go.bytebuilders.dev/nats-logger` — utilities for piping logs through NATS. Comprises a small producer (`main.go`) that streams logs into NATS and a consumer that drains them. Used as a building block for SSH-exec-style demos and log forwarding in the AppsCode platform.

## Architecture

- `main.go` — producer entry point.
- `consumer/` — NATS consumer that drains the stream.
- `internal/` — shared internal packages.
- `nats-logger` — pre-built binary checked in.
- `hack/`, `Makefile` — build harness.

## Common commands

- `make build` — Go build.
- `go run .` — run the producer directly.

## Conventions

- Module path is `go.bytebuilders.dev/nats-logger` (vanity URL).
- License: `LICENSE.md`. Sign off commits (`git commit -s`).
- Small utility; resist scope creep. If this evolves into a broader logging product, separate producer and consumer into their own modules first.
- README points at a personal gist for the original SSH-exec context — keep that link or expand it inline if rewriting.
