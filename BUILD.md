# 源码构建文档

本文档详细说明如何从源代码构建 **兴宜街道红旗路输入法切换工具**。

---

## 目录

- [环境要求](#环境要求)
- [依赖安装](#依赖安装)
- [构建步骤](#构建步骤)
- [构建选项](#构建选项)
- [常见问题](#常见问题)
- [开发调试](#开发调试)

---

## 环境要求

### 操作系统
- **Windows 10** 或更高版本
- **Windows 11** (推荐)

### 编译工具
- **Go 语言** 版本 1.16 或更高
  - 推荐使用 Go 1.18+ 以获得更好的性能和泛型支持
  - 下载地址: https://golang.org/dl/

### 必需工具
- **Git** (可选,用于克隆仓库)
  - 下载地址: https://git-scm.com/download/win
- **im-select.exe** (已包含在项目中)
  - 输入法检测和切换工具

---

## 依赖安装

### 1. 安装 Go 语言环境

**方法一: 使用安装包 (推荐)**

1. 访问 [Go 官方下载页](https://golang.org/dl/)
2. 下载 Windows 版本的 MSI 安装包
3. 运行安装程序,使用默认设置
4. 验证安装:
   ```bash
   go version
   ```
   应输出类似: `go version go1.21.0 windows/amd64`

**方法二: 使用包管理器**

如果你安装了 Chocolatey:
```bash
choco install golang
```

### 2. 配置 Go 环境变量

通常安装程序会自动配置,手动检查:

```bash
# 查看 GOPATH
go env GOPATH

# 查看 GOROOT
go env GOROOT
```

如需手动设置(以管理员身份运行):
```bash
setx GOPATH "%USERPROFILE%\go"
setx PATH "%PATH%;%GOPATH%\bin"
```

### 3. 配置 Go 模块代理 (中国大陆用户推荐)

加速依赖下载:
```bash
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
```

### 4. 下载项目依赖

进入项目目录后执行:
```bash
go mod download
```

这将下载以下依赖:
- `github.com/getlantern/systray` - 系统托盘支持
- `golang.org/x/sys/windows` - Windows API 访问

---

## 构建步骤

### 基础构建

```bash
# 1. 进入项目目录
cd d:\safe\switch-input-method

# 2. 构建可执行文件
go build -o switch-input-method.exe

# 3. 运行程序
.\switch-input-method.exe
```

### 托盘版构建 (无控制台窗口)

```bash
go build -ldflags="-H windowsgui" -o switch-input-method.exe
```

参数说明:
- `-ldflags="-H windowsgui"` - 隐藏控制台窗口,作为后台程序运行
- `-o` - 指定输出文件名

### 优化构建 (减小文件体积)

```bash
go build -ldflags="-s -w -H windowsgui" -o switch-input-method.exe
```

参数说明:
- `-s` - 去除符号表 (减小约 30% 体积)
- `-w` - 去除 DWARF 调试信息 (进一步减小体积)
- `-H windowsgui` - 隐藏控制台窗口

**文件大小对比:**
- 普通构建: ~8-10 MB
- 优化构建: ~5-6 MB
- UPX 压缩后: ~2-3 MB (可选)

### 跨版本编译

指定 Go 版本特性:
```bash
go build -gcflags="-N -l" -o switch-input-method.exe
```

---

## 构建选项

### 完整构建命令参数

```bash
go build \
  -ldflags="-s -w -H windowsgui -X main.version=2.4.0" \
  -trimpath \
  -o switch-input-method.exe
```

**参数详解:**

| 参数                               | 说明         | 效果             |
| ---------------------------------- | ------------ | ---------------- |
| `-ldflags="-s -w"`                 | 去除调试信息 | 减小 30-40% 体积 |
| `-ldflags="-H windowsgui"`         | 隐藏控制台   | 作为后台程序运行 |
| `-ldflags="-X main.version=2.4.0"` | 设置版本变量 | 可在代码中使用   |
| `-trimpath`                        | 移除文件路径 | 提升安全性       |
| `-race`                            | 启用竞态检测 | 调试并发问题     |
| `-v`                               | 显示详细输出 | 查看编译过程     |

### 条件编译

项目使用 Go 的构建标签:

```go
//go:build windows
// +build windows
```

确保代码只在 Windows 平台编译。

---

## 常见问题

### Q1: 提示 "go: command not found"

**原因:** Go 未正确安装或环境变量未配置

**解决方案:**
1. 重新安装 Go
2. 检查环境变量:
   ```bash
   echo %PATH% | findstr go
   ```
3. 重启终端或重新登录

### Q2: 依赖下载失败

**错误信息:**
```
go: downloading github.com/getlantern/systray failed
```

**解决方案:**
```bash
# 设置国内镜像
go env -w GOPROXY=https://goproxy.cn,direct

# 清理缓存后重试
go clean -modcache
go mod download
```

### Q3: 编译错误 "undefined: syscall"

**原因:** 平台不匹配或构建标签问题

**解决方案:**
确保在 Windows 系统上编译,检查文件头部:
```go
//go:build windows
// +build windows
```

### Q4: 运行时窗口闪现

**原因:** 未使用 `-H windowsgui` 参数

**解决方案:**
```bash
go build -ldflags="-H windowsgui" -o switch-input-method.exe
```

### Q5: im-select.exe 找不到

**原因:** im-select.exe 不在可执行文件同目录

**解决方案:**
1. 确保 `im-select.exe` 与 `switch-input-method.exe` 在同一目录
2. 或修改代码中的路径为绝对路径

### Q6: 钩子注册失败 (错误代码 1429)

**原因:** DLL 注入限制或权限不足

**解决方案:**
1. 以管理员身份运行
2. 检查杀毒软件是否拦截
3. 在 Windows 安全中心添加排除项

### Q7: 编译后体积过大

**解决方案:**
```bash
# 方法1: 使用优化参数
go build -ldflags="-s -w -H windowsgui"

# 方法2: 使用 UPX 压缩 (需安装 UPX)
upx --best switch-input-method.exe
```

---

## 开发调试

### 启用调试日志

修改代码,添加日志输出:

```go
import "log"

log.SetFlags(log.LstdFlags | log.Lshortfile)
log.Printf("调试信息: %v", someValue)
```

### 调试构建 (保留符号表)

```bash
go build -gcflags="all=-N -l" -o switch-input-method-debug.exe
```

参数说明:
- `-N` - 禁用优化
- `-l` - 禁用内联
- 便于使用调试器 (如 Delve)

### 使用 Delve 调试器

```bash
# 安装 Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# 启动调试
dlv debug
```

### 检查依赖版本

```bash
# 查看依赖树
go mod graph

# 查看可更新的依赖
go list -u -m all
```

### 更新依赖

```bash
# 更新所有依赖到最新版本
go get -u ./...

# 整理依赖
go mod tidy
```

### 单元测试

```bash
# 运行所有测试
go test ./...

# 运行指定测试
go test -v -run TestFunctionName

# 测试覆盖率
go test -cover ./...
```

### 性能分析

```bash
# CPU 性能分析
go build -o switch-input-method.exe
go tool pprof cpu.prof

# 内存分析
go build -o switch-input-method.exe
go tool pprof mem.prof
```

---

## 快速参考

### 快速构建命令

| 用途         | 命令                                      |
| ------------ | ----------------------------------------- |
| **开发测试** | `go build`                                |
| **发布版本** | `go build -ldflags="-s -w -H windowsgui"` |
| **调试版本** | `go build -gcflags="all=-N -l"`           |
| **最小体积** | `go build -ldflags="-s -w" && upx --best` |

### 环境检查清单

- [ ] Go 版本 >= 1.16
- [ ] `go env` 输出正常
- [ ] 能访问 GOPROXY
- [ ] `go mod download` 成功
- [ ] im-select.exe 存在
- [ ] 编译无错误
- [ ] 运行无崩溃

### 构建脚本示例

创建 `build.bat`:
```batch
@echo off
echo 开始构建 switch-input-method...
go build -ldflags="-s -w -H windowsgui -X main.version=2.4.0" -trimpath -o switch-input-method.exe
if %errorlevel% == 0 (
    echo 构建成功!
    echo 文件: switch-input-method.exe
    dir switch-input-method.exe
) else (
    echo 构建失败!
)
pause
```

---

## 相关资源

- **Go 官方文档:** https://golang.org/doc/
- **systray 库:** https://github.com/getlantern/systray
- **Windows API:** https://learn.microsoft.com/windows/win32/api/
- **虚拟键码参考:** https://learn.microsoft.com/windows/win32/inputdev/virtual-key-codes
- **im-select 项目:** https://github.com/daipeihust/im-select

---

## 贡献指南

欢迎提交问题和改进建议!

构建问题请提供:
1. Go 版本 (`go version`)
2. 操作系统版本
3. 完整错误信息
4. 使用的构建命令

---

**最后更新:** 2024年 (v2.4.0)
**维护者:** switch-input-method 开发团队
