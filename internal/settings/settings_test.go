package settings

import (
	"encoding/json"
	"os"
	"testing"
)

func TestLoadSettings(t *testing.T) {
	// Clean up any existing settings file
	os.Remove("settings.json")
	defer os.Remove("settings.json")

	// Test loading default settings when file doesn't exist
	t.Run("Default Settings", func(t *testing.T) {
		settings := LoadSettings()
		if settings == nil {
			t.Fatal("Expected non-nil settings")
		}
		if settings.Port != 3004 {
			t.Errorf("Expected default port 3004, got %d", settings.Port)
		}

		// Verify file was created
		if _, err := os.Stat("settings.json"); os.IsNotExist(err) {
			t.Error("Settings file was not created")
		}
	})

	// Test loading custom settings
	t.Run("Custom Settings", func(t *testing.T) {
		customSettings := &Settings{
			Port: 8080,
		}

		// Write custom settings to file
		data, err := json.MarshalIndent(customSettings, "", "\t")
		if err != nil {
			t.Fatalf("Failed to marshal settings: %v", err)
		}

		if err := os.WriteFile("settings.json", data, 0666); err != nil {
			t.Fatalf("Failed to write settings file: %v", err)
		}

		// Load settings
		settings := LoadSettings()
		if settings == nil {
			t.Fatal("Expected non-nil settings")
		}
		if settings.Port != 8080 {
			t.Errorf("Expected port 8080, got %d", settings.Port)
		}
	})
}

func TestLoadSettings_InvalidFile(t *testing.T) {
	// Clean up any existing settings file
	os.Remove("settings.json")
	defer os.Remove("settings.json")

	// Test loading with invalid JSON
	t.Run("Invalid JSON", func(t *testing.T) {
		if err := os.WriteFile("settings.json", []byte("invalid json"), 0666); err != nil {
			t.Fatalf("Failed to write invalid settings file: %v", err)
		}

		settings := LoadSettings()
		if settings == nil {
			t.Fatal("Expected non-nil settings")
		}
		if settings.Port != 3004 {
			t.Errorf("Expected default port 3004, got %d", settings.Port)
		}
	})

	// Test loading with directory instead of file
	t.Run("Directory Instead of File", func(t *testing.T) {
		os.Remove("settings.json")
		if err := os.Mkdir("settings.json", 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		settings := LoadSettings()
		if settings == nil {
			t.Fatal("Expected non-nil settings")
		}
		if settings.Port != 3004 {
			t.Errorf("Expected default port 3004, got %d", settings.Port)
		}

		os.RemoveAll("settings.json")
	})
}

func TestWriteSettings(t *testing.T) {
	// Clean up any existing settings file
	os.Remove("settings.json")
	defer os.Remove("settings.json")

	settings := &Settings{
		Port: 9090,
	}

	if err := writeSettings(settings); err != nil {
		t.Fatalf("Failed to write settings: %v", err)
	}

	// Read and verify file contents
	data, err := os.ReadFile("settings.json")
	if err != nil {
		t.Fatalf("Failed to read settings file: %v", err)
	}

	var loadedSettings Settings
	if err := json.Unmarshal(data, &loadedSettings); err != nil {
		t.Fatalf("Failed to unmarshal settings: %v", err)
	}

	if loadedSettings.Port != settings.Port {
		t.Errorf("Expected port %d, got %d", settings.Port, loadedSettings.Port)
	}
}

func TestDefaultSettings(t *testing.T) {
	// Clean up any existing settings file
	os.Remove("settings.json")
	defer os.Remove("settings.json")

	settings := defaultSettings()
	if settings == nil {
		t.Fatal("Expected non-nil settings")
	}

	if settings.Port != 3004 {
		t.Errorf("Expected default port 3004, got %d", settings.Port)
	}

	// Verify file was created with default settings
	data, err := os.ReadFile("settings.json")
	if err != nil {
		t.Fatalf("Failed to read settings file: %v", err)
	}

	var loadedSettings Settings
	if err := json.Unmarshal(data, &loadedSettings); err != nil {
		t.Fatalf("Failed to unmarshal settings: %v", err)
	}

	if loadedSettings.Port != settings.Port {
		t.Errorf("Expected port %d, got %d", settings.Port, loadedSettings.Port)
	}
}
