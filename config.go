package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const (
	configFilename = "config.json"
)

type Config struct {
	GroqAPIKey string `json:"groqAPIKey"`
	Model      string `json:"model"` // Add the model field
}

var (
	config     Config
	configOnce sync.Once
	configPath string
)

func init() {
	// Determine the config file path
	configDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("Error getting user config dir:", err)
		// Fallback to current directory
		configDir = "."
	}
	configPath = filepath.Join(configDir, "whisgo", configFilename)
}

func GetConfig() Config {
	configOnce.Do(func() {
		// Load the config file
		file, err := os.Open(configPath)
		if err == nil {
			defer file.Close()
			decoder := json.NewDecoder(file)
			err = decoder.Decode(&config)
			if err != nil {
				fmt.Println("Error decoding config:", err)
			}
		} else {
			fmt.Println("Error opening config file:", err)
		}
	})
	return config
}

func SaveConfig(cfg Config) error {
	// Ensure the directory exists
	dir := filepath.Dir(configPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	// Save the config file
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(cfg)
	if err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	config = cfg // Update the global config
	return nil
}
