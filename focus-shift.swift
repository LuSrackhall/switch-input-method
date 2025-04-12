// 运行命令: swift focus-shift.swift
// 编译命令: swiftc focus-shift.swift -o focus-shift

import AppKit

class StatusBarController: NSObject, NSApplicationDelegate {
    func getCurrentMousePosition() -> CGPoint {
        let currentPos = CGEvent(source: nil)?.location ?? .zero
        // CGEvent坐标系(左下角为原点)转换为屏幕坐标系(左上角为原点)
        return CGPoint(x: currentPos.x, y: currentPos.y)
    }
    
    func moveMouseTo(_ position: CGPoint) {
        // 屏幕坐标系(左上角为原点)转换为CGEvent坐标系(左下角为原点)
        let adjustedPosition = CGPoint(x: position.x, y: position.y)
        let moveEvent = CGEvent(mouseEventSource: nil, mouseType: .mouseMoved, mouseCursorPosition: adjustedPosition, mouseButton: .left)
        moveEvent?.post(tap: .cghidEventTap)
    }
    
    func performClick(at position: CGPoint) {
        let screenHeight = NSScreen.main?.frame.height ?? 0
        // 屏幕坐标系转换为CGEvent坐标系
        let adjustedPosition = CGPoint(x: position.x, y: position.y)
        
        // 鼠标按下事件
        let clickDown = CGEvent(mouseEventSource: nil, mouseType: .leftMouseDown, mouseCursorPosition: adjustedPosition, mouseButton: .left)
        clickDown?.post(tap: .cghidEventTap)
        
        // 鼠标松开事件
        let clickUp = CGEvent(mouseEventSource: nil, mouseType: .leftMouseUp, mouseCursorPosition: adjustedPosition, mouseButton: .left)
        clickUp?.post(tap: .cghidEventTap)
    }

    func applicationDidFinishLaunching(_ notification: Notification) {
        // 1. 记录当前位置
        let originalPosition = getCurrentMousePosition()
        print("1. 记录的原始位置: \(originalPosition)")
        
        
        // 执行点击(这是本脚本的核心, 改变x的值, 使其点击位置不会触发任何菜单即可)
        self.performClick(at: CGPoint(x: 900, y: 0))
        print("2. 在(900,0)位置执行了点击")
        
        // 3. 2秒后移回原位置
        // DispatchQueue.main.asyncAfter(deadline: .now() + 2.0) {
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.1) {
            print("3. 准备移回原始位置: \(originalPosition)")
            self.moveMouseTo(originalPosition)
            
            // 添加短暂延迟确保移动完成后再获取位置
            DispatchQueue.main.asyncAfter(deadline: .now() + 0.1) {
                let finalPosition = self.getCurrentMousePosition()
                print("4. 移回后的实际位置: \(finalPosition)")
                
                // 程序退出
                print("程序退出")
                exit(0)
            }
        }
    }
}

let app = NSApplication.shared
let controller = StatusBarController()
app.delegate = controller
app.run()
