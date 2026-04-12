package security

import (
	"strings"
	"testing"
)

func TestWrapPrompt(t *testing.T) {
	sys, wrapped, err := WrapPrompt("draw a cat")
	if err != nil {
		t.Fatalf("WrapPrompt: %v", err)
	}

	// System prompt should contain expanded tag references
	if strings.Contains(sys, "{{DATA_TAG}}") {
		t.Error("system prompt still contains {{DATA_TAG}} placeholder")
	}
	if !strings.Contains(sys, "user_data_") {
		t.Error("system prompt should contain expanded tag name")
	}

	// Wrapped prompt should be enclosed in XML tags
	if !strings.Contains(wrapped, "<user_data_") {
		t.Error("wrapped prompt should start with <user_data_...")
	}
	if !strings.Contains(wrapped, "draw a cat") {
		t.Error("wrapped prompt should contain original text")
	}
}

func TestWrapPrompt_DifferentTagsPerCall(t *testing.T) {
	_, wrapped1, _ := WrapPrompt("test1")
	_, wrapped2, _ := WrapPrompt("test2")

	// Each call should generate a different nonce tag
	if wrapped1 == wrapped2 {
		t.Error("consecutive calls should generate different tags")
	}
}
