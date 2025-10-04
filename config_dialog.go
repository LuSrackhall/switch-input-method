//go:build windows
// +build windows

package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

// ShowConfigDialog 显示配置对话框(简化版 - 使用消息框)
func ShowConfigDialog() {
	config := GetCurrentConfig()
	if config == nil {
		ShowMessageBox("错误", "配置未初始化", 0x10)
		return
	}

	// 构建配置信息
	message := "当前按键绑定:\n\n"
	for i, binding := range config.KeyBindings {
		modifierName := GetKeyName(binding.ModifierKey)
		functionName := GetKeyName(binding.FunctionKey)
		message += fmt.Sprintf("%d. %s+%s -> %s\n   (IMKey: %s)\n\n",
			i+1, modifierName, functionName, binding.Description, binding.IMKey)
	}

	message += "\n要修改配置,请编辑程序目录下的 config.json 文件,\n"
	message += "然后重启程序使配置生效。\n\n"
	message += "配置文件路径:\n" + GetConfigPath()

	ShowMessageBox("按键绑定配置", message, 0x40)
}

// ShowAddBindingDialog 显示添加绑定对话框
func ShowAddBindingDialog() {
	message := "添加新的按键绑定:\n\n"
	message += "1. 按下要设置的按键组合(例如: Win+L)\n"
	message += "2. 切换到想要绑定的输入法\n"
	message += "3. 点击\"捕获当前设置\"\n\n"
	message += "注意: 此功能需要进一步开发\n"
	message += "目前请直接编辑 config.json 文件"

	ShowMessageBox("添加按键绑定", message, 0x40)
}

// CapturKeyAndIME 捕获按键组合和当前输入法(示意)
func CapturKeyAndIME() (uint32, uint32, string, error) {
	// 1. 获取当前输入法
	currentIM, err := GetCurrentInputMethod()
	if err != nil {
		return 0, 0, "", err
	}

	// 2. TODO: 捕获用户按下的按键组合
	// 这需要临时的键盘监听逻辑
	// 暂时返回错误提示
	return 0, 0, currentIM, fmt.Errorf("按键捕获功能待实现")
}

// ShowQuickBindingMenu 显示快速绑定菜单
func ShowQuickBindingMenu() {
	// 获取当前输入法
	currentIMStr, err := GetCurrentInputMethod()
	if err != nil {
		ShowMessageBox("错误", fmt.Sprintf("获取当前输入法失败: %v", err), 0x10)
		return
	}

	// 单窗口显示输入法信息和配置示例
	message := "━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	message += "    快速绑定当前输入法\n"
	message += "━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"
	message += fmt.Sprintf("✓ 当前输入法标识:\n  %s\n\n", currentIMStr)
	message += "━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"
	message += "配置示例 (添加到 config.json):\n\n"
	message += "{\n"
	message += "  \"modifier_key\": 91,\n"
	message += "  \"function_key\": 74,\n"
	message += fmt.Sprintf("  \"im_key\": \"%s\",\n", currentIMStr)
	message += fmt.Sprintf("  \"description\": \"%s\"\n", getIMDescription(currentIMStr))
	message += "}\n\n"
	message += "━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"
	message += "常用键码:\n"
	message += "  Win=91  Ctrl=162  Alt=164\n"
	message += "  J=74  K=75  L=76  I=73\n\n"
	message += "修改后点击「配置管理」->「重新加载配置」\n\n"
	message += "━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"
	message += "点击「确定」打开配置文件\n"
	message += "点击「取消」查看完整键码表"

	ret := ShowMessageBoxWithButton("快速绑定当前输入法", message, 0x40|0x1)
	if ret == 1 {
		// 用户点击确定,打开配置文件
		openConfigFile()
	} else if ret == 2 {
		// 用户点击取消,显示键码表
		ShowKeyCodeReference()
	}
}

// getIMDescription 根据输入法路径生成友好的描述
func getIMDescription(imKey string) string {
	if imKey == "1033" || imKey == "0x0409" {
		return "切换到英文"
	}
	if imKey == "2052" || imKey == "0x0804" {
		return "切换到中文"
	}
	// 尝试从路径中提取有意义的信息
	if len(imKey) > 50 {
		return "切换输入法"
	}
	return imKey
}

// ShowKeyCodeReference 显示虚拟键码参考
func ShowKeyCodeReference() {
	message := "━━━━ 虚拟键码参考(十进制) ━━━━\n\n"

	message += "【修饰键】\n"
	message += " 91-LWin   92-RWin   160-LShift 161-RShift\n"
	message += " 162-LCtrl 163-RCtrl 164-LAlt   165-RAlt\n\n"

	message += "【字母键 A-Z】\n"
	message += " A=65 B=66 C=67 D=68 E=69 F=70 G=71\n"
	message += " H=72 I=73 J=74 K=75 L=76 M=77 N=78\n"
	message += " O=79 P=80 Q=81 R=82 S=83 T=84 U=85\n"
	message += " V=86 W=87 X=88 Y=89 Z=90\n\n"

	message += "【数字键 0-9】\n"
	message += " 0=48 1=49 2=50 3=51 4=52\n"
	message += " 5=53 6=54 7=55 8=56 9=57\n\n"

	message += "【功能键 F1-F12】\n"
	message += " F1=112  F2=113  F3=114  F4=115\n"
	message += " F5=116  F6=117  F7=118  F8=119\n"
	message += " F9=120  F10=121 F11=122 F12=123\n\n"

	message += "【方向键】    【编辑键】\n"
	message += " 37-Left←     33-PageUp   35-End\n"
	message += " 38-Up↑       34-PageDn   36-Home\n"
	message += " 39-Right→    45-Insert   46-Delete\n"
	message += " 40-Down↓\n\n"

	message += "【数字小键盘】\n"
	message += " Num0-9: 96-105  *=106 +=107\n"
	message += " -=109 /=111 .=110\n\n"

	message += "【特殊键】\n"
	message += " 8-Backspace 9-Tab    13-Enter  27-Esc\n"
	message += " 32-Space    20-Caps  144-Num   145-Scroll\n\n"

	message += "【符号键】\n"
	message += " ;=186 +=187 ,=188 -=189 .=190 /=191\n"
	message += " `=192 [=219 \\=220 ]=221 '=222\n\n"

	message += "━━━━━━━━━━━━━━━━━━━━━━━━\n"
	message += "提示: 点击「确定」打开官方完整文档"

	// 显示消息框
	ret := ShowMessageBoxWithButton(
		"虚拟键码参考",
		message,
		0x40|0x1, // MB_ICONINFORMATION | MB_OKCANCEL
	)

	// 如果用户点击了确定按钮(IDOK=1),则打开浏览器
	if ret == 1 {
		openURL("https://learn.microsoft.com/windows/win32/inputdev/virtual-key-codes")
	}
}

// ShowMessageBoxWithButton 显示带按钮的消息框并返回用户选择
func ShowMessageBoxWithButton(title, message string, flags uint) int {
	user32 := syscall.NewLazyDLL("user32.dll")
	messageBox := user32.NewProc("MessageBoxW")

	titlePtr, _ := syscall.UTF16PtrFromString(title)
	messagePtr, _ := syscall.UTF16PtrFromString(message)

	ret, _, _ := messageBox.Call(
		uintptr(0),
		uintptr(unsafe.Pointer(messagePtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(flags),
	)

	return int(ret)
}

// openURL 使用默认浏览器打开URL
func openURL(url string) {
	shell32 := syscall.NewLazyDLL("shell32.dll")
	shellExecute := shell32.NewProc("ShellExecuteW")

	operation, _ := syscall.UTF16PtrFromString("open")
	urlPtr, _ := syscall.UTF16PtrFromString(url)

	shellExecute.Call(
		0,
		uintptr(unsafe.Pointer(operation)),
		uintptr(unsafe.Pointer(urlPtr)),
		0,
		0,
		1, // SW_SHOWNORMAL
	)
}
