package main

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"

	hook "github.com/robotn/gohook"
)

// Store 定义事件存储结构
type Store struct {
	Keycode uint16 `json:"keycode"`
	State   string `json:"state"`
}

var Clients_sse_stores sync.Map
var once_stores sync.Once
var mutex sync.Mutex
var OPTION = false

func main() {
	// 	fmt.Println("Listening for keyboard events... Press Ctrl+C to exit.")
	// 	fmt.Println("All keycode values will be printed to help you identify the desired key combination.")

	// 检查是否需要以守护进程方式运行
	if len(os.Args) == 1 {
		// 启动守护进程
		cmd := exec.Command(os.Args[0], "daemon")
		// 设置进程属性
		cmd.Stdin = nil
		cmd.Stdout = nil
		cmd.Stderr = nil
		// 启动时分离进程
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid: true, // 创建新的会话
		}

		cmd.Start()

		fmt.Printf("程序已在后台启动，进程 PID: %d\n", cmd.Process.Pid)
		fmt.Println("要结束程序，请执行: kill", cmd.Process.Pid)
		fmt.Println("----------, 或执行: kill -9", cmd.Process.Pid)

		cmd.Process.Release()
	}

	// 实际的程序逻辑
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
					go switchInputIfNeeded("com.apple.keylayout.UnicodeHexInput")
				}
				// 检查是否是目标按键组合（比如 Option+K）
				if OPTION == true && ev.Keycode == 37 { // 这里的38需要根据实际观察到的keycode调整
					// go switchInputIfNeeded("com.apple.inputmethod.SCIM.Shuangpin")
					// go switchInputIfNeeded("im.rime.inputmethod.Squirrel.Hans")

					go func() {
						switchInputIfNeeded("im.rime.inputmethod.Squirrel.Hans")
						if !mutex.TryLock() {
							// 锁可以保证鼠标回到原有位置
							fmt.Println("锁被占用，放弃执行")
							return
						}
						err := exec.Command("swift", "/Users/srackhalllu/Desktop/资源管理器/safe/输入法按键绑定脚本/focus-shift.swift").Run()
						if err != nil {
							fmt.Println("焦点转移失败", err)
							return
						}
						mutex.Unlock()
					}()
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
// * 传入的参数为 可以判断输入法的UUID(或称input method key), 可通过手动切换到你需要的输入法，然后执行 `im-select` 命令获取
func switchInputIfNeeded(imkey string) {
	err := exec.Command("im-select", imkey).Run()
	if err != nil {
		fmt.Println("切换失败", err)
		return
	}
}
