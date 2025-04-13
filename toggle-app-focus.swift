import Cocoa

class StatusBarController: NSObject, NSApplicationDelegate {
    func applicationDidFinishLaunching(_ notification: Notification) {
        let window = NSWindow(
            contentRect: NSRect(x: 0, y: 0, width: 200, height: 100),
            styleMask: [.titled, .closable, .texturedBackground],
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
        
        let label = NSTextField(labelWithString: "输入法切换成功")
        label.alignment = .center
        label.font = NSFont.systemFont(ofSize: 15, weight: .regular)
        label.textColor = .labelColor
        label.backgroundColor = .clear
        label.isBezeled = false
        label.isEditable = false
        
        label.sizeToFit()
        let labelFrame = NSRect(
            x: (window.contentView!.bounds.width - label.frame.width) / 2,
            y: window.contentView!.bounds.height - 35,
            width: label.frame.width,
            height: label.frame.height
        )
        label.frame = labelFrame
        
        // 使用 NSTextField（基于之前的建议）
        let textField = NSTextField(frame: NSRect(
            x: 20,
            y: 15,
            width: window.contentView!.bounds.width - 40,
            height: 24
        ))
        textField.font = NSFont.systemFont(ofSize: 13)
        textField.bezelStyle = .roundedBezel
        textField.isEditable = true
        textField.isSelectable = true
        textField.placeholderString = "请输入文本..."
        
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
        window.contentView?.addSubview(label)
        window.contentView?.addSubview(textField)
        
        // 设置初始响应者
        window.initialFirstResponder = textField
        
        // 居中窗口
        if let screen = NSScreen.main {
            let x = (screen.frame.width - window.frame.width) / 2
            let y = (screen.frame.height - window.frame.height) / 2
            window.setFrameOrigin(NSPoint(x: x, y: y))
        }
        
        // 显示窗口并确保焦点
        window.makeKeyAndOrderFront(nil)
        NSApp.activate(ignoringOtherApps: true)
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.05) {
            window.makeFirstResponder(textField)
            textField.becomeFirstResponder()
        }
        
        // 延迟关闭
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.05) {
            NSApplication.shared.terminate(nil)
        }
    }
}

let app = NSApplication.shared
NSApp.setActivationPolicy(.regular)
NSApp.activate(ignoringOtherApps: true) // 确保应用启动时激活
let controller = StatusBarController()
app.delegate = controller
// Thread.sleep(forTimeInterval: 8.0) . // 这行注释很重要, 可以协助调试
app.run()