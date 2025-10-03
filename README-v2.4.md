# Switch Input Method - Windows 版本

一个轻量级的 Windows 输入法快速切换工具,支持自定义按键绑定,可在任意输入法之间快速切换。

## ✨ 核心特性

- 🎯 **可配置按键绑定**: 自定义任意快捷键组合(Win/Ctrl/Alt/Shift + 任意键)
- 🎨 **多输入法支持**: 可绑定任意数量的输入法
- 🖥️ **系统托盘运行**: 支持后台运行,无需保持终端窗口
- 🚫 **阻止系统快捷键冲突**: 完全拦截自定义快捷键,不触发系统原有功能
- 🔒 **底层键盘钩子**: 使用 Windows Low-Level Keyboard Hook API
- 💪 **稳定可靠**: 异步切换 + 互斥锁,防止并发问题
- ⚙️ **配置管理**: 托盘菜单支持查看、编辑、热重载配置
- 📝 **配置文件**: JSON 格式配置,简单易懂

## 🆕 v2.4 更新 - 可配置按键绑定

现在支持完全自定义的按键绑定系统:
- ✅ 自定义修饰键(Win/Ctrl/Alt/Shift等)
- ✅ 自定义功能键(字母/数字/功能键等)
- ✅ 绑定任意输入法
- ✅ 支持多个按键组合
- ✅ 托盘菜单配置管理
- ✅ 热重载配置,无需重启

详见 [v2.4更新日志-可配置按键.md](./v2.4更新日志-可配置按键.md) 和 [配置指南.md](./配置指南.md)

## 📦 依赖

- Windows 10/11
- [im-select.exe](https://github.com/daipeihust/im-select) - 输入法切换工具(已内嵌)

## 🚀 快速开始

### 直接运行

1. 下载或编译 `switch-input-method.exe`
2. 双击运行(会自动隐藏到系统托盘)
3. 首次运行会自动创建默认配置文件 `config.json`
4. 默认快捷键:
   - **Win+J**: 切换到英文输入法
   - **Win+K**: 切换到中文输入法

### 自定义配置

1. 右键系统托盘图标
2. 选择 "配置管理" -> "打开配置文件"
3. 编辑 `config.json` 文件
4. 保存后选择 "重新加载配置" 或重启程序

详细配置说明请参考 [配置指南.md](./配置指南.md)

## ⚙️ 配置示例

### 默认配置 (Win+J/K)
```json
{
  "key_bindings": [
    {
      "modifier_key": 91,
      "function_key": 74,
      "im_key": "1033",
      "description": "切换到英文输入法"
    },
    {
      "modifier_key": 91,
      "function_key": 75,
      "im_key": "2052",
      "description": "切换到中文输入法"
    }
  ]
}
```

### 使用 Ctrl+1/2 切换
```json
{
  "key_bindings": [
    {
      "modifier_key": 162,
      "function_key": 49,
      "im_key": "1033",
      "description": "切换到英文"
    },
    {
      "modifier_key": 162,
      "function_key": 50,
      "im_key": "2052",
      "description": "切换到中文"
    }
  ]
}
```

## 📖 托盘菜单功能

- **查看当前配置**: 显示当前所有按键绑定
- **快速绑定当前输入法**: 获取当前输入法的配置示例
- **虚拟键码参考**: 查看常用按键的虚拟键码
- **打开配置文件**: 在记事本中直接编辑配置
- **重新加载配置**: 应用新配置,无需重启程序
- **退出**: 退出程序

## 🔧 开发和编译

### 环境要求
- Go 1.16+
- Windows 10/11

### 编译
```bash
go build -o switch-input-method.exe
```

### 开发调试
```bash
go run .
```

## 📚 相关文档

- [配置指南.md](./配置指南.md) - 详细的配置说明
- [v2.4更新日志-可配置按键.md](./v2.4更新日志-可配置按键.md) - 最新更新说明
- [托盘版使用说明.md](./托盘版使用说明.md) - 托盘功能说明
- [config.json.example](./config.json.example) - 配置文件示例

## 🎯 常用虚拟键码

### 修饰键
- 91 - 左 Win 键
- 92 - 右 Win 键
- 162/163 - 左/右 Ctrl
- 164/165 - 左/右 Alt
- 160/161 - 左/右 Shift

### 字母键
- 65-90 - A到Z (例如: J=74, K=75)

### 数字键
- 48-57 - 0到9

## 💡 获取输入法 Key

在命令行运行:
```bash
im-select.exe
```
输出的就是当前输入法的标识符,例如:
- `1033` - 英语(美国)
- `2052` - 中文(中国)

## 🐛 故障排除

### 配置文件损坏
删除 `config.json` 文件,程序会自动创建默认配置。

### 按键不生效
1. 检查配置文件 JSON 格式是否正确
2. 使用托盘菜单的"重新加载配置"功能
3. 确认虚拟键码是否正确
4. 验证输入法 key 是否正确(运行 `im-select.exe`)

## 📜 License

MIT License

## 🙏 致谢

- [im-select](https://github.com/daipeihust/im-select) - 输入法切换工具
- [systray](https://github.com/getlantern/systray) - 系统托盘库
