#!/bin/bash
# v2.5 功能测试脚本

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  v2.5 增强版 im-select 功能测试"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 检查文件是否存在
echo "1. 检查文件..."
if [ -f "im-select-enhanced.exe" ]; then
    echo "  ✓ im-select-enhanced.exe 存在"
else
    echo "  ✗ im-select-enhanced.exe 不存在"
    exit 1
fi

if [ -f "switch-input-method.exe" ]; then
    echo "  ✓ switch-input-method.exe 存在"
else
    echo "  ✗ switch-input-method.exe 不存在"
    exit 1
fi

echo ""
echo "2. 测试 im-select-enhanced..."
echo ""

# 显示当前输入法
echo "━━━ 当前输入法 ━━━"
./im-select-enhanced.exe -v
echo ""

# 列出所有输入法
echo "━━━ 所有输入法 ━━━"
./im-select-enhanced.exe -l -v
echo ""

# 测试切换
echo "━━━ 测试切换功能 ━━━"
echo ""

echo "→ 切换到英语..."
./im-select-enhanced.exe -s 0x04090409 -v
echo ""

sleep 1

echo "→ 切换到中文标准键盘..."
./im-select-enhanced.exe -s 0x08040804 -v
echo ""

sleep 1

echo "→ 显示最终状态..."
./im-select-enhanced.exe -v
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  ✓ 所有测试完成!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
