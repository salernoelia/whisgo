package main

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.design/x/hotkey"
	_ "golang.design/x/hotkey"
	_ "golang.design/x/hotkey/mainthread"
	"golang.design/x/mainthread"
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

    mainthread.Init(a.RegisterHotKey)
}

func (a *App) shutdown(ctx context.Context) {
    if a.isRecording {
        a.StopRecordingMicrophone()
    }
}

func (a *App) RegisterHotKey() {
	registerHotkey(a)
}


func registerHotkey(a *App) {
	// the actual shortcut keybind - Ctrl + Shift + S
	// for more info - refer to the golang.design/x/hotkey documentation
	hk := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl}, hotkey.KeySpace)
	err := hk.Register()
	if err != nil {
		return
	}
	<-hk.Keydown()
    a.toggleRecord()

	hk.Unregister()

	registerHotkey(a)
}

func (a *App) Greet(name string) string {
    return fmt.Sprintf("Hello %s, It's show time!", name)
}


// func (a *App) toggleVisibility() {
//     a.mu.Lock()
//     defer a.mu.Unlock()
    
//     if a.isVisible {
//         wailsRuntime.WindowHide(a.ctx)
//         a.isVisible = false
//     } else {
//         wailsRuntime.WindowShow(a.ctx)
//         wailsRuntime.WindowSetAlwaysOnTop(a.ctx, true)
//         time.Sleep(100 * time.Millisecond)
//         wailsRuntime.WindowSetAlwaysOnTop(a.ctx, false)
//         a.isVisible = true
//     }
// }


func (a *App) toggleRecord() {
    a.mu.Lock()
    defer a.mu.Unlock()
    
    if a.isRecording {
        a.StopRecordingMicrophone()
        wailsRuntime.EventsEmit(a.ctx, "recording-stopped")
    } else {
        go func() {
            result := a.StartRecordingMicrophone()
            wailsRuntime.EventsEmit(a.ctx, "recording-started", result)
        }()
    }
}