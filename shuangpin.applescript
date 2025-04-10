tell application "System Events"
	-- 获取当前输入法标识符
	set currentInput to do shell script "defaults read ~/Library/Preferences/com.apple.HIToolbox.plist AppleSelectedInputSources"
	
	-- 检查是否包含 "KeyboardLayout Name" = ABC
	set isABCLayout to (currentInput contains "\"Input Mode\" = \"com.apple.inputmethod.SCIM.Shuangpin\"")
	
	-- 调试：显示当前输入法内容
	display dialog "Current Input: " & currentInput & return & "Is ABC Layout? " & isABCLayout buttons {"OK"} default button "OK"
	
	-- 如果不是 ABC，则切换输入法
	if isABCLayout is false then
		key code 49 using {control down} -- 触发 Control+空格
		display dialog "Switched to next input." buttons {"OK"} default button "OK"
	else
		display dialog "Already on ABC, no switch needed." buttons {"OK"} default button "OK"
	end if
end tell