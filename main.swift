#!/usr/bin/env swift

/*
 * 构建方式:
 * swiftc main.swift -o main
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

// MARK: - 辅助类与扩展

class AppDelegate: NSObject, NSApplicationDelegate {
    var statusItem: NSStatusItem!
    
    func applicationDidFinishLaunching(_ notification: Notification) {
        // 设置托盘图标
        statusItem = NSStatusBar.system.statusItem(withLength: NSStatusItem.variableLength)
        if let button = statusItem.button {
            button.title = "⌨️" // 使用字符作为图标
            button.toolTip = "输入法切换助手"
        }
        
        // 创建菜单
        let menu = NSMenu()
        menu.addItem(NSMenuItem(title: "输入法切换助手正在运行", action: nil, keyEquivalent: ""))
        menu.addItem(NSMenuItem.separator())
        menu.addItem(NSMenuItem(title: "退出", action: #selector(quit), keyEquivalent: "q"))
        statusItem.menu = menu
        
        // 启动键盘监听
        startKeyEventListen()
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

// 检查是否需要以守护进程方式运行
// 为了方便调试，默认改为前台运行。
// 如果需要后台运行，请使用: ./main.swift daemon
if args.contains("daemon") {
    // --- 子进程 (守护进程) 逻辑 ---
    // 创建新的会话
    setsid()
} else {
    // --- 前台运行逻辑 ---
    print("程序以前台模式启动 (PID: \(ProcessInfo.processInfo.processIdentifier))...")
    print("如需后台运行，请执行: ./main.swift daemon &")
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
                        switchInputIfNeeded("com.apple.keylayout.UnicodeHexInput")
                    }
                }
                if keyCode == KEY_TARGET_2 { // K
                    DispatchQueue.global().async {
                        switchInputIfNeeded("im.rime.inputmethod.Squirrel.Hans")
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
