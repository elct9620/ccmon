package entity

import (
	"testing"
)

func TestNewModel(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid_haiku_model",
			input:    "claude-3-5-haiku-20241022",
			expected: "claude-3-5-haiku-20241022",
		},
		{
			name:     "valid_sonnet_model",
			input:    "claude-3-5-sonnet-20241022",
			expected: "claude-3-5-sonnet-20241022",
		},
		{
			name:     "valid_opus_model",
			input:    "claude-3-opus-20240229",
			expected: "claude-3-opus-20240229",
		},
		{
			name:     "valid_short_model",
			input:    "gpt",
			expected: "gpt",
		},
		{
			name:     "empty_string",
			input:    "",
			expected: "unknown",
		},
		{
			name:     "whitespace_only",
			input:    "   ",
			expected: "unknown",
		},
		{
			name:     "whitespace_padded_valid",
			input:    "  claude-3-5-haiku-20241022  ",
			expected: "claude-3-5-haiku-20241022",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			model := NewModel(tc.input)
			if model.String() != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, model.String())
			}
		})
	}
}

func TestModel_IsBase(t *testing.T) {
	testCases := []struct {
		name     string
		model    string
		expected bool
	}{
		{
			name:     "haiku_lowercase",
			model:    "claude-3-haiku-20240307",
			expected: true,
		},
		{
			name:     "haiku_uppercase",
			model:    "CLAUDE-3-HAIKU-20240307",
			expected: true,
		},
		{
			name:     "haiku_mixed_case",
			model:    "Claude-3-Haiku-20240307",
			expected: true,
		},
		{
			name:     "haiku_in_middle",
			model:    "some-haiku-model",
			expected: true,
		},
		{
			name:     "sonnet_model",
			model:    "claude-3-5-sonnet-20241022",
			expected: false,
		},
		{
			name:     "opus_model",
			model:    "claude-3-opus-20240229",
			expected: false,
		},
		{
			name:     "other_model",
			model:    "gpt-4",
			expected: false,
		},
		{
			name:     "empty_model",
			model:    "",
			expected: false,
		},
		{
			name:     "haiku_partial_match",
			model:    "haik", // Should not match
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			model := NewModel(tc.model)
			result := model.IsBase()
			if result != tc.expected {
				t.Errorf("Expected IsBase() to return %v for model %q, got %v", tc.expected, tc.model, result)
			}
		})
	}
}

func TestModel_String(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple_model",
			input:    "claude-3-haiku",
			expected: "claude-3-haiku",
		},
		{
			name:     "trimmed_input",
			input:    "  claude-3-sonnet  ",
			expected: "claude-3-sonnet",
		},
		{
			name:     "complex_model_name",
			input:    "claude-3-5-haiku-20241022",
			expected: "claude-3-5-haiku-20241022",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			model := NewModel(tc.input)
			result := model.String()
			if result != tc.expected {
				t.Errorf("Expected String() to return %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestModel_Integration(t *testing.T) {
	// Test that a model created with NewModel works correctly with business logic
	model := NewModel("claude-3-5-haiku-20241022")

	// Test that it correctly identifies as base model
	if !model.IsBase() {
		t.Errorf("Expected haiku model to be identified as base model")
	}

	// Test string representation
	if model.String() != "claude-3-5-haiku-20241022" {
		t.Errorf("Expected correct string representation")
	}

	// Test with premium model
	premiumModel := NewModel("claude-3-5-sonnet-20241022")
	if premiumModel.IsBase() {
		t.Errorf("Expected sonnet model to NOT be identified as base model")
	}

	// Test with unknown model (invalid input)
	unknownModel := NewModel("")
	if unknownModel.String() != "unknown" {
		t.Errorf("Expected empty input to result in 'unknown' model")
	}
	if unknownModel.IsBase() {
		t.Errorf("Expected 'unknown' model to NOT be identified as base model")
	}
}
