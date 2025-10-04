//go:build windows
// +build windows

package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// IMSelectCommand im-select 命令行工具入口
func IMSelectCommand() {
	// 定义命令行参数
	listFlag := flag.Bool("l", false, "列出所有可用的输入法")
	verboseFlag := flag.Bool("v", false, "显示详细信息")
	switchFlag := flag.String("s", "", "切换到指定的输入法 (HKL 或名称)")
	helpFlag := flag.Bool("h", false, "显示帮助信息")

	flag.Parse()

	// 显示帮助
	if *helpFlag {
		showHelp()
		return
	}

	// 列出所有输入法
	if *listFlag {
		listInputMethods(*verboseFlag)
		return
	}

	// 切换输入法
	if *switchFlag != "" {
		switchInputMethod(*switchFlag)
		return
	}

	// 默认: 显示当前输入法
	showCurrentInputMethod(*verboseFlag)
}

// showHelp 显示帮助信息
func showHelp() {
	help := `
增强版 im-select - Windows 输入法切换工具

用法:
    im-select [选项]

选项:
    (无参数)    显示当前输入法的 HKL
    -l          列出所有可用的输入法
    -v          显示详细信息 (与其他选项配合使用)
    -s <HKL>    切换到指定的输入法
    -h          显示此帮助信息

示例:
    # 显示当前输入法
    im-select
    
    # 显示当前输入法详细信息
    im-select -v
    
    # 列出所有输入法
    im-select -l
    
    # 列出所有输入法(详细)
    im-select -l -v
    
    # 切换到英语
    im-select -s 0x04090409
    
    # 切换到中文
    im-select -s 0x08040804
    
    # 通过名称切换
    im-select -s 00000804

说明:
    - HKL (Handle to Keyboard Layout) 是完整的 32/64 位值
    - 低 16 位是语言 ID (LANGID)
    - 高 16 位是布局 ID,可以区分不同的中文输入法
    - 例如: 0xE00E0804 (微软拼音) vs 0x00000804 (标准中文键盘)

完整文档:
    https://github.com/YourRepo/switch-input-method
`
	fmt.Println(help)
}

// showCurrentInputMethod 显示当前输入法
func showCurrentInputMethod(verbose bool) {
	info, err := GetCurrentInputMethodEnhanced()
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	if verbose {
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("当前输入法信息:")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println(FormatInputMethodInfo(info, true))
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	} else {
		// 兼容原版 im-select: 只输出 HKL
		fmt.Printf("0x%08X\n", uint32(info.HKL))
	}
}

// listInputMethods 列出所有输入法
func listInputMethods(verbose bool) {
	methods, err := GetAllInputMethods()
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	if verbose {
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("检测到 %d 个输入法:\n", len(methods))
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println()

		for i, info := range methods {
			fmt.Printf("[%d] %s\n", i+1, info.DisplayName)
			fmt.Println(FormatInputMethodInfo(info, true))
			fmt.Println()
		}
	} else {
		fmt.Println("HKL        语言ID  布局ID  描述")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		for _, info := range methods {
			fmt.Printf("0x%08X  0x%04X  0x%04X  %s\n",
				info.HKL,
				info.LangID,
				info.LayoutID,
				info.Description,
			)
		}
	}
}

// switchInputMethod 切换输入法
func switchInputMethod(target string) {
	var err error

	// 尝试解析为 HKL (十六进制数)
	if strings.HasPrefix(target, "0x") || strings.HasPrefix(target, "0X") {
		// 去掉 0x 前缀
		hexStr := target[2:]
		hkl, parseErr := strconv.ParseUint(hexStr, 16, 64)
		if parseErr == nil {
			err = SwitchInputMethod(uintptr(hkl))
			if err == nil {
				if isVerbose() {
					fmt.Printf("✓ 已切换到输入法: 0x%08X\n", hkl)

					// 验证切换结果
					current, _ := GetCurrentInputMethodEnhanced()
					if current != nil {
						fmt.Println("\n当前输入法:")
						fmt.Println(FormatInputMethodInfo(current, true))
					}
				} else {
					fmt.Printf("0x%08X\n", hkl)
				}
				return
			}
		}
	}

	// 尝试作为布局名称
	err = SwitchInputMethodByName(target)
	if err == nil {
		if isVerbose() {
			fmt.Printf("✓ 已切换到输入法: %s\n", target)

			// 验证切换结果
			current, _ := GetCurrentInputMethodEnhanced()
			if current != nil {
				fmt.Println("\n当前输入法:")
				fmt.Println(FormatInputMethodInfo(current, true))
			}
		} else {
			current, _ := GetCurrentInputMethodEnhanced()
			if current != nil {
				fmt.Printf("0x%08X\n", uint32(current.HKL))
			}
		}
		return
	}

	// 切换失败
	fmt.Fprintf(os.Stderr, "错误: 无法切换到输入法 '%s'\n", target)
	fmt.Fprintf(os.Stderr, "提示: 使用 'im-select -l' 查看所有可用的输入法\n")
	os.Exit(1)
}

// isVerbose 检查是否启用详细模式
func isVerbose() bool {
	for _, arg := range os.Args {
		if arg == "-v" {
			return true
		}
	}
	return false
}
