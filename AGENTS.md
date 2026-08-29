# AGENTS.md — gem-image

## Project summary

CLI tool for image generation and editing using Vertex AI Gemini
(native image generation). Part of util-series.

## Build commands

```bash
make build          # Build → dist/gem-image
make test           # Run all tests
make build-all      # Cross-compile for 5 platforms
make verify-release  # gate: .notarized marker + freshness (run before upload)
make check          # vet → test → build
make clean          # Remove dist/
```

## Module path

`github.com/nlink-jp/gem-image`

## Key structure

```
gem-image/
├── main.go                 ← entry point
├── cmd/root.go             ← cobra command, flag definitions, orchestration
├── internal/
│   ├── config/             ← TOML + GEMIMAGE_* env var loading
│   ├── client/             ← Vertex AI Gemini client (GenerateContent)
│   ├── image/              ← image file I/O, format conversion
│   └── security/           ← nlk/guard wrapper, path/magic-byte validation
├── Makefile
└── docs/                   ← RFP, architecture (ADR), detailed design
```

## Environment variables

- `GEMIMAGE_PROJECT` (required) — GCP project ID
- `GEMIMAGE_LOCATION` (optional, default: global) — Vertex AI location
- `GEMIMAGE_MODEL` (optional, default: gemini-3.1-flash-image) — model name
- Falls back to `GOOGLE_CLOUD_PROJECT` / `GOOGLE_CLOUD_LOCATION`

## Gotchas

- **Gemini 3 models are served from the `global` endpoint only** — a regional
  location 404s. `internal/client` appends a hint to such 404s. Gemini 2.5
  models still answer regionally.
- Response encoding varies by model: PNG for most, **JPEG for
  `gemini-3.1-flash-lite-image`**. `internal/image/output.go` transcodes in
  both directions so the written file matches the requested format.
- `-o` is required — no stdout binary output (terminal corruption risk).
- Prompt injection defense uses nlk/guard (nonce-tagged XML), not the
  project-local isolation package pattern from gem-cli.
- Authentication via ADC (`gcloud auth application-default login`).
