package main

import (
	"context"
	"fmt"
	"sync"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.design/x/hotkey"
	_ "golang.design/x/hotkey"
	_ "golang.design/x/hotkey/mainthread"
	"golang.design/x/mainthread"
)

type App struct {
    ctx        context.Context
    mu         sync.Mutex
}

func NewApp() *App {
    return &App{}
}

func (a *App) startup(ctx context.Context) {
    a.ctx = ctx
    mainthread.Init(a.RegisterHotKey)
}

func (a *App) shutdown(ctx context.Context) {
    // No cleanup needed
}

func (a *App) RegisterHotKey() {
	registerHotkey(a)
}

func registerHotkey(a *App) {
	// Register Ctrl+Space hotkey
	// Use numeric constants for hotkey instead of named constants
	// to avoid platform-specific compilation issues
	modifier := 1 // Ctrl key (ModCtrl)
	key := 49     // Space key (KeySpace)
	
	hk := hotkey.New([]hotkey.Modifier{hotkey.Modifier(modifier)}, hotkey.Key(key))
	err := hk.Register()
	if err != nil {
		return
	}
	<-hk.Keydown()
    
    // Emit an event to the frontend
    wailsRuntime.EventsEmit(a.ctx, "hotkey-triggered")

	hk.Unregister()

	registerHotkey(a)
}

func (a *App) Greet(name string) string {
    return fmt.Sprintf("Hello %s, It's show time!", name)
}