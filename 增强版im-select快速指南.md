# 增强版 im-select 快速指南

## ✅ 测试成功!

已成功实现增强版 `im-select`,能够区分不同的中文输入法!

## 核心改进

### 原版 im-select 的问题
```bash
# 原版 im-select 只返回语言 ID
im-select.exe
# 输出: 2052  (无法区分具体哪个中文输入法)
```

### 增强版的解决方案
```bash
# 增强版返回完整 HKL
im-select-enhanced.exe
# 输出: 0x08040804  或  0xE0200804  (可以区分不同输入法!)
```

## 测试结果

在你的系统上检测到 **3 个输入法**:

1. **0x08040804** - 中文(简体) - 标准键盘
2. **0x04090409** - 英语(美国) - 标准键盘
3. **0xE0200804** - 中文(简体) - 布局 0xE020 (可能是某个中文输入法)

**关键发现:**
- `0x0804` 和 `0xE020` 都是中文简体,但是**布局 ID 不同**!
- 原版 im-select 会将它们都识别为 `2052`,无法区分
- 增强版使用完整 HKL,可以精确区分!

## 基本使用

### 1. 查看当前输入法
```bash
# 简洁模式 (兼容原版)
im-select-enhanced.exe
# 输出: 0xE0200804

# 详细模式
im-select-enhanced.exe -v
# 输出:
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 当前输入法:
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# HKL:      0xE0200804
# 语言ID:   0x0804
# 布局ID:   0xE020
# ...
```

### 2. 列出所有输入法
```bash
# 简洁列表
im-select-enhanced.exe -l

# 详细列表
im-select-enhanced.exe -l -v
```

### 3. 切换输入法
```bash
# 切换到英语
im-select-enhanced.exe -s 0x04090409

# 切换到中文标准键盘
im-select-enhanced.exe -s 0x08040804

# 切换到布局 0xE020 的中文输入法
im-select-enhanced.exe -s 0xE0200804

# 使用详细模式查看切换结果
im-select-enhanced.exe -s 0x04090409 -v
```

## 在主程序中使用

### 方法 1: 使用托盘菜单获取 HKL

1. 切换到你想要的输入法(如微软拼音)
2. 右键托盘图标 → "配置管理" → "快速绑定当前输入法"
3. 记下显示的 HKL (如 `0xE00E0804`)

### 方法 2: 使用命令行获取 HKL

```bash
# 步骤 1: 列出所有输入法
im-select-enhanced.exe -l -v

# 步骤 2: 找到你想要的输入法的 HKL
# 例如: 0xE0200804

# 步骤 3: 在 config.json 中使用这个 HKL
```

### 配置文件示例

```json
{
  "key_bindings": [
    {
      "modifier_key": 91,
      "function_key": 74,
      "im_key": "0x04090409",
      "description": "Win+J → 英语"
    },
    {
      "modifier_key": 91,
      "function_key": 75,
      "im_key": "0x08040804",
      "description": "Win+K → 中文标准键盘"
    },
    {
      "modifier_key": 91,
      "function_key": 76,
      "im_key": "0xE0200804",
      "description": "Win+L → 中文输入法 (0xE020)"
    }
  ]
}
```

## 技术原理

### HKL 结构

```
HKL = 32/64 位键盘布局句柄

┌────────────────┬────────────────┐
│   高 16 位     │   低 16 位     │
│  布局 ID       │  语言 ID       │
│ (Layout ID)    │  (LANGID)      │
└────────────────┴────────────────┘

示例 1: 0x04090409
        │    └─ 0x0409 = 1033 (英语美国)
        └────── 0x0409 = 标准键盘

示例 2: 0x08040804
        │    └─ 0x0804 = 2052 (中文简体)
        └────── 0x0804 = 标准键盘

示例 3: 0xE0200804
        │    └─ 0x0804 = 2052 (中文简体)
        └────── 0xE020 = 特定输入法实现
```

### 为什么可以区分中文输入法?

- 原版 im-select: 只看**低 16 位**(语言 ID)
  - 微软拼音 → `2052`
  - 搜狗拼音 → `2052`
  - 标准键盘 → `2052`
  - **无法区分** ❌

- 增强版 im-select: 使用**完整 32 位** HKL
  - 微软拼音 → `0xE00E0804`
  - 搜狗拼音 → `0xE0200804` (示例)
  - 标准键盘 → `0x08040804`
  - **可以区分** ✅

## 常用布局 ID

| 布局 ID | 描述                       |
| ------- | -------------------------- |
| 0x0409  | 英语标准键盘               |
| 0x0804  | 中文标准键盘               |
| 0xE00E  | 微软拼音新体验             |
| 0xE010  | 微软五笔新体验             |
| 0xE001  | 微软拼音(旧版)             |
| 0xE002  | 微软五笔(旧版)             |
| 0xE020  | (你系统上的某个中文输入法) |

## 兼容性

完全兼容原版 im-select 的使用方式:

```bash
# 原版用法(仍然支持)
im-select-enhanced.exe          # 显示当前
im-select-enhanced.exe 1033     # 切换到英语

# 增强用法(推荐)
im-select-enhanced.exe -v               # 详细显示
im-select-enhanced.exe -l -v            # 列出所有
im-select-enhanced.exe -s 0x04090409    # 精确切换
```

## 下一步

1. ✅ **已完成:** 实现增强版 im-select
2. ✅ **已完成:** 编译和测试成功
3. 🔄 **建议:** 更新主程序使用新 API
4. 🔄 **建议:** 创建 GUI 输入法选择器
5. 🔄 **建议:** 从注册表读取输入法友好名称

## 命令行参数

```bash
im-select-enhanced.exe              # 显示当前输入法
im-select-enhanced.exe -v           # 显示当前输入法(详细)
im-select-enhanced.exe -l           # 列出所有输入法
im-select-enhanced.exe -l -v        # 列出所有输入法(详细)
im-select-enhanced.exe -s <HKL>     # 切换到指定输入法
im-select-enhanced.exe -s <HKL> -v  # 切换并显示详细信息
im-select-enhanced.exe -h           # 显示帮助信息
```

## 文件说明

- `im-select-enhanced.exe` - 独立命令行工具
- `cmd/im-select/main.go` - 工具源代码
- `im_selector_windows.go` - 核心 API 实现
- `IM-SELECT-ENHANCED.md` - 完整文档

---

**总结:** 增强版 im-select 通过使用完整的 HKL(键盘布局句柄),成功解决了原版无法区分不同中文输入法的问题!
