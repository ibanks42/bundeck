package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestNewRunner(t *testing.T) {
	runner, err := NewRunner()
	if err != nil {
		t.Fatalf("Failed to create new runner: %v", err)
	}

	// Check if temp directory exists
	if _, err := os.Stat(runner.tempDir); os.IsNotExist(err) {
		t.Error("Temp directory was not created")
	}

	// Cleanup
	os.RemoveAll(runner.tempDir)
}

func TestRunner_Run(t *testing.T) {
	runner, err := NewRunner()
	if err != nil {
		t.Fatalf("Failed to create new runner: %v", err)
	}
	defer os.RemoveAll(runner.tempDir)

	tests := []struct {
		name    string
		code    string
		want    string
		wantErr bool
	}{
		{
			name: "Valid TypeScript code",
			code: `console.log("Hello, World!")`,
			want: "Hello, World!\n",
		},
		{
			name:    "Invalid TypeScript code",
			code:    `console.log(undefined.property)`,
			wantErr: true,
		},
		{
			name: "Multiple console logs",
			code: `
				console.log("First line");
				console.log("Second line");
			`,
			want: "First line\nSecond line\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := runner.Run(1, tt.code)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Normalize line endings and trim spaces
			got = strings.TrimSpace(got)
			want := strings.TrimSpace(tt.want)

			if got != want {
				t.Errorf("Run() output = %q, want %q", got, want)
			}

			// Verify cleanup
			files, err := os.ReadDir(runner.tempDir)
			if err != nil {
				t.Fatalf("Failed to read temp directory: %v", err)
			}
			if len(files) > 0 {
				t.Error("Temporary files were not cleaned up")
			}
		})
	}
}

func TestRunner_TempFileHandling(t *testing.T) {
	runner, err := NewRunner()
	if err != nil {
		t.Fatalf("Failed to create new runner: %v", err)
	}
	defer os.RemoveAll(runner.tempDir)

	// Test concurrent runs
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		t.Run("Concurrent run "+string(rune('A'+i)), func(t *testing.T) {
			defer wg.Done()
			code := fmt.Sprintf(`console.log("Test %c")`, rune('A'+i))
			_, err := runner.Run(i, code)
			if err != nil {
				t.Errorf("Failed to run concurrent code: %v", err)
			}
		})
	}
	wg.Wait()

	// Verify temp file naming
	tempFile := filepath.Join(runner.tempDir, "1.ts")
	code := `console.log("test")`
	_, err = runner.Run(1, code)
	if err != nil {
		t.Fatalf("Failed to run code: %v", err)
	}

	// Verify file was cleaned up
	if _, err := os.Stat(tempFile); !os.IsNotExist(err) {
		t.Error("Temporary file was not cleaned up")
	}
}

func TestRunner_LargeCode(t *testing.T) {
	runner, err := NewRunner()
	if err != nil {
		t.Fatalf("Failed to create new runner: %v", err)
	}
	defer os.RemoveAll(runner.tempDir)

	// Generate large code
	var code strings.Builder
	code.WriteString("const letters = [];\n")
	for i := 0; i < 1000; i++ {
		code.WriteString(fmt.Sprintf("letters.push('%c');\n", rune('A'+i%26)))
	}
	code.WriteString("letters.forEach(letter => console.log(letter));\n")

	_, err = runner.Run(1, code.String())
	if err != nil {
		t.Fatalf("Failed to run large code: %v", err)
	}
}
