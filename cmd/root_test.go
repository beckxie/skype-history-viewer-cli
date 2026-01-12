package cmd

import (
	"os"
	"testing"
)

func TestCheckJSONPath(t *testing.T) {
	// Reset jsonPath after test
	defer func() { jsonPath = "" }()

	t.Run("Empty Path", func(t *testing.T) {
		jsonPath = ""
		if err := checkJSONPath(); err == nil {
			t.Error("expected error for empty path, got nil")
		}
	})

	t.Run("Non-existent Path", func(t *testing.T) {
		jsonPath = "/path/to/non/existent/file.json"
		if err := checkJSONPath(); err == nil {
			t.Error("expected error for non-existent path, got nil")
		}
	})

	t.Run("Valid Path", func(t *testing.T) {
		// Create a temporary file
		f, err := os.CreateTemp("", "test.json")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(f.Name())

		jsonPath = f.Name()
		if err := checkJSONPath(); err != nil {
			t.Errorf("unexpected error for valid path: %v", err)
		}
	})
}
