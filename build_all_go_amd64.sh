#!/bin/bash
set -e

# 获取脚本所在目录作为项目根目录
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 构建输出目录（相对于项目根目录）
BUILD_DIR="$ROOT_DIR/linglong-scdtool/files_x86/bin"

# 子项目列表
SUBPROJECTS=("txtmaker" "scdeditor" "scdparser" "scdmaker")

echo "========== Go Multi-project Build Script =========="
echo "Root:       $ROOT_DIR"
echo "Build dir:  $BUILD_DIR"
echo "Projects:   ${SUBPROJECTS[*]}"
echo "==================================================="

# 删除旧的构建结果
if [ -d "$BUILD_DIR" ]; then
    echo "🧹 Removing old build directory..."
    rm -rf "$BUILD_DIR"
fi

# 重新创建输出目录
mkdir -p "$BUILD_DIR"

# 设置 Go 国内代理（防止网络问题）
go env -w GOPROXY=https://goproxy.cn,direct

# 遍历并构建子项目
for proj in "${SUBPROJECTS[@]}"; do
    SRC_DIR="$ROOT_DIR/$proj"
    echo "---- Building $proj ----"
    if [ -d "$SRC_DIR" ]; then
        cd "$SRC_DIR"
        go mod tidy
        go build -ldflags="-s -w" -o "$BUILD_DIR/$proj" .
        echo "✔ Built: $BUILD_DIR/$proj"
    else
        echo "⚠ Warning: directory $proj not found, skipped."
    fi
done

echo "================ All builds finished ================"
echo "✅ Output binaries are located in: $BUILD_DIR"
