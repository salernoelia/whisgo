package main

import (
	"fmt"
	"os/exec"
	"runtime"
)

func (a *App) GetAudioDevices() []AudioDevice {
    devices := []AudioDevice{}
    
    if runtime.GOOS == "darwin" {
        cmd := exec.Command("system_profiler", "SPAudioDataType", "-json")
        _, err := cmd.Output()
        if err != nil {
            fmt.Printf("Error getting audio devices: %v\n", err)
            devices = append(devices, AudioDevice{
                ID:   "default",
                Name: "System Default",
            })
            return devices
        }
        
        availableDevices, err := exec.Command("ffmpeg", "-f", "avfoundation", "-list_devices", "true", "-i", "").CombinedOutput()
        if err != nil {
            deviceList := string(availableDevices)
            fmt.Printf("Available devices:\n%s\n", deviceList)
            
            devices = append(devices, AudioDevice{
                ID:   "0",
                Name: "Default Input Device",
            })
            
            devices = append(devices, AudioDevice{
                ID:   "1",
                Name: "Built-in Microphone",
            })
        }
    } else if runtime.GOOS == "windows" {
        devices = append(devices, AudioDevice{
            ID:   "default",
            Name: "System Default",
        })
    } else {
        devices = append(devices, AudioDevice{
            ID:   "default",
            Name: "System Default",
        })
    }
    
    return devices
}

func (a *App) SetSelectedDevice(deviceID string) {
    a.selectedDeviceID = deviceID
    fmt.Println("Selected device ID:", deviceID)
}