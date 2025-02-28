package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	hook "github.com/robotn/gohook"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
    ctx              context.Context
    isVisible        bool
    mu               sync.Mutex
    isRecording      bool
    recordingMutex   sync.Mutex
    stopRecording    chan struct{}
    recordingProcess *exec.Cmd
    selectedDeviceID string
    stdin            io.WriteCloser
    groqAPIKey       string
    transcriptionHistory []string
    model            string
}

type AudioDevice struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

func NewApp() *App {
    return &App{
        isVisible:     true,
        isRecording:   false,
        stopRecording: make(chan struct{}),
        transcriptionHistory: []string{},
    }
}

func (a *App) startup(ctx context.Context) {
    a.ctx = ctx
    
    config := GetConfig()
    a.groqAPIKey = config.GroqAPIKey
    a.model = config.Model
    go a.registerHotkey()
}

func (a *App) shutdown(ctx context.Context) {
    if a.isRecording {
        a.StopRecordingMicrophone()
    }
}

func (a *App) registerHotkey() {
    hook.Register(hook.KeyDown, []string{"alt", "space"}, func(e hook.Event) {
        a.toggleVisibilityAndRecord()
    })

    hook.Register(hook.KeyDown, []string{"control", "m"}, func(e hook.Event) {
        a.toggleVisibility()
    })
    
    s := hook.Start()
    <-hook.Process(s)
}

func (a *App) IsRecording() bool {
    a.recordingMutex.Lock()
    defer a.recordingMutex.Unlock()
    return a.isRecording
}

func (a *App) toggleVisibilityAndRecord() {
    a.mu.Lock()
    defer a.mu.Unlock()
    
    if a.isVisible {
        wailsRuntime.WindowHide(a.ctx)
        if a.isRecording {
            a.StopRecordingMicrophone()
            wailsRuntime.EventsEmit(a.ctx, "recording-stopped")
        }
        a.isVisible = false
    } else {
        wailsRuntime.WindowShow(a.ctx)
        wailsRuntime.WindowSetAlwaysOnTop(a.ctx, true)
        time.Sleep(100 * time.Millisecond)
        wailsRuntime.WindowSetAlwaysOnTop(a.ctx, false)
        a.isVisible = true
        
        go func() {
            result := a.StartRecordingMicrophone()
            wailsRuntime.EventsEmit(a.ctx, "recording-started", result)
        }()
    }
}

func (a *App) toggleVisibility() {
    a.mu.Lock()
    defer a.mu.Unlock()
    
    if a.isVisible {
        wailsRuntime.WindowHide(a.ctx)
        a.isVisible = false
    } else {
        wailsRuntime.WindowShow(a.ctx)
        wailsRuntime.WindowSetAlwaysOnTop(a.ctx, true)
        time.Sleep(100 * time.Millisecond)
        wailsRuntime.WindowSetAlwaysOnTop(a.ctx, false)
        a.isVisible = true
    }
}

func (a *App) Greet(name string) string {
    return fmt.Sprintf("Hello %s, It's show time!", name)
}
func (a *App) GetAudioDevices() []AudioDevice {
    devices := []AudioDevice{}
    
    if runtime.GOOS == "darwin" {
        cmd := exec.Command("system_profiler", "SPAudioDataType", "-json")
        _, err := cmd.Output()
        if err != nil {
            fmt.Printf("Error getting audio devices: %v\n", err)
            devices = append(devices, AudioDevice{
                ID:   "default",
                Name: "System Default",
            })
            return devices
        }
        
        availableDevices, err := exec.Command("ffmpeg", "-f", "avfoundation", "-list_devices", "true", "-i", "").CombinedOutput()
        if err != nil {
            deviceList := string(availableDevices)
            fmt.Printf("Available devices:\n%s\n", deviceList)
            
            devices = append(devices, AudioDevice{
                ID:   "0",
                Name: "Default Input Device",
            })
            
            devices = append(devices, AudioDevice{
                ID:   "1",
                Name: "Built-in Microphone",
            })
        }
    } else if runtime.GOOS == "windows" {
        devices = append(devices, AudioDevice{
            ID:   "default",
            Name: "System Default",
        })
    } else {
        devices = append(devices, AudioDevice{
            ID:   "default",
            Name: "System Default",
        })
    }
    
    return devices
}

func (a *App) SetSelectedDevice(deviceID string) {
    a.selectedDeviceID = deviceID
    fmt.Println("Selected device ID:", deviceID)
}


func (a *App) StartRecordingMicrophone() string {
    a.recordingMutex.Lock()
    defer a.recordingMutex.Unlock()

    if a.groqAPIKey == "" {
        return "Please set your Groq API key"
    }
    
    if a.isRecording {
        return "Already recording"
    }
    
    recordingsDir := "recordings"
    if _, err := os.Stat(recordingsDir); os.IsNotExist(err) {
        err = os.MkdirAll(recordingsDir, 0755)
        if err != nil {
            errMsg := fmt.Sprintf("Failed to create recordings directory: %v", err)
            fmt.Println(errMsg)
            return errMsg
        }
    }
    
    timestamp := time.Now().Format("2006-01-02_15-04-05")
    filename := filepath.Join(recordingsDir, timestamp+".wav")
    
    var cmd *exec.Cmd
    
    if runtime.GOOS == "darwin" {
        deviceID := "0"
        if a.selectedDeviceID != "" {
            deviceID = a.selectedDeviceID
        }
        
        exec.Command("pkill", "ffmpeg").Run()
        
        cmd = exec.Command("ffmpeg", 
            "-f", "avfoundation",
            "-i", fmt.Sprintf(":%s", deviceID),
            "-ac", "1",
            "-ar", "44100",
            "-y", filename)
        
        cmd.Stderr = os.Stderr
        
        stdin, err := cmd.StdinPipe()
        if err != nil {
            return fmt.Sprintf("Failed to create stdin pipe: %v", err)
        }
        a.stdin = stdin
        
        err = cmd.Start()
        if err != nil {
            a.stdin.Close()
            return fmt.Sprintf("Failed to start recording: %v", err)
        }
        
        a.recordingProcess = cmd
        a.isRecording = true
        
        go func() {
            err := cmd.Wait()
            if err != nil {
                fmt.Printf("Recording process ended with error: %v\n", err)
            }
        }()
        
        return "Recording started"
    } else if runtime.GOOS == "windows" {
        return "Not implemented for Windows yet"
    } else {
        return "Not implemented for Linux yet"
    }
}

func (a *App) StopRecordingMicrophone() string {
    a.recordingMutex.Lock()
    defer a.recordingMutex.Unlock()
    
    if !a.isRecording {
        return "Not currently recording"
    }
    
    if a.stdin != nil {
        a.stdin.Close()
        a.stdin = nil
    }
    
    if a.recordingProcess != nil && a.recordingProcess.Process != nil {
        if err := a.recordingProcess.Process.Signal(os.Interrupt); err != nil {
            fmt.Printf("Failed to send interrupt signal: %v\n", err)
            a.recordingProcess.Process.Kill()
        }
        
        done := make(chan error, 1)
        go func() {
            done <- a.recordingProcess.Wait()
        }()
        
        select {
        case <-time.After(3 * time.Second):
            a.recordingProcess.Process.Kill()
        case err := <-done:
            if err != nil {
                fmt.Printf("Process ended with error: %v\n", err)
            }
        }
        
        a.recordingProcess = nil
    }
    
    exec.Command("pkill", "ffmpeg").Run()
    
    a.isRecording = false

    recordingsDir := "recordings"
    files, err := filepath.Glob(filepath.Join(recordingsDir, "*.wav"))
    if err != nil {
        fmt.Printf("Failed to list recording files: %v\n", err)
        return "Recording stopped, failed to process"
    }

    if len(files) == 0 {
        fmt.Println("No recording files found")
        return "Recording stopped, no audio found"
    }

    filename := files[len(files)-1]
    audioData, err := os.ReadFile(filename)
    if err != nil {
        fmt.Printf("Failed to read audio file: %v\n", err)
        return "Recording stopped, failed to read audio"
    }

    transcription, err := GenerateWhisperTranscription(audioData, "en", a.groqAPIKey)
    if err != nil {
        fmt.Printf("Failed to generate transcription: %v\n", err)
        return "Recording stopped, transcription failed"
    }

    err = AddTranscription(transcription)
    if err != nil {
        fmt.Printf("Failed to add transcription to database: %v\n", err)
    }

    err = wailsRuntime.ClipboardSetText(a.ctx, transcription)
    if err != nil {
        fmt.Printf("Failed to copy to clipboard: %v\n", err)
    }

    a.emitTranscriptionHistory()

    return transcription
}

func (a *App) ClearRecordingsDir() string {
    recordingsDir := "recordings"
    if _, err := os.Stat(recordingsDir); os.IsNotExist(err) {
        return "No recordings directory found"
    }
    
    err := os.RemoveAll(recordingsDir)
    if err != nil {
        fmt.Printf("Failed to remove recordings directory: %v\n", err)
        return fmt.Sprintf("Failed to remove recordings directory: %v", err)
    }

    err = ClearTranscriptions()
    if (err != nil) {
        fmt.Printf("Failed to clear transcriptions from database: %v\n", err)
        return fmt.Sprintf("Failed to clear transcriptions from database: %v", err)
    }

    a.emitTranscriptionHistory()

    return "Recordings directory cleared"
}

// GetGroqAPIKey returns the Groq API key
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

func (a *App) GetTranscriptionHistory() ([]Transcription, error) {
    return GetTranscriptions()
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

func (a *App) emitTranscriptionHistory() {
    transcriptions, err := GetTranscriptions()
    if err != nil {
        fmt.Printf("Failed to get transcriptions: %v\n", err)
        return
    }
    wailsRuntime.EventsEmit(a.ctx, "transcription-history-changed", transcriptions)
}