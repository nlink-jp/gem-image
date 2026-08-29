# CLAUDE.md — gem-image

**Organization rules (mandatory): https://github.com/nlink-jp/.github/blob/main/CONVENTIONS.md**

## Project overview

CLI tool for image generation and editing using Vertex AI Gemini
(native image generation). Accepts text prompts and optional input images,
outputs generated/edited images to files. Part of util-series.

## Non-negotiable rules

- **Tests are mandatory** — write them with the implementation.
- **Never `go build` directly** — always `make build` (outputs to `dist/`).
- **Docs in sync** — update `README.md` and `README.ja.md` together.
- **Small, typed commits** — `feat:`, `fix:`, `test:`, `chore:`, etc.
- **Security first** — prompt injection defense (nlk/guard), no secrets in code.

## Build & test

```bash
make build          # → dist/gem-image
make test           # or: go test ./...
make build-all      # cross-compile 4 platforms (darwin arm64 only; no Intel)
make check          # vet → test → build
```

## Environment variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `GEMIMAGE_PROJECT` | Yes | — | GCP project ID |
| `GEMIMAGE_LOCATION` | No | `global` | Vertex AI location (Gemini 3 is global-only) |
| `GEMIMAGE_MODEL` | No | `gemini-3.1-flash-image` | Model name |

## Key dependencies

- `google.golang.org/genai` — Google Gemini SDK (Vertex AI backend)
- `github.com/nlink-jp/nlk` — guard (prompt injection protection)
- `github.com/spf13/cobra` — CLI framework
- `github.com/BurntSushi/toml` — config file parsing

## Architecture

- `internal/config/` — TOML + environment variable configuration
- `internal/client/` — Vertex AI Gemini client (GenerateContent with image modality)
- `internal/image/` — image file I/O, PNG/JPEG transcoding
- `internal/security/` — nlk/guard wrapper, path validation, magic byte verification
