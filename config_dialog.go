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
	// 步骤1: 获取当前输入法
	currentIM, err := GetCurrentInputMethod()
	if err != nil {
		ShowMessageBox("错误", fmt.Sprintf("获取当前输入法失败: %v", err), 0x10)
		return
	}

	// 步骤2: 显示当前输入法并询问是否继续
	message := "━━━━━━━━━━━━━━━━━━━━━━\n"
	message += "    快速绑定输入法向导\n"
	message += "━━━━━━━━━━━━━━━━━━━━━━\n\n"
	message += fmt.Sprintf("✓ 检测到当前输入法:\n  %s\n\n", currentIM)
	message += "请按照以下步骤操作:\n\n"
	message += "【步骤1】切换到目标输入法\n"
	message += "  请在系统中切换到您想要绑定的\n"
	message += "  输入法(例如:英文/中文)\n\n"
	message += "【步骤2】点击「确定」继续\n"
	message += "  切换好后,点击确定按钮\n\n"
	message += "【步骤3】选择快捷键\n"
	message += "  在下一步选择要绑定的快捷键\n\n"
	message += "━━━━━━━━━━━━━━━━━━━━━━\n\n"
	message += "提示: 点击「取消」退出向导"

	ret := ShowMessageBoxWithButton("快速绑定 - 步骤 1/3", message, 0x40|0x1) // MB_ICONINFORMATION | MB_OKCANCEL
	if ret != 1 {                                                       // 用户点击取消
		return
	}

	// 步骤3: 再次获取输入法(用户可能已切换)
	targetIM, err := GetCurrentInputMethod()
	if err != nil {
		ShowMessageBox("错误", fmt.Sprintf("获取输入法失败: %v", err), 0x10)
		return
	}

	// 步骤4: 选择快捷键
	message = "━━━━━━━━━━━━━━━━━━━━━━\n"
	message += "    快速绑定 - 选择快捷键\n"
	message += "━━━━━━━━━━━━━━━━━━━━━━\n\n"
	message += fmt.Sprintf("✓ 目标输入法:\n  %s\n\n", targetIM)
	message += "请选择快捷键组合:\n\n"
	message += "【推荐组合】\n"
	message += "  Win+J / Win+K / Win+L\n"
	message += "  Win+I / Win+U / Win+O\n\n"
	message += "【其他组合】\n"
	message += "  Ctrl+Alt+字母键\n"
	message += "  Win+数字键\n\n"
	message += "━━━━━━━━━━━━━━━━━━━━━━\n\n"
	message += "点击「确定」查看详细配置说明"

	ret = ShowMessageBoxWithButton("快速绑定 - 步骤 2/3", message, 0x40|0x1)
	if ret != 1 {
		return
	}

	// 步骤5: 显示配置示例
	message = "━━━━━━━━━━━━━━━━━━━━━━\n"
	message += "    配置文件编辑说明\n"
	message += "━━━━━━━━━━━━━━━━━━━━━━\n\n"
	message += "请在 config.json 中添加:\n\n"
	message += "{\n"
	message += "  \"key_bindings\": [\n"
	message += "    {\n"
	message += "      \"modifier_key\": 91,\n"
	message += "      \"function_key\": 74,\n"
	message += fmt.Sprintf("      \"im_key\": \"%s\",\n", targetIM)
	message += fmt.Sprintf("      \"description\": \"%s\"\n", getIMDescription(targetIM))
	message += "    }\n"
	message += "  ]\n"
	message += "}\n\n"
	message += "━━━━━━━━━━━━━━━━━━━━━━\n\n"
	message += "【按键代码参考】\n"
	message += "  修饰键: Win=91, Ctrl=162, Alt=164\n"
	message += "  功能键: J=74, K=75, L=76\n\n"
	message += "保存后点击托盘菜单中的\n"
	message += "「重新加载配置」即可生效!\n\n"
	message += "━━━━━━━━━━━━━━━━━━━━━━\n\n"
	message += "点击「确定」查看虚拟键码表"

	ret = ShowMessageBoxWithButton("快速绑定 - 步骤 3/3", message, 0x40|0x1)
	if ret == 1 {
		// 用户想查看虚拟键码表
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
	message := "常用虚拟键码参考(十进制):\n\n"
	message += "【修饰键】\n"
	message += "  91  - 左 Win 键 (LWin)\n"
	message += "  92  - 右 Win 键 (RWin)\n"
	message += "  160 - 左 Shift (LShift)\n"
	message += "  161 - 右 Shift (RShift)\n"
	message += "  162 - 左 Ctrl (LCtrl)\n"
	message += "  163 - 右 Ctrl (RCtrl)\n"
	message += "  164 - 左 Alt (LAlt)\n"
	message += "  165 - 右 Alt (RAlt)\n\n"

	message += "【字母键 A-Z】\n"
	message += "  A=65  B=66  C=67  D=68  E=69\n"
	message += "  F=70  G=71  H=72  I=73  J=74\n"
	message += "  K=75  L=76  M=77  N=78  O=79\n"
	message += "  P=80  Q=81  R=82  S=83  T=84\n"
	message += "  U=85  V=86  W=87  X=88  Y=89\n"
	message += "  Z=90\n\n"

	message += "【数字键 0-9】\n"
	message += "  0=48  1=49  2=50  3=51  4=52\n"
	message += "  5=53  6=54  7=55  8=56  9=57\n\n"

	message += "【功能键 F1-F12】\n"
	message += "  F1=112  F2=113  F3=114  F4=115\n"
	message += "  F5=116  F6=117  F7=118  F8=119\n"
	message += "  F9=120  F10=121 F11=122 F12=123\n\n"

	message += "【方向键】\n"
	message += "  37 - Left (←)\n"
	message += "  38 - Up (↑)\n"
	message += "  39 - Right (→)\n"
	message += "  40 - Down (↓)\n\n"

	message += "【编辑键】\n"
	message += "  33 - Page Up\n"
	message += "  34 - Page Down\n"
	message += "  35 - End\n"
	message += "  36 - Home\n"
	message += "  45 - Insert\n"
	message += "  46 - Delete\n\n"

	message += "【数字小键盘】\n"
	message += "  96=Num0  97=Num1  98=Num2\n"
	message += "  99=Num3  100=Num4 101=Num5\n"
	message += "  102=Num6 103=Num7 104=Num8\n"
	message += "  105=Num9\n"
	message += "  106=Multiply(*) 107=Add(+)\n"
	message += "  109=Subtract(-) 111=Divide(/)\n"
	message += "  110=Decimal(.)\n\n"

	message += "【特殊键】\n"
	message += "  8  - Backspace\n"
	message += "  9  - Tab\n"
	message += "  13 - Enter\n"
	message += "  27 - Escape\n"
	message += "  32 - Space\n"
	message += "  20 - Caps Lock\n"
	message += "  144 - Num Lock\n"
	message += "  145 - Scroll Lock\n\n"

	message += "【符号键】\n"
	message += "  186=; 187=+ 188=, 189=-\n"
	message += "  190=. 191=/ 192=` 219=[\n"
	message += "  220=\\ 221=] 222='\n\n"

	message += "点击「打开完整参考」查看所有虚拟键码"

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
