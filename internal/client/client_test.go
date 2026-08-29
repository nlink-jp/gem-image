package client

import (
	"errors"
	"fmt"
	"strings"
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

func TestHintGemini3Location(t *testing.T) {
	notFound := errors.New("Error 404, Message: Publisher Model `projects/p/locations/us-central1/publishers/google/models/gemini-3.1-flash-image` was not found: NOT_FOUND")
	quota := errors.New("Error 429: quota exceeded")

	tests := []struct {
		name     string
		err      error
		model    string
		location string
		wantHint bool
	}{
		{"gemini3 image regional 404 gets hint", notFound, "gemini-3.1-flash-image", "us-central1", true},
		{"google/ prefix still recognized", notFound, "google/gemini-3.1-flash-image", "us-central1", true},
		{"pro image regional 404 gets hint", notFound, "gemini-3-pro-image", "asia-northeast1", true},
		{"global location needs no hint", notFound, "gemini-3.1-flash-image", "global", false},
		{"gemini 2.5 regional 404 is genuine", notFound, "gemini-2.5-flash-image", "us-central1", false},
		{"non-404 error passes through", quota, "gemini-3.1-flash-image", "us-central1", false},
		{"nil error stays nil", nil, "gemini-3.1-flash-image", "us-central1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hintGemini3Location(tt.err, tt.model, tt.location)
			if tt.err == nil {
				if got != nil {
					t.Fatalf("hintGemini3Location(nil) = %v, want nil", got)
				}
				return
			}
			if !errors.Is(got, tt.err) {
				t.Errorf("original error not wrapped: %v", got)
			}
			hinted := strings.Contains(got.Error(), "global endpoint")
			if hinted != tt.wantHint {
				t.Errorf("hint present = %v, want %v (%q)", hinted, tt.wantHint, got)
			}
		})
	}
}
