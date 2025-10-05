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
	procKeybd_event         = user32.NewProc("keybd_event")

	keyboardHook      windows.Handle
	modifierKeyStates map[uint32]bool // 修饰键状态映射
	// 当某个修饰键（尤其是 Win）被我们拦截按下后，
	// 如果检测到非配置组合的普通按键，我们会为该修饰键启动一次“透传模式”，
	// 通过模拟该修饰键的按下，让系统能识别诸如 Win+D/Win+V 等快捷键。
	modifierPassthrough map[uint32]bool // 修饰键透传模式标记
	switchMutex         sync.Mutex
	switchTriggered     bool // 标志：是否触发了输入法切换
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

	KEYEVENTF_KEYUP = 0x0002 // keybd_event 的释放标志

	// 用于标记模拟事件的魔术数字，避免递归
	SIMULATED_EVENT_MARKER = 0x12345678
)

// KBDLLHOOKSTRUCT 键盘钩子结构
type KBDLLHOOKSTRUCT struct {
	VkCode      uint32
	ScanCode    uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

// 模拟按下并释放修饰键（用于保持单独修饰键功能）
func simulateModifierKeyPress(vkCode uint32) {
	fmt.Printf("模拟修饰键 %s 按下和释放以保持单独修饰键功能 (VK: %d)\n", GetKeyName(vkCode), vkCode)
	// 按下修饰键，使用 dwExtraInfo 标记这是模拟事件
	procKeybd_event.Call(
		uintptr(vkCode),
		0,
		0,
		uintptr(SIMULATED_EVENT_MARKER),
	)
	// 释放修饰键，同样标记
	procKeybd_event.Call(
		uintptr(vkCode),
		0,
		uintptr(KEYEVENTF_KEYUP),
		uintptr(SIMULATED_EVENT_MARKER),
	)
}

// 键盘钩子回调函数
func keyboardHookProc(nCode int, wParam uintptr, lParam uintptr) uintptr {
	if nCode >= 0 {
		kbdStruct := (*KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam))
		vkCode := kbdStruct.VkCode

		// 检查是否是模拟事件，如果是则直接放行，避免递归
		if kbdStruct.DwExtraInfo == SIMULATED_EVENT_MARKER {
			ret, _, _ := procCallNextHookEx.Call(
				uintptr(keyboardHook),
				uintptr(nCode),
				wParam,
				lParam,
			)
			return ret
		}

		// 获取当前配置
		config := GetCurrentConfig()
		if config == nil {
			// 如果配置未初始化，直接传递事件
			ret, _, _ := procCallNextHookEx.Call(
				uintptr(keyboardHook),
				uintptr(nCode),
				wParam,
				lParam,
			)
			return ret
		}

		// 收集所有可能的修饰键
		allModifierKeys := make(map[uint32]bool)
		for _, binding := range config.KeyBindings {
			allModifierKeys[binding.ModifierKey] = true
		}

		// 检测修饰键按下
		if wParam == WM_KEYDOWN || wParam == WM_SYSKEYDOWN {
			if _, isModifier := allModifierKeys[vkCode]; isModifier {
				if modifierKeyStates == nil {
					modifierKeyStates = make(map[uint32]bool)
				}
				if modifierPassthrough == nil {
					modifierPassthrough = make(map[uint32]bool)
				}
				modifierKeyStates[vkCode] = true
				modifierPassthrough[vkCode] = false // 新一次按下，重置透传模式
				switchTriggered = false             // 重置切换标志
				fmt.Printf("修饰键 %s 按下 - 阻止传递\n", GetKeyName(vkCode))
				return 1 // 阻止修饰键按下事件的传递
			}

			// 检测按键组合是否匹配配置
			matched := false
			for _, binding := range config.KeyBindings {
				// 当某修饰键处于透传模式时，不再触发切换，以免与系统快捷键冲突
				if modifierKeyStates[binding.ModifierKey] && !modifierPassthrough[binding.ModifierKey] && vkCode == binding.FunctionKey {
					fmt.Printf("检测到 %s+%s，切换到 %s\n",
						GetKeyName(binding.ModifierKey),
						GetKeyName(binding.FunctionKey),
						binding.Description)
					switchTriggered = true // 标记已触发切换

					// 复制 imKey 以避免闭包问题
					imKey := binding.IMKey
					go func() {
						switchMutex.Lock()
						defer switchMutex.Unlock()
						switchInputIfNeeded(imKey)
					}()
					matched = true
					// 返回 1 阻止事件传递给系统
					return 1
				}
			}

			// 若未匹配到任何配置组合，但此时有修饰键（例如 Win）处于按下状态，
			// 则为该修饰键启动透传模式：模拟一次该修饰键按下，让系统识别如 Win+D/Win+V。
			if !matched && modifierKeyStates != nil {
				for mk, pressed := range modifierKeyStates {
					if !pressed {
						continue
					}
					// 仅当还未开启透传时模拟一次按下
					if modifierPassthrough != nil && !modifierPassthrough[mk] {
						// 仅对 Win 键做透传是最核心的诉求，其他修饰键保持原逻辑
						if mk == VK_LWIN || mk == VK_RWIN {
							fmt.Printf("为 %s 启动透传模式（模拟按下），以允许系统快捷键生效\n", GetKeyName(mk))
							procKeybd_event.Call(
								uintptr(mk),
								0,
								0,
								uintptr(SIMULATED_EVENT_MARKER),
							)
							modifierPassthrough[mk] = true
						}
					}
				}
				// 不返回 1，允许该普通按键事件继续传递给系统
			}
		}

		// 检测修饰键释放
		if wParam == WM_KEYUP || wParam == WM_SYSKEYUP {
			if _, isModifier := allModifierKeys[vkCode]; isModifier {
				wasPressed := modifierKeyStates[vkCode]
				triggered := switchTriggered
				passthrough := false
				if modifierPassthrough != nil {
					passthrough = modifierPassthrough[vkCode]
				}
				if modifierKeyStates != nil {
					modifierKeyStates[vkCode] = false
				}
				if modifierPassthrough != nil {
					// 重置透传标记（稍后可能需要用它判断是否模拟弹起）
					// 此处先读取到局部变量 passthrough，再清空标记
					modifierPassthrough[vkCode] = false
				}

				fmt.Printf("修饰键 %s 释放 (切换操作: %v)\n", GetKeyName(vkCode), triggered)

				// 如果触发了切换操作，不做任何动作（阻止释放事件）
				if wasPressed && triggered {
					fmt.Println("触发了切换操作 - 阻止释放事件传递")
					return 1 // 阻止原始释放事件传递
				}

				// 未触发切换：
				// - 若此前为 Win 开启过透传（系统已感知到 Win 被按下），此时仅模拟 Win 弹起即可；
				// - 若未开启透传：
				//   * 对 Win：不再模拟点按，直接阻止释放事件，避免开始菜单弹出；
				//   * 对其它修饰键：保留原有“单键点按”的功能，仍然模拟一次按下/弹起。
				if wasPressed {
					if passthrough && (vkCode == VK_LWIN || vkCode == VK_RWIN) {
						fmt.Println("Win 透传模式结束 - 模拟 Win 弹起")
						procKeybd_event.Call(
							uintptr(vkCode),
							0,
							uintptr(KEYEVENTF_KEYUP),
							uintptr(SIMULATED_EVENT_MARKER),
						)
						return 1 // 阻止原始释放事件传递
					}
					if vkCode == VK_LWIN || vkCode == VK_RWIN {
						// 未透传且未切换：完全吞掉 Win 的释放，不做任何模拟，避免开始菜单
						fmt.Println("未触发切换且未透传 - 吞掉 Win 释放，避免开始菜单")
						return 1
					}
					// 非 Win 修饰键延续旧逻辑：模拟一次点按保持单键功能
					fmt.Println("未触发切换操作 - 模拟非 Win 修饰键点按以保持功能")
					go simulateModifierKeyPress(vkCode) // 异步执行，避免阻塞钩子
					return 1                            // 阻止原始释放事件传递
				}
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

	// 显示当前配置的快捷键
	config := GetCurrentConfig()
	if config != nil && len(config.KeyBindings) > 0 {
		fmt.Println("快捷键设置:")
		for _, binding := range config.KeyBindings {
			modifierName := GetKeyName(binding.ModifierKey)
			functionName := GetKeyName(binding.FunctionKey)
			fmt.Printf("  %s+%s - %s (IMKey: %s)\n",
				modifierName, functionName, binding.Description, binding.IMKey)
		}
	} else {
		fmt.Println("警告: 未加载配置，使用默认配置")
	}

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
