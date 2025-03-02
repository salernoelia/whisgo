package main

import "fmt"


func (a *App) GetGroqAPIKey() string {
    return a.groqAPIKey
}

// SetGroqAPIKey sets the Groq API key
func (a *App) SetGroqAPIKey(apiKey string) string {
    a.groqAPIKey = apiKey
    config := GetConfig()
    config.GroqAPIKey = apiKey
    err := SaveConfig(config)
    if err != nil {
        fmt.Printf("Failed to save config: %v\n", err)
        return "Failed to save config"
    }
    return "API key saved"
}



func (a *App) GetModel() string {
    return a.model
}

// SetModel sets the selected model
func (a *App) SetModel(model string) string {
    a.model = model
    config := GetConfig()
    config.Model = model
    err := SaveConfig(config)
    if err != nil {
        fmt.Printf("Failed to save config: %v\n", err)
        return "Failed to save config"
    }
    return "Model saved"
}

