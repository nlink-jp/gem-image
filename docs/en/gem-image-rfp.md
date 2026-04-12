# RFP: gem-image

> Generated: 2026-04-12
> Status: Draft

## 1. Problem Statement

Gemini 2.5 Flash (Nano banana) on Vertex AI offers native image generation and
editing capabilities, but the web-based Gemini interface is impractical for
batch-style usage. gem-image exposes these capabilities through a CLI, enabling
integration into shell scripts and pipelines.

- **Target user:** Personal use (nlink-jp developer)
- **Complementary tool:** gem-cli handles text generation and image description.
  Image description is explicitly out of scope for gem-image

## 2. Functional Specification

### Commands / API Surface

```
gem-image -p "prompt" [-i input.png ...] -o output.png [--format png|jpeg]
```

No subcommands. Image generation vs. editing is determined by the presence of
input images (`-i`).

### Flags

| Flag | Required | Description |
|------|----------|-------------|
| `-p` | * | Prompt. Falls back to stdin if omitted. `-p` takes precedence |
| `-i` | No | Input image path. Repeatable for multiple images |
| `-o` | Yes | Output file path |
| `--format` | No | Output format `png`\|`jpeg` (default: `png`). `-o` extension takes precedence if present |

\* Error if `-p` is omitted and stdin is not a pipe.

### Input / Output

- **Input:** Prompt (`-p` or stdin), image files (`-i`)
- **Output:** Image file (`-o`, required)
- **Status output:** Token consumption displayed on stderr (for cost tracking)

### Configuration

Follows gem-cli conventions:

- Environment variables > config file (in order of precedence)
- Vertex AI project ID, region, etc.

### External Dependencies

- Vertex AI API (`aiplatform.googleapis.com`)
- Gemini 2.5 Flash (image generation model)
- nlk library (guard package — prompt injection protection)

## 3. Design Decisions

### Language / Framework

- **Go** — Consistent with gem-cli. Maintains build and distribution uniformity

### Design Principles

- **UNIX philosophy** — Do one thing well. Implemented as a standalone tool
- **Relationship with gem-cli** — Independent tool. Configuration and authentication follow gem-cli conventions
- **One request, one operation** — Batch processing is controlled by the shell. The tool itself stays simple

### Security Design (Security First)

This tool accepts user input and sends it to an external API. Security is the
top design priority.

| Risk | Mitigation |
|------|------------|
| Prompt injection | nlk/guard package: nonce-tagged XML wrapping. New Tag per request |
| `-i` path traversal | Path normalization, symlink resolution, existence check |
| `-o` path traversal | Path normalization and validation |
| Malicious image files | Magic byte validation, file size limits |
| Prompt length abuse | Input length limits |
| Config file tampering | File permission verification |
| Credential leakage | ADC only. Never log tokens or project IDs |
| Output file permissions | Created with 0644 |
| Safety filter blocks | Clear error messages with appropriate exit codes |

#### nlk/guard Integration

```go
tag := guard.NewTag()
wrapped, err := tag.Wrap(userPrompt)
systemPrompt := tag.Expand("Image generation request is in {{DATA_TAG}} tags. ...")
```

- User prompts (`-p` / stdin) are wrapped with nonce-tagged XML via `Wrap()`
- System prompt uses `Expand()` to declare tag boundaries
- Tags are generated fresh per request (never reused)

### Out of Scope

- **Image description** — Handled by gem-cli
- **Video generation** — Different model responsibility (may become gem-video in the future)
- **Image processing library features** — No pixel-level resize, filters, etc.
- **Binary output to stdout** — Excluded due to terminal corruption risk

## 4. Development Plan

### Phase 1: Design

- Detailed design document
- Development plan
- Architecture document (focus on decision rationale / ADRs)
- Document placement: `docs/en/` + `docs/ja/`

### Phase 2: Core Implementation

- Text-to-image generation
- Image editing with input images (`-i`)
- File output (`-o`, `--format`)
- Configuration loading (env vars > config file)
- Security implementation (nlk/guard integration, input validation)
- Token consumption display on stderr
- Tests (unit tests + E2E tests)
- E2E testing with all features integrated

### Phase 3: Release

- README.md / README.ja.md (project root)
- CHANGELOG.md
- Release process (tag, binary build, upload)
- Register as util-series submodule

Each phase can be reviewed independently.

## 5. Required API Scopes / Permissions

Same as gem-cli:

- **Authentication:** Google Cloud ADC (Application Default Credentials)
- **IAM role:** `roles/aiplatform.user`
- **API:** `aiplatform.googleapis.com` enabled
- **Additional scopes:** None (uses Model Garden models, no extra permissions)

## 6. Series Placement

- **Series:** util-series
- **Reason:** Same lineage as gem-cli — a pipe-friendly data processing CLI.
  Placed in util-series as a standalone tool following UNIX philosophy

## 7. External Platform Constraints

| Constraint | Impact | Mitigation |
|-----------|--------|------------|
| Vertex AI quota limits (RPM/TPM) | May hit rate limits during batch execution | Clear error messages. Sleep/retry controlled by shell |
| Model output resolution limits | Generated image size is model-dependent | Document limitations |
| Safety filters | Some prompts may be blocked | Appropriate error codes and messages |
| Token consumption | Cost management needed | Display token usage on stderr |

---

## Discussion Log

### Tool Name & Problem Statement
- Confirmed the need to use Gemini 2.5 Flash native image generation from CLI
- Web-based Gemini is impractical for batch usage, motivating CLI implementation
- Image description excluded from scope (handled by gem-cli)

### Command Design
- No subcommands needed (input types don't change between generation and editing)
- `-o` required (stdout binary output excluded due to terminal corruption risk)
- Multiple input images via `-i` repeat flag
- Prompt via `-p` flag + stdin fallback (`-p` takes precedence)
- Output format via `--format` flag + `-o` extension auto-detection

### Design Principles
- Go language, consistent with gem-cli
- Standalone tool, UNIX philosophy ("do one thing well")
- One request, one operation. Batch processing controlled by shell
- Prompt-based editing is in scope; image processing library features are not

### Security
- Security First design due to user input being sent to external API
- nlk/guard package nonce-tagged XML wrapping adopted for prompt injection protection
- Path traversal prevention, magic byte validation, permission management incorporated

### Development Plan
- Phase 1 produces design documents before implementation (detailed design, dev plan, architecture ADRs)
- Phase 2 is a single implementation phase (E2E testing requires all features)
- Phase 3 covers documentation and release

### Token Consumption Display
- Added stderr output of token usage for cost tracking purposes
