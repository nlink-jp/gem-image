# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [0.2.1] - 2026-05-22

### Changed

- **Darwin releases are now Developer ID signed and Apple-notarized.**
  `gem-image-v0.2.1-darwin-{amd64,arm64}.zip` carry full Apple
  Developer ID Application signatures and notarization tickets from
  Apple. End users on macOS no longer need to bypass Gatekeeper
  with right-click → Open or `xattr -d com.apple.quarantine` on
  first launch; local users who place `gem-image` under
  Dropbox-synced (or any other FileProvider-managed) paths are no
  longer killed by macOS's ad-hoc + provenance distrust policy.
  Pipeline: `scripts/codesign-darwin.sh` +
  `scripts/notarize-darwin.sh`, driven by `make package`. Adopts
  the org-wide convention in `nlink-jp/.github` CONVENTIONS.md
  §Code Signing.
- **Release zip filenames now embed the version**
  (`gem-image-vX.Y.Z-<os>-<arch>.zip`), aligning with the
  sibling util-series tools (json-filter, gem-search, gem-summary).
  Previous v0.2.0 assets used a version-less name.

No behaviour change to the binary itself — feature-wise this is
identical to v0.2.0.

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
