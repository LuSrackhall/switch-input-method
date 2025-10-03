# Switch Input Method - Windows 版本

一个轻量级的 Windows 输入法快速切换工具，使用全局快捷键（Win+J/K）在中英文输入法之间快速切换。

## ✨ 核心特性

- 🎯 **全局快捷键**: Win+J 切换英文，Win+K 切换中文
- 🚫 **阻止系统快捷键冲突**: 完全拦截 Win+J/K，不触发系统原有功能
- 🔒 **底层键盘钩子**: 使用 Windows Low-Level Keyboard Hook API
- 💪 **稳定可靠**: 异步切换 + 互斥锁，防止并发问题

## 📦 依赖

- Windows 10/11
- [im-select.exe](https://github.com/daipeihust/im-select) - 输入法切换工具

## 🚀 快速开始

### 安装 im-select

```bash
# 下载 im-select.exe 并放到 PATH 路径或项目目录
# https://github.com/daipeihust/im-select
```

### 运行方式

**开发调试：**
```bash
go run .
```

**编译运行（推荐）：**
```bash
go build -o switch-input-method.exe
./switch-input-method.exe
```

**使用脚本：**
```bash
# Windows
run.bat

# Linux/Mac (WSL)
./run.sh
```

## 📖 使用说明

### 快捷键

- **Win + J**: 切换到英文输入法（LCID: 1033）
- **Win + K**: 切换到中文输入法（LCID: 2052）

### 查看当前输入法

```bash
im-select.exe
```

输出示例：
- `1033` - 英语(美国)
- `2052` - 中文(简体，中国)

## ⚙️ 自定义配置

### 修改 im-select.exe 路径

编辑 `main.go` 第 30 行：

```go
err := exec.Command("你的路径\\im-select.exe", imkey).Run()
```

### 修改快捷键

编辑 `keyboard_hook_windows.go`：

```go
const (
    VK_J = 0x4A  // 改为其他键，如 VK_H = 0x48
    VK_K = 0x4B  // 改为其他键，如 VK_L = 0x4C
)
```

### 修改输入法代码

编辑 `keyboard_hook_windows.go` 第 68 和 80 行：

```go
switchInputIfNeeded("1033")  // 改为你的输入法代码
switchInputIfNeeded("2052")  // 改为你的输入法代码
```

## 📂 项目结构

```
switch-input-method/
├── main.go                      # 主程序入口
├── keyboard_hook_windows.go     # Windows 键盘钩子实现
├── daemon_windows.go            # Windows 守护进程配置
├── daemon_unix.go               # Unix 守护进程配置（保留兼容）
├── run.bat                      # Windows 启动脚本
├── run.sh                       # Linux/Mac 启动脚本
├── 使用说明.md                   # 详细使用文档
└── GO_RUN说明.md                 # Go Run 命令说明
```

## 🔧 技术实现

### Low-Level Keyboard Hook

使用 Windows API `SetWindowsHookEx` 创建低级键盘钩子：

```go
// 安装钩子
hook := SetWindowsHookEx(WH_KEYBOARD_LL, callback, 0, 0)

// 回调函数中拦截事件
func keyboardHookProc(...) uintptr {
    if winKeyPressed && vkCode == VK_J {
        switchInputIfNeeded("1033")
        return 1  // 阻止事件传递
    }
    return CallNextHookEx(...)  // 其他事件正常传递
}
```

### 事件拦截流程

1. 用户按下 Win+J/K
2. 键盘钩子捕获事件
3. 检测到目标组合键
4. 异步执行输入法切换
5. 返回 1 阻止事件传递给系统
6. 系统原有的 Win+J/K 功能不会被触发

## ❓ 常见问题

### Q: `go run main.go` 报错 undefined

**A**: 请使用 `go run .` 运行整个包，或查看 [GO_RUN说明.md](GO_RUN说明.md)

### Q: 快捷键不生效

**A**: 
1. 以管理员权限运行
2. 检查终端是否显示 "键盘钩子安装成功"

### Q: 输入法切换失败

**A**:
1. 确认 `im-select.exe` 路径正确
2. 手动运行 `im-select.exe` 测试
3. 检查输入法代码是否匹配系统已安装的输入法

### Q: 与其他程序冲突

**A**: 
某些安全软件可能阻止键盘钩子，请添加程序到信任列表

## 📝 开发说明

### 编译

```bash
# Windows
go build -o switch-input-method.exe

# 指定平台交叉编译
GOOS=windows GOARCH=amd64 go build -o switch-input-method.exe
```

### 调试

在 `keyboard_hook_windows.go` 中添加调试输出：

```go
fmt.Printf("VK=%d, wParam=%d\n", vkCode, wParam)
```

### 依赖管理

```bash
go mod tidy
```

## 📄 许可证

MIT License

## 🔗 相关链接

- [im-select](https://github.com/daipeihust/im-select) - 输入法切换工具
- [Windows Hooks](https://learn.microsoft.com/en-us/windows/win32/winmsg/hooks) - 官方文档
- [Virtual-Key Codes](https://learn.microsoft.com/en-us/windows/win32/inputdev/virtual-key-codes) - 虚拟键码表

## 🙏 致谢

感谢所有为这个项目做出贡献的开发者！
