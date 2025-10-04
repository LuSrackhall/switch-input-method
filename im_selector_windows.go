//go:build windows
// +build windows

package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	user32DLL                  = syscall.NewLazyDLL("user32.dll")
	procGetKeyboardLayoutList  = user32DLL.NewProc("GetKeyboardLayoutList")
	procGetKeyboardLayout      = user32DLL.NewProc("GetKeyboardLayout")
	procActivateKeyboardLayout = user32DLL.NewProc("ActivateKeyboardLayout")
	procLoadKeyboardLayoutW    = user32DLL.NewProc("LoadKeyboardLayoutW")
	procGetKeyboardLayoutNameW = user32DLL.NewProc("GetKeyboardLayoutNameW")
)

const (
	KLF_ACTIVATE      = 0x00000001
	KLF_SETFORPROCESS = 0x00000100
	KLF_REORDER       = 0x00000008
)

// InputMethodInfo 输入法信息
type InputMethodInfo struct {
	HKL         uintptr // 完整的键盘布局句柄
	LangID      uint16  // 语言ID (低16位)
	LayoutID    uint16  // 布局ID (高16位)
	Name        string  // 布局名称 (如 "00000804")
	DisplayName string  // 友好显示名称
	Description string  // 描述
}

// GetCurrentInputMethodEnhanced 获取当前输入法 (增强版)
func GetCurrentInputMethodEnhanced() (*InputMethodInfo, error) {
	// 获取当前线程ID (使用0表示当前线程)
	threadID := uintptr(0)

	// 获取当前键盘布局
	ret, _, _ := procGetKeyboardLayout.Call(threadID)
	if ret == 0 {
		return nil, fmt.Errorf("获取当前键盘布局失败")
	}

	hkl := ret

	// 获取布局名称
	var layoutName [260]uint16
	r1, _, _ := procGetKeyboardLayoutNameW.Call(uintptr(unsafe.Pointer(&layoutName[0])))
	if r1 == 0 {
		return nil, fmt.Errorf("获取键盘布局名称失败")
	}

	name := syscall.UTF16ToString(layoutName[:])

	// 解析 HKL
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

// GetAllInputMethods 获取所有可用的输入法
func GetAllInputMethods() ([]*InputMethodInfo, error) {
	// 第一次调用获取数量
	count, _, _ := procGetKeyboardLayoutList.Call(0, 0)
	if count == 0 {
		return nil, fmt.Errorf("获取键盘布局列表失败")
	}

	// 分配缓冲区
	hkls := make([]uintptr, count)

	// 第二次调用获取实际数据
	ret, _, _ := procGetKeyboardLayoutList.Call(
		count,
		uintptr(unsafe.Pointer(&hkls[0])),
	)

	if ret == 0 {
		return nil, fmt.Errorf("获取键盘布局列表失败")
	}

	// 构建输入法信息列表
	var methods []*InputMethodInfo
	for _, hkl := range hkls[:ret] {
		langID := uint16(hkl & 0xFFFF)
		layoutID := uint16((hkl >> 16) & 0xFFFF)

		// 加载布局以获取名称
		layoutName := getLayoutName(hkl)

		info := &InputMethodInfo{
			HKL:         hkl,
			LangID:      langID,
			LayoutID:    layoutID,
			Name:        layoutName,
			DisplayName: getDisplayName(hkl, langID, layoutID),
			Description: getDescription(langID, layoutID),
		}

		methods = append(methods, info)
	}

	return methods, nil
}

// SwitchInputMethod 切换到指定的输入法
func SwitchInputMethod(hkl uintptr) error {
	// 激活键盘布局
	ret, _, _ := procActivateKeyboardLayout.Call(
		hkl,
		KLF_SETFORPROCESS,
	)

	if ret == 0 {
		return fmt.Errorf("激活键盘布局失败")
	}

	return nil
}

// SwitchInputMethodByName 通过布局名称切换输入法
func SwitchInputMethodByName(name string) error {
	// 将名称转换为 UTF16
	namePtr, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return err
	}

	// 加载键盘布局
	ret, _, _ := procLoadKeyboardLayoutW.Call(
		uintptr(unsafe.Pointer(namePtr)),
		KLF_ACTIVATE|KLF_SETFORPROCESS,
	)

	if ret == 0 {
		return fmt.Errorf("加载键盘布局失败: %s", name)
	}

	return nil
}

// FindInputMethodByHKL 通过HKL查找输入法
func FindInputMethodByHKL(targetHKL uintptr) (*InputMethodInfo, error) {
	methods, err := GetAllInputMethods()
	if err != nil {
		return nil, err
	}

	for _, method := range methods {
		if method.HKL == targetHKL {
			return method, nil
		}
	}

	return nil, fmt.Errorf("未找到 HKL: 0x%X", targetHKL)
}

// FindInputMethodByLangID 通过语言ID查找输入法(可能有多个)
func FindInputMethodByLangID(langID uint16) ([]*InputMethodInfo, error) {
	methods, err := GetAllInputMethods()
	if err != nil {
		return nil, err
	}

	var matches []*InputMethodInfo
	for _, method := range methods {
		if method.LangID == langID {
			matches = append(matches, method)
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("未找到语言ID: 0x%04X", langID)
	}

	return matches, nil
}

// getLayoutName 获取布局名称
func getLayoutName(hkl uintptr) string {
	// HKL 高16位为0时,布局名称就是语言ID的16进制
	// 否则是完整的8位16进制
	langID := uint16(hkl & 0xFFFF)
	layoutID := uint16((hkl >> 16) & 0xFFFF)

	if layoutID == 0 {
		return fmt.Sprintf("%08X", langID)
	}

	return fmt.Sprintf("%08X", hkl)
}

// getDisplayName 获取友好显示名称
func getDisplayName(hkl uintptr, langID, layoutID uint16) string {
	// 完整的 HKL 显示
	if layoutID == 0 {
		return fmt.Sprintf("0x%04X (%s)", langID, getLanguageName(langID))
	}

	return fmt.Sprintf("0x%08X (%s - %s)",
		hkl,
		getLanguageName(langID),
		getLayoutDescription(layoutID),
	)
}

// getDescription 获取描述信息
func getDescription(langID, layoutID uint16) string {
	langName := getLanguageName(langID)

	if layoutID == 0 {
		return langName
	}

	layoutDesc := getLayoutDescription(layoutID)
	return fmt.Sprintf("%s - %s", langName, layoutDesc)
}

// getLanguageName 根据语言ID获取语言名称
func getLanguageName(langID uint16) string {
	// 常见语言映射
	languages := map[uint16]string{
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

	if name, ok := languages[langID]; ok {
		return name
	}

	return fmt.Sprintf("语言0x%04X", langID)
}

// getLayoutDescription 根据布局ID获取布局描述
func getLayoutDescription(layoutID uint16) string {
	// 常见输入法布局映射
	// 高位标识不同的输入法实现
	layouts := map[uint16]string{
		0x0000: "默认键盘",
		0xE001: "微软拼音输入法",
		0xE002: "微软五笔输入法",
		0xE00E: "微软拼音新体验",
		0xE010: "微软五笔新体验",
		// 可以继续添加其他输入法的标识
	}

	if desc, ok := layouts[layoutID]; ok {
		return desc
	}

	// 如果是用户自定义或第三方输入法
	if layoutID >= 0xE000 {
		return fmt.Sprintf("输入法0x%04X", layoutID)
	}

	return fmt.Sprintf("布局0x%04X", layoutID)
}

// FormatInputMethodInfo 格式化输入法信息为字符串
func FormatInputMethodInfo(info *InputMethodInfo, verbose bool) string {
	if !verbose {
		// 简洁模式: 只显示完整的HKL
		return fmt.Sprintf("0x%08X", info.HKL)
	}

	// 详细模式
	return fmt.Sprintf(
		"HKL:      0x%08X\n"+
			"语言ID:   0x%04X\n"+
			"布局ID:   0x%04X\n"+
			"名称:     %s\n"+
			"显示名:   %s\n"+
			"描述:     %s",
		info.HKL,
		info.LangID,
		info.LayoutID,
		info.Name,
		info.DisplayName,
		info.Description,
	)
}
