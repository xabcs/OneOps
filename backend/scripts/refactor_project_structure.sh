#!/bin/bash

# OneOps 后端目录结构改造脚本（方案 A：最小改动）
# 遵循 golang-standards/project-layout 标准

set -e

PROJECT_ROOT="/Users/mpm/Desktop/doc/go/OneOpsV2/OneOps"
BACKUP_BRANCH="refactor/golang-project-structure"

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

echo_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

echo_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

echo_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

echo_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# 显示改造方案
show_plan() {
    cat << 'EOF'
╔──────────────────────────────────────────────────────────────────────────────┘
                        OneOps 后端目录结构改造方案
                         遵循 golang-standards/project-layout
╔──────────────────────────────────────────────────────────────────────────────┘

当前结构:
  backend/
  ├── main.go
  ├── controller/
  ├── services/
  ├── models/
  ├── middleware/
  └── ...

改造后结构:
  backend/
  ├── cmd/
  │   └── server/
  │       └── main.go          # 应用入口
  ├── internal/
  │   ├── app/                 # 应用层
  │   │   ├── controller/
  │   │   ├── services/
  │   │   ├── models/
  │   │   ├── routes/
  │   │   └── handler/
  │   └── pkg/                 # 内部共享库
  │       ├── middleware/
  │       ├── logger/
  │       ├── utils/          # 重命名为 common
  │       ├── dto/
  │       ├── config/
  │       ├── container/
  │       └── errors/
  ├── configs/                  # 配置文件
  │   ├── config.yaml
  │   ├── config.yaml.example
  │   └── ...
  ├── migrations/              # 数据库迁移
  ├── scripts/                  # 脚本工具
  ├── docs/                     # 文档
  └── (其他文件)

改动说明:
  ✓ 添加 cmd/ 目录，移动 main.go
  ✓ 添加 internal/ 目录保护私有代码
  ✓ 重命名 utils/ 为 common/
  ✓ 统一配置文件位置到 configs/
  ✓ 更新所有导入路径

EOF
}

# 创建备份分支
create_backup_branch() {
    echo_step "创建备份分支"

    cd "$PROJECT_ROOT"

    if git show-ref --verify --quiet refs/heads/$BACKUP_BRANCH; then
        echo_warn "分支 $BACKUP_BRANCH 已存在"
        read -p "是否删除并重新创建？(y/n): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            git branch -D $BACKUP_BRANCH
            git checkout -b $BACKUP_BRANCH
        else
            echo_info "切换到现有分支"
            git checkout $BACKUP_BRANCH
        fi
    else
        git checkout -b $BACKUP_BRANCH
    fi

    echo_info "备份分支创建成功"
}

# 第一阶段：添加 cmd 和 internal 目录
phase_1_add_directories() {
    echo_step "第一阶段：添加 cmd 和 internal 目录"

    cd "$PROJECT_ROOT/backend"

    # 1. 创建 cmd/server 目录
    echo_info "创建 cmd/server 目录"
    mkdir -p cmd/server

    # 2. 移动 main.go
    if [ -f "main.go" ]; then
        echo_info "移动 main.go 到 cmd/server/"
        git mv main.go cmd/server/main.go
    else
        echo_warn "main.go 不在根目录，可能已移动"
    fi

    # 3. 创建 internal 目录结构
    echo_info "创建 internal 目录结构"
    mkdir -p internal/app
    mkdir -p internal/pkg

    # 4. 移动应用层代码到 internal/app
    echo_info "移动应用层代码到 internal/app/"
    for dir in controller services models routes handler; do
        if [ -d "$dir" ]; then
            echo_info "  移动 $dir/"
            git mv "$dir" internal/app/
        fi
    done

    echo_info "第一阶段完成"
}

# 第二阶段：重组内部库
phase_2_reorganize_libs() {
    echo_step "第二阶段：重组内部库"

    cd "$PROJECT_ROOT/backend"

    # 1. 移动支撑库到 internal/pkg
    echo_info "移动支撑库到 internal/pkg/"
    for dir in middleware logger dto config container errors; do
        if [ -d "$dir" ]; then
            echo_info "  移动 $dir/"
            git mv "$dir" internal/pkg/
        fi
    done

    # 2. 重命名 utils 为 common
    if [ -d "utils" ]; then
        echo_info "重命名 utils/ 为 common/"
        git mv utils common
        git mv common internal/pkg/common
    fi

    echo_info "第二阶段完成"
}

# 第三阶段：统一配置文件
phase_3_unify_configs() {
    echo_step "第三阶段：统一配置文件"

    cd "$PROJECT_ROOT/backend"

    # 创建 configs 目录
    mkdir -p configs

    # 移动配置文件
    if [ -d "internal/pkg/config" ]; then
        echo_info "移动配置文件到 configs/"
        find internal/pkg/config -maxdepth 1 -type f \( -name "*.yaml" -o -name "*.yml" -o -name "*.conf" -o -name "*.sql" \) -exec git mv {} configs/ \; 2>/dev/null || true
    fi

    echo_info "第三阶段完成"
}

# 第四阶段：更新导入路径
phase_4_update_imports() {
    echo_step "第四阶段：更新导入路径"

    cd "$PROJECT_ROOT/backend"

    # 更新导入路径
    echo_info "更新导入路径..."

    # main.go 中的导入
    find cmd -name "*.go" -exec sed -i '' 's|oneops/backend/|oneops/backend/internal/|g' {} +

    # internal/app 中的导入
    find internal/app -name "*.go" -exec sed -i '' 's|oneops/backend/controller|oneops/backend/internal/app/controller|g' {} +
    find internal/app -name "*.go" -exec sed -i '' 's|oneops/backend/services|oneops/backend/internal/app/services|g' {} +
    find internal/app -name "*.go" -exec sed -i '' 's|oneops/backend/models|oneops/backend/internal/app/models|g' {} +
    find internal/app -name "*.go" -exec sed -i '' 's|oneops/backend/routes|oneops/backend/internal/app/routes|g' {} +
    find internal/app -name "*.go" -exec sed -i '' 's|oneops/backend/handler|oneops/backend/internal/app/handler|g' {} +

    # internal/pkg 中的导入
    find internal/pkg -name "*.go" -exec sed -i '' 's|oneops/backend/middleware|oneops/backend/internal/pkg/middleware|g' {} +
    find internal/pkg -name "*.go" -exec sed -i '' 's|oneops/backend/logger|oneops/backend/internal/pkg/logger|g' {} +
    find internal/pkg -name "*.go" -exec sed -i '' 's|oneops/backend/utils|oneops/backend/internal/pkg/common|g' {} +
    find internal/pkg -name "*.go" -exec sed -i '' 's|oneops/backend/dto|oneops/backend/internal/pkg/dto|g' {} +
    find internal/pkg -name "*.go" -exec sed -i '' 's|oneops/backend/errors|oneops/backend/internal/pkg/errors|g' {} +

    echo_info "第四阶段完成"
}

# 第五阶段：验证编译
phase_5_verify() {
    echo_step "第五阶段：验证编译"

    cd "$PROJECT_ROOT/backend"

    echo_info "尝试编译..."
    if go build -o /tmp/oneops-test ./cmd/server 2>&1; then
        echo_info "✅ 编译成功"
        rm -f /tmp/oneops-test
        return 0
    else
        echo_error "❌ 编译失败，请检查错误信息"
        return 1
    fi
}

# 提交更改
commit_changes() {
    echo_step "提交更改"

    cd "$PROJECT_ROOT"

    git add backend

    git commit -m "refactor: 遵循 golang-standards/project-layout 标准

目录结构改造：
- 添加 cmd/ 目录，移动 main.go
- 添加 internal/ 目录保护私有代码
- 重命名 utils/ 为 common/
- 统一配置文件到 configs/
- 更新所有导入路径

改造方案：最小改动（方案 A）
参考标准：golang-standards/project-layout

影响范围：所有后端代码"

    echo_info "更改已提交到分支: $BACKUP_BRANCH"
}

# 显示总结
show_summary() {
    cat << 'EOF'
╔──────────────────────────────────────────────────────────────────────────────┘
                              改造完成总结
╔──────────────────────────────────────────────────────────────────────────────┘

✅ 改造内容:
  1. 添加 cmd/server/ 目录，移动 main.go
  2. 添加 internal/ 目录，保护私有代码
  3. 重命名 utils/ 为 common/
  4. 统一配置文件到 configs/
  5. 更新所有导入路径

📋 下一步操作:
  1. 检查代码: git diff main
  2. 编译测试: go build ./cmd/server
  3. 运行测试: go test ./...
  4. 确认无误后合并

🔄 回滚方式:
  git checkout main
  git branch -D $BACKUP_BRANCH

📚 相关文档:
  - backend/docs/目录结构对比分析.md
  - backend/docs/目录结构总结.md

EOF
}

# 主函数
main() {
    echo_info "OneOps 后端目录结构改造"
    echo ""

    show_plan

    read -p "是否继续执行改造？(y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo_info "已取消改造"
        exit 0
    fi

    cd "$PROJECT_ROOT"

    create_backup_branch
    phase_1_add_directories
    phase_2_reorganize_libs
    phase_3_unify_configs
    phase_4_update_imports
    phase_5_verify
    commit_changes
    show_summary

    echo_info "改造完成！"
}

# 执行主函数
main
