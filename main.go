package main

import (
    "embed"

    "github.com/wailsapp/wails/v2"
    "github.com/wailsapp/wails/v2/pkg/options"
    "github.com/wailsapp/wails/v2/pkg/options/assetserver"
    "github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
    app := NewApp()

    // Create application with options
    err := wails.Run(&options.App{
        Title:  "whisgo",
        Width:  500,
        Height: 200,
        AlwaysOnTop:        true,
        DisableResize:      true,
        AssetServer: &assetserver.Options{
            Assets: assets,
        },
        BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
        OnStartup:        app.startup,
        Bind: []interface{}{
            app,
        },
        
        Mac: &mac.Options{
            TitleBar: &mac.TitleBar{
                TitlebarAppearsTransparent: false,
                HideTitle:                  false,
                HideTitleBar:               false,
                FullSizeContent:            false,
                UseToolbar:                 false,
                HideToolbarSeparator:       true,
            },
            Appearance: mac.NSAppearanceNameDarkAqua,
            // Enable this to continue running in the background when window is closed
            WebviewIsTransparent:         true,
            WindowIsTranslucent:          true,
            About: &mac.AboutInfo{
                Title:   "Whisgo",
                Message: "Global keyboard shortcut: Option+Shift+W",
                Icon:    nil,
            },
        },
        // This will keep the app running in the background
        HideWindowOnClose: true,
    })

    if err != nil {
        println("Error:", err.Error())
    }
}