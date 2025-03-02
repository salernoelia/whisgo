package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)


func (a *App) IsRecording() bool {
    a.recordingMutex.Lock()
    defer a.recordingMutex.Unlock()
    return a.isRecording
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
    
   recordingsDir, err := a.getRecordingsDir()
    if err != nil {
        errMsg := fmt.Sprintf("Failed to get recordings directory: %v", err)
        fmt.Println(errMsg)
        return errMsg
    }

        
    timestamp := time.Now().Format("2006-01-02_15-04-05")
    filename := filepath.Join(recordingsDir, timestamp+".wav")
    
    var cmd *exec.Cmd
    
    if runtime.GOOS == "darwin" {
    deviceID := "0"
    if a.selectedDeviceID != "" {
        deviceID = a.selectedDeviceID
    }
    
    exec.Command("/opt/homebrew/bin/pkill", "ffmpeg").Run() 
    
    cmd = exec.Command("/opt/homebrew/bin/ffmpeg", 
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

    recordingsDir, err := a.getRecordingsDir()
    if err != nil {
        return fmt.Sprintf("Failed to get recordings directory: %v", err)
    }

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
    recordingsDir, err := a.getRecordingsDir()
    if err != nil {
        return fmt.Sprintf("Failed to get recordings directory: %v", err)
    }
    err = os.RemoveAll(recordingsDir)
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

func (a *App) getRecordingsDir() (string, error) {
    userHome, err := os.UserHomeDir()
    if err != nil {
        return "", fmt.Errorf("failed to get user home directory: %v", err)
    }
    
    // Create the application support directory path
    appSupportDir := filepath.Join(userHome, "Library", "Application Support", "Whisgo")
    recordingsDir := filepath.Join(appSupportDir, "recordings")
    
    // Create the directory if it doesn't exist
    if err := os.MkdirAll(recordingsDir, 0755); err != nil {
        return "", fmt.Errorf("failed to create recordings directory: %v", err)
    }
    
    return recordingsDir, nil
}