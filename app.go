package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.design/x/hotkey"
	_ "golang.design/x/hotkey"
	_ "golang.design/x/hotkey/mainthread"
	"golang.design/x/mainthread"
)

type App struct {
	ctx context.Context
	mu  sync.Mutex
	hk  *hotkey.Hotkey
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	mainthread.Init(a.RegisterHotKey)
}

func (a *App) shutdown(ctx context.Context) {
	if a.hk != nil {
		a.hk.Unregister()
	}
}

func (a *App) CopyToClipboard(text string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return wailsRuntime.ClipboardSetText(a.ctx, text)
}

func (a *App) ShowWindow() {
	// wailsRuntime.WindowShow(a.ctx)
	wailsRuntime.WindowUnminimise(a.ctx)
}

func (a *App) HideWindow() {
	// wailsRuntime.WindowHide(a.ctx)
	wailsRuntime.WindowMinimise(a.ctx)
}

func (a *App) RegisterHotKey() {
	go a.monitorHotkey()
}

func (a *App) monitorHotkey() {
	a.mu.Lock()

	// Use Alt+Space as the hotkey combination
	// This matches what is shown in the About info
	var modifiers []hotkey.Modifier

	if runtime.GOOS == "darwin" {
		modifiers = []hotkey.Modifier{hotkey.ModCtrl}
	} else {
		modifiers = []hotkey.Modifier{hotkey.ModCtrl}
	}

	a.hk = hotkey.New(modifiers, hotkey.KeySpace)
	a.mu.Unlock()

	err := a.hk.Register()
	if err != nil {
		fmt.Printf("Failed to register hotkey: %v\n", err)
		return
	}

	for {
		select {
		case <-a.hk.Keydown():
			if runtime.GOOS == "darwin" {
				fmt.Println("Hotkey detected: Alt+Space")
			} else {
				fmt.Println("Hotkey detected: Ctrl+Space")
			}

			// Show and focus the window when hotkey is pressed
			wailsRuntime.WindowShow(a.ctx)
			wailsRuntime.WindowSetAlwaysOnTop(a.ctx, true)
			wailsRuntime.WindowSetAlwaysOnTop(a.ctx, false)

			// Then emit the event
			wailsRuntime.EventsEmit(a.ctx, "hotkey-triggered")
		}
	}
}

func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
