package cmd

import (
	"testing"
	"unicode/utf8"
)

func TestTruncateDescriptionUnicode(t *testing.T) {
	description := "😀😀😀😀😀😀"
	truncated := TruncateDescription(description, 5)

	if truncated != "😀😀..." {
		t.Errorf("TruncateDescription(%q, 5) = %q, want %q", description, truncated, "😀😀...")
	}
	if !utf8.ValidString(truncated) {
		t.Errorf("TruncateDescription(%q, 5) returned invalid UTF-8: %q", description, truncated)
	}
	if utf8.RuneCountInString(truncated) != 5 {
		t.Errorf("TruncateDescription(%q, 5) returned %d runes, want 5", description, utf8.RuneCountInString(truncated))
	}
}

func TestTruncateDescriptionShortLimit(t *testing.T) {
	tests := []struct {
		name     string
		maxLen   int
		expected string
	}{
		{name: "negative limit", maxLen: -1, expected: ""},
		{name: "zero limit", maxLen: 0, expected: ""},
		{name: "limit shorter than ellipsis", maxLen: 2, expected: "ab"},
		{name: "limit equal to ellipsis", maxLen: 3, expected: "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TruncateDescription("abcdef", tt.maxLen); got != tt.expected {
				t.Errorf("TruncateDescription(%q, %d) = %q, want %q", "abcdef", tt.maxLen, got, tt.expected)
			}
		})
	}
}
