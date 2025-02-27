package main

import (
    "context"
    "fmt"
    "sync"
    "time"

    "github.com/robotn/gohook"
    "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
    ctx       context.Context
    isVisible bool
    mu        sync.Mutex
}

// NewApp creates a new App application struct
func NewApp() *App {
    return &App{
        isVisible: true,
    }
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methodsÁÁ
func (a *App) startup(ctx context.Context) {
    a.ctx = ctx
    
    // Start a goroutine to listen for the hotkey
    go a.registerHotkey()
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
        runtime.WindowHide(a.ctx)
        a.isVisible = false
    } else {
        // Show window and bring to front
        runtime.WindowShow(a.ctx)
        runtime.WindowSetAlwaysOnTop(a.ctx, true)
        time.Sleep(100 * time.Millisecond)
        runtime.WindowSetAlwaysOnTop(a.ctx, false)
        a.isVisible = true
    }
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
    return fmt.Sprintf("Hello %s, It's show time!", name)
}