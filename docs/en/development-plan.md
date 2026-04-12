# Development Plan: gem-image

> Generated: 2026-04-12
> Status: Draft

## Phase Overview

| Phase | Content | Deliverables |
|-------|---------|-------------|
| Phase 1 | Design | RFP, architecture, detailed design, development plan |
| Phase 2 | Core implementation | Working binary + tests |
| Phase 3 | Release | README, CHANGELOG, release binaries |

---

## Phase 1: Design (Current Phase)

### Tasks

| # | Task | Deliverable | Status |
|---|------|------------|--------|
| 1.1 | RFP creation | `docs/ja/gem-image-rfp.ja.md`, `docs/en/gem-image-rfp.md` | Done |
| 1.2 | Architecture document | `docs/ja/architecture.ja.md`, `docs/en/architecture.md` | Done |
| 1.3 | Detailed design | `docs/ja/design.ja.md`, `docs/en/design.md` | Done |
| 1.4 | Development plan | `docs/ja/development-plan.ja.md`, `docs/en/development-plan.md` | Done |
| 1.5 | Review & approval | — | Pending |

### Completion Criteria

- All design documents available in both Japanese and English
- Review feedback incorporated

---

## Phase 2: Core Implementation

### Prerequisites

- Phase 1 documents approved
- Vertex AI API enabled
- ADC authentication configured

### Tasks

Implementation follows this order. Each task targets one commit.

| # | Task | Depends On | Description |
|---|------|-----------|-------------|
| 2.1 | Project scaffold | — | go mod init, Makefile, directory structure, .gitignore |
| 2.2 | config package | 2.1 | TOML loading + env override + tests |
| 2.3 | security package | 2.1 | Path validation, magic byte verification, nlk/guard wrapper + tests |
| 2.4 | image package | 2.3 | Image read/write, format resolution + tests |
| 2.5 | client package | 2.2 | Gemini API client + tests |
| 2.6 | CLI integration | 2.3, 2.4, 2.5 | cmd/root.go flag parsing and orchestration |
| 2.7 | E2E tests | 2.6 | Integration tests with actual API |
| 2.8 | AGENTS.md / CLAUDE.md | 2.7 | Project meta documentation |

### Dependency Graph

```
2.1 (scaffold)
 |-- 2.2 (config)
 |    +-- 2.5 (client) --+
 +-- 2.3 (security)      |
      +-- 2.4 (image) ---+
                          v
                     2.6 (CLI integration)
                          |
                          v
                     2.7 (E2E)
                          |
                          v
                     2.8 (meta docs)
```

### Completion Criteria

- `make build` produces a binary
- `go test ./...` passes all tests
- E2E tests pass with real data:
  - Text to image generation
  - Image + text to edited image
  - Multiple input images
  - Invalid input rejection
  - Token info displayed on stderr

---

## Phase 3: Release

### Prerequisites

- Phase 2 E2E tests passed
- Binary simulation on actual machine completed

### Tasks

| # | Task | Description |
|---|------|-------------|
| 3.1 | README.md | English README (project root) |
| 3.2 | README.ja.md | Japanese README (project root) |
| 3.3 | CHANGELOG.md | v0.1.0 entry |
| 3.4 | Binary simulation test | Final verification with `make build` binary |
| 3.5 | Git repository creation | nlink-jp/gem-image |
| 3.6 | Release | Tag v0.1.0, `gh release create`, binary upload |
| 3.7 | util-series submodule registration | Add submodule + update pointer |
| 3.8 | Org profile update | Add gem-image to `.github/profile/README.md` |
| 3.9 | Run check-org.sh | Final verification |

### Completion Criteria

- GitHub release is published
- util-series submodule pointer updated
- gem-image listed in org profile
- `check-org.sh` passes

---

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Gemini 2.5 Flash Image API spec changes | Implementation rework | Keep preview model name configurable |
| Excessive safety filter blocking | Usability degradation | Clear error messages indicating cause |
| nlk guard package API changes | Build errors | Pin version in go.mod |
| genai SDK major version upgrade | Compatibility issues | Coordinate with gem-cli updates |
