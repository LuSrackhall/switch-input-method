@echo off
chcp 65001 >nul
echo ========================================
echo   输入法快速切换工具 - 测试
echo ========================================
echo.
echo 正在启动程序(调试模式)...
echo.
echo 提示:
echo   - 程序将显示详细日志
echo   - 按 Ctrl+C 可以退出
echo   - Win+J: 切换到英文
echo   - Win+K: 切换到中文
echo.
echo ========================================
echo.

REM 使用普通编译版本以显示控制台输出
go build -o switch-input-method-debug.exe
switch-input-method-debug.exe

pause
