# gem-image

CLI tool for image generation and editing using Vertex AI Gemini (native image generation).

Generate images from text prompts or edit existing images via the command line.
Designed for batch workflows through shell scripts and pipelines.

## Prerequisites

- **Google Cloud project** with the Vertex AI API enabled
- **Application Default Credentials** — run `gcloud auth application-default login`

## Installation

```bash
git clone https://github.com/nlink-jp/gem-image.git
cd gem-image
make build
# Binary: dist/gem-image
```

## Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `GEMIMAGE_PROJECT` | Yes | — | GCP project ID |
| `GEMIMAGE_LOCATION` | No | `global` | Vertex AI location |
| `GEMIMAGE_MODEL` | No | `gemini-3.1-flash-image` | Gemini model name |

Falls back to `GOOGLE_CLOUD_PROJECT` / `GOOGLE_CLOUD_LOCATION` if tool-specific
variables are not set.

Alternatively, create `~/.config/gem-image/config.toml`:

```toml
[gcp]
project  = "your-project-id"
location = "global"

[model]
name = "gemini-3.1-flash-image"
```

Priority: CLI flags > environment variables > config file > defaults.

### Models

| Model | Endpoint | Notes |
|-------|----------|-------|
| `gemini-3.1-flash-image` (default) | `global` only | Balanced quality and cost; returns PNG |
| `gemini-3-pro-image` | `global` only | Highest quality, roughly 2x the price; returns PNG |
| `gemini-3.1-flash-lite-image` | `global` only | Cheapest; **returns JPEG**, transcoded when PNG output is requested |
| `gemini-2.5-flash-image` | `global`, `us-central1` | Previous default; the Gemini 2.5 family retires on Vertex AI from 2026-10-16 |

**Vertex AI serves the Gemini 3 family from the global endpoint only** — a
regional `location` returns `404 NOT_FOUND` for them. Set
`location = "us-central1"` explicitly if you pin the model back to
`gemini-2.5-flash-image`.

## Usage

```bash
# Generate an image from a text prompt
gem-image -p "A cat sitting on a windowsill" -o cat.png

# Edit an existing image
gem-image -p "Add a rainbow in the sky" -i photo.png -o edited.png

# Multiple input images
gem-image -p "Combine these into a collage" -i a.png -i b.png -o collage.png

# JPEG output (auto-detected from extension)
gem-image -p "A sunset over the ocean" -o sunset.jpg

# Explicit format flag
gem-image -p "A mountain landscape" -o landscape.bin --format jpeg

# Stdin prompt (pipeline)
echo "A minimalist logo for a coffee shop" | gem-image -o logo.png

# Override model
gem-image -p "A watercolor painting" -o art.png -m gemini-3-pro-image
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--prompt` | `-p` | — | Image generation prompt (stdin if omitted) |
| `--input` | `-i` | — | Input image path (repeatable) |
| `--output` | `-o` | — | Output file path (required) |
| `--format` | — | `png` | Output format: `png` or `jpeg` |
| `--config` | `-c` | — | Config file path |
| `--model` | `-m` | — | Model name override |
| `--force` | — | `false` | Overwrite existing output file |
| `--debug` | — | `false` | Enable debug output |

### Output format resolution

1. If `-o` has a `.png`/`.jpg`/`.jpeg` extension → use that format
2. Otherwise, use `--format` flag value
3. Default: `png`

Each model picks the encoding of what it returns — PNG for most, JPEG for
`gemini-3.1-flash-lite-image`. Whatever comes back is transcoded client-side so
the written file always matches the resolved format.

### Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Input validation error |
| 3 | API error |
| 4 | Safety filter block |

### Token usage

Token consumption is displayed on stderr after each request:

```
tokens: input=218 output=1290 total=1508
```

## Security

- **Prompt injection protection** — user prompts are wrapped with [nlk/guard](https://github.com/nlink-jp/nlk) nonce-tagged XML before API submission
- **Input validation** — image files are verified by magic bytes (not just extension)
- **Path traversal prevention** — all file paths are normalized and validated
- **Overwrite protection** — existing files are not overwritten unless `--force` is specified
- **Image Bomb prevention** — image dimensions are checked before decoding to prevent OOM attacks
- **No secrets in output** — project IDs and tokens are never logged

## License

MIT
