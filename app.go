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
	// Using correct hotkey combination for Ctrl+Space
	hk := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl}, hotkey.KeySpace)
	
	err := hk.Register()
	if err != nil {
		fmt.Printf("Failed to register hotkey: %v\n", err)
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