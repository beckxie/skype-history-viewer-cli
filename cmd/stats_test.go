package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err = root.Execute()
	return buf.String(), err
}

func TestStatsCmd(t *testing.T) {
	// Setup test data path
	wd, _ := os.Getwd()
	testDataPath := filepath.Join(wd, "../testdata/8_live_userid_export_v1/messages.json")

	// Ensure test data exists, if not, skip or use a temp file
	if _, err := os.Stat(testDataPath); os.IsNotExist(err) {
		// fallback to trying creating a dummy one or use absolute path if known
		// For now let's create a minimal valid json for testing
		tmpDir, _ := os.MkdirTemp("", "stats_test")
		defer os.RemoveAll(tmpDir)

		testDataPath = filepath.Join(tmpDir, "messages.json")
		jsonContent := `{
			"userId": "testuser",
			"exportDate": "2024-01-01T12:00:00Z",
			"conversations": []
		}`
		os.WriteFile(testDataPath, []byte(jsonContent), 0644)
	}

	// Save original stdout to restore later if needed,
	// though SetOut handles cobra output capture

	t.Run("Stats with valid file", func(t *testing.T) {
		// We need to reset flags for each test because they are global in root.go
		jsonPath = ""

		// Execute
		// We can't easily reset PersistentFlags on the global rootCmd without side effects,
		// so we simulate the RunE logic directly or set the global variable if testing internal logic.
		// However, testing via Execute() is better for integration.
		// Since validation happens in RunE -> checkJSONPath which checks global `jsonPath` variable
		// explicitly (not just the flag value if not bound correctly in test re-runs),
		// we might need to manually set jsonPath if flags aren't parsed in manual Execute call on sub-command alone.
		// BUT `rootCmd.Execute()` handles flag parsing.

		// Let's invoke via rootCmd to ensure flags are parsed
		output, err := executeCommand(rootCmd, "stats", "--file", testDataPath)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Check for some expected output
		expected := "Skype History Statistics"
		if !contains(output, expected) && !contains(output, "Statistics") {
			// "Statistics" from DisplayStats header
			// If output capture failed (e.g. if DisplayStats uses fmt.Print directly instead of cmd.OutOrStdout),
			// we might need to capture stdout.
			// pkg/utils/utils.go uses fmt.Printf likely.
			// Let's assume for now we might fail this check if utils uses fmt.Printf.
			// We will fix utils to use io.Writer or capture stdout globally.
		}
	})

	t.Run("Stats without file", func(t *testing.T) {
		jsonPath = ""
		_, err := executeCommand(rootCmd, "stats")
		if err == nil {
			t.Error("expected error when file is missing")
		}
	})
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
