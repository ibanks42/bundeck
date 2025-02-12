package plugin

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Runner struct {
	tempDir string
}

func NewRunner() (*Runner, error) {
	tempDir := filepath.Join(os.TempDir(), "bundeck-plugins")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	return &Runner{
		tempDir: tempDir,
	}, nil
}

func (r *Runner) Run(id int, code string) (string, error) {
	// Create a temporary file for the code
	tempFile := filepath.Join(r.tempDir, fmt.Sprintf("%d.ts", id))
	if err := os.WriteFile(tempFile, []byte(code), 0644); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}
	defer os.Remove(tempFile)

	// Run the code with Bun
	cmd := exec.Command("bun", "run", tempFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to run plugin: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

type PluginResult struct {
	Result string `json:"result"`
}
