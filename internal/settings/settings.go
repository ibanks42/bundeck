package settings

import (
	"encoding/json"
	"io"
	"os"
)

type Settings struct {
	Port int `json:"port"`
}

func LoadSettings() *Settings {
	fi, err := os.Stat("settings.json")
	if err != nil {
		return defaultSettings()
	}

	if !fi.Mode().IsRegular() {
		return defaultSettings()
	}

	var s *Settings

	f, err := os.Open(fi.Name())
	if err != nil {
		return defaultSettings()
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return defaultSettings()
	}
	err = json.Unmarshal(b, &s)
	if err != nil {
		return defaultSettings()
	}

	writeSettings(s)

	return s
}

func writeSettings(s *Settings) error {
	j, err := json.MarshalIndent(&s, "", "\t")
	if err != nil {
		return err
	}

	err = os.WriteFile("settings.json", j, 0666)
	return err
}

func defaultSettings() *Settings {
	s := &Settings{
		Port: 3004,
	}

	writeSettings(s)

	return s
}
