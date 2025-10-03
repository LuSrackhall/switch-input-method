//go:build windows
// +build windows

package main

import (
	"fmt"
	"log"
	"os/exec"
)

func main() {
	fmt.Println("===========================================")
	fmt.Println("  输入法快速切换工具 - Windows 版本")
	fmt.Println("===========================================")
	fmt.Println("正在初始化系统托盘...")

	// 在后台启动键盘钩子
	go func() {
		err := StartKeyboardHook()
		if err != nil {
			log.Fatal("启动键盘钩子失败:", err)
		}
	}()

	// 等待托盘初始化完成后再启动钩子
	// 这样可以确保托盘图标先出现
	fmt.Println("程序已在系统托盘中运行")
	fmt.Println("右键托盘图标可以退出程序")

	// 启动系统托盘(这会阻塞主线程)
	InitTray()

	// 程序退出时会自动调用 onExit 清理资源
}

// 切换输入法
// * 1033 = 英语(美国), 2052 = 中文(中国)
func switchInputIfNeeded(imkey string) {
	err := exec.Command("C:\\Users\\Public\\Downloads\\插件\\vscode插件\\vim插件\\im-select.exe", imkey).Run()
	if err != nil {
		fmt.Printf("❌ 切换输入法失败: %v\n", err)
		return
	}

	inputMethodName := "未知"
	if imkey == "1033" {
		inputMethodName = "英文"
	} else if imkey == "2052" {
		inputMethodName = "中文"
	}
	fmt.Printf("✅ 已切换到%s输入法\n", inputMethodName)
}

// 以下为旧的 gohook 实现，已废弃
/*
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
				if ev.Keycode == 3675 {
					OPTION = true
				}
				// 检查是否是目标按键组合（比如 Option+J）
				if OPTION == true && ev.Keycode == 36 { // 这里的38需要根据实际观察到的keycode调整
					go switchInputIfNeeded("1033")
				}
				// 检查是否是目标按键组合（比如 Option+K）
				if OPTION == true && ev.Keycode == 37 { // 这里的38需要根据实际观察到的keycode调整
					// go switchInputIfNeeded("com.apple.inputmethod.SCIM.Shuangpin")
					go switchInputIfNeeded("2052")

					// go func() {
					// 	switchInputIfNeeded("im.rime.inputmethod.Squirrel.Hans")
					// 	if !mutex.TryLock() {
					// 		// 锁可以保证鼠标回到原有位置
					// 		fmt.Println("锁被占用，放弃执行")
					// 		return
					// 	}
					// 	// err := exec.Command("/Users/srackhalllu/Desktop/资源管理器/safe/输入法按键绑定脚本/focus-shift").Run()
					// 	// err := exec.Command("swift", "/Users/srackhalllu/Desktop/资源管理器/safe/输入法按键绑定脚本/toggle-app-focus.swift").Run()
					// 	err := exec.Command("/Users/srackhalllu/Desktop/资源管理器/safe/输入法按键绑定脚本/toggle-app-focus").Run()
					// 	if err != nil {
					// 		fmt.Println("焦点转移失败", err)
					// 		return
					// 	}
					// 	mutex.Unlock()
					// }()
				}
				key_down_soundIsRun = true
			}
		}

		if ev.Kind == 5 { // KeyUp
			fmt.Printf("\nKeyUp - Keycode: %d\n", ev.Keycode)
			if ev.Keycode == 3675 {
				OPTION = false
			}
			key_down_soundIsRun = false
		}
	}
}

// setupDaemonProcess 设置 Windows 平台的守护进程属性
func setupDaemonProcess(cmd *exec.Cmd) {
	// Windows 平台使用 CREATE_NEW_PROCESS_GROUP 和 DETACHED_PROCESS 标志
	// 来创建独立的后台进程
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | 0x00000008, // DETACHED_PROCESS
	}
}
*/
