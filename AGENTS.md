# AGENTS.md — gem-image

## Project summary

CLI tool for image generation and editing using Vertex AI Gemini 2.5 Flash
(native image generation). Part of util-series.

## Build commands

```bash
make build          # Build → dist/gem-image
make test           # Run all tests
make build-all      # Cross-compile for 5 platforms
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
- `GEMIMAGE_LOCATION` (optional, default: us-central1) — Vertex AI region
- `GEMIMAGE_MODEL` (optional, default: gemini-2.5-flash-image) — model name
- Falls back to `GOOGLE_CLOUD_PROJECT` / `GOOGLE_CLOUD_LOCATION`

## Gotchas

- Model always returns PNG regardless of requested format. JPEG conversion
  is handled client-side in `internal/image/output.go`.
- `-o` is required — no stdout binary output (terminal corruption risk).
- Prompt injection defense uses nlk/guard (nonce-tagged XML), not the
  project-local isolation package pattern from gem-cli.
- Authentication via ADC (`gcloud auth application-default login`).
