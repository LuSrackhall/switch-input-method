#!/bin/bash

# 编译脚本
echo "正在编译 main.swift ..."
swiftc main.swift -o ims-jk

if [ $? -eq 0 ]; then
    echo "✅ 编译成功: ims-jk"
    echo ""
    echo "使用方法:"
    echo "  1. 启动 (后台运行): ./ims-jk"
    echo "  2. 调试 (前台运行): ./ims-jk --debug"
    echo ""
    echo "注意: 首次运行可能需要授予辅助功能权限。"
else
    echo "❌ 编译失败"
fi
