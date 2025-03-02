package main

import (
	"context"
	"fmt"
	"io"
	"os/exec"
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

func (a *App) Greet(name string) string {
    return fmt.Sprintf("Hello %s, It's show time!", name)
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

