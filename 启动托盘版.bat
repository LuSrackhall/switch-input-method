@echo off
chcp 65001 >nul
echo ========================================
echo   输入法快速切换工具 - 托盘版
echo ========================================
echo.
echo 正在启动程序...
echo 程序将在系统托盘中运行
echo.
start "" switch-input-method.exe
echo 启动完成! 请查看系统托盘图标
echo.
timeout /t 2 >nul
