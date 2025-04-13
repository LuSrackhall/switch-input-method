import Cocoa

class StatusBarController: NSObject, NSApplicationDelegate {
    func applicationDidFinishLaunching(_ notification: Notification) {
        // 创建临时状态栏项
        let statusItem = NSStatusBar.system.statusItem(withLength: NSStatusItem.variableLength)
        
        // 创建一个简单的视图控制器
        let viewController = NSViewController()
        let label = NSTextField(labelWithString: "输入法切换成功")
        label.alignment = .center
        viewController.view = label
        
        // 创建弹出窗口
        let popover = NSPopover()
        popover.contentViewController = viewController
        popover.behavior = .transient
        
        // 显示弹出窗口
        if let button = statusItem.button {
            popover.show(relativeTo: button.bounds, of: button, preferredEdge: .minY)
            
            // 添加全局事件监听
            NSEvent.addGlobalMonitorForEvents(matching: [.keyDown, .leftMouseDown, .rightMouseDown]) { _ in
                popover.close()
                NSApplication.shared.terminate(nil)
            }
            
            // 设置自动关闭
            DispatchQueue.main.asyncAfter(deadline: .now() + 0.5) {
                popover.close()
                NSApplication.shared.terminate(nil)
            }
        }
    }
}

// 创建并运行应用
let app = NSApplication.shared
let controller = StatusBarController()
app.delegate = controller
app.run()
