# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [0.2.0] - 2026-04-12

### Added

- Automatic retry with exponential backoff (max 3 retries) for transient API errors (429/500/503)
- Progress feedback on stderr (`Generating image... done.`) during API calls
- `Generator` interface for testable client design

### Security

- Prevent accidental overwrite of existing files; require `--force` to overwrite
- Add Image Bomb prevention: check PNG dimensions before decoding to prevent OOM attacks

## [0.1.0] - 2026-04-12

### Added

- Text-to-image generation via Vertex AI Gemini 2.5 Flash
- Image editing with input images (`-i`, repeatable)
- PNG and JPEG output with automatic format detection from file extension
- PNG-to-JPEG conversion when model returns PNG but JPEG is requested
- Prompt injection protection using nlk/guard nonce-tagged XML wrapping
- Input image validation (magic bytes, file size, path traversal prevention)
- Token usage display on stderr
- Configuration via environment variables (`GEMIMAGE_*`) and TOML config file
- gem-cli compatible configuration pattern (env > config file > defaults)
