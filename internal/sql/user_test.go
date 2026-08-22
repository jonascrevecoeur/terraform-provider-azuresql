package sql

import "testing"

func TestQuoteIdentifier(t *testing.T) {
	tests := map[string]string{
		"dbo":        "[dbo]",
		"group]name": "[group]]name]",
	}

	for input, expected := range tests {
		if actual := quoteIdentifier(input); actual != expected {
			t.Errorf("quoteIdentifier(%q) = %q, want %q", input, actual, expected)
		}
	}
}
