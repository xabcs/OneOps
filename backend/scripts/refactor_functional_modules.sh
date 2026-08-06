#!/bin/bash

# OneOps 后端目录结构改造脚本（修正版）
# 基于功能模块分层方案_修正版.md
# 实现按业务模块的DDD分层架构

set -e

PROJECT_ROOT="/Users/mpm/Desktop/doc/go/OneOpsV2/OneOps"
BACKUP_BRANCH="refactor/functional-module-structure"

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
                    OneOps 后端功能模块分层改造方案（修正版）
                        基于 DDD + Clean Architecture
╔──────────────────────────────────────────────────────────────────────────────┘

核心模块划分:
  1. CMDB 模块（含堡垒机） - 资产管理 + SSH连接
  2. K8s 模块 - Kubernetes 管理
  3. Monitoring 模块 - 监控告警
  4. Audit 模块 - 审计日志
  5. System 模块（IAM） - 用户、角色、菜单、权限
  6. Authorization 模块（授权中心） - 外部应用权限

改造后结构:
  backend/
  ├── cmd/
  │   └── api/
  │       └── main.go              # 应用入口
  ├── internal/
  │   ├── cmdb/                    # CMDB 模块（含堡垒机）
  │   │   ├── domain/
  │   │   ├── repository/
  │   │   ├── service/
  │   │   ├── handler/
  │   │   └── routes.go
  │   ├── k8s/                     # K8s 模块
  │   ├── monitoring/              # 监控模块
  │   ├── audit/                   # 审计模块
  │   ├── system/                  # 系统管理（IAM）
  │   ├── authorization/           # 授权中心（外部应用）
  │   ├── auth/                    # 认证服务（JWT）
  │   ├── diagnostic/              # 诊断工具
  │   ├── notification/            # 通知服务
  │   └── pkg/                     # 共享库
  │       ├── middleware/
  │       ├── database/
  │       ├── cache/
  │       ├── auth/
  │       ├── crypto/
  │       ├── http/
  │       ├── validator/
  │       ├── logger/
  │       ├── errors/
  │       ├── config/
  │       ├── dto/
  │       └── container/
  ├── configs/                     # 配置文件
  ├── migrations/                  # 数据库迁移
  ├── scripts/                     # 脚本工具
  └── docs/                        # 文档

关键改进:
  ✓ CMDB + Bastion 合并为同一模块
  ✓ System = IAM（用户、角色、菜单、权限）
  ✓ Authorization 独立处理外部应用权限
  ✓ 按业务功能分层，而非技术分层
  ✓ 遵循 DDD 领域驱动设计

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

# 第一阶段：创建基础目录结构
phase_1_create_directories() {
    echo_step "第一阶段：创建基础目录结构"

    cd "$PROJECT_ROOT/backend"

    # 1. 创建 cmd 目录
    echo_info "创建 cmd/api 目录"
    mkdir -p cmd/api

    # 2. 移动 main.go
    if [ -f "main.go" ]; then
        echo_info "移动 main.go 到 cmd/api/"
        git mv main.go cmd/api/main.go
    else
        echo_warn "main.go 不在根目录，可能已移动"
    fi

    # 3. 创建 internal 目录
    echo_info "创建 internal 模块目录"
    mkdir -p internal/{cmdb,k8s,monitoring,audit,system,authorization,auth,diagnostic,notification}
    mkdir -p internal/pkg/{middleware,database,cache,auth,crypto,http,validator,logger,errors,config,dto,container}

    # 4. 创建 configs 目录
    mkdir -p configs

    echo_info "第一阶段完成"
}

# 第二阶段：迁移 CMDB 模块（含堡垒机）
phase_2_migrate_cmdb() {
    echo_step "第二阶段：迁移 CMDB 模块（含堡垒机）"

    cd "$PROJECT_ROOT/backend"

    echo_info "创建 CMDB 模块目录结构"
    mkdir -p internal/cmdb/{domain,repository,service,handler}

    echo_info "迁移 CMDB 相关文件"

    # Models -> Domain
    if [ -f "models/cmdb.go" ]; then
        git mv models/cmdb.go internal/cmdb/domain/server.go
    fi
    if [ -f "models/bastion.go" ]; then
        git mv models/bastion.go internal/cmdb/domain/session.go
    fi
    if [ -f "models/permission_mapping.go" ]; then
        git mv models/permission_mapping.go internal/cmdb/domain/policy.go
    fi

    # Controllers -> Handlers
    if [ -f "controller/cmdb.go" ]; then
        git mv controller/cmdb.go internal/cmdb/handler/server_handler.go
    fi
    if [ -f "controller/bastion.go" ]; then
        git mv controller/bastion.go internal/cmdb/handler/session_handler.go
    fi

    # Services
    if [ -f "services/cmdb.go" ]; then
        git mv services/cmdb.go internal/cmdb/service/server_service.go
    fi
    if [ -f "services/bastion.go" ]; then
        git mv services/bastion.go internal/cmdb/service/bastion_service.go
    fi

    # Routes
    if [ -f "routes/cmdb_routes.go" ]; then
        git mv routes/cmdb_routes.go internal/cmdb/routes.go
    fi

    echo_info "CMDB 模块迁移完成"
}

# 第三阶段：迁移 System 模块（IAM）
phase_3_migrate_system() {
    echo_step "第三阶段：迁移 System 模块（IAM）"

    cd "$PROJECT_ROOT/backend"

    echo_info "创建 System 模块目录结构"
    mkdir -p internal/system/{domain,repository,service,handler}

    echo_info "迁移 System 相关文件"

    # Models -> Domain
    if [ -f "models/user.go" ]; then
        git mv models/user.go internal/system/domain/user.go
    fi
    if [ -f "models/role.go" ]; then
        git mv models/role.go internal/system/domain/role.go
    fi
    if [ -f "models/menu.go" ]; then
        git mv models/menu.go internal/system/domain/menu.go
    fi

    # Controllers -> Handlers
    if [ -f "controller/user.go" ]; then
        git mv controller/user.go internal/system/handler/user_handler.go
    fi
    if [ -f "controller/role.go" ]; then
        git mv controller/role.go internal/system/handler/role_handler.go
    fi
    if [ -f "controller/menu.go" ]; then
        git mv controller/menu.go internal/system/handler/menu_handler.go
    fi
    if [ -f "controller/auth.go" ]; then
        git mv controller/auth.go internal/system/handler/auth_handler.go
    fi

    # Services
    if [ -f "services/auth.go" ]; then
        git mv services/auth.go internal/system/service/auth_service.go
    fi
    if [ -f "services/rbac.go" ]; then
        git mv services/rbac.go internal/system/service/rbac_service.go
    fi

    # Routes
    if [ -f "routes/system_routes.go" ]; then
        git mv routes/system_routes.go internal/system/routes.go
    fi

    echo_info "System 模块迁移完成"
}

# 第四阶段：迁移 Authorization 模块
phase_4_migrate_authorization() {
    echo_step "第四阶段：迁移 Authorization 模块（授权中心）"

    cd "$PROJECT_ROOT/backend"

    echo_info "创建 Authorization 模块目录结构"
    mkdir -p internal/authorization/{domain,repository,service,handler}

    echo_info "迁移 Authorization 相关文件"

    # Controllers -> Handlers
    if [ -f "controller/application_permission.go" ]; then
        git mv controller/application_permission.go internal/authorization/handler/app_handler.go
    fi

    # Services
    if [ -f "services/application_permission_service.go" ]; then
        git mv services/application_permission_service.go internal/authorization/service/application_service.go
    fi

    # Adapters
    mkdir -p internal/authorization/service/adapter
    for file in services/jumpserver_adapter*.go; do
        if [ -f "$file" ]; then
            git mv "$file" internal/authorization/service/adapter/
        fi
    done

    # Models -> Domain
    for file in models/application_permission*.go; do
        if [ -f "$file" ]; then
            git mv "$file" internal/authorization/domain/
        fi
    done

    # Routes (从 auth_routes.go 提取授权部分)
    if [ -f "routes/auth_routes.go" ]; then
        git mv routes/auth_routes.go internal/authorization/routes.go
    fi

    echo_info "Authorization 模块迁移完成"
}

# 第五阶段：迁移其他模块
phase_5_migrate_other_modules() {
    echo_step "第五阶段：迁移其他模块"

    cd "$PROJECT_ROOT/backend"

    # K8s 模块
    echo_info "迁移 K8s 模块"
    mkdir -p internal/k8s/{domain,repository,service,handler}
    if [ -f "controller/k8s_cluster.go" ]; then
        git mv controller/k8s_cluster.go internal/k8s/handler/cluster_handler.go
    fi
    if [ -f "controller/k8s_resource.go" ]; then
        git mv controller/k8s_resource.go internal/k8s/handler/resource_handler.go
    fi
    if [ -f "routes/k8s_routes.go" ]; then
        git mv routes/k8s_routes.go internal/k8s/routes.go
    fi

    # Monitoring 模块
    echo_info "迁移 Monitoring 模块"
    mkdir -p internal/monitoring/{domain,repository,service,handler}
    if [ -f "controller/monitoring.go" ]; then
        git mv controller/monitoring.go internal/monitoring/handler/metric_handler.go
    fi
    if [ -f "routes/monitoring_routes.go" ]; then
        git mv routes/monitoring_routes.go internal/monitoring/routes.go
    fi

    # Audit 模块
    echo_info "迁移 Audit 模块"
    mkdir -p internal/audit/{domain,repository,service,handler}
    if [ -f "controller/audit.go" ]; then
        git mv controller/audit.go internal/audit/handler/audit_handler.go
    fi
    if [ -f "routes/audit_routes.go" ]; then
        git mv routes/audit_routes.go internal/audit/routes.go
    fi

    # Diagnostic 模块
    echo_info "迁移 Diagnostic 模块"
    mkdir -p internal/diagnostic/{service,handler}
    if [ -f "controller/diagnostic.go" ]; then
        git mv controller/diagnostic.go internal/diagnostic/handler/diagnostic_handler.go
    fi

    echo_info "其他模块迁移完成"
}

# 第六阶段：迁移共享库到 internal/pkg
phase_6_migrate_shared_libs() {
    echo_step "第六阶段：迁移共享库到 internal/pkg"

    cd "$PROJECT_ROOT/backend"

    echo_info "迁移支撑库到 internal/pkg/"

    # 移动现有的支撑库
    for dir in middleware logger dto config container errors; do
        if [ -d "$dir" ]; then
            echo_info "  移动 $dir/"
            git mv "$dir" internal/pkg/
        fi
    done

    # 重命名 utils 为 common
    if [ -d "utils" ]; then
        echo_info "重命名 utils/ 为 common/"
        # 创建 common 子目录
        mkdir -p internal/pkg/{auth,crypto,http,security,ua}

        # 移动文件到对应子目录
        if [ -f "utils/jwt.go" ]; then
            git mv utils/jwt.go internal/pkg/auth/jwt.go
        fi
        if [ -f "utils/password.go" ]; then
            git mv utils/password.go internal/pkg/auth/password.go
        fi
        if [ -f "utils/encryption.go" ]; then
            git mv utils/encryption.go internal/pkg/crypto/encryption.go
        fi
        if [ -f "utils/security.go" ]; then
            git mv utils/security.go internal/pkg/security/security.go
        fi
        if [ -f "utils/pagination.go" ]; then
            git mv utils/pagination.go internal/pkg/http/pagination.go
        fi
        if [ -f "utils/response.go" ]; then
            git mv utils/response.go internal/pkg/http/response.go
        fi
        if [ -f "utils/user_agent.go" ]; then
            git mv utils/user_agent.go internal/pkg/ua/user_agent.go
        fi

        # 移动剩余文件到 common
        find utils -maxdepth 1 -type f -name "*.go" -exec git mv {} internal/pkg/common/ \; 2>/dev/null || true

        # 删除空目录
        rmdir utils 2>/dev/null || true
    fi

    # 移动 handler -> internal/pkg/handler
    if [ -d "handler" ]; then
        git mv handler internal/pkg/handler
    fi

    echo_info "共享库迁移完成"
}

# 第七阶段：统一配置文件
phase_7_unify_configs() {
    echo_step "第七阶段：统一配置文件"

    cd "$PROJECT_ROOT/backend"

    # 移动配置文件到 configs/
    if [ -d "internal/pkg/config" ]; then
        echo_info "移动配置文件到 configs/"
        find internal/pkg/config -maxdepth 1 -type f \( -name "*.yaml" -o -name "*.yml" -o -name "*.conf" -o -name "*.sql" \) -exec git mv {} configs/ \; 2>/dev/null || true
    fi

    echo_info "配置文件统一完成"
}

# 第八阶段：更新导入路径
phase_8_update_imports() {
    echo_step "第八阶段：更新导入路径"

    cd "$PROJECT_ROOT/backend"

    echo_info "更新导入路径..."

    # cmd/api/main.go
    find cmd -name "*.go" -exec sed -i '' 's|oneops/backend/|oneops/backend/internal/|g' {} +

    # CMDB 模块
    find internal/cmdb -name "*.go" -exec sed -i '' 's|oneops/backend/controller|oneops/backend/internal/cmdb/handler|g' {} +
    find internal/cmdb -name "*.go" -exec sed -i '' 's|oneops/backend/services|oneops/backend/internal/cmdb/service|g' {} +
    find internal/cmdb -name "*.go" -exec sed -i '' 's|oneops/backend/models|oneops/backend/internal/cmdb/domain|g' {} +

    # System 模块
    find internal/system -name "*.go" -exec sed -i '' 's|oneops/backend/controller|oneops/backend/internal/system/handler|g' {} +
    find internal/system -name "*.go" -exec sed -i '' 's|oneops/backend/services|oneops/backend/internal/system/service|g' {} +
    find internal/system -name "*.go" -exec sed -i '' 's|oneops/backend/models|oneops/backend/internal/system/domain|g' {} +

    # Authorization 模块
    find internal/authorization -name "*.go" -exec sed -i '' 's|oneops/backend/controller|oneops/backend/internal/authorization/handler|g' {} +
    find internal/authorization -name "*.go" -exec sed -i '' 's|oneops/backend/services|oneops/backend/internal/authorization/service|g' {} +
    find internal/authorization -name "*.go" -exec sed -i '' 's|oneops/backend/models|oneops/backend/internal/authorization/domain|g' {} +

    # K8s 模块
    find internal/k8s -name "*.go" -exec sed -i '' 's|oneops/backend/controller|oneops/backend/internal/k8s/handler|g' {} +
    find internal/k8s -name "*.go" -exec sed -i '' 's|oneops/backend/routes|oneops/backend/internal/k8s|g' {} +

    # Monitoring 模块
    find internal/monitoring -name "*.go" -exec sed -i '' 's|oneops/backend/controller|oneops/backend/internal/monitoring/handler|g' {} +

    # Audit 模块
    find internal/audit -name "*.go" -exec sed -i '' 's|oneops/backend/controller|oneops/backend/internal/audit/handler|g' {} +

    # Diagnostic 模块
    find internal/diagnostic -name "*.go" -exec sed -i '' 's|oneops/backend/controller|oneops/backend/internal/diagnostic/handler|g' {} +

    # internal/pkg
    find internal/pkg -name "*.go" -exec sed -i '' 's|oneops/backend/middleware|oneops/backend/internal/pkg/middleware|g' {} +
    find internal/pkg -name "*.go" -exec sed -i '' 's|oneops/backend/logger|oneops/backend/internal/pkg/logger|g' {} +
    find internal/pkg -name "*.go" -exec sed -i '' 's|oneops/backend/dto|oneops/backend/internal/pkg/dto|g' {} +
    find internal/pkg -name "*.go" -exec sed -i '' 's|oneops/backend/errors|oneops/backend/internal/pkg/errors|g' {} +
    find internal/pkg -name "*.go" -exec sed -i '' 's|oneops/backend/handler|oneops/backend/internal/pkg/handler|g' {} +

    echo_info "导入路径更新完成"
}

# 第九阶段：验证编译
phase_9_verify() {
    echo_step "第九阶段：验证编译"

    cd "$PROJECT_ROOT/backend"

    echo_info "尝试编译..."
    if go build -o /tmp/oneops-test ./cmd/api 2>&1; then
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

    git commit -m "refactor: 按功能模块分层重构目录结构（修正版）

基于功能模块分层方案_修正版.md实现：
- CMDB 模块（含堡垒机）：资产管理 + SSH连接
- System 模块（IAM）：用户、角色、菜单、权限管理
- Authorization 模块（授权中心）：外部应用权限
- K8s 模块：Kubernetes 管理
- Monitoring 模块：监控告警
- Audit 模块：审计日志
- 添加 cmd/api/ 目录，移动 main.go
- 添加 internal/ 目录保护私有代码
- 拆分 utils/ 为具体功能包
- 统一配置文件到 configs/

关键改进：
- CMDB + Bastion 合并为同一模块
- System = IAM（用户、角色、菜单、权限）
- Authorization 独立处理外部应用权限
- 按业务功能分层，而非技术分层
- 遵循 DDD 领域驱动设计

参考标准：golang-standards/project-layout + DDD"

    echo_info "更改已提交到分支: $BACKUP_BRANCH"
}

# 显示总结
show_summary() {
    cat << 'EOF'
╔──────────────────────────────────────────────────────────────────────────────┘
                              改造完成总结
╔──────────────────────────────────────────────────────────────────────────────┘

✅ 模块划分:
  1. CMDB 模块（含堡垒机） - internal/cmdb/
  2. System 模块（IAM） - internal/system/
  3. Authorization 模块 - internal/authorization/
  4. K8s 模块 - internal/k8s/
  5. Monitoring 模块 - internal/monitoring/
  6. Audit 模块 - internal/audit/
  7. Diagnostic 模块 - internal/diagnostic/
  8. Notification 模块 - internal/notification/
  9. Auth 模块 - internal/auth/
  10. 共享库 - internal/pkg/

📋 下一步操作:
  1. 检查代码: git diff main
  2. 编译测试: go build ./cmd/api
  3. 运行测试: go test ./...
  4. 检查路由: 确保 API 路由正确
  5. 确认无误后合并

🔄 回滚方式:
  git checkout main
  git branch -D $BACKUP_BRANCH

📚 相关文档:
  - backend/docs/功能模块分层方案_修正版.md
  - backend/docs/目录结构对比分析.md
  - backend/docs/目录结构总结.md

EOF
}

# 主函数
main() {
    echo_info "OneOps 后端功能模块分层改造（修正版）"
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
    phase_1_create_directories
    phase_2_migrate_cmdb
    phase_3_migrate_system
    phase_4_migrate_authorization
    phase_5_migrate_other_modules
    phase_6_migrate_shared_libs
    phase_7_unify_configs
    phase_8_update_imports
    phase_9_verify
    commit_changes
    show_summary

    echo_info "改造完成！"
}

# 执行主函数
main
