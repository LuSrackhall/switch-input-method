package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	hook "github.com/robotn/gohook"
)

// Store 定义事件存储结构
type Store struct {
	Keycode uint16 `json:"keycode"`
	State   string `json:"state"`
}

var Clients_sse_stores sync.Map
var once_stores sync.Once

var OPTION = false

func main() {
	fmt.Println("Listening for keyboard events... Press Ctrl+C to exit.")
	fmt.Println("All keycode values will be printed to help you identify the desired key combination.")

	KeyEventListen()
}

func KeyEventListen() {
	evChan := hook.Start()
	defer hook.End()

	keycode_keycodeChan_map := make(map[uint16]chan hook.Event)

	for ev := range evChan {
		// 防止Keycode为0的未知按键触发
		if ev.Keycode != 0 {
			if ev.Kind == 4 || ev.Kind == 5 { // 只处理 KeyHold(4) 和 KeyUp(5) 事件
				if _, exists := keycode_keycodeChan_map[ev.Keycode]; exists {
					keycode_keycodeChan_map[ev.Keycode] <- ev
				} else {
					keycode_keycodeChan_map[ev.Keycode] = make(chan hook.Event)
					go handleKeyEvent(keycode_keycodeChan_map[ev.Keycode])
					keycode_keycodeChan_map[ev.Keycode] <- ev
				}
			}
		}
	}
}

func handleKeyEvent(evChan chan hook.Event) {
	var key_down_soundIsRun bool = false

	for ev := range evChan {
		if ev.Kind == 4 { // KeyHold
			if !key_down_soundIsRun {
				fmt.Printf("\nKeyHold - Keycode: %d\n", ev.Keycode)
				if ev.Keycode == 56 {
					OPTION = true
				}
				// 检查是否是目标按键组合（比如 Option+J）
				if OPTION == true && ev.Keycode == 36 { // 这里的38需要根据实际观察到的keycode调整
					go switchInputIfNeeded()
				}
				key_down_soundIsRun = true
			}
		}

		if ev.Kind == 5 { // KeyUp
			fmt.Printf("\nKeyUp - Keycode: %d\n", ev.Keycode)
			if ev.Keycode == 56 {
				OPTION = false
			}
			key_down_soundIsRun = false
		}
	}
}

// 检查当前输入法并切换
func switchInputIfNeeded() {
	homeDir := os.Getenv("HOME")
	cmd := exec.Command("defaults", "read", homeDir+"/Library/Preferences/com.apple.HIToolbox.plist", "AppleSelectedInputSources")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error reading input source:", err)
		return
	}

	currentInput := string(output)
	if !strings.Contains(currentInput, `"KeyboardLayout Name" = ABC`) {
		err := exec.Command("osascript", "-e",
			`tell application "System Events" to key code 49 using {control down}`).Run()
		if err != nil {
			fmt.Println("Error switching input:", err)
		} else {
			fmt.Println("Switched input source.")
		}
	}
}
