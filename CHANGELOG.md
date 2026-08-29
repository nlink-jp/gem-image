# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Changed

- **Default model is now `gemini-3.1-flash-image`** (Nano Banana 2, was
  `gemini-2.5-flash-image`), ahead of the Vertex AI Gemini 2.5 retirement
  from 2026-10-16. Gemini 2.5 still works when set explicitly via config,
  `GEMIMAGE_MODEL`, or `-m`.
- **Default location is now `global`** (was `us-central1`): Vertex AI serves
  the Gemini 3 family only from the global endpoint — regional endpoints
  return 404 for them. Pin a regional location explicitly when going back to
  `gemini-2.5-flash-image`.
- Updated google.golang.org/genai to v1.70.0.

### Added

- Actionable error hint: requesting a Gemini 3 model from a regional endpoint
  used to fail with a bare `404 NOT_FOUND`; the error now explains that
  Gemini 3 models require `location = "global"`.
- ADR-009 in `docs/{en,ja}/architecture*.md` records the migration and the
  global-endpoint constraint.

### Fixed

- **Output format is now honoured for every model.** Conversion only ran
  PNG to JPEG, so a model that returns JPEG (`gemini-3.1-flash-lite-image`)
  wrote JPEG bytes into a `.png` file. Both directions are converted now, and
  JPEG-to-PNG is written at 8 bits per channel instead of being widened to 16.

## [0.3.0] - 2026-07-12

### Removed

- **darwin/amd64 (Intel) pre-built binary.** macOS releases now ship
  **arm64 only**, per the org-wide policy (darwin is Apple-Silicon only; no
  universal binaries). Intel Mac users can build from source.

### Changed

- **Linux release archives are now `.tar.gz`** (darwin/windows remain `.zip`),
  per `nlink-jp/.github` CONVENTIONS.md §Release Archive Standard.
- **`LICENSE` is now bundled** in every release archive alongside `README.md`.
- **darwin code-signature identifier** is now the canonical `gem-image`.

No change to the binary's behaviour — a packaging / build-config release.

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
