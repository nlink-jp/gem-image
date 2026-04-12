// Package client provides a Gemini API client for image generation.
package client

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nlink-jp/gem-image/internal/config"
	"github.com/nlink-jp/nlk/backoff"
	"google.golang.org/genai"
)

const maxRetries = 3

var (
	ErrNoImage     = errors.New("no image in response")
	ErrSafetyBlock = errors.New("request blocked by safety filter")
)

// UsageInfo holds token consumption data.
type UsageInfo struct {
	InputTokens  int32
	OutputTokens int32
	TotalTokens  int32
}

// GenerateResult holds the API response with extracted image data.
type GenerateResult struct {
	ImageData []byte
	MIMEType  string
	Text      string
	Usage     *UsageInfo
}

// GenerateOpts holds parameters for image generation.
type GenerateOpts struct {
	SystemPrompt string
	UserPrompt   string
	Images       []*ImageInput
	OutputFormat string // "image/png" or "image/jpeg"
}

// ImageInput holds image data for API submission.
type ImageInput struct {
	Data     []byte
	MIMEType string
}

// Generator is the interface for generating images. Extracted for testability.
type Generator interface {
	Generate(ctx context.Context, opts *GenerateOpts) (*GenerateResult, error)
	Close() error
}

// Client wraps the Gemini genai client for image generation.
type Client struct {
	inner *genai.Client
	model string
}

// Verify Client implements Generator at compile time.
var _ Generator = (*Client)(nil)

// New creates a Client configured for Vertex AI.
func New(ctx context.Context, cfg *config.Config) (*Client, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Backend:  genai.BackendVertexAI,
		Project:  cfg.GCP.Project,
		Location: cfg.GCP.Location,
	})
	if err != nil {
		return nil, fmt.Errorf("create genai client: %w", err)
	}
	return &Client{inner: client, model: cfg.Model.Name}, nil
}

// Close releases the client resources.
func (c *Client) Close() error {
	return nil
}

// Generate calls the Gemini API for image generation/editing with automatic retry.
func (c *Client) Generate(ctx context.Context, opts *GenerateOpts) (*GenerateResult, error) {
	gcConfig := &genai.GenerateContentConfig{
		ResponseModalities: []string{
			string(genai.ModalityImage),
		},
	}

	if opts.SystemPrompt != "" {
		gcConfig.SystemInstruction = genai.NewContentFromText(opts.SystemPrompt, "")
	}

	// Build content parts
	var parts []*genai.Part
	parts = append(parts, genai.NewPartFromText(opts.UserPrompt))

	for _, img := range opts.Images {
		parts = append(parts, genai.NewPartFromBytes(img.Data, img.MIMEType))
	}

	contents := []*genai.Content{
		genai.NewContentFromParts(parts, "user"),
	}

	bo := backoff.New(
		backoff.WithBase(2*time.Second),
		backoff.WithMax(30*time.Second),
	)

	var lastErr error
	for attempt := range maxRetries + 1 {
		resp, err := c.inner.Models.GenerateContent(ctx, c.model, contents, gcConfig)
		if err == nil {
			return extractResult(resp)
		}

		lastErr = err
		if !isRetryable(err) || attempt == maxRetries {
			return nil, fmt.Errorf("generate content: %w", err)
		}

		wait := bo.Duration(attempt)
		log.Printf("API call failed (attempt %d/%d), retrying in %v: %v",
			attempt+1, maxRetries+1, wait.Round(time.Second), err)
		time.Sleep(wait)
	}

	return nil, fmt.Errorf("generate content after %d retries: %w", maxRetries, lastErr)
}

func isRetryable(err error) bool {
	errStr := strings.ToLower(err.Error())
	for _, k := range []string{"429", "503", "500", "unavailable", "timeout", "connection refused", "eof"} {
		if strings.Contains(errStr, k) {
			return true
		}
	}
	return false
}

func extractResult(resp *genai.GenerateContentResponse) (*GenerateResult, error) {
	if resp == nil || len(resp.Candidates) == 0 {
		return nil, ErrNoImage
	}

	candidate := resp.Candidates[0]

	// Check for safety filter block
	if candidate.FinishReason == genai.FinishReasonSafety {
		return nil, ErrSafetyBlock
	}

	if candidate.Content == nil {
		return nil, ErrNoImage
	}

	result := &GenerateResult{}

	// Extract image and text from parts
	for _, part := range candidate.Content.Parts {
		if part.InlineData != nil && result.ImageData == nil {
			result.ImageData = part.InlineData.Data
			result.MIMEType = part.InlineData.MIMEType
		}
		if part.Text != "" {
			result.Text = part.Text
		}
	}

	if result.ImageData == nil {
		return nil, ErrNoImage
	}

	// Extract usage metadata
	if resp.UsageMetadata != nil {
		result.Usage = &UsageInfo{
			InputTokens:  resp.UsageMetadata.PromptTokenCount,
			OutputTokens: resp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:  resp.UsageMetadata.TotalTokenCount,
		}
	}

	return result, nil
}
