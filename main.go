//go:build windows
// +build windows

package main

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// 嵌入 im-select.exe 到程序中
//
//go:embed im-select.exe
var imSelectData []byte

var (
	// 全局互斥锁名称 - 确保单例运行
	mutexName = "Global\\XingYiStreetHongQiRoadIMESwitcher"
)

func main() {
	fmt.Println("===========================================")
	fmt.Println("  兴宜街道红旗路输入法切换工具")
	fmt.Println("===========================================")

	// 检查是否已有实例在运行
	if !ensureSingleInstance() {
		fmt.Println("⚠️  程序已在运行中!")
		// 使用系统弹窗提示用户
		ShowMessageBox(
			"兴宜街道红旗路输入法切换工具",
			"程序已在运行中！\n\n请检查系统托盘图标。\n如需退出，请右键托盘图标选择退出。",
			0x30, // MB_ICONWARNING
		)
		return
	}

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

// ensureSingleInstance 确保程序只运行一个实例
func ensureSingleInstance() bool {
	kernel32 := windows.NewLazyDLL("kernel32.dll")
	procCreateMutex := kernel32.NewProc("CreateMutexW")

	mutexNamePtr, err := windows.UTF16PtrFromString(mutexName)
	if err != nil {
		log.Printf("创建互斥锁名称失败: %v", err)
		return true
	}

	// 创建互斥锁
	handle, _, err := procCreateMutex.Call(
		uintptr(0),
		uintptr(0),
		uintptr(unsafe.Pointer(mutexNamePtr)),
	)

	if handle == 0 {
		log.Printf("创建互斥锁失败: %v", err)
		return true
	}

	// 检查是否已存在
	if err != nil && err.Error() == "The operation completed successfully." {
		return true
	}

	// ERROR_ALREADY_EXISTS = 183
	if err != nil {
		errno, ok := err.(syscall.Errno)
		if ok && errno == 183 {
			return false // 已有实例在运行
		}
	}

	return true
}

// 切换输入法
// * 1033 = 英语(美国), 2052 = 中文(中国)
func switchInputIfNeeded(imkey string) {
	// 获取im-select.exe路径 (与程序同目录)
	imSelectPath := getIMSelectPath()
	cmd := exec.Command(imSelectPath, imkey)

	// 隐藏控制台窗口 - 这是关键!
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}

	err := cmd.Run()
	if err != nil {
		errMsg := fmt.Sprintf("切换输入法失败\n\n错误信息：%v\n\n请确认 im-select.exe 正常工作。", err)
		fmt.Printf("❌ 切换输入法失败: %v\n", err)
		// 显示错误弹窗
		ShowMessageBox(
			"输入法切换失败",
			errMsg,
			0x10, // MB_ICONERROR
		)
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

// getIMSelectPath 获取im-select.exe的路径
// 从嵌入的资源中提取到临时目录
func getIMSelectPath() string {
	// 获取临时目录
	tempDir := os.TempDir()
	imSelectPath := filepath.Join(tempDir, "im-select-xingyijiedao.exe")

	// 检查临时文件是否已存在
	if _, err := os.Stat(imSelectPath); err == nil {
		return imSelectPath
	}

	// 从嵌入的数据中提取文件
	err := ioutil.WriteFile(imSelectPath, imSelectData, 0755)
	if err != nil {
		fmt.Printf("❌ 提取 im-select.exe 失败: %v\n", err)
		// 降级：尝试使用同目录的文件
		exePath, _ := os.Executable()
		exeDir := filepath.Dir(exePath)
		return filepath.Join(exeDir, "im-select.exe")
	}

	return imSelectPath
} // 以下为旧的 gohook 实现，已废弃
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
