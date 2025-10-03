//go:build windows
// +build windows

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// KeyBinding 按键绑定配置
type KeyBinding struct {
	ModifierKey uint32 `json:"modifier_key"` // 修饰键,如 VK_LWIN(0x5B)
	FunctionKey uint32 `json:"function_key"` // 功能键,如 VK_J(0x4A)
	IMKey       string `json:"im_key"`       // 输入法标识,如 "1033" 或 "2052"
	Description string `json:"description"`  // 描述,如 "切换到英文"
}

// Config 应用程序配置
type Config struct {
	KeyBindings []KeyBinding `json:"key_bindings"`
}

var (
	// 全局配置
	appConfig *Config
)

// GetDefaultConfig 获取默认配置
func GetDefaultConfig() *Config {
	return &Config{
		KeyBindings: []KeyBinding{
			{
				ModifierKey: VK_LWIN, // 左 Win 键
				FunctionKey: VK_J,    // J 键
				IMKey:       "1033",  // 英文
				Description: "切换到英文输入法",
			},
			{
				ModifierKey: VK_LWIN, // 左 Win 键
				FunctionKey: VK_K,    // K 键
				IMKey:       "2052",  // 中文
				Description: "切换到中文输入法",
			},
		},
	}
}

// GetConfigPath 获取配置文件路径
func GetConfigPath() string {
	// 获取可执行文件所在目录
	exePath, err := os.Executable()
	if err != nil {
		// 如果获取失败,使用临时目录
		return filepath.Join(os.TempDir(), "switch-input-method-config.json")
	}
	exeDir := filepath.Dir(exePath)
	return filepath.Join(exeDir, "config.json")
}

// LoadConfig 加载配置文件
func LoadConfig() (*Config, error) {
	configPath := GetConfigPath()

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 配置文件不存在,创建默认配置
		fmt.Println("配置文件不存在,创建默认配置...")
		defaultConfig := GetDefaultConfig()
		if err := SaveConfig(defaultConfig); err != nil {
			fmt.Printf("保存默认配置失败: %v\n", err)
			return defaultConfig, nil // 返回默认配置但不报错
		}
		return defaultConfig, nil
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析配置
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 验证配置
	if len(config.KeyBindings) == 0 {
		fmt.Println("配置文件中没有按键绑定,使用默认配置")
		return GetDefaultConfig(), nil
	}

	return &config, nil
}

// SaveConfig 保存配置文件
func SaveConfig(config *Config) error {
	configPath := GetConfigPath()

	// 序列化配置
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	fmt.Printf("配置已保存到: %s\n", configPath)
	return nil
}

// InitConfig 初始化配置
func InitConfig() error {
	config, err := LoadConfig()
	if err != nil {
		return err
	}
	appConfig = config

	// 打印当前配置
	fmt.Println("当前按键绑定配置:")
	for i, binding := range appConfig.KeyBindings {
		modifierName := GetKeyName(binding.ModifierKey)
		functionName := GetKeyName(binding.FunctionKey)
		fmt.Printf("  %d. %s+%s -> %s (IMKey: %s)\n",
			i+1, modifierName, functionName, binding.Description, binding.IMKey)
	}

	return nil
}

// GetKeyName 获取键码对应的名称
func GetKeyName(vkCode uint32) string {
	keyNames := map[uint32]string{
		// 修饰键
		91:  "Left Win",  // VK_LWIN
		92:  "Right Win", // VK_RWIN
		160: "Left Shift",
		161: "Right Shift",
		162: "Left Ctrl",
		163: "Right Ctrl",
		164: "Left Alt",
		165: "Right Alt",
		// 字母键 A-Z (65-90)
		65: "A",
		66: "B",
		67: "C",
		68: "D",
		69: "E",
		70: "F",
		71: "G",
		72: "H",
		73: "I",
		74: "J",
		75: "K",
		76: "L",
		77: "M",
		78: "N",
		79: "O",
		80: "P",
		81: "Q",
		82: "R",
		83: "S",
		84: "T",
		85: "U",
		86: "V",
		87: "W",
		88: "X",
		89: "Y",
		90: "Z",
		// 数字键 0-9 (48-57)
		48: "0",
		49: "1",
		50: "2",
		51: "3",
		52: "4",
		53: "5",
		54: "6",
		55: "7",
		56: "8",
		57: "9",
		// 功能键 F1-F12 (112-123)
		112: "F1",
		113: "F2",
		114: "F3",
		115: "F4",
		116: "F5",
		117: "F6",
		118: "F7",
		119: "F8",
		120: "F9",
		121: "F10",
		122: "F11",
		123: "F12",
	}

	if name, ok := keyNames[vkCode]; ok {
		return name
	}
	// 使用十进制显示未知键码
	return fmt.Sprintf("VK_%d", vkCode)
}

// AddKeyBinding 添加新的按键绑定
func AddKeyBinding(modifierKey, functionKey uint32, imKey, description string) error {
	if appConfig == nil {
		return fmt.Errorf("配置未初始化")
	}

	// 检查是否已存在相同的按键组合
	for _, binding := range appConfig.KeyBindings {
		if binding.ModifierKey == modifierKey && binding.FunctionKey == functionKey {
			return fmt.Errorf("按键组合 %s+%s 已存在",
				GetKeyName(modifierKey), GetKeyName(functionKey))
		}
	}

	// 添加新绑定
	newBinding := KeyBinding{
		ModifierKey: modifierKey,
		FunctionKey: functionKey,
		IMKey:       imKey,
		Description: description,
	}
	appConfig.KeyBindings = append(appConfig.KeyBindings, newBinding)

	// 保存配置
	return SaveConfig(appConfig)
}

// RemoveKeyBinding 删除按键绑定
func RemoveKeyBinding(index int) error {
	if appConfig == nil {
		return fmt.Errorf("配置未初始化")
	}

	if index < 0 || index >= len(appConfig.KeyBindings) {
		return fmt.Errorf("索引越界")
	}

	// 删除绑定
	appConfig.KeyBindings = append(
		appConfig.KeyBindings[:index],
		appConfig.KeyBindings[index+1:]...,
	)

	// 保存配置
	return SaveConfig(appConfig)
}

// GetCurrentConfig 获取当前配置
func GetCurrentConfig() *Config {
	return appConfig
}
