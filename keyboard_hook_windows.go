//go:build windows
// +build windows

package main

import (
	"fmt"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32                  = windows.NewLazyDLL("user32.dll")
	procSetWindowsHookEx    = user32.NewProc("SetWindowsHookExW")
	procCallNextHookEx      = user32.NewProc("CallNextHookEx")
	procUnhookWindowsHookEx = user32.NewProc("UnhookWindowsHookEx")
	procGetMessage          = user32.NewProc("GetMessageW")
	procGetAsyncKeyState    = user32.NewProc("GetAsyncKeyState")

	keyboardHook  windows.Handle
	winKeyPressed bool
	switchMutex   sync.Mutex
)

const (
	WH_KEYBOARD_LL = 13
	WM_KEYDOWN     = 0x0100
	WM_KEYUP       = 0x0101
	WM_SYSKEYDOWN  = 0x0104
	WM_SYSKEYUP    = 0x0105

	VK_LWIN = 0x5B // 左 Win 键
	VK_RWIN = 0x5C // 右 Win 键
	VK_J    = 0x4A // J 键
	VK_K    = 0x4B // K 键
)

// KBDLLHOOKSTRUCT 键盘钩子结构
type KBDLLHOOKSTRUCT struct {
	VkCode      uint32
	ScanCode    uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

// 键盘钩子回调函数
func keyboardHookProc(nCode int, wParam uintptr, lParam uintptr) uintptr {
	if nCode >= 0 {
		kbdStruct := (*KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam))
		vkCode := kbdStruct.VkCode

		// 检测 Win 键状态
		if wParam == WM_KEYDOWN || wParam == WM_SYSKEYDOWN {
			if vkCode == VK_LWIN || vkCode == VK_RWIN {
				winKeyPressed = true
				fmt.Println("Win 键按下")
			}

			// 检测 Win+J (切换到英文)
			if winKeyPressed && vkCode == VK_J {
				fmt.Println("检测到 Win+J，切换到英文输入法")
				go func() {
					switchMutex.Lock()
					defer switchMutex.Unlock()
					switchInputIfNeeded("1033")
				}()
				// 返回 1 阻止事件传递给系统
				return 1
			}

			// 检测 Win+K (切换到中文)
			if winKeyPressed && vkCode == VK_K {
				fmt.Println("检测到 Win+K，切换到中文输入法")
				go func() {
					switchMutex.Lock()
					defer switchMutex.Unlock()
					switchInputIfNeeded("2052")
				}()
				// 返回 1 阻止事件传递给系统
				return 1
			}
		}

		// 检测 Win 键释放
		if wParam == WM_KEYUP || wParam == WM_SYSKEYUP {
			if vkCode == VK_LWIN || vkCode == VK_RWIN {
				winKeyPressed = false
				fmt.Println("Win 键释放")
			}
		}
	}

	// 其他按键正常传递
	ret, _, _ := procCallNextHookEx.Call(
		uintptr(keyboardHook),
		uintptr(nCode),
		wParam,
		lParam,
	)
	return ret
}

// 启动键盘钩子
func StartKeyboardHook() error {
	fmt.Println("正在安装键盘钩子...")
	fmt.Println("快捷键设置:")
	fmt.Println("  Win+J - 切换到英文输入法 (1033)")
	fmt.Println("  Win+K - 切换到中文输入法 (2052)")
	fmt.Println("按 Ctrl+C 退出程序")
	fmt.Println("----------------------------------------")

	// 创建回调函数指针
	callback := syscall.NewCallback(keyboardHookProc)

	// 安装钩子
	hook, _, err := procSetWindowsHookEx.Call(
		uintptr(WH_KEYBOARD_LL),
		callback,
		0,
		0,
	)

	if hook == 0 {
		return fmt.Errorf("安装键盘钩子失败: %v", err)
	}

	keyboardHook = windows.Handle(hook)
	fmt.Println("键盘钩子安装成功！")

	// 消息循环
	var msg struct {
		hwnd    uintptr
		message uint32
		wParam  uintptr
		lParam  uintptr
		time    uint32
		pt      struct{ x, y int32 }
	}

	for {
		ret, _, _ := procGetMessage.Call(
			uintptr(unsafe.Pointer(&msg)),
			0,
			0,
			0,
		)

		if ret == 0 {
			break
		}
	}

	return nil
}

// 停止键盘钩子
func StopKeyboardHook() {
	if keyboardHook != 0 {
		procUnhookWindowsHookEx.Call(uintptr(keyboardHook))
		keyboardHook = 0
		fmt.Println("键盘钩子已卸载")
	}
}
