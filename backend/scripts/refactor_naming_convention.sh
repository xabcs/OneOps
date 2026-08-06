#!/bin/bash

# Golang 命名规范改造脚本
# 用途：自动重命名目录并更新所有导入路径

set -e  # 遇到错误立即退出

PROJECT_ROOT="/Users/mpm/Desktop/doc/go/OneOpsV2/OneOps"
BACKEND_ROOT="$PROJECT_ROOT/backend"
BACKUP_BRANCH="refactor/golang-naming-convention"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

echo_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

echo_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查是否在 git 仓库中
check_git_repo() {
    cd "$PROJECT_ROOT"
    if [ ! -d ".git" ]; then
        echo_error "不在 Git 仓库中"
        exit 1
    fi
    echo_info "检测到 Git 仓库"
}

# 创建备份分支
create_backup_branch() {
    echo_info "创建备份分支: $BACKUP_BRANCH"

    # 检查分支是否已存在
    if git show-ref --verify --quiet refs/heads/$BACKUP_BRANCH; then
        echo_warn "分支 $BACKUP_BRANCH 已存在"
        read -p "是否删除并重新创建？(y/n): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            git branch -D $BACKUP_BRANCH
            git checkout -b $BACKUP_BRANCH
        else
            echo_info "切换到现有分支: $BACKUP_BRANCH"
            git checkout $BACKUP_BRANCH
        fi
    else
        git checkout -b $BACKUP_BRANCH
    fi

    echo_info "备份分支创建成功"
}

# 重命名目录
rename_directories() {
    echo_info "开始重命名目录..."

    cd "$BACKEND_ROOT"

    # 重命名 middlewares → middleware
    if [ -d "middlewares" ]; then
        echo_info "重命名: middlewares → middleware"
        git mv middlewares middleware
    fi

    # 重命名 controllers → controller
    if [ -d "controllers" ]; then
        echo_info "重命名: controllers → controller"
        git mv controllers controller
    fi

    # 重命名 handlers → handler
    if [ -d "handlers" ]; then
        echo_info "重命名: handlers → handler"
        git mv handlers handler
    fi

    echo_info "目录重命名完成"
}

# 更新导入路径
update_imports() {
    echo_info "开始更新导入路径..."

    cd "$BACKEND_ROOT"

    # 更新 middlewares → middleware
    find . -type f -name "*.go" -exec sed -i '' 's|oneops/backend/middlewares|oneops/backend/middleware|g' {} +

    # 更新 controllers → controller
    find . -type f -name "*.go" -exec sed -i '' 's|oneops/backend/controllers|oneops/backend/controller|g' {} +

    # 更新 handlers → handler
    find . -type f -name "*.go" -exec sed -i '' 's|oneops/backend/handlers|oneops/backend/handler|g' {} +

    echo_info "导入路径更新完成"
}

# 更新包声明
update_package_declarations() {
    echo_info "开始更新包声明..."

    cd "$BACKEND_ROOT"

    # 更新 middleware 包中的包声明
    if [ -d "middleware" ]; then
        find middleware -type f -name "*.go" -exec sed -i '' 's|^package middlewares|package middleware|g' {} +
    fi

    # 更新 controller 包中的包声明
    if [ -d "controller" ]; then
        find controller -type f -name "*.go" -exec sed -i '' 's|^package controllers|package controller|g' {} +
    fi

    # 更新 handler 包中的包声明
    if [ -d "handler" ]; then
        find handler -type f -name "*.go" -exec sed -i '' 's|^package handlers|package handler|g' {} +
    fi

    echo_info "包声明更新完成"
}

# 验证修改
verify_changes() {
    echo_info "验证修改..."

    cd "$BACKEND_ROOT"

    echo_info "检查是否有遗漏的旧导入路径..."
    if grep -r "oneops/backend/middlewares" --include="*.go" . 2>/dev/null; then
        echo_error "发现未更新的导入: oneops/backend/middlewares"
        return 1
    fi

    if grep -r "oneops/backend/controllers" --include="*.go" . 2>/dev/null; then
        echo_error "发现未更新的导入: oneops/backend/controllers"
        return 1
    fi

    if grep -r "oneops/backend/handlers" --include="*.go" . 2>/dev/null; then
        echo_error "发现未更新的导入: oneops/backend/handlers"
        return 1
    fi

    echo_info "验证通过，所有导入路径已更新"
    return 0
}

# 运行测试
run_tests() {
    echo_info "运行测试..."

    cd "$BACKEND_ROOT"

    if [ -f "go.mod" ]; then
        go mod tidy
        if go test ./... -v; then
            echo_info "测试通过"
        else
            echo_error "测试失败，请检查代码"
            return 1
        fi
    else
        echo_warn "未找到 go.mod 文件，跳过测试"
    fi
}

# 提交更改
commit_changes() {
    echo_info "提交更改..."

    git add .
    git commit -m "refactor: 遵循 Golang 命名规范

- 重命名 middlewares → middleware
- 重命名 controllers → controller
- 重命名 handlers → handler
- 更新所有导入路径和包声明

参考规范: backend/docs/golang命名规范.txt"

    echo_info "更改已提交到分支: $BACKUP_BRANCH"
}

# 显示总结
show_summary() {
    echo ""
    echo_info "==================== 改造总结 ===================="
    echo_info "改造前目录：middlewares, controllers, handlers"
    echo_info "改造后目录：middleware, controller, handler"
    echo_info "当前分支：$BACKUP_BRANCH"
    echo ""
    echo_info "下一步操作："
    echo "  1. 检查代码: git diff main"
    echo "  2. 运行测试: go test ./..."
    echo "  3. 确认无误后合并: git checkout main && git merge $BACKUP_BRANCH"
    echo "  4. 或者回滚: git checkout main && git branch -D $BACKUP_BRANCH"
    echo_info "=================================================="
}

# 主函数
main() {
    echo_info "开始 Golang 命名规范改造..."
    echo ""

    cd "$PROJECT_ROOT"

    check_git_repo
    create_backup_branch
    rename_directories
    update_imports
    update_package_declarations
    verify_changes
    run_tests
    commit_changes
    show_summary

    echo_info "改造完成！"
}

# 执行主函数
main
