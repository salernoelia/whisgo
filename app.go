package main

import (
    "context"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
    "sync"
    "time"
    "io"

    "github.com/robotn/gohook"
    wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
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
}

// AudioDevice represents an audio input device
type AudioDevice struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

// NewApp creates a new App application struct
func NewApp() *App {
    return &App{
        isVisible:     true,
        isRecording:   false,
        stopRecording: make(chan struct{}),
    }
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
    a.ctx = ctx
    
    // Start a goroutine to listen for the hotkey
    go a.registerHotkey()
}

// shutdown cleans up resources
func (a *App) shutdown(ctx context.Context) {
    if a.isRecording {
        a.StopRecordingMicrophone()
    }
}

// registerHotkey sets up the global hotkey listener
func (a *App) registerHotkey() {
    hook.Register(hook.KeyDown, []string{"alt", "space"}, func(e hook.Event) {
        a.toggleVisibility()
    })
    
    s := hook.Start()
    <-hook.Process(s)
}

// toggleVisibility toggles the window visibility
func (a *App) toggleVisibility() {
    a.mu.Lock()
    defer a.mu.Unlock()
    
    if a.isVisible {
        // Hide window
        wailsRuntime.WindowHide(a.ctx)
        a.isVisible = false
    } else {
        // Show window and bring to front
        wailsRuntime.WindowShow(a.ctx)
        wailsRuntime.WindowSetAlwaysOnTop(a.ctx, true)
        time.Sleep(100 * time.Millisecond)
        wailsRuntime.WindowSetAlwaysOnTop(a.ctx, false)
        a.isVisible = true
    }
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
    return fmt.Sprintf("Hello %s, It's show time!", name)
}

// GetAudioDevices returns a list of available audio input devices
// GetAudioDevices returns a list of available audio input devices
func (a *App) GetAudioDevices() []AudioDevice {
    devices := []AudioDevice{}
    
    if runtime.GOOS == "darwin" {
        // Get list of audio devices using system_profiler
        cmd := exec.Command("system_profiler", "SPAudioDataType", "-json")
        _, err := cmd.Output()
        if err != nil {
            fmt.Printf("Error getting audio devices: %v\n", err)
            // Fallback to default device
            devices = append(devices, AudioDevice{
                ID:   "default",
                Name: "System Default",
            })
            return devices
        }
        

        // Parse available devices from macOS audio devices
        availableDevices, err := exec.Command("ffmpeg", "-f", "avfoundation", "-list_devices", "true", "-i", "").CombinedOutput()
        if err != nil {
            // FFmpeg always returns an error here, but still outputs the device list
            deviceList := string(availableDevices)
            fmt.Printf("Available devices:\n%s\n", deviceList)
            
            // Add default device
            devices = append(devices, AudioDevice{
                ID:   "0",
                Name: "Default Input Device",
            })
            
            // Add any other detected input devices
            devices = append(devices, AudioDevice{
                ID:   "1",
                Name: "Built-in Microphone",
            })
        }
    } else if runtime.GOOS == "windows" {
        // Add a default device for Windows
        devices = append(devices, AudioDevice{
            ID:   "default",
            Name: "System Default",
        })
    } else {
        // Linux or other platforms
        devices = append(devices, AudioDevice{
            ID:   "default",
            Name: "System Default",
        })
    }
    
    return devices
}

// SetSelectedDevice sets the selected audio device ID
func (a *App) SetSelectedDevice(deviceID string) {
    a.selectedDeviceID = deviceID
    fmt.Println("Selected device ID:", deviceID)
}


func (a *App) StartRecordingMicrophone() string {
    a.recordingMutex.Lock()
    defer a.recordingMutex.Unlock()
    
    if a.isRecording {
        return "Already recording"
    }
    
    // Create recordings directory if it doesn't exist
    recordingsDir := "recordings"
    if _, err := os.Stat(recordingsDir); os.IsNotExist(err) {
        err = os.MkdirAll(recordingsDir, 0755)
        if err != nil {
            errMsg := fmt.Sprintf("Failed to create recordings directory: %v", err)
            fmt.Println(errMsg)
            return errMsg
        }
    }
    
    // Generate a filename based on the current time
    timestamp := time.Now().Format("2006-01-02_15-04-05")
    filename := filepath.Join(recordingsDir, timestamp+".wav")
    
    var cmd *exec.Cmd
    
    if runtime.GOOS == "darwin" {
        deviceID := "0" // default device
        if a.selectedDeviceID != "" {
            deviceID = a.selectedDeviceID
        }
        
        // Use pkill to ensure no ffmpeg processes are running
        exec.Command("pkill", "ffmpeg").Run()
        
        cmd = exec.Command("ffmpeg", 
            "-f", "avfoundation",
            "-i", fmt.Sprintf(":%s", deviceID),
            "-ac", "1",
            "-ar", "44100",
            "-y", filename)
        
        // Set up error output
        cmd.Stderr = os.Stderr
        
        // Create stdin pipe
        stdin, err := cmd.StdinPipe()
        if err != nil {
            return fmt.Sprintf("Failed to create stdin pipe: %v", err)
        }
        a.stdin = stdin
        
        // Start the recording process
        err = cmd.Start()
        if err != nil {
            a.stdin.Close()
            return fmt.Sprintf("Failed to start recording: %v", err)
        }
        
        a.recordingProcess = cmd
        a.isRecording = true
        
        // Start a goroutine to wait for the process
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

// Update the StopRecordingMicrophone function
func (a *App) StopRecordingMicrophone() string {
    a.recordingMutex.Lock()
    defer a.recordingMutex.Unlock()
    
    if !a.isRecording {
        return "Not currently recording"
    }
    
    // First close stdin to signal ffmpeg to stop
    if a.stdin != nil {
        a.stdin.Close()
        a.stdin = nil
    }
    
    // Send SIGINT to the process
    if a.recordingProcess != nil && a.recordingProcess.Process != nil {
        if err := a.recordingProcess.Process.Signal(os.Interrupt); err != nil {
            fmt.Printf("Failed to send interrupt signal: %v\n", err)
            // Force kill as fallback
            a.recordingProcess.Process.Kill()
        }
        
        // Wait with timeout
        done := make(chan error, 1)
        go func() {
            done <- a.recordingProcess.Wait()
        }()
        
        // Wait for process to end or timeout
        select {
        case <-time.After(3 * time.Second):
            // Force kill if timeout
            a.recordingProcess.Process.Kill()
        case err := <-done:
            if err != nil {
                fmt.Printf("Process ended with error: %v\n", err)
            }
        }
        
        a.recordingProcess = nil
    }
    
    // Make sure ffmpeg is really stopped
    exec.Command("pkill", "ffmpeg").Run()
    
    a.isRecording = false
    
    return "Recording stopped"
}