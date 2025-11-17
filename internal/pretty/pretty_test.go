package pretty_test

import (
	"os"
	"testing"

	"github.com/ptdewey/freeze/internal/pretty"
)

func TestColorFunctionsWithColor(t *testing.T) {
	os.Unsetenv("NO_COLOR")

	tests := []struct {
		name string
		fn   func(string) string
		text string
	}{
		{"Red", pretty.Red, "error"},
		{"Green", pretty.Green, "success"},
		{"Yellow", pretty.Yellow, "warning"},
		{"Blue", pretty.Blue, "info"},
		{"Gray", pretty.Gray, "gray"},
		{"Bold", pretty.Bold, "bold"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.text)
			if result == "" {
				t.Errorf("%s returned empty string", tt.name)
			}
			if result == tt.text {
				t.Errorf("%s did not add color codes", tt.name)
			}
			if !contains(result, tt.text) {
				t.Errorf("%s does not contain original text", tt.name)
			}
		})
	}
}

func TestColorFunctionsNoColor(t *testing.T) {
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")

	tests := []struct {
		name string
		fn   func(string) string
		text string
	}{
		{"Red", pretty.Red, "error"},
		{"Green", pretty.Green, "success"},
		{"Yellow", pretty.Yellow, "warning"},
		{"Blue", pretty.Blue, "info"},
		{"Gray", pretty.Gray, "gray"},
		{"Bold", pretty.Bold, "bold"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.text)
			if result != tt.text {
				t.Errorf("%s should return plain text when NO_COLOR is set", tt.name)
			}
		})
	}
}

func TestHeader(t *testing.T) {
	os.Unsetenv("NO_COLOR")

	result := pretty.Header("test header")
	if result == "" {
		t.Error("Header returned empty string")
	}
	if result == "test header" {
		t.Error("Header should apply formatting")
	}
	if !contains(result, "test header") {
		t.Error("Header should contain original text")
	}
}

func TestSuccess(t *testing.T) {
	os.Unsetenv("NO_COLOR")

	result := pretty.Success("success message")
	if result == "" {
		t.Error("Success returned empty string")
	}
	if !contains(result, "success message") {
		t.Error("Success should contain original text")
	}
}

func TestError(t *testing.T) {
	os.Unsetenv("NO_COLOR")

	result := pretty.Error("error message")
	if result == "" {
		t.Error("Error returned empty string")
	}
	if !contains(result, "error message") {
		t.Error("Error should contain original text")
	}
}

func TestWarning(t *testing.T) {
	os.Unsetenv("NO_COLOR")

	result := pretty.Warning("warning message")
	if result == "" {
		t.Error("Warning returned empty string")
	}
	if !contains(result, "warning message") {
		t.Error("Warning should contain original text")
	}
}

func TestTerminalWidth(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected int
	}{
		{"default", "", 80},
		{"valid width", "120", 120},
		{"invalid width", "invalid", 80},
		{"zero width", "0", 80},
		{"negative width", "-10", 80},
		{"large width", "1000", 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue == "" {
				os.Unsetenv("COLUMNS")
			} else {
				os.Setenv("COLUMNS", tt.envValue)
			}
			defer os.Unsetenv("COLUMNS")

			result := pretty.TerminalWidth()
			if result != tt.expected {
				t.Errorf("TerminalWidth() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestClearScreen(t *testing.T) {
	pretty.ClearScreen()
}

func TestClearLine(t *testing.T) {
	pretty.ClearLine()
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
