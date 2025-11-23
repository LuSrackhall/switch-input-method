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

// MARK: - 主程序入口

// 获取命令行参数
let args = ProcessInfo.processInfo.arguments

// // 检查是否需要以守护进程方式运行
// if args.count == 1 {
//     // 启动守护进程
//     let executablePath = args[0]
//     let task = Process()
//     task.launchPath = executablePath
//     task.arguments = ["daemon"]
    
//     // 设置进程属性
//     // cmd.Stdin = nil ... (Swift Process 默认不连接)
    
//     // 启动时分离进程
//     // Go: cmd.SysProcAttr = &syscall.SysProcAttr{ Setsid: true }
//     // Swift Process 无法直接设置 setsid，我们在子进程启动后立即调用 setsid()
    
//     do {
//         try task.run()
        
//         print("程序已在后台启动，进程 PID: \(task.processIdentifier)")
//         print("要结束程序，请执行: kill \(task.processIdentifier)")
//         print("----------, 或执行: kill -9 \(task.processIdentifier)")
        
//         // cmd.Process.Release()
        
//         print("---")
//         print("---")
//         print("下方读秒只是为了方便用户判断当前终端是否卡死, 但主要功能是防止主要逻辑被重复执行")
//         Thread.sleep(forTimeInterval: 1)
        
//         // 计时以告知用户当前终端是没有卡死的
//         var i = 0
//         while true {
//             print("\r-------------- \(i) 秒...", terminator: "")
//             fflush(stdout)
//             Thread.sleep(forTimeInterval: 1)
//             i += 1
//         }
//     } catch {
//         print("启动失败: \(error)")
//         exit(1)
//     }
// }

// 实际的程序逻辑
KeyEventListen()

// MARK: - 函数定义

func KeyEventListen() {
    // 对应 Go: Setsid: true 的效果，子进程调用 setsid 创建新会话
    setsid()
    
    // evChan := hook.Start() ...
    
    let eventMask = (1 << CGEventType.keyDown.rawValue) | (1 << CGEventType.keyUp.rawValue) | (1 << CGEventType.flagsChanged.rawValue)
    
    guard let eventTap = CGEvent.tapCreate(
        tap: .cgSessionEventTap,
        place: .headInsertEventTap,
        options: .defaultTap,
        eventsOfInterest: CGEventMask(eventMask),
        callback: eventCallback,
        userInfo: nil
    ) else {
        print("无法创建事件监听器。请确保已授予辅助功能权限。")
        exit(1)
    }
    
    let runLoopSource = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, eventTap, 0)
    CFRunLoopAddSource(CFRunLoopGetCurrent(), runLoopSource, CFRunLoopMode.commonModes)
    CGEvent.tapEnable(tap: eventTap, enable: true)
    
    CFRunLoopRun()
}

func eventCallback(proxy: CGEventTapProxy, type: CGEventType, event: CGEvent, refcon: UnsafeMutableRawPointer?) -> Unmanaged<CGEvent>? {
    let keyCode = UInt16(event.getIntegerValueField(.keyboardEventKeycode))
    
    // handleKeyEvent 逻辑
    
    // 处理修饰键 (FlagsChanged)
    if type == .flagsChanged {
        if pressedModifiers.contains(keyCode) {
            // 之前已按下，现在触发说明是松开 (修饰键不重复触发)
            pressedModifiers.remove(keyCode)
            print("\nKeyUp - Keycode: \(keyCode)")
            if keyCode == KEY_OPTION {
                OPTION = false
            }
        } else {
            // 之前未按下，现在触发说明是按下
            pressedModifiers.insert(keyCode)
            print("\nKeyHold - Keycode: \(keyCode)")
            if keyCode == KEY_OPTION {
                OPTION = true
            }
        }
        return Unmanaged.passUnretained(event)
    }
    
    if type == .keyDown { // KeyHold (Go: Kind == 4)
        // Go: if !key_down_soundIsRun { ... }
        // Swift: check autorepeat
        let isRepeat = event.getIntegerValueField(.keyboardEventAutorepeat) != 0
        
        if !isRepeat {
            print("\nKeyHold - Keycode: \(keyCode)")
            
            // if keyCode == KEY_OPTION {
            //    OPTION = true
            // }
            
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
                    // 注释掉的焦点转移逻辑 ...
                }
            }
        }
    }
    
    if type == .keyUp { // KeyUp (Go: Kind == 5)
        print("\nKeyUp - Keycode: \(keyCode)")
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
