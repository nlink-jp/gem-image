package client

import (
	"errors"
	"fmt"
	"testing"

	"google.golang.org/genai"
)

func TestExtractResult_WithImage(t *testing.T) {
	resp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []*genai.Part{
						{
							InlineData: &genai.Blob{
								Data:     []byte("fake-image-data"),
								MIMEType: "image/png",
							},
						},
					},
				},
			},
		},
		UsageMetadata: &genai.GenerateContentResponseUsageMetadata{
			PromptTokenCount:     100,
			CandidatesTokenCount: 1290,
			TotalTokenCount:      1390,
		},
	}

	result, err := extractResult(resp)
	if err != nil {
		t.Fatalf("extractResult: %v", err)
	}
	if string(result.ImageData) != "fake-image-data" {
		t.Errorf("ImageData = %q, want fake-image-data", result.ImageData)
	}
	if result.MIMEType != "image/png" {
		t.Errorf("MIMEType = %q, want image/png", result.MIMEType)
	}
	if result.Usage == nil {
		t.Fatal("Usage is nil")
	}
	if result.Usage.InputTokens != 100 {
		t.Errorf("InputTokens = %d, want 100", result.Usage.InputTokens)
	}
	if result.Usage.OutputTokens != 1290 {
		t.Errorf("OutputTokens = %d, want 1290", result.Usage.OutputTokens)
	}
}

func TestExtractResult_WithTextAndImage(t *testing.T) {
	resp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []*genai.Part{
						{Text: "Here is your image:"},
						{
							InlineData: &genai.Blob{
								Data:     []byte("img"),
								MIMEType: "image/jpeg",
							},
						},
					},
				},
			},
		},
	}

	result, err := extractResult(resp)
	if err != nil {
		t.Fatalf("extractResult: %v", err)
	}
	if result.Text != "Here is your image:" {
		t.Errorf("Text = %q", result.Text)
	}
	if result.MIMEType != "image/jpeg" {
		t.Errorf("MIMEType = %q, want image/jpeg", result.MIMEType)
	}
}

func TestExtractResult_NoImage(t *testing.T) {
	resp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []*genai.Part{
						{Text: "I cannot generate that image."},
					},
				},
			},
		},
	}

	_, err := extractResult(resp)
	if !errors.Is(err, ErrNoImage) {
		t.Errorf("expected ErrNoImage, got %v", err)
	}
}

func TestExtractResult_SafetyBlock(t *testing.T) {
	resp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				FinishReason: genai.FinishReasonSafety,
				Content:      &genai.Content{},
			},
		},
	}

	_, err := extractResult(resp)
	if !errors.Is(err, ErrSafetyBlock) {
		t.Errorf("expected ErrSafetyBlock, got %v", err)
	}
}

func TestExtractResult_NilResponse(t *testing.T) {
	_, err := extractResult(nil)
	if !errors.Is(err, ErrNoImage) {
		t.Errorf("expected ErrNoImage, got %v", err)
	}
}

func TestExtractResult_EmptyCandidates(t *testing.T) {
	resp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{},
	}
	_, err := extractResult(resp)
	if !errors.Is(err, ErrNoImage) {
		t.Errorf("expected ErrNoImage, got %v", err)
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		err  error
		want bool
	}{
		{fmt.Errorf("Error 429: quota exceeded"), true},
		{fmt.Errorf("Error 503: service unavailable"), true},
		{fmt.Errorf("Error 500: internal server error"), true},
		{fmt.Errorf("connection refused"), true},
		{fmt.Errorf("timeout"), true},
		{fmt.Errorf("unexpected EOF"), true},
		{fmt.Errorf("Error 400: bad request"), false},
		{fmt.Errorf("Error 404: not found"), false},
		{fmt.Errorf("Error 403: forbidden"), false},
	}

	for _, tt := range tests {
		t.Run(tt.err.Error(), func(t *testing.T) {
			got := isRetryable(tt.err)
			if got != tt.want {
				t.Errorf("isRetryable(%q) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
