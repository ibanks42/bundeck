package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
)

// Script represents a task script with an ID and filename
type Script struct {
	ID   int    `json:"id"`
	File string `json:"file"`
}

var getScriptsPath = func() string {
	return filepath.Join(os.Getenv("HOME"), ".bundeck", "scripts")
}

func globExtensions(dir string, extensions []string) ([]string, error) {
	var matches []string

	for _, ext := range extensions {
		pattern := filepath.Join(dir, "*"+ext)
		files, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}
		matches = append(matches, files...)
	}

	return matches, nil
}

// getScripts reads and returns all available scripts
func getScripts() []Script {
	path := getScriptsPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0755)
	}

	scripts_json := filepath.Join(path, "scripts.json")
	if _, err := os.Stat(scripts_json); os.IsNotExist(err) {
		// get all .js and .ts files in the path
		files, err := globExtensions(path, []string{".js", ".ts", ".jsx", ".tsx", ".mjs", ".cjs"})
		if err != nil {
			log.Fatal(err)
		}
		scripts := make([]Script, len(files))
		for i, file := range files {
			scripts[i] = Script{
				ID:   i + 1,
				File: filepath.Base(file),
			}
		}
		if err := writeScriptsJson(scripts); err != nil {
			log.Fatal(err)
		}
		return scripts
	}

	data, err := os.ReadFile(scripts_json)
	if err != nil {
		log.Fatal(err)
	}

	var scripts []Script
	err = json.Unmarshal(data, &scripts)

	if err != nil || len(scripts) == 0 {
		files, _ := globExtensions(path, []string{".js", ".ts", ".jsx", ".tsx", ".mjs", ".cjs"})
		scripts = make([]Script, len(files))
		for i, file := range files {
			scripts[i] = Script{
				ID:   i + 1,
				File: filepath.Base(file),
			}
		}
		if err := writeScriptsJson(scripts); err != nil {
			log.Fatal(err)
		}
	}

	return scripts
}

// writeScript creates a new script file
func writeScript(id int, filename string, content string) error {
	path := getScriptsPath()
	scripts := getScripts()

	// Check if ID already exists
	if slices.ContainsFunc(scripts, func(s Script) bool { return s.ID == id }) {
		return fmt.Errorf("script with ID %d already exists", id)
	}

	// Write script file
	err := os.WriteFile(filepath.Join(path, filename), []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write script file: %w", err)
	}

	// Update scripts.json
	scripts = append(scripts, Script{ID: id, File: filename})
	return writeScriptsJson(scripts)
}

// updateScript modifies an existing script
func updateScript(script Script, newId int, content string) error {
	path := getScriptsPath()
	scripts := getScripts()

	// Remove old script from list
	scripts = slices.DeleteFunc(scripts, func(s Script) bool {
		return s.ID == script.ID
	})

	// Write updated script file
	err := os.WriteFile(filepath.Join(path, script.File), []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write script file: %w", err)
	}

	// Update scripts.json with new ID
	scripts = append(scripts, Script{ID: newId, File: script.File})
	return writeScriptsJson(scripts)
}

// readScript reads the content of a script file
func readScript(filename string) (string, error) {
	path := getScriptsPath()
	data, err := os.ReadFile(filepath.Join(path, filename))
	if err != nil {
		return "", fmt.Errorf("failed to read script file: %w", err)
	}
	return string(data), nil
}

// writeScriptsJson updates the scripts.json file
func writeScriptsJson(scripts []Script) error {
	path := getScriptsPath()
	data, err := json.MarshalIndent(scripts, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal scripts: %w", err)
	}

	err = os.WriteFile(filepath.Join(path, "scripts.json"), data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write scripts.json: %w", err)
	}

	return nil
}

// getMaxScriptId returns the highest script ID
func getMaxScriptId() int {
	maxId := 0
	for _, v := range getScripts() {
		if v.ID > maxId {
			maxId = v.ID
		}
	}
	return maxId
}
