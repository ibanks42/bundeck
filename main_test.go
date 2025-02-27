package main

import (
	"bundeck/internal/api"
	"bundeck/internal/settings"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"os"
	"testing"
	"testing/fstest"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Mock plugins filesystem for testing
var mockPluginsEmbedFS = fstest.MapFS{
	"plugins/list.json": &fstest.MapFile{
		Data: []byte(`{
			"utility": {
				"name": "Utility",
				"plugins": [
					{
						"id": "test-plugin",
						"name": "Test Plugin",
						"description": "A test plugin",
						"file": "test-plugin.ts"
					}
				]
			}
		}`),
	},
	"plugins/test-plugin.ts": &fstest.MapFile{
		Data: []byte(`// Test plugin
export default {
	run: async () => {
		return "This is a test plugin";
	}
};`),
	},
}

func TestMainIntegration(t *testing.T) {
	// Skip this test by default as it starts a real server
	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("Skipping integration test. Set RUN_INTEGRATION_TESTS=1 to run")
	}

	// Create a temporary database file
	dbPath = "test_plugins.db"
	defer os.Remove(dbPath)

	// Set up mock plugins filesystem
	subFS, err := fs.Sub(mockPluginsEmbedFS, "plugins")
	if err != nil {
		t.Fatalf("Failed to create sub filesystem: %v", err)
	}
	api.PluginsFS = subFS

	// Start the server in a goroutine
	go func() {
		os.Args = []string{"cmd", "-test.run=TestMainIntegration"}
		main()
	}()

	// Wait for the server to start
	time.Sleep(2 * time.Second)

	// Create an HTTP client
	client := &http.Client{}

	// Test creating a plugin
	t.Run("Create Plugin", func(t *testing.T) {
		// Create multipart form data
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// Add form fields
		_ = writer.WriteField("name", "Test Plugin")
		_ = writer.WriteField("code", "console.log('test')")
		_ = writer.WriteField("order_num", "1")

		// Add image file
		part, err := writer.CreateFormFile("image/png", "test.png")
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}
		// Create a small PNG file (1x1 transparent pixel)
		pngData := []byte{
			0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D,
			0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
			0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4, 0x89, 0x00, 0x00, 0x00,
			0x0A, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9C, 0x63, 0x00, 0x00, 0x00, 0x05,
			0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45,
			0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
		}
		_, err = part.Write(pngData)
		if err != nil {
			t.Fatalf("Failed to write image data: %v", err)
		}

		writer.Close()

		// Create request
		req, err := http.NewRequest("POST", "http://localhost:3004/api/plugins", body)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to create plugin: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != fiber.StatusCreated {
			respBody, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected status %d, got %d. Response: %s", fiber.StatusCreated, resp.StatusCode, string(respBody))
		}
	})

	// Test getting all plugins
	t.Run("Get All Plugins", func(t *testing.T) {
		req, err := http.NewRequest("GET", "http://localhost:3004/api/plugins", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to get plugins: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
		}

		var plugins []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&plugins); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(plugins) == 0 {
			t.Error("Expected at least one plugin")
		}
	})
}

func TestDatabaseConnection(t *testing.T) {
	// Create a temporary database file
	dbFile := "test_plugins.db"
	defer os.Remove(dbFile)

	pragmas := "?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)&_pragma=journal_size_limit(200000000)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)&_pragma=temp_store(MEMORY)&_pragma=cache_size(-16000)"
	db, err := sql.Open("sqlite", dbFile+pragmas)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	// Test database operations
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS test_table (
			id INTEGER PRIMARY KEY,
			name TEXT
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Test insert
	result, err := db.Exec("INSERT INTO test_table (name) VALUES (?)", "test")
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get last insert ID: %v", err)
	}

	// Test select
	var name string
	err = db.QueryRow("SELECT name FROM test_table WHERE id = ?", id).Scan(&name)
	if err != nil {
		t.Fatalf("Failed to query test data: %v", err)
	}

	if name != "test" {
		t.Errorf("Expected name 'test', got '%s'", name)
	}
}

func TestSettingsLoading(t *testing.T) {
	// Create a temporary settings file
	settingsFile := "settings.json"
	defer os.Remove(settingsFile)

	// Test default settings
	s := settings.LoadSettings()
	if s.Port != 3004 {
		t.Errorf("Expected default port 3004, got %d", s.Port)
	}

	// Test custom settings
	customSettings := map[string]interface{}{
		"port": 8080,
	}

	data, err := json.Marshal(customSettings)
	if err != nil {
		t.Fatalf("Failed to marshal settings: %v", err)
	}

	if err := os.WriteFile(settingsFile, data, 0644); err != nil {
		t.Fatalf("Failed to write settings file: %v", err)
	}

	loadedSettings := settings.LoadSettings()
	if loadedSettings.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", loadedSettings.Port)
	}
}

func TestServerStartup(t *testing.T) {
	// Create a channel to signal server startup
	ready := make(chan bool)

	// Start the server in a goroutine
	go func() {
		app := fiber.New()
		app.Get("/health", func(c *fiber.Ctx) error {
			return c.SendString("OK")
		})

		// Signal that the server is ready
		ready <- true

		if err := app.Listen(":0"); err != nil {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	// Wait for the server to start or timeout
	select {
	case <-ready:
		// Server started successfully
	case <-time.After(5 * time.Second):
		t.Fatal("Server failed to start within timeout")
	}
}
