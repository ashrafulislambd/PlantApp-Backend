package ai

import (
	"context"
	"strings"
	"testing"
)

func TestFallbackChatProvider_NoEntriesReturnsError(t *testing.T) {
	provider := NewFallbackChatProvider()

	_, err := provider.Reply(context.Background(), nil, "hello", "en")
	if err == nil {
		t.Fatal("Reply() error = nil, want error when no providers are configured")
	}
	if !strings.Contains(err.Error(), "no providers configured") {
		t.Fatalf("Reply() error = %v, want message containing %q", err, "no providers configured")
	}
}

func TestFallbackDiagnosisProvider_NoEntriesReturnsError(t *testing.T) {
	provider := NewFallbackDiagnosisProvider()

	_, err := provider.Analyze(context.Background(), nil)
	if err == nil {
		t.Fatal("Analyze() error = nil, want error when no providers are configured")
	}
	if !strings.Contains(err.Error(), "no providers configured") {
		t.Fatalf("Analyze() error = %v, want message containing %q", err, "no providers configured")
	}
}
