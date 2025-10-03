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
		VK_LWIN: "Left Win",
		VK_RWIN: "Right Win",
		VK_J:    "J",
		VK_K:    "K",
		0x41:    "A",
		0x42:    "B",
		0x43:    "C",
		0x44:    "D",
		0x45:    "E",
		0x46:    "F",
		0x47:    "G",
		0x48:    "H",
		0x49:    "I",
		0x4C:    "L",
		0x4D:    "M",
		0x4E:    "N",
		0x4F:    "O",
		0x50:    "P",
		0x51:    "Q",
		0x52:    "R",
		0x53:    "S",
		0x54:    "T",
		0x55:    "U",
		0x56:    "V",
		0x57:    "W",
		0x58:    "X",
		0x59:    "Y",
		0x5A:    "Z",
		0xA0:    "Left Shift",
		0xA1:    "Right Shift",
		0xA2:    "Left Ctrl",
		0xA3:    "Right Ctrl",
		0xA4:    "Left Alt",
		0xA5:    "Right Alt",
	}

	if name, ok := keyNames[vkCode]; ok {
		return name
	}
	return fmt.Sprintf("VK_%X", vkCode)
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
