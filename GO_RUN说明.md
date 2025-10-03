# Go Run 命令说明

## 问题说明

Go 的 `go run` 命令在指定单个文件时，只会编译和运行该文件，不会包含同一包中的其他文件。

由于我们的项目分为多个文件：
- `main.go` - 主入口和输入法切换逻辑
- `keyboard_hook_windows.go` - Windows 键盘钩子实现
- `daemon_windows.go` - 守护进程配置
- `daemon_unix.go` - Unix 守护进程配置

## 正确的运行方式

### ✅ 方式 1：运行整个包（推荐）

```bash
go run .
```

这会编译当前目录下所有属于 `main` 包的 Go 文件。

### ✅ 方式 2：使用启动脚本

**Windows:**
```bash
run.bat
```

**Linux/Mac:**
```bash
./run.sh
```

### ✅ 方式 3：明确指定所有文件

```bash
go run main.go keyboard_hook_windows.go daemon_windows.go
```

### ✅ 方式 4：编译后运行（推荐用于生产环境）

```bash
# 编译
go build -o switch-input-method.exe

# 运行
./switch-input-method.exe
```

## ❌ 错误的运行方式

```bash
go run main.go  # ❌ 只运行 main.go，缺少其他文件的函数定义
```

这会导致错误：`undefined: StartKeyboardHook`

## 为什么这样设计？

将代码分成多个文件有以下好处：

1. **代码组织更清晰**：每个文件负责特定功能
2. **平台特定代码分离**：使用构建标签区分 Windows/Unix 实现
3. **易于维护**：修改某个功能时只需编辑对应文件
4. **支持条件编译**：Go 会根据平台自动选择正确的文件

## 构建标签说明

文件顶部的构建标签控制文件何时被编译：

```go
//go:build windows
// +build windows
```

- `keyboard_hook_windows.go` 和 `daemon_windows.go` 只在 Windows 平台编译
- `daemon_unix.go` 只在非 Windows 平台编译

使用 `go run .` 可以让 Go 自动处理这些构建标签。

## 总结

**开发阶段**：使用 `go run .` 或 `run.bat`
**生产部署**：使用 `go build` 编译后运行可执行文件
