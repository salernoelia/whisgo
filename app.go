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

func (a *App) RegisterHotKey() {
	go a.monitorHotkey()
}

func (a *App) monitorHotkey() {

	a.mu.Lock()
	a.hk = hotkey.New([]hotkey.Modifier{hotkey.ModCtrl}, hotkey.KeySpace)
	a.mu.Unlock()

	err := a.hk.Register()
	if err != nil {
		fmt.Printf("Failed to register hotkey: %v\n", err)
		return
	}

	for {
		select {
		case <-a.hk.Keydown():
			fmt.Println("Hotkey detected: Ctrl+Space")

			wailsRuntime.EventsEmit(a.ctx, "hotkey-triggered")
		}
	}
}

func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
