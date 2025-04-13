import Cocoa

class StatusBarController: NSObject, NSApplicationDelegate {
    func applicationDidFinishLaunching(_ notification: Notification) {
        let window = NSWindow(
            contentRect: NSRect(x: 0, y: 0, width: 200, height: 40), // 减小窗口高度
            styleMask: [.titled, .closable, .fullSizeContentView],
            backing: .buffered,
            defer: false
        )
        
        window.backgroundColor = NSColor(white: 1.0, alpha: 0.95)
        window.isOpaque = false
        window.level = .screenSaver
        window.hasShadow = true
        window.alphaValue = 1.0  // 设置固定透明度为1.0
        window.titlebarAppearsTransparent = true
        window.isMovableByWindowBackground = true
        window.titleVisibility = .hidden
        
        let effectView = NSVisualEffectView(frame: window.contentView!.bounds)
        effectView.material = .popover
        effectView.state = .active
        effectView.blendingMode = .withinWindow
        effectView.wantsLayer = true
        effectView.layer?.cornerRadius = 12
        effectView.layer?.masksToBounds = true
        
        // 调整文本框位置和大小
        let textField = NSTextField(frame: NSRect(
            x: 10,
            y: 8,
            width: window.contentView!.bounds.width - 20,
            height: 24
        ))
        textField.font = NSFont.systemFont(ofSize: 13)
        textField.bezelStyle = .roundedBezel
        textField.isEditable = true
        textField.isSelectable = true
        textField.placeholderString = "请输入密码..."
        (textField.cell as? NSSecureTextFieldCell)?.echosBullets = true
        textField.cell = NSSecureTextFieldCell()  // 设置为密码输入模式
        
        class TextFieldDelegate: NSObject, NSTextFieldDelegate {
            func controlTextDidChange(_ obj: Notification) {
                print("Text changed: \((obj.object as? NSTextField)?.stringValue ?? "")")
            }
            
            func control(_ control: NSControl, textView: NSTextView, doCommandBy commandSelector: Selector) -> Bool {
                if commandSelector == #selector(NSResponder.cancelOperation(_:)) {
                    NSApplication.shared.terminate(nil)
                    return true
                }
                return false
            }
        }
        
        let delegate = TextFieldDelegate()
        textField.delegate = delegate
        
        window.contentView?.addSubview(effectView)
        window.contentView?.addSubview(textField)
        
        // 设置初始响应者
        window.initialFirstResponder = textField
        
        // 设置固定位置显示窗口
        window.setFrameOrigin(NSPoint(x: 0, y: 0))
        
        // 显示窗口并确保焦点
        window.makeKeyAndOrderFront(nil)
        NSApp.activate(ignoringOtherApps: true)
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.05) {
            window.makeFirstResponder(textField)
            textField.becomeFirstResponder()
        }
        
        // 延迟关闭
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.01) {
            NSApplication.shared.terminate(nil)
        }
    }
}

let app = NSApplication.shared
NSApp.setActivationPolicy(.regular)
NSApp.activate(ignoringOtherApps: true) // 确保应用启动时激活
let controller = StatusBarController()
app.delegate = controller
// Thread.sleep(forTimeInterval: 8.0) // 这行注释很重要, 可以协助调试
app.run()