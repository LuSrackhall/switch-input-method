# im-select-enhanced - 增强版输入法切换工具

## 概述

这是一个 **增强版的 im-select 工具**,专为 Windows 平台设计,能够:

✅ **区分不同的中文输入法** (微软拼音、搜狗拼音、RIME 等)  
✅ **使用完整的 HKL** (Handle to Keyboard Layout)  
✅ **直接调用 Windows API**,无需外部依赖  
✅ **完全兼容原版 im-select** 的使用方式  

---

## 背景

### 原版 im-select 的局限性

原版 `im-select` 在 Windows 下只返回 **语言 ID (LANGID)** 的低 16 位,例如:

```
1033  # 英语
2052  # 中文(简体)
```

**问题:** 无法区分不同的中文输入法!

如果你的系统安装了:
- 微软拼音输入法
- 搜狗拼音输入法
- RIME 输入法

它们都会显示为 `2052`,无法区分具体是哪一个。

### 增强版的改进

使用完整的 **HKL (32/64 位值)**:

```
0x04090409  # 英语(美国) - 标准键盘
0x08040804  # 中文(简体) - 标准键盘
0xE00E0804  # 中文(简体) - 微软拼音输入法
0xE0100804  # 中文(简体) - 微软五笔输入法
```

**高 16 位** 标识了具体的输入法实现,因此能够精确区分!

---

## 技术原理

### HKL 结构

```
HKL (32/64 位)
┌────────────────┬────────────────┐
│   高 16 位     │   低 16 位     │
│  布局 ID       │  语言 ID       │
│ (Layout ID)    │  (LANGID)      │
└────────────────┴────────────────┘

示例: 0xE00E0804
      │    │
      │    └─ 0x0804 = 2052 (中文简体)
      └────── 0xE00E = 微软拼音输入法
```

### 语言 ID 映射

| 十进制 | 十六进制 | 语言       |
| ------ | -------- | ---------- |
| 1033   | 0x0409   | 英语(美国) |
| 2052   | 0x0804   | 中文(简体) |
| 1028   | 0x0404   | 中文(繁体) |
| 1041   | 0x0411   | 日语       |
| 1042   | 0x0412   | 韩语       |

### 布局 ID 示例

| 布局 ID | 描述                 |
| ------- | -------------------- |
| 0x0000  | 标准键盘             |
| 0xE00E  | 微软拼音输入法       |
| 0xE010  | 微软五笔输入法       |
| 0xE001  | 微软拼音输入法(旧版) |

---

## 安装和编译

### 方式 1: 独立命令行工具

编译独立的 `im-select-enhanced.exe`:

```bash
cd d:\safe\switch-input-method
go build -ldflags="-s -w" -o im-select-enhanced.exe cmd/im-select/main.go im_selector_windows.go
```

### 方式 2: 集成到主程序

增强版功能已集成到主程序 `switch-input-method.exe` 中:

```bash
go build -ldflags="-s -w -H windowsgui" -o switch-input-method.exe
```

---

## 使用方法

### 基础用法 (兼容原版)

```bash
# 显示当前输入法
im-select-enhanced
# 输出: 0x04090409

# 切换到英语
im-select-enhanced 0x04090409

# 切换到中文
im-select-enhanced 0x08040804
```

### 增强功能

#### 1. 列出所有输入法

```bash
im-select-enhanced -l
```

输出:
```
HKL        语言ID  布局ID  描述
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
0x04090409  0x0409  0x0000  英语(美国)
0x08040804  0x0804  0x0000  中文(简体)
0xE00E0804  0x0804  0xE00E  中文(简体) - 微软拼音新体验
```

#### 2. 显示详细信息

```bash
im-select-enhanced -v
```

输出:
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
当前输入法:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
HKL:      0xE00E0804
语言ID:   0x0804
布局ID:   0xE00E
名称:     E00E0804
显示名:   0xE00E0804 (中文(简体) - 微软拼音新体验)
描述:     中文(简体) - 微软拼音新体验
```

#### 3. 列出所有输入法(详细)

```bash
im-select-enhanced -l -v
```

#### 4. 切换输入法

支持多种格式:

```bash
# 完整 HKL (十六进制)
im-select-enhanced 0xE00E0804

# 简短格式
im-select-enhanced -s 0x0804

# 十进制语言 ID
im-select-enhanced 2052

# 使用 -s 参数
im-select-enhanced -s 0x04090409
```

---

## 在主程序中使用

### 配置文件示例

```json
{
  "key_bindings": [
    {
      "modifier_key": 91,
      "function_key": 74,
      "im_key": "0x04090409",
      "description": "切换到英语"
    },
    {
      "modifier_key": 91,
      "function_key": 75,
      "im_key": "0xE00E0804",
      "description": "切换到微软拼音"
    },
    {
      "modifier_key": 91,
      "function_key": 76,
      "im_key": "0x08040804",
      "description": "切换到中文标准键盘"
    }
  ]
}
```

### 获取当前输入法的 HKL

1. **使用托盘菜单:**
   - 右键托盘图标
   - 点击"配置管理" → "快速绑定当前输入法"
   - 会显示当前输入法的完整 HKL

2. **使用命令行:**
   ```bash
   im-select-enhanced -v
   ```

3. **使用程序代码:**
   ```go
   info, err := GetCurrentInputMethod()
   if err == nil {
       fmt.Printf("HKL: 0x%08X\n", info.HKL)
   }
   ```

---

## API 接口

### Go 语言接口

```go
// 获取当前输入法
info, err := GetCurrentInputMethod()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("HKL: 0x%08X\n", info.HKL)

// 获取所有输入法
methods, err := GetAllInputMethods()
for _, method := range methods {
    fmt.Println(method.DisplayName)
}

// 切换输入法
err = SwitchInputMethod(0x04090409)
if err != nil {
    log.Fatal(err)
}

// 通过名称切换
err = SwitchInputMethodByName("04090409")

// 查找输入法
method, err := FindInputMethodByHKL(0xE00E0804)
if err == nil {
    fmt.Println(method.Description)
}
```

---

## 常见问题

### Q1: 如何获取我系统中所有输入法的 HKL?

```bash
im-select-enhanced -l -v
```

### Q2: 我的搜狗拼音的 HKL 是多少?

每个系统可能不同,使用以下方法获取:

1. 切换到搜狗拼音
2. 运行 `im-select-enhanced -v`
3. 查看输出的 HKL 值

### Q3: 为什么我的配置文件还是用 `2052`?

旧版配置仍然兼容,但建议升级为完整 HKL:

```json
// 旧版
"im_key": "2052"

// 新版 (推荐)
"im_key": "0xE00E0804"
```

### Q4: 如何区分微软拼音和搜狗拼音?

使用完整 HKL:

```
0xE00E0804  → 微软拼音
0x????????  → 搜狗拼音 (具体值需在你的系统上查询)
```

### Q5: 兼容原版 im-select 吗?

完全兼容! 所有原版用法都支持:

```bash
# 原版用法
im-select           # 显示当前
im-select 1033      # 切换到英语

# 增强用法
im-select-enhanced -v           # 显示详细
im-select-enhanced 0x04090409   # 精确切换
```

---

## 对比原版 im-select

| 特性           | 原版 im-select   | 增强版 im-select-enhanced |
| -------------- | ---------------- | ------------------------- |
| 返回值         | 语言 ID (低16位) | 完整 HKL (32位)           |
| 区分中文输入法 | ❌ 不支持         | ✅ 支持                    |
| 依赖           | im-select.exe    | 无需外部依赖              |
| API 调用       | 外部进程         | 直接 Windows API          |
| 详细信息       | ❌ 无             | ✅ 有 (-v 参数)            |
| 列出所有输入法 | ❌ 不支持         | ✅ 支持 (-l 参数)          |
| 性能           | 进程启动开销     | 直接 API,更快             |
| 兼容性         | ✅                | ✅ 完全兼容                |

---

## 技术细节

### Windows API 调用

```go
// 获取当前布局
GetKeyboardLayout(0)

// 获取所有布局
GetKeyboardLayoutList()

// 激活布局
ActivateKeyboardLayout(hkl, KLF_SETFORPROCESS)

// 加载布局
LoadKeyboardLayoutW(layoutName, KLF_ACTIVATE)
```

### HKL 解析

```go
hkl := uintptr(0xE00E0804)

// 提取语言 ID (低 16 位)
langID := uint16(hkl & 0xFFFF)  // 0x0804

// 提取布局 ID (高 16 位)
layoutID := uint16((hkl >> 16) & 0xFFFF)  // 0xE00E
```

---

## 开发计划

- [x] 基础 HKL 支持
- [x] 完整 Windows API 集成
- [x] 命令行工具
- [x] 主程序集成
- [ ] 输入法名称自动识别(通过注册表)
- [ ] GUI 配置界面中的输入法选择器
- [ ] 支持 TSF (Text Services Framework) API
- [ ] 跨平台支持(macOS, Linux)

---

## 许可证

MIT License

---

## 贡献

欢迎提交 Issue 和 Pull Request!

特别欢迎:
- 补充更多输入法的布局 ID 映射
- 优化输入法名称识别
- 增加更多语言支持

---

## 致谢

- 原版 [im-select](https://github.com/daipeihust/im-select)
- Windows Input Method Framework
- Go Windows API 封装

---

**项目地址:** https://github.com/LuSrackhall/switch-input-method  
**版本:** v2.5.0  
**更新日期:** 2025年10月4日
