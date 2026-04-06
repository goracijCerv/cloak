package cmd

import (
	"encoding/json"
	"io"
	"os"
	"testing"
)

// This function redirects what fmt.Println would print into a string variable.
func captureStdout(f func()) string {
	// Save the original console output
	oldStdout := os.Stdout

	// Create a virtual "pipe" in the operating system
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute the function that prints things
	f()

	// Close the writer and read everything that went through the pipe
	w.Close()
	out, _ := io.ReadAll(r)

	// Restore the console to normal so the test doesn't break your terminal
	os.Stdout = oldStdout

	return string(out)
}

// 2. THE COBRA INTEGRATION TEST
func TestListCmd_JSON(t *testing.T) {
	// STEP 1: Isolation (Sandbox)
	// Create a temporary folder so we don't affect your real project
	tempDir := t.TempDir()

	// Trick your logger and config into thinking the user's "HOME"
	// is this temporary folder. This way we don't mess up your real ~/.config/cloak/
	t.Setenv("HOME", tempDir)        // For Mac/Linux
	t.Setenv("USERPROFILE", tempDir) // For Windows

	// Move into the temporary folder (which obviously has no Git backups)
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	// STEP 2: Command Execution
	// Capture the output while Cobra does its job
	output := captureStdout(func() {
		// Tell Cobra: "The user just typed 'list --json'"
		rootCmd.SetArgs([]string{"list", "--json"})

		// Execute the entire tool
		rootCmd.Execute()
	})

	// STEP 3: Verification (Asserts)
	// Since we requested --json, we expect 'output' to be valid JSON
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(output), &response); err != nil {
		t.Fatalf("The command did not return valid JSON. Error: %v\nOutput received:\n%s", err, output)
	}

	// Since we're in a new folder with no backups, the "status" should be "error"
	// (Remember that in your cmd/list.go we catch ErrNoBackUpDir and return error in JSON)
	if response["status"] != "error" {
		t.Errorf("Expected status 'error', got: %v", response["status"])
	}

	// And the message should be user-friendly
	expectedMsg := "Failed to retrieve backups"
	if response["message"] != expectedMsg {
		t.Errorf("Expected message %q, got %q", expectedMsg, response["message"])
	}
}
