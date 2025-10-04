//go:build windows
// +build windows

package main

import (
	"flag"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

// Windows API
var (
	user32                     = syscall.NewLazyDLL("user32.dll")
	procGetKeyboardLayoutList  = user32.NewProc("GetKeyboardLayoutList")
	procGetKeyboardLayout      = user32.NewProc("GetKeyboardLayout")
	procActivateKeyboardLayout = user32.NewProc("ActivateKeyboardLayout")
	procLoadKeyboardLayoutW    = user32.NewProc("LoadKeyboardLayoutW")
	procGetKeyboardLayoutNameW = user32.NewProc("GetKeyboardLayoutNameW")
)

const (
	KLF_ACTIVATE      = 0x00000001
	KLF_SETFORPROCESS = 0x00000100
	KLF_REORDER       = 0x00000008
)

// InputMethodInfo 输入法信息
type InputMethodInfo struct {
	HKL         uintptr
	LangID      uint16
	LayoutID    uint16
	Name        string
	DisplayName string
	Description string
}

// GetCurrentInputMethod 获取当前输入法
func GetCurrentInputMethod() (*InputMethodInfo, error) {
	threadID := uintptr(0)
	ret, _, _ := procGetKeyboardLayout.Call(threadID)
	if ret == 0 {
		return nil, fmt.Errorf("获取当前键盘布局失败")
	}

	hkl := ret
	var layoutName [260]uint16
	r1, _, _ := procGetKeyboardLayoutNameW.Call(uintptr(unsafe.Pointer(&layoutName[0])))
	if r1 == 0 {
		return nil, fmt.Errorf("获取键盘布局名称失败")
	}

	name := syscall.UTF16ToString(layoutName[:])
	langID := uint16(hkl & 0xFFFF)
	layoutID := uint16((hkl >> 16) & 0xFFFF)

	info := &InputMethodInfo{
		HKL:         hkl,
		LangID:      langID,
		LayoutID:    layoutID,
		Name:        name,
		DisplayName: getDisplayName(hkl, langID, layoutID),
		Description: getDescription(langID, layoutID),
	}

	return info, nil
}

// GetAllInputMethods 获取所有输入法
func GetAllInputMethods() ([]*InputMethodInfo, error) {
	count, _, _ := procGetKeyboardLayoutList.Call(0, 0)
	if count == 0 {
		return nil, fmt.Errorf("获取键盘布局列表失败")
	}

	hkls := make([]uintptr, count)
	ret, _, _ := procGetKeyboardLayoutList.Call(
		count,
		uintptr(unsafe.Pointer(&hkls[0])),
	)

	if ret == 0 {
		return nil, fmt.Errorf("获取键盘布局列表失败")
	}

	var methods []*InputMethodInfo
	for _, hkl := range hkls[:ret] {
		langID := uint16(hkl & 0xFFFF)
		layoutID := uint16((hkl >> 16) & 0xFFFF)

		info := &InputMethodInfo{
			HKL:         hkl,
			LangID:      langID,
			LayoutID:    layoutID,
			DisplayName: getDisplayName(hkl, langID, layoutID),
			Description: getDescription(langID, layoutID),
		}

		methods = append(methods, info)
	}

	return methods, nil
}

// SwitchInputMethod 切换输入法
func SwitchInputMethod(hkl uintptr) error {
	// 获取所有已安装的输入法
	methods, err := GetAllInputMethods()
	if err != nil {
		return fmt.Errorf("无法获取已安装的输入法列表")
	}

	// 查找匹配的 HKL (比较 32 位值)
	target32 := uint32(hkl)
	var targetHKL uintptr
	found := false

	for _, method := range methods {
		if uint32(method.HKL) == target32 {
			targetHKL = method.HKL
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("HKL 0x%08X 未在系统中找到", target32)
	}

	// 使用 KLF_ACTIVATE 标志激活
	ret, _, _ := procActivateKeyboardLayout.Call(
		targetHKL,
		uintptr(KLF_ACTIVATE),
	)

	if ret == 0 {
		return fmt.Errorf("激活输入法失败")
	}

	return nil
}

// FormatInputMethodInfo 格式化输入法信息
func FormatInputMethodInfo(info *InputMethodInfo, verbose bool) string {
	// 确保 HKL 显示为 32 位值
	hkl32 := uint32(info.HKL)

	if verbose {
		return fmt.Sprintf(`━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
HKL:      0x%08X
语言ID:   0x%04X
布局ID:   0x%04X
名称:     %s
显示名:   %s
描述:     %s`,
			hkl32,
			info.LangID,
			info.LayoutID,
			info.Name,
			info.DisplayName,
			info.Description,
		)
	}

	return fmt.Sprintf("0x%08X", hkl32)
}

func getDisplayName(hkl uintptr, langID, layoutID uint16) string {
	langName := getLanguageName(langID)
	layoutDesc := getLayoutDescription(layoutID)

	if layoutDesc != "" {
		return fmt.Sprintf("0x%04X%04X (%s - %s)", layoutID, langID, langName, layoutDesc)
	}

	return fmt.Sprintf("0x%04X%04X (%s)", layoutID, langID, langName)
}

func getDescription(langID, layoutID uint16) string {
	langName := getLanguageName(langID)
	layoutDesc := getLayoutDescription(layoutID)

	if layoutDesc != "" {
		return fmt.Sprintf("%s - %s", langName, layoutDesc)
	}

	return langName
}

func getLanguageName(langID uint16) string {
	names := map[uint16]string{
		0x0409: "英语(美国)",
		0x0804: "中文(简体)",
		0x0404: "中文(繁体)",
		0x0411: "日语",
		0x0412: "韩语",
		0x040C: "法语",
		0x0407: "德语",
		0x0410: "意大利语",
		0x040A: "西班牙语",
		0x0419: "俄语",
	}

	if name, ok := names[langID]; ok {
		return name
	}

	return fmt.Sprintf("语言 0x%04X", langID)
}

func getLayoutDescription(layoutID uint16) string {
	descriptions := map[uint16]string{
		0x0000: "",
		0xE001: "微软拼音输入法",
		0xE002: "微软五笔输入法",
		0xE00E: "微软拼音新体验",
		0xE010: "微软五笔新体验",
	}

	if desc, ok := descriptions[layoutID]; ok {
		return desc
	}

	if layoutID != 0 {
		return fmt.Sprintf("布局 0x%04X", layoutID)
	}

	return ""
}

// 独立的 im-select 命令行工具
// 编译命令: go build -o im-select-enhanced.exe cmd/im-select/main.go

func main() {
	// 如果没有参数,显示当前输入法
	if len(os.Args) == 1 {
		showCurrent(false)
		return
	}

	// 解析命令行参数
	listFlag := flag.Bool("l", false, "列出所有可用的输入法")
	verboseFlag := flag.Bool("v", false, "显示详细信息")
	switchFlag := flag.String("s", "", "切换到指定的输入法")
	helpFlag := flag.Bool("h", false, "显示帮助信息")

	flag.Parse()

	// 处理帮助
	if *helpFlag {
		showHelp()
		return
	}

	// 列出所有输入法
	if *listFlag {
		listAll(*verboseFlag)
		return
	}

	// 切换输入法
	if *switchFlag != "" {
		switchTo(*switchFlag, *verboseFlag)
		return
	}

	// 检查是否有位置参数 (兼容原版 im-select 用法)
	args := flag.Args()
	if len(args) > 0 {
		// im-select <HKL> 形式切换
		switchTo(args[0], *verboseFlag)
		return
	}

	// 默认显示当前输入法
	showCurrent(*verboseFlag)
}

func showCurrent(verbose bool) {
	info, err := GetCurrentInputMethod()
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	if verbose {
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("当前输入法:")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println(FormatInputMethodInfo(info, true))
	} else {
		// 兼容原版输出格式
		fmt.Printf("0x%08X\n", uint32(info.HKL))
	}
}

func listAll(verbose bool) {
	methods, err := GetAllInputMethods()
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	if verbose {
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("共检测到 %d 个输入法\n", len(methods))
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println()

		for i, info := range methods {
			fmt.Printf("【%d】 %s\n", i+1, info.DisplayName)
			fmt.Println(FormatInputMethodInfo(info, true))
			if i < len(methods)-1 {
				fmt.Println()
			}
		}
	} else {
		fmt.Println("HKL        语言ID  布局ID  描述")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		for _, info := range methods {
			fmt.Printf("0x%08X  0x%04X  0x%04X  %s\n",
				uint32(info.HKL),
				info.LangID,
				info.LayoutID,
				info.Description,
			)
		}
	}
}

func switchTo(target string, verbose bool) {
	// 解析 HKL
	var hkl uintptr
	var err error

	// 支持多种格式
	if len(target) > 2 && (target[:2] == "0x" || target[:2] == "0X") {
		// 十六进制: 0x04090409
		var value uint64
		_, err = fmt.Sscanf(target, "0x%X", &value)
		if err == nil {
			hkl = uintptr(value)
		}
	} else if len(target) == 8 {
		// 十六进制字符串: 04090409
		var value uint64
		_, err = fmt.Sscanf(target, "%X", &value)
		if err == nil {
			hkl = uintptr(value)
		}
	} else {
		// 十进制: 1033
		var value uint64
		_, err = fmt.Sscanf(target, "%d", &value)
		if err == nil {
			// 转换为完整 HKL
			// 如果是 4 位数,可能是语言 ID,需要构造完整 HKL
			if value < 0x10000 {
				hkl = uintptr(value | (value << 16))
			} else {
				hkl = uintptr(value)
			}
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无效的输入法标识 '%s'\n", target)
		os.Exit(1)
	}

	// 切换
	err = SwitchInputMethod(hkl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	// 输出结果
	if verbose {
		fmt.Printf("✓ 已切换到: 0x%08X\n\n", uint32(hkl))

		// 显示切换后的当前输入法
		current, err := GetCurrentInputMethod()
		if err == nil {
			fmt.Println("当前输入法:")
			fmt.Println(FormatInputMethodInfo(current, true))
		}
	} else {
		// 兼容原版: 输出当前 HKL
		current, err := GetCurrentInputMethod()
		if err == nil {
			fmt.Printf("0x%08X\n", uint32(current.HKL))
		}
	}
}

func showHelp() {
	help := `
im-select-enhanced - Windows 输入法切换工具 (增强版)

用法:
    im-select-enhanced [选项] [HKL]

选项:
    无参数      显示当前输入法的 HKL
    -l          列出所有可用的输入法
    -v          显示详细信息
    -s <HKL>    切换到指定的输入法
    -h          显示此帮助信息
    <HKL>       直接切换到指定的输入法 (兼容原版用法)

HKL 格式:
    0x04090409  完整的 32 位十六进制 HKL
    04090409    8 位十六进制字符串
    0x0409      16 位十六进制语言 ID
    1033        十进制语言 ID

示例:
    # 显示当前输入法
    im-select-enhanced
    
    # 显示详细信息
    im-select-enhanced -v
    
    # 列出所有输入法
    im-select-enhanced -l
    
    # 列出所有输入法(详细)
    im-select-enhanced -l -v
    
    # 切换到英语 (美国)
    im-select-enhanced 0x04090409
    im-select-enhanced -s 0x04090409
    im-select-enhanced 1033
    
    # 切换到中文 (简体)
    im-select-enhanced 0x08040804
    im-select-enhanced 2052

说明:
    HKL (Handle to Keyboard Layout) 是 Windows 的键盘布局句柄
    - 低 16 位: 语言 ID (LANGID)
    - 高 16 位: 布局 ID (可区分不同的中文输入法)
    
    例如:
    - 0x04090409 = 英语 (美国) 标准键盘
    - 0x08040804 = 中文 (简体) 标准键盘
    - 0xE00E0804 = 微软拼音输入法
    - 0xE0100804 = 微软五笔输入法

与原版 im-select 的区别:
    ✓ 使用完整的 HKL,能区分不同的中文输入法
    ✓ 直接调用 Windows API,无需外部依赖
    ✓ 提供详细的输入法信息显示
    ✓ 完全兼容原版 im-select 的用法

项目地址:
    https://github.com/YourRepo/switch-input-method
`
	fmt.Println(help)
}
