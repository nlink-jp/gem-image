# Detailed Design: gem-image

> Generated: 2026-04-12
> Status: Draft

## Overview

gem-image is a CLI tool for using Vertex AI Gemini 2.5 Flash's
image generation and editing capabilities.

---

## CLI Interface

### Basic Syntax

```
gem-image -p <prompt> [-i <image>...] -o <output> [--format png|jpeg] [--debug]
```

### Flag Details

| Flag | Short | Type | Required | Default | Description |
|------|-------|------|----------|---------|-------------|
| `--prompt` | `-p` | string | * | — | Image generation prompt |
| `--input` | `-i` | []string | No | — | Input image path (repeatable) |
| `--output` | `-o` | string | Yes | — | Output file path |
| `--format` | — | string | No | `png` | Output format (`png` \| `jpeg`) |
| `--config` | `-c` | string | No | `~/.config/gem-image/config.toml` | Config file path |
| `--model` | `-m` | string | No | (from config) | Model name override |
| `--debug` | — | bool | No | false | Enable debug output |
| `--version` | `-v` | bool | No | — | Show version |

\* If `-p` is omitted, prompt is read from stdin. When both `-p` and stdin are available, `-p` takes precedence.

### Output Format Resolution Logic

```
if -o extension is .png or .jpeg/.jpg:
    -> follow extension
elif --format is specified:
    -> follow --format
else:
    -> png (default)
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error (config issues, file I/O errors, etc.) |
| 2 | Input validation error (invalid file format, path traversal, etc.) |
| 3 | API error (auth failure, quota exceeded, network error, etc.) |
| 4 | Safety filter block |

---

## Package Design

### cmd/root.go

Responsibility: CLI flag parsing, input assembly, orchestration of package calls.

```go
func Execute(version string)
func runGenerate(cmd *cobra.Command, args []string) error
```

**Processing Flow:**

1. Load configuration (config.Load)
2. Obtain prompt (-p / stdin)
3. Validate input images (security.ValidateImagePath)
4. Protect prompt (security.WrapPrompt)
5. Call API (client.Generate)
6. Write image (image.WriteFile)
7. Display token info (stderr)

### internal/config

Same pattern as gem-cli. TOML config file + environment variable overrides.

```go
type Config struct {
    GCP   GCPConfig
    Model ModelConfig
}

type GCPConfig struct {
    Project  string `toml:"project"`
    Location string `toml:"location"`
}

type ModelConfig struct {
    Name string `toml:"name"`
}

func Load(path string) (*Config, error)
func (c *Config) ApplyFlags(model string)
```

**Config file example (config.example.toml):**

```toml
[gcp]
project  = "your-project-id"
location = "us-central1"

[model]
name = "gemini-2.5-flash-image"
```

### internal/client

Gemini API client. Configures and calls `GenerateContent` for image generation.

```go
type Client struct {
    inner *genai.Client
    model string
}

type GenerateResult struct {
    ImageData []byte         // Generated image binary
    MIMEType  string         // "image/png" or "image/jpeg"
    Text      string         // Text response (if any)
    Usage     *UsageInfo
}

type UsageInfo struct {
    InputTokens  int64
    OutputTokens int64
    TotalTokens  int64
}

func New(ctx context.Context, cfg *config.Config) (*Client, error)
func (c *Client) Generate(ctx context.Context, opts *GenerateOpts) (*GenerateResult, error)
func (c *Client) Close() error
```

**GenerateOpts:**

```go
type GenerateOpts struct {
    SystemPrompt string
    UserPrompt   string       // Wrapped prompt
    Images       []*ImageInput
    OutputFormat string       // "image/png" or "image/jpeg"
}

type ImageInput struct {
    Data     []byte
    MIMEType string
}
```

**API Configuration:**

```go
config := &genai.GenerateContentConfig{
    ResponseModalities: []string{
        string(genai.ModalityImage),
    },
}
```

### internal/image

Handles image file I/O.

**input.go:**

```go
// ReadImageFile reads an image file and returns a validated ImageInput
func ReadImageFile(path string) (*client.ImageInput, error)

// detectMIME determines MIME type from magic bytes
func detectMIME(data []byte) (string, error)
```

**output.go:**

```go
// WriteFile writes image data to a file (permission 0644)
func WriteFile(path string, data []byte) error

// ResolveFormat determines output MIME type from -o extension and --format
func ResolveFormat(outputPath string, formatFlag string) string
```

### internal/security

Consolidates security-related processing.

**guard.go:**

```go
// WrapPrompt wraps user prompt with nlk/guard
func WrapPrompt(userPrompt string) (systemPrompt string, wrappedUser string, err error)
```

**validate.go:**

```go
// ValidateImagePath validates file path safety
func ValidateImagePath(path string) (string, error)

// ValidateOutputPath validates output path safety
func ValidateOutputPath(path string) error

// ValidateImageData validates magic bytes and size
func ValidateImageData(data []byte) error
```

**Validation Items:**

| Validation | Implementation |
|-----------|----------------|
| Path traversal | `filepath.Abs()` + `filepath.EvalSymlinks()` |
| Magic bytes | PNG: `\x89PNG\r\n\x1a\n`, JPEG: `\xFF\xD8\xFF` |
| File size | Limit check (aligned with model input constraints) |
| Output directory | Existence check, write permission check |

---

## Error Handling

### API Errors

| Error Type | Response |
|-----------|----------|
| Auth error (401/403) | Display message prompting ADC setup |
| Quota exceeded (429) | Notify rate limit exceeded; retry controlled by shell |
| Safety filter (FinishReasonSafety) | Exit code 4 with explicit notification |
| Model unsupported | Display message prompting model name check |
| No image generated | Error when response contains no InlineData |

### Input Errors

| Error Type | Response |
|-----------|----------|
| `-o` not specified | Required flag error (handled by Cobra) |
| No prompt input | Error when stdin is terminal and `-p` is absent |
| Invalid image file | Error on magic byte mismatch (exit code 2) |
| Path traversal | Error on detection (exit code 2) |

---

## Test Design

### Unit Tests

| Package | Test Content |
|---------|-------------|
| config | TOML loading, env var overrides, required field checks |
| client | GenerateOpts construction, image extraction from response, UsageMetadata |
| image/input | Magic byte detection, MIME detection, invalid file rejection |
| image/output | File writing, format resolution logic |
| security/guard | nlk/guard wrapping, system prompt expansion |
| security/validate | Path normalization, traversal detection, size limits |

### E2E Tests

- Text to image generation -> file save -> magic byte verification
- Input image + prompt -> edited image -> file save
- Multiple input images operation
- Safety filter block exit code verification
- Token info stderr output verification
- Invalid input rejection (path traversal, spoofed files)

---

## Dependencies

### Direct Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/spf13/cobra` | latest | CLI framework |
| `github.com/BurntSushi/toml` | latest | Config file parsing |
| `google.golang.org/genai` | latest | Gemini API SDK |
| `github.com/nlink-jp/nlk` | latest | guard (prompt injection protection) |

### Indirect Dependencies

genai SDK pulls Google Cloud related packages (same as gem-cli).
