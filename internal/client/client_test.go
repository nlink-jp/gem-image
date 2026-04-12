package client

import (
	"errors"
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
