# GUI 界面实现可行性分析

## 目录
- [概述](#概述)
- [Go语言GUI框架对比](#go语言gui框架对比)
- [推荐方案](#推荐方案)
- [实现设计](#实现设计)
- [代码示例](#代码示例)
- [优缺点分析](#优缺点分析)
- [实施建议](#实施建议)

---

## 概述

当前项目是一个纯 Go 语言实现的 Windows 输入法切换工具,使用系统托盘和 JSON 配置文件。本文档分析实现图形化配置界面的可行性和具体方案。

**当前状态:**
- 配置方式: 手动编辑 `config.json` 文件
- 用户界面: 系统托盘菜单 + MessageBox
- 文件体积: 4.5 MB (优化构建)
- 依赖: systray, golang.org/x/sys/windows

**目标:**
- 提供可视化的配置编辑界面
- 支持添加/编辑/删除按键绑定
- 可视化选择虚拟键码和输入法
- 保持文件体积小,性能高

---

## Go语言GUI框架对比

### 1. walk (推荐) ⭐⭐⭐⭐⭐

**项目地址:** https://github.com/lxn/walk

**简介:**
- Windows Application Library Kit
- 纯 Go 实现,使用 Windows 原生控件
- 无需 CGO,编译简单

**特点:**
```
✅ 优点:
  - 纯 Go,无需 CGO
  - Windows 原生控件,性能好
  - 文件体积增加 2-3 MB
  - API 简单易用
  - 活跃维护,文档完善
  - 支持数据绑定

❌ 缺点:
  - 仅支持 Windows
  - 界面风格传统 (Win32)
  - 自定义样式有限
```

**适配度:** 95%
- 与当前项目完美契合
- 已使用 Windows API,迁移成本低
- 体积控制好

**示例代码:**
```go
import "github.com/lxn/walk"

func showConfigWindow() {
    mw := new(walk.MainWindow)
    
    if err := (MainWindow{
        AssignTo: &mw,
        Title:    "配置管理",
        Size:     Size{600, 400},
    }.Create()); err != nil {
        log.Fatal(err)
    }
    
    mw.Run()
}
```

---

### 2. fyne ⭐⭐⭐

**项目地址:** https://github.com/fyne-io/fyne

**简介:**
- 跨平台 GUI 框架
- 使用 OpenGL 渲染
- Material Design 风格

**特点:**
```
✅ 优点:
  - 跨平台(Windows/Mac/Linux)
  - 现代化 UI 设计
  - 丰富的控件
  - 活跃开发

❌ 缺点:
  - 文件体积大 (10+ MB)
  - OpenGL 依赖,兼容性问题
  - 渲染开销高
  - CGO 依赖,编译复杂
```

**适配度:** 60%
- 文件体积超标
- 性能开销大
- 对于简单配置界面过于重量级

---

### 3. webview ⭐⭐⭐⭐

**项目地址:** https://github.com/webview/webview

**简介:**
- 嵌入系统 WebView
- 使用 HTML/CSS/JS 构建界面
- 轻量级

**特点:**
```
✅ 优点:
  - 使用 Web 技术,灵活度高
  - 文件体积适中 (6-8 MB)
  - 可使用现代 CSS 框架
  - 跨平台

❌ 缺点:
  - 需要 CGO
  - 依赖系统 WebView
  - 调试相对复杂
  - 前后端通信开销
```

**适配度:** 70%
- 灵活但复杂度高
- 适合需要复杂交互的场景

---

### 4. wails ⭐⭐

**项目地址:** https://github.com/wailsapp/wails

**简介:**
- 类似 Electron 的框架
- Go 后端 + Web 前端
- 现代化开发体验

**特点:**
```
✅ 优点:
  - 现代化开发体验
  - 可使用 React/Vue/Svelte
  - 丰富的工具链

❌ 缺点:
  - 文件体积巨大 (20+ MB)
  - 打包复杂
  - 学习曲线陡峭
  - 过于重量级
```

**适配度:** 30%
- 对于简单配置工具过度设计

---

### 5. 手动 Win32 API ⭐⭐⭐

**简介:**
- 直接使用 Windows API
- 通过 syscall 调用

**特点:**
```
✅ 优点:
  - 完全控制
  - 无额外依赖
  - 文件体积最小

❌ 缺点:
  - 开发工作量大
  - 代码冗长
  - 容易出错
  - 维护困难
```

**适配度:** 50%
- 性能和体积最优
- 但开发成本太高

---

## 推荐方案

### 首选: walk

**理由:**
1. **兼容性好** - 与当前项目技术栈一致
2. **体积可控** - 增加 2-3 MB,总体积约 6-7 MB
3. **开发效率** - API 简单,开发快速
4. **原生体验** - Windows 原生控件,用户熟悉
5. **无 CGO** - 编译简单,交叉编译方便

**安装:**
```bash
go get github.com/lxn/walk
```

**文件结构:**
```
switch-input-method/
├── main.go                 # 主程序入口
├── config.go               # 配置管理
├── keyboard_hook_windows.go # 键盘钩子
├── tray_windows.go         # 系统托盘
├── config_dialog.go        # 原有对话框
├── gui_config.go           # 新增: GUI配置窗口
└── gui_models.go           # 新增: 数据模型
```

---

## 实现设计

### 界面布局

```
┌─────────────────────────────────────────┐
│  兴宜街道红旗路输入法切换工具 - 配置    │
├─────────────────────────────────────────┤
│                                         │
│  当前按键绑定:                          │
│  ┌───────────────────────────────────┐ │
│  │ 修饰键  │ 功能键 │ 输入法 │ 描述  │ │
│  ├───────────────────────────────────┤ │
│  │ Win    │ J     │ 1033  │ 英文   │ │
│  │ Win    │ K     │ 2052  │ 中文   │ │
│  │                                   │ │
│  └───────────────────────────────────┘ │
│                                         │
│  [添加] [编辑] [删除] [上移] [下移]     │
│                                         │
│  ─────────────────────────────────────  │
│                                         │
│  [虚拟键码参考] [获取当前输入法]        │
│                                         │
│  [保存] [应用] [重置] [取消]            │
│                                         │
└─────────────────────────────────────────┘
```

### 编辑对话框

```
┌─────────────────────────────────┐
│  编辑按键绑定                   │
├─────────────────────────────────┤
│                                 │
│  修饰键:  [▼ 左 Win 键 (91)]    │
│                                 │
│  功能键:  [▼ J 键 (74)]         │
│                                 │
│  输入法:  [获取当前]            │
│           1033                  │
│                                 │
│  描述:    [切换到英文输入法___] │
│                                 │
│                                 │
│        [确定]  [取消]           │
│                                 │
└─────────────────────────────────┘
```

### 功能模块

#### 1. 主配置窗口
- `TableView` 显示所有绑定
- 按钮工具栏
- 实时预览

#### 2. 编辑对话框
- `ComboBox` 选择修饰键
- `ComboBox` 选择功能键
- `LineEdit` 输入 im_key
- `LineEdit` 输入描述

#### 3. 虚拟键码选择器
- 分类列表(修饰键/字母/数字/功能键等)
- 搜索功能
- 键码和名称显示

#### 4. 输入法检测
- 调用 `GetCurrentInputMethod()`
- 实时显示当前输入法
- 一键填充

---

## 代码示例

### 主配置窗口

```go
package main

import (
    "github.com/lxn/walk"
    . "github.com/lxn/walk/declarative"
)

type ConfigWindow struct {
    *walk.MainWindow
    model *ConfigModel
    table *walk.TableView
}

func ShowGUIConfig() {
    cw := new(ConfigWindow)
    cw.model = NewConfigModel()
    
    if err := (MainWindow{
        AssignTo: &cw.MainWindow,
        Title:   "配置管理 - 输入法切换工具",
        Size:    Size{Width: 700, Height: 500},
        Layout:  VBox{},
        Children: []Widget{
            Label{
                Text: "当前按键绑定:",
                Font: Font{PointSize: 10, Bold: true},
            },
            TableView{
                AssignTo: &cw.table,
                Columns: []TableViewColumn{
                    {Title: "修饰键", Width: 100},
                    {Title: "功能键", Width: 100},
                    {Title: "输入法", Width: 150},
                    {Title: "描述", Width: 200},
                },
                Model: cw.model,
            },
            Composite{
                Layout: HBox{},
                Children: []Widget{
                    PushButton{
                        Text: "添加",
                        OnClicked: cw.addBinding,
                    },
                    PushButton{
                        Text: "编辑",
                        OnClicked: cw.editBinding,
                    },
                    PushButton{
                        Text: "删除",
                        OnClicked: cw.deleteBinding,
                    },
                    HSpacer{},
                    PushButton{
                        Text: "虚拟键码参考",
                        OnClicked: func() {
                            ShowKeyCodeReference()
                        },
                    },
                },
            },
            Composite{
                Layout: HBox{},
                Children: []Widget{
                    HSpacer{},
                    PushButton{
                        Text: "保存",
                        OnClicked: cw.save,
                    },
                    PushButton{
                        Text: "应用",
                        OnClicked: cw.apply,
                    },
                    PushButton{
                        Text: "取消",
                        OnClicked: func() {
                            cw.Close()
                        },
                    },
                },
            },
        },
    }.Create()); err != nil {
        walk.MsgBox(nil, "错误", "创建窗口失败: "+err.Error(), walk.MsgBoxIconError)
        return
    }
    
    cw.Run()
}

func (cw *ConfigWindow) addBinding() {
    // 显示编辑对话框
    ShowEditDialog(cw.MainWindow, nil, func(binding *KeyBinding) {
        cw.model.AddBinding(binding)
        cw.table.PublishRowsReset()
    })
}

func (cw *ConfigWindow) editBinding() {
    index := cw.table.CurrentIndex()
    if index < 0 {
        walk.MsgBox(cw, "提示", "请选择要编辑的项", walk.MsgBoxIconInformation)
        return
    }
    
    binding := cw.model.GetBinding(index)
    ShowEditDialog(cw.MainWindow, binding, func(updated *KeyBinding) {
        cw.model.UpdateBinding(index, updated)
        cw.table.PublishRowsReset()
    })
}

func (cw *ConfigWindow) deleteBinding() {
    index := cw.table.CurrentIndex()
    if index < 0 {
        walk.MsgBox(cw, "提示", "请选择要删除的项", walk.MsgBoxIconInformation)
        return
    }
    
    ret := walk.MsgBox(cw, "确认", "确定要删除此绑定吗?", walk.MsgBoxYesNo|walk.MsgBoxIconQuestion)
    if ret == walk.DlgCmdYes {
        cw.model.RemoveBinding(index)
        cw.table.PublishRowsReset()
    }
}

func (cw *ConfigWindow) save() {
    config := cw.model.ToConfig()
    if err := SaveConfig(config); err != nil {
        walk.MsgBox(cw, "错误", "保存配置失败: "+err.Error(), walk.MsgBoxIconError)
        return
    }
    walk.MsgBox(cw, "成功", "配置已保存", walk.MsgBoxIconInformation)
    cw.Close()
}

func (cw *ConfigWindow) apply() {
    config := cw.model.ToConfig()
    if err := SaveConfig(config); err != nil {
        walk.MsgBox(cw, "错误", "保存配置失败: "+err.Error(), walk.MsgBoxIconError)
        return
    }
    
    // 重新加载配置
    if err := InitConfig(); err != nil {
        walk.MsgBox(cw, "错误", "重新加载配置失败: "+err.Error(), walk.MsgBoxIconError)
        return
    }
    
    // 重启键盘钩子
    StopKeyboardHook()
    go StartKeyboardHook()
    
    walk.MsgBox(cw, "成功", "配置已应用", walk.MsgBoxIconInformation)
}
```

### 数据模型

```go
package main

import (
    "github.com/lxn/walk"
)

type ConfigModel struct {
    walk.TableModelBase
    bindings []*KeyBinding
}

func NewConfigModel() *ConfigModel {
    m := &ConfigModel{}
    
    // 加载现有配置
    config := GetCurrentConfig()
    if config != nil {
        for i := range config.KeyBindings {
            m.bindings = append(m.bindings, &config.KeyBindings[i])
        }
    }
    
    return m
}

func (m *ConfigModel) RowCount() int {
    return len(m.bindings)
}

func (m *ConfigModel) Value(row, col int) interface{} {
    binding := m.bindings[row]
    
    switch col {
    case 0: // 修饰键
        return GetKeyName(binding.ModifierKey)
    case 1: // 功能键
        return GetKeyName(binding.FunctionKey)
    case 2: // 输入法
        return binding.IMKey
    case 3: // 描述
        return binding.Description
    }
    
    return nil
}

func (m *ConfigModel) AddBinding(binding *KeyBinding) {
    m.bindings = append(m.bindings, binding)
}

func (m *ConfigModel) UpdateBinding(index int, binding *KeyBinding) {
    if index >= 0 && index < len(m.bindings) {
        m.bindings[index] = binding
    }
}

func (m *ConfigModel) RemoveBinding(index int) {
    if index >= 0 && index < len(m.bindings) {
        m.bindings = append(m.bindings[:index], m.bindings[index+1:]...)
    }
}

func (m *ConfigModel) GetBinding(index int) *KeyBinding {
    if index >= 0 && index < len(m.bindings) {
        return m.bindings[index]
    }
    return nil
}

func (m *ConfigModel) ToConfig() *Config {
    config := &Config{
        KeyBindings: make([]KeyBinding, len(m.bindings)),
    }
    
    for i, b := range m.bindings {
        config.KeyBindings[i] = *b
    }
    
    return config
}
```

### 编辑对话框

```go
package main

import (
    "github.com/lxn/walk"
    . "github.com/lxn/walk/declarative"
)

func ShowEditDialog(owner walk.Form, binding *KeyBinding, onSave func(*KeyBinding)) {
    var dlg *walk.Dialog
    var modifierCombo, functionCombo *walk.ComboBox
    var imKeyEdit, descEdit *walk.LineEdit
    
    // 准备下拉列表数据
    modifierKeys := []string{
        "左 Win 键 (91)", "右 Win 键 (92)",
        "左 Ctrl (162)", "右 Ctrl (163)",
        "左 Alt (164)", "右 Alt (165)",
        "左 Shift (160)", "右 Shift (161)",
    }
    
    functionKeys := []string{
        "J 键 (74)", "K 键 (75)", "L 键 (76)", "I 键 (73)",
        "U 键 (85)", "O 键 (79)", "P 键 (80)",
    }
    
    // 默认值
    modifierIndex := 0
    functionIndex := 0
    imKey := ""
    desc := ""
    
    if binding != nil {
        // 编辑模式,填充现有值
        imKey = binding.IMKey
        desc = binding.Description
        // TODO: 根据键码找到对应的索引
    }
    
    title := "添加按键绑定"
    if binding != nil {
        title = "编辑按键绑定"
    }
    
    Dialog{
        AssignTo: &dlg,
        Title:    title,
        Size:     Size{Width: 400, Height: 300},
        Layout:   VBox{},
        Children: []Widget{
            Composite{
                Layout: Grid{Columns: 2},
                Children: []Widget{
                    Label{Text: "修饰键:"},
                    ComboBox{
                        AssignTo:      &modifierCombo,
                        Model:        modifierKeys,
                        CurrentIndex: modifierIndex,
                    },
                    
                    Label{Text: "功能键:"},
                    ComboBox{
                        AssignTo:      &functionCombo,
                        Model:        functionKeys,
                        CurrentIndex: functionIndex,
                    },
                    
                    Label{Text: "输入法标识:"},
                    Composite{
                        Layout: HBox{},
                        Children: []Widget{
                            LineEdit{
                                AssignTo: &imKeyEdit,
                                Text:     imKey,
                            },
                            PushButton{
                                Text: "获取当前",
                                OnClicked: func() {
                                    im, err := GetCurrentInputMethod()
                                    if err == nil {
                                        imKeyEdit.SetText(im)
                                    }
                                },
                            },
                        },
                    },
                    
                    Label{Text: "描述:"},
                    LineEdit{
                        AssignTo: &descEdit,
                        Text:     desc,
                    },
                },
            },
            Composite{
                Layout: HBox{},
                Children: []Widget{
                    HSpacer{},
                    PushButton{
                        Text: "确定",
                        OnClicked: func() {
                            // 创建绑定对象
                            newBinding := &KeyBinding{
                                ModifierKey: getKeyCodeFromComboIndex(modifierCombo.CurrentIndex(), true),
                                FunctionKey: getKeyCodeFromComboIndex(functionCombo.CurrentIndex(), false),
                                IMKey:       imKeyEdit.Text(),
                                Description: descEdit.Text(),
                            }
                            
                            // 验证
                            if newBinding.IMKey == "" {
                                walk.MsgBox(dlg, "错误", "请输入输入法标识", walk.MsgBoxIconError)
                                return
                            }
                            
                            onSave(newBinding)
                            dlg.Accept()
                        },
                    },
                    PushButton{
                        Text: "取消",
                        OnClicked: func() {
                            dlg.Cancel()
                        },
                    },
                },
            },
        },
    }.Run(owner)
}

func getKeyCodeFromComboIndex(index int, isModifier bool) uint32 {
    if isModifier {
        codes := []uint32{91, 92, 162, 163, 164, 165, 160, 161}
        if index >= 0 && index < len(codes) {
            return codes[index]
        }
    } else {
        codes := []uint32{74, 75, 76, 73, 85, 79, 80}
        if index >= 0 && index < len(codes) {
            return codes[index]
        }
    }
    return 0
}
```

---

## 优缺点分析

### 使用 walk 实现 GUI 的优点

1. **用户体验提升**
   - 可视化配置,降低学习成本
   - 避免 JSON 语法错误
   - 实时验证和预览
   - 友好的错误提示

2. **功能增强**
   - 下拉选择虚拟键码
   - 一键获取当前输入法
   - 拖拽排序
   - 导入/导出配置

3. **开发效率**
   - API 简单,学习曲线平缓
   - 丰富的控件和布局
   - 数据绑定机制

4. **维护性**
   - 代码结构清晰
   - 易于扩展新功能
   - 统一的界面风格

### 缺点

1. **文件体积**
   - 增加 2-3 MB (总计 6-7 MB)
   - 对于配置工具可接受

2. **依赖增加**
   - 新增 walk 依赖
   - 需要更新文档

3. **开发工作量**
   - 初次开发需要 1-2 天
   - 后续维护成本

4. **仅支持 Windows**
   - 与当前项目一致
   - 不是问题

---

## 实施建议

### 方案 A: 完全替换 (推荐)

**实施步骤:**
1. 保留托盘菜单作为主入口
2. "查看当前配置"改为打开 GUI 窗口
3. 移除现有的 MessageBox 对话框
4. 统一使用 walk 界面

**优点:**
- 体验一致
- 代码简洁
- 用户友好

**缺点:**
- 开发工作量稍大
- 需要充分测试

### 方案 B: 混合模式

**实施步骤:**
1. 保留现有 MessageBox 功能
2. 新增 "高级配置" 菜单项打开 GUI
3. 用户可选择使用方式

**优点:**
- 兼容性好
- 渐进式迁移
- 风险低

**缺点:**
- 代码冗余
- 两套逻辑维护

### 方案 C: 暂不实施 (适用于当前)

**保持现状:**
- 继续使用 JSON + 注释
- 托盘菜单 + MessageBox
- 文档完善

**适用场景:**
- 用户不频繁修改配置
- 追求最小体积
- 开发资源有限

**当前评估:**
考虑到:
1. 配置文件已有详细注释
2. 快速绑定功能已简化
3. 托盘菜单操作便捷
4. 用户修改配置频率低

**建议: 暂不实施完整 GUI**
- 当前的 JSON + 工具菜单已足够友好
- 保持轻量级特性
- 未来有需求时再考虑

---

## 总结

### GUI 实现可行性: ✅ 完全可行

**技术可行性:** 95%
- walk 框架成熟稳定
- API 简单易用
- 与现有代码兼容

**经济可行性:** 70%
- 开发成本: 1-2 天
- 体积增加: 2-3 MB
- 维护成本: 低

**用户需求:** 60%
- 高级用户更喜欢配置文件
- 新手用户需要 GUI
- 当前方案已较友好

### 最终建议

**短期 (当前版本 v2.4):**
- ✅ 保持现有方案
- ✅ 完善配置文件注释
- ✅ 优化工具菜单
- ✅ 简化快速绑定

**中期 (v2.5-v3.0):**
- 🔄 观察用户反馈
- 🔄 评估 GUI 需求
- 🔄 制作原型演示

**长期 (v3.0+):**
- 🚀 若用户强烈需求,实施 GUI
- 🚀 使用 walk 框架
- 🚀 完全替换模式

---

## 附录

### walk 学习资源

- **官方文档:** https://github.com/lxn/walk
- **示例代码:** https://github.com/lxn/walk/tree/master/examples
- **中文教程:** https://studygolang.com/articles/12489

### 参考项目

使用 walk 的成功案例:
- **Notepad++** 插件管理器
- **Various Windows 工具** 
- **企业内部管理工具**

---

**文档版本:** 1.0  
**最后更新:** 2025年10月3日  
**维护者:** switch-input-method 开发团队
