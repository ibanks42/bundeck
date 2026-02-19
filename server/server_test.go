package main

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"net/http"

	"fyne.io/fyne/v2/app"
	"github.com/gofiber/fiber/v2"
)

func init() {
	// Initialize Fyne app for tests with a proper ID
	app.NewWithID("dev.ibanks.bundeck-server.test")
}

func TestFiberGetScripts(t *testing.T) {
	// Setup test directory and files
	tmpDir := t.TempDir()
	originalPath := getScriptsPath
	getScriptsPath = func() string { return tmpDir }
	defer func() { getScriptsPath = originalPath }()

	// Create test scripts
	testScripts := []struct {
		file    string
		content string
	}{
		{"test1.ts", "console.log('test1')"},
		{"test2.js", "console.log('test2')"},
	}

	for _, ts := range testScripts {
		err := os.WriteFile(filepath.Join(tmpDir, ts.file), []byte(ts.content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Create scripts.json
	scripts := []Script{
		{ID: 1, File: "test1.ts"},
		{ID: 2, File: "test2.js"},
	}
	scriptsData, _ := json.MarshalIndent(scripts, "", "\t")
	err := os.WriteFile(filepath.Join(tmpDir, "scripts.json"), scriptsData, 0644)
	if err != nil {
		t.Fatalf("Failed to create scripts.json: %v", err)
	}

	// Setup Fiber app
	app := fiber.New()
	app.Get("/scripts", fiberGetScripts)

	// Test GET /scripts
	req := httptest.NewRequest("GET", "/scripts", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	var responseScripts []string
	if err := json.NewDecoder(resp.Body).Decode(&responseScripts); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	expectedScripts := []string{"test1", "test2"} // Note: extensions are removed in the response
	if len(responseScripts) != len(expectedScripts) {
		t.Errorf("Expected %d scripts, got %d", len(expectedScripts), len(responseScripts))
	}

	for i, script := range expectedScripts {
		if responseScripts[i] != script {
			t.Errorf("Expected script %s, got %s", script, responseScripts[i])
		}
	}
}

func TestExecuteScript(t *testing.T) {
	// Skip if bun is not installed
	if _, err := os.Stat("/usr/bin/bun"); os.IsNotExist(err) {
		t.Skip("Skipping test: bun is not installed")
	}

	// Setup test directory and files
	tmpDir := t.TempDir()
	originalPath := getScriptsPath
	getScriptsPath = func() string { return tmpDir }
	defer func() { getScriptsPath = originalPath }()

	// Create test script
	testScript := `console.log("Hello, World!")`
	err := os.WriteFile(filepath.Join(tmpDir, "test.js"), []byte(testScript), 0644)
	if err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	// Setup Fiber app
	app := fiber.New()
	app.Get("/scripts/:id", executeScript)

	// Test script execution
	req := httptest.NewRequest("GET", "/scripts/test.js", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	// Read response body
	body := make([]byte, 1024)
	n, err := resp.Body.Read(body)
	if err != nil && err.Error() != "EOF" {
		t.Fatalf("Failed to read response body: %v", err)
	}

	output := string(body[:n])
	expected := "Hello, World!"
	if output != expected {
		t.Errorf("Expected output %q, got %q", expected, output)
	}
}

func TestStartServer(t *testing.T) {
	// Setup test directory
	tmpDir := t.TempDir()
	originalPath := getScriptsPath
	getScriptsPath = func() string { return tmpDir }
	defer func() { getScriptsPath = originalPath }()

	// Create test scripts.json
	scripts := []Script{}
	scriptsData, _ := json.MarshalIndent(scripts, "", "\t")
	err := os.WriteFile(filepath.Join(tmpDir, "scripts.json"), scriptsData, 0644)
	if err != nil {
		t.Fatalf("Failed to create scripts.json: %v", err)
	}

	// Start server
	StartServer()

	// Wait for server to be ready with timeout
	select {
	case <-serverReady:
		// Server is ready
	case <-time.After(5 * time.Second):
		t.Fatal("Server failed to start within timeout")
	}

	if fiberApp == nil {
		t.Error("Server did not start properly")
	}

	// Test if server is actually responding
	resp, err := http.Get("http://localhost:9212/scripts")
	if err != nil {
		t.Errorf("Server is not responding: %v", err)
	} else {
		resp.Body.Close()
	}

	// Cleanup
	if fiberApp != nil {
		fiberApp.Shutdown()
	}
}
