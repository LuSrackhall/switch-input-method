//go:build windows
// +build windows

package main

import (
	"fmt"
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
	currentIM, err := GetCurrentInputMethod()
	if err != nil {
		ShowMessageBox("错误", fmt.Sprintf("获取当前输入法失败: %v", err), 0x10)
		return
	}

	message := fmt.Sprintf("当前输入法: %s\n\n", currentIM)
	message += "请选择要绑定的快捷键:\n\n"
	message += "注意: 此功能正在开发中\n"
	message += "当前请手动编辑 config.json 文件:\n\n"
	message += "示例配置:\n"
	message += "{\n"
	message += "  \"key_bindings\": [\n"
	message += "    {\n"
	message += "      \"modifier_key\": 91,  // 91=Win键\n"
	message += "      \"function_key\": 74,  // 74=J键\n"
	message += fmt.Sprintf("      \"im_key\": \"%s\",\n", currentIM)
	message += "      \"description\": \"切换到当前输入法\"\n"
	message += "    }\n"
	message += "  ]\n"
	message += "}\n\n"
	message += "修改后重启程序生效。"

	ShowMessageBox("快速绑定", message, 0x40)
}

// ShowKeyCodeReference 显示虚拟键码参考
func ShowKeyCodeReference() {
	message := "常用虚拟键码参考(十进制):\n\n"
	message += "修饰键:\n"
	message += "  91  - 左 Win 键\n"
	message += "  92  - 右 Win 键\n"
	message += "  160 - 左 Shift\n"
	message += "  161 - 右 Shift\n"
	message += "  162 - 左 Ctrl\n"
	message += "  163 - 右 Ctrl\n"
	message += "  164 - 左 Alt\n"
	message += "  165 - 右 Alt\n\n"
	message += "字母键 A-Z:\n"
	message += "  A=65  B=66  C=67  D=68  E=69\n"
	message += "  F=70  G=71  H=72  I=73  J=74\n"
	message += "  K=75  L=76  M=77  N=78  O=79\n"
	message += "  P=80  Q=81  R=82  S=83  T=84\n"
	message += "  U=85  V=86  W=87  X=88  Y=89\n"
	message += "  Z=90\n\n"
	message += "数字键 0-9:\n"
	message += "  0=48  1=49  2=50  3=51  4=52\n"
	message += "  5=53  6=54  7=55  8=56  9=57\n\n"
	message += "功能键 F1-F12:\n"
	message += "  F1=112  F2=113  F3=114  F4=115\n"
	message += "  F5=116  F6=117  F7=118  F8=119\n"
	message += "  F9=120  F10=121 F11=122 F12=123\n\n"
	message += "完整参考请访问:\n"
	message += "https://learn.microsoft.com/\n"
	message += "windows/win32/inputdev/\n"
	message += "virtual-key-codes"

	ShowMessageBox("虚拟键码参考", message, 0x40)
}
