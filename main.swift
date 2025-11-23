#!/usr/bin/env swift

/*
 * 构建方式:
 * swiftc main.swift -o main // 已弃用, 改为使用build.sh脚本
 *
 * 调试方式:
 * 1. 赋予执行权限: chmod +x main.swift
 * 2. 直接运行: ./main.swift
 * 3. 编译后运行: ./main
 *
 * 注意: 首次运行需要授予终端或应用辅助功能权限 (Accessibility)
 */

import Foundation
import AppKit
import Carbon

// MARK: - 全局变量与常量 (保持与 main.go 一致)

// Store 定义事件存储结构 (保留注释以保持一致)
// type Store struct { ... }

// var Clients_sse_stores sync.Map
// var once_stores sync.Once
// var mutex sync.Mutex

var OPTION = false
var pressedModifiers = Set<UInt16>()

// 定义按键码
let KEY_OPTION: UInt16 = 56
let KEY_TARGET_1: UInt16 = 38 // J
let KEY_TARGET_2: UInt16 = 40 // K

// 默认输入法 ID
let DEFAULT_IM_1 = "com.apple.keylayout.UnicodeHexInput"
let DEFAULT_IM_2 = "im.rime.inputmethod.Squirrel.Hans"

// 当前配置 (优先从 UserDefaults 加载)
var targetIM1 = UserDefaults.standard.string(forKey: "TargetIM1") ?? DEFAULT_IM_1
var targetIM2 = UserDefaults.standard.string(forKey: "TargetIM2") ?? DEFAULT_IM_2

// MARK: - 辅助类与扩展

struct InputSource {
    let id: String
    let name: String
}

func getInstalledInputSources() -> [InputSource] {
    let properties = [kTISPropertyInputSourceCategory: kTISCategoryKeyboardInputSource] as CFDictionary
    guard let sources = TISCreateInputSourceList(properties, false)?.takeRetainedValue() as? [TISInputSource] else {
        return []
    }
    
    var result: [InputSource] = []
    for source in sources {
        // 获取 ID
        let ptrID = TISGetInputSourceProperty(source, kTISPropertyInputSourceID)
        guard let ptrID = ptrID else { continue }
        let id = Unmanaged<CFString>.fromOpaque(ptrID).takeUnretainedValue() as String
        
        // 获取名称
        let ptrName = TISGetInputSourceProperty(source, kTISPropertyLocalizedName)
        let name: String
        if let ptrName = ptrName {
            name = Unmanaged<CFString>.fromOpaque(ptrName).takeUnretainedValue() as String
        } else {
            name = id
        }
        
        // 过滤掉一些非输入法的源 (可选)
        result.append(InputSource(id: id, name: name))
    }
    return result.sorted { $0.name < $1.name }
}

class AppDelegate: NSObject, NSApplicationDelegate {
    var statusItem: NSStatusItem!
    var menu: NSMenu!
    
    func applicationDidFinishLaunching(_ notification: Notification) {
        // 设置托盘图标
        statusItem = NSStatusBar.system.statusItem(withLength: NSStatusItem.variableLength)
        if let button = statusItem.button {
            button.title = "⌨️"
            button.toolTip = "输入法切换助手"
        }
        
        buildMenu()
        
        // 启动键盘监听
        startKeyEventListen()
    }
    
    func buildMenu() {
        menu = NSMenu()
        
        // 标题
        let titleItem = NSMenuItem(title: "输入法切换助手", action: nil, keyEquivalent: "")
        titleItem.isEnabled = false
        menu.addItem(titleItem)
        menu.addItem(NSMenuItem.separator())
        
        // Option + J 设置
        let im1Menu = NSMenu()
        let im1Item = NSMenuItem(title: "Option + J 切换至...", action: nil, keyEquivalent: "")
        im1Item.submenu = im1Menu
        menu.addItem(im1Item)
        
        // Option + K 设置
        let im2Menu = NSMenu()
        let im2Item = NSMenuItem(title: "Option + K 切换至...", action: nil, keyEquivalent: "")
        im2Item.submenu = im2Menu
        menu.addItem(im2Item)
        
        // 填充子菜单
        let sources = getInstalledInputSources()
        
        // 填充 Option + J 列表
        for source in sources {
            let item = NSMenuItem(title: source.name, action: #selector(selectIM1(_:)), keyEquivalent: "")
            item.target = self
            item.representedObject = source.id
            if source.id == targetIM1 { item.state = .on }
            im1Menu.addItem(item)
        }
        
        // 填充 Option + K 列表
        for source in sources {
            let item = NSMenuItem(title: source.name, action: #selector(selectIM2(_:)), keyEquivalent: "")
            item.target = self
            item.representedObject = source.id
            if source.id == targetIM2 { item.state = .on }
            im2Menu.addItem(item)
        }
        
        menu.addItem(NSMenuItem.separator())
        menu.addItem(NSMenuItem(title: "退出", action: #selector(quit), keyEquivalent: "q"))
        
        statusItem.menu = menu
    }
    
    @objc func selectIM1(_ sender: NSMenuItem) {
        guard let id = sender.representedObject as? String else { return }
        targetIM1 = id
        UserDefaults.standard.set(id, forKey: "TargetIM1")
        print("Option + J 已设置为: \(sender.title) (\(id))")
        buildMenu() // 重建菜单以更新勾选状态
    }
    
    @objc func selectIM2(_ sender: NSMenuItem) {
        guard let id = sender.representedObject as? String else { return }
        targetIM2 = id
        UserDefaults.standard.set(id, forKey: "TargetIM2")
        print("Option + K 已设置为: \(sender.title) (\(id))")
        buildMenu() // 重建菜单以更新勾选状态
    }
    
    @objc func quit() {
        NSApplication.shared.terminate(nil)
    }
}

func getModifiersString(_ flags: CGEventFlags) -> String {
    var mods: [String] = []
    if flags.contains(.maskAlphaShift) { mods.append("Caps") }
    if flags.contains(.maskShift) { mods.append("Shift") }
    if flags.contains(.maskControl) { mods.append("Ctrl") }
    if flags.contains(.maskAlternate) { mods.append("Opt") }
    if flags.contains(.maskCommand) { mods.append("Cmd") }
    if flags.contains(.maskSecondaryFn) { mods.append("Fn") }
    return mods.isEmpty ? "None" : mods.joined(separator: "+")
}

func getEventName(_ type: CGEventType) -> String {
    switch type {
    case .keyDown: return "KeyDown"
    case .keyUp: return "KeyUp"
    case .flagsChanged: return "FlagsChanged"
    default: return "Unknown"
    }
}

// MARK: - 主程序入口

// 获取命令行参数
let args = ProcessInfo.processInfo.arguments

// 检查运行模式
// 默认模式: 启动后台守护进程并退出
// 调试模式: ./main.swift --debug (前台运行)
// 内部模式: ./main.swift --daemon (后台实际运行的进程)

if args.contains("--daemon") {
    // --- 子进程 (守护进程) 逻辑 ---
    // 创建新的会话，脱离控制终端
    setsid()
    
    // 重定向标准输入输出到 /dev/null (防止向已关闭的终端输出导致 SIGPIPE)
    freopen("/dev/null", "r", stdin)
    freopen("/dev/null", "w", stdout)
    freopen("/dev/null", "w", stderr)
    
} else if args.contains("--debug") {
    // --- 调试模式 (前台运行) ---
    print("程序以调试模式启动 (PID: \(ProcessInfo.processInfo.processIdentifier))...")
    print("日志将直接输出到终端。")
    
} else {
    // --- 默认模式 (启动后台进程) ---
    let executablePath = args[0]
    let task = Process()
    task.launchPath = executablePath
    task.arguments = ["--daemon"]
    
    do {
        try task.run()
        print("✅ 程序已在后台启动 (PID: \(task.processIdentifier))")
        print("⌨️  您可以在菜单栏找到图标进行管理")
        exit(0)
    } catch {
        print("❌ 启动失败: \(error)")
        exit(1)
    }
}

// 初始化应用
let app = NSApplication.shared
let delegate = AppDelegate()
app.delegate = delegate

// 设置为 Accessory 应用 (隐藏 Dock 图标，只显示菜单栏图标)
app.setActivationPolicy(.accessory)

// 运行应用 (这将启动主 RunLoop)
app.run()

// MARK: - 函数定义

func startKeyEventListen() {
    let eventMask = (1 << CGEventType.keyDown.rawValue) | (1 << CGEventType.keyUp.rawValue) | (1 << CGEventType.flagsChanged.rawValue)
    
    guard let eventTap = CGEvent.tapCreate(
        tap: .cgSessionEventTap,
        place: .headInsertEventTap,
        options: .defaultTap,
        eventsOfInterest: CGEventMask(eventMask),
        callback: eventCallback,
        userInfo: nil
    ) else {
        print("❌ 无法创建事件监听器。")
        print("请确保已授予终端或此程序辅助功能权限 (Accessibility)。")
        print("设置路径: 系统设置 -> 隐私与安全性 -> 辅助功能")
        exit(1)
    }
    
    let runLoopSource = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, eventTap, 0)
    CFRunLoopAddSource(CFRunLoopGetCurrent(), runLoopSource, CFRunLoopMode.commonModes)
    CGEvent.tapEnable(tap: eventTap, enable: true)
    
    print("✅ 键盘监听已启动...")
    print("⌨️  托盘图标已加载")
}

func eventCallback(proxy: CGEventTapProxy, type: CGEventType, event: CGEvent, refcon: UnsafeMutableRawPointer?) -> Unmanaged<CGEvent>? {
    let keyCode = UInt16(event.getIntegerValueField(.keyboardEventKeycode))
    let flags = event.flags
    let eventName = getEventName(type)
    let modsString = getModifiersString(flags)
    
    // 打印详细信息
    // 格式: [时间] 事件类型 | KeyCode: XX | Mods: [Ctrl+Opt...]
    let dateFormatter = DateFormatter()
    dateFormatter.dateFormat = "HH:mm:ss.SSS"
    let timeString = dateFormatter.string(from: Date())
    
    print("[\(timeString)] \(eventName.padding(toLength: 12, withPad: " ", startingAt: 0)) | KeyCode: \(String(format: "%3d", keyCode)) | Mods: \(modsString)")
    
    // handleKeyEvent 逻辑
    
    // 处理修饰键 (FlagsChanged)
    if type == .flagsChanged {
        if pressedModifiers.contains(keyCode) {
            pressedModifiers.remove(keyCode)
            if keyCode == KEY_OPTION { OPTION = false }
        } else {
            pressedModifiers.insert(keyCode)
            if keyCode == KEY_OPTION { OPTION = true }
        }
        return Unmanaged.passUnretained(event)
    }
    
    if type == .keyDown {
        let isRepeat = event.getIntegerValueField(.keyboardEventAutorepeat) != 0
        
        if !isRepeat {
            // 检查是否是目标按键组合
            // 使用 flags 检查 Option 键是否按下 (支持左右 Option 键)
            let isOptionHeld = event.flags.contains(.maskAlternate)
            
            if isOptionHeld {
                if keyCode == KEY_TARGET_1 { // J
                    DispatchQueue.global().async {
                        switchInputIfNeeded(targetIM1)
                    }
                }
                if keyCode == KEY_TARGET_2 { // K
                    DispatchQueue.global().async {
                        switchInputIfNeeded(targetIM2)
                    }
                }
            }
        }
    }
    
    if type == .keyUp {
        if keyCode == KEY_OPTION {
            OPTION = false
        }
    }
    
    return Unmanaged.passUnretained(event)
}

func switchInputIfNeeded(_ imkey: String) {
    let task = Process()
    task.launchPath = "/usr/bin/env"
    task.arguments = ["ims-mac", imkey]
    
    do {
        try task.run()
        task.waitUntilExit()
        if task.terminationStatus != 0 {
             print("切换失败 (Exit Code: \(task.terminationStatus))")
        }
    } catch {
        print("切换失败 \(error)")
    }
}
