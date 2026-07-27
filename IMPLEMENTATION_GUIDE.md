# 权限映射系统改造实施指南

## 实施路线图

```
第一阶段: 数据库基础 (7天)
  ├── 创建 user_identity_mappings 表
  ├── 创建 application_permission_templates 表  
  ├── 优化 auth_group_permission_mappings 表
  ├── 数据迁移脚本
  └── 单元测试

第二阶段: 核心服务 (14天)
  ├── 用户身份服务
  │   ├── CreateExternalIdentityForMember
  │   ├── GetUserExternalIdentities
  │   ├── SyncUserStatusToApps
  │   └── DeleteUserIdentity
  ├── 权限映射服务
  │   ├── AssignPermissionToAuthGroup
  │   ├── GetAuthGroupPermissions
  │   ├── GetUserEffectivePermissions
  │   └── ProcessPendingPermissions
  ├── 适配器接口扩展
  │   ├── JumpServer适配器
  │   ├── Jenkins适配器
  │   └── GitLab适配器
  └── 集成测试

第三阶段: API接口 (7天)
  ├── 用户管理API
  ├── 权限管理API
  ├── 权限模板API
  └── API文档

第四阶段: 前端改造 (14天)
  ├── 用户管理页面
  ├── 权限分配页面
  ├── 权限展示页面
  └── 用户体验优化

第五阶段: 集成测试 (7天)
  ├── 端到端场景测试
  ├── 性能测试
  ├── 安全测试
  └── 问题修复

第六阶段: 部署上线 (7天)
  ├── 数据库备份
  ├── 灰度发布
  ├── 监控配置
  ├── 用户培训
  └── 正式上线

总计: 约10周（70天）
```

## 第一阶段详细实施指南

### 1.1 数据库表创建

#### user_identity_mappings 表
```bash
# 在 backend/migrations/ 目录下创建文件
# 20240101_create_user_identity_mappings.sql

-- 用户身份映射表
CREATE TABLE `user_identity_mappings` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `auth_user_id` bigint unsigned NOT NULL COMMENT '授权中心用户ID',
  `app_id` bigint unsigned NOT NULL COMMENT '外部应用ID',
  `external_username` varchar(100) NOT NULL COMMENT '外部应用用户名',
  `external_user_id` varchar(100) DEFAULT NULL COMMENT '外部应用用户ID',
  `external_email` varchar(100) DEFAULT NULL COMMENT '外部应用邮箱',
  `mapping_type` varchar(20) DEFAULT 'auto' COMMENT '映射类型:auto,manual,linked',
  `mapping_status` varchar(20) DEFAULT 'active' COMMENT '映射状态:active,inactive,orphaned,pending',
  `sync_method` varchar(20) DEFAULT 'api' COMMENT '同步方法:api,import,manual',
  `last_synced_at` datetime DEFAULT NULL COMMENT '最后同步时间',
  `sync_status` varchar(20) DEFAULT 'success' COMMENT '同步状态:success,failed,pending',
  `sync_error_message` text COMMENT '同步错误信息',
  `created_by` varchar(50) DEFAULT NULL COMMENT '创建人',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_user_app` (`auth_user_id`, `app_id`),
  KEY `idx_app_status` (`app_id`, `mapping_status`),
  KEY `idx_external_user` (`app_id`, `external_username`),
  CONSTRAINT `fk_identity_user` FOREIGN KEY (`auth_user_id`) REFERENCES `auth_users` (`id`),
  CONSTRAINT `fk_identity_app` FOREIGN KEY (`app_id`) REFERENCES `applications` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='用户身份映射表';
```

#### application_permission_templates 表
```bash
# 20240102_create_permission_templates.sql

CREATE TABLE `application_permission_templates` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `app_id` bigint unsigned NOT NULL COMMENT '外部应用ID',
  `permission_code` varchar(100) NOT NULL COMMENT '权限代码',
  `permission_name` varchar(100) NOT NULL COMMENT '权限名称',
  `permission_type` varchar(50) NOT NULL COMMENT '权限类型:role,rule,preset',
  `description` varchar(200) DEFAULT NULL COMMENT '描述',
  `template` json COMMENT '权限内容模板(JSON)',
  `is_system_preset` tinyint(1) DEFAULT '0' COMMENT '是否系统预设',
  `sort_order` int DEFAULT '0' COMMENT '排序',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_app` (`app_id`),
  KEY `idx_code` (`permission_code`),
  CONSTRAINT `fk_template_app` FOREIGN KEY (`app_id`) REFERENCES `applications` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='外部应用权限定义模板表';
```

### 1.2 Go模型创建

#### models/user_identity_mapping.go
```go
package models

import (
	"time"
)

// UserIdentityMapping 用户身份映射表
type UserIdentityMapping struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	AuthUserID  uint       `json:"authUserId" gorm:"not null;uniqueIndex:idx_user_app;index:idx_app_status;comment:授权中心用户ID"`
	AuthUserIDField AuthUser `json:"authUser,omitempty" gorm:"foreignKey:AuthUserID"`
	AppID       uint       `json:"appId" gorm:"not null;uniqueIndex:idx_user_app;index:idx_app_status;index:idx_external_user;comment:外部应用ID"`
	AppIDField  Application `json:"-" gorm:"foreignKey:AppID"`

	// 外部身份信息
	ExternalUsername string `json:"externalUsername" gorm:"size:100;not null;comment:外部应用用户名"`
	ExternalUserID   string `json:"externalUserId" gorm:"size:100;comment:外部应用用户ID"`
	ExternalEmail    string `json:"externalEmail" gorm:"size:100;comment:外部应用邮箱"`

	// 映射信息
	MappingType   string `json:"mappingType" gorm:"size:20;default:auto;comment:映射类型:auto,manual,linked"`
	MappingStatus string `json:"mappingStatus" gorm:"size:20;default:active;index:idx_app_status;comment:映射状态:active,inactive,orphaned,pending"`
	SyncMethod    string `json:"syncMethod" gorm:"size:20;default:api;comment:同步方法:api,import,manual"`

	// 同步状态
	LastSyncedAt      *time.Time `json:"lastSyncedAt;comment:最后同步时间"`
	SyncStatus        string     `json:"syncStatus;gorm:size:20;default:success;comment:同步状态:success,failed,pending"`
	SyncErrorMessage  string     `json:"syncErrorMessage;gorm:type:text;comment:同步错误信息"`

	// 审计信息
	CreatedBy  string    `json:"createdBy;gorm:size:50;comment:创建人"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	DeletedAt  *time.Time `json:"deletedAt,gorm:"index""`
}

func (UserIdentityMapping) TableName() string {
	return "user_identity_mappings"
}
```

### 1.3 数据迁移脚本

#### 初始化系统预设模板
```bash
# backend/scripts/init_permission_templates.sql

-- JumpServer 预设模板
INSERT INTO `application_permission_templates` (`app_id`, `permission_code`, `permission_name`, `permission_type`, `description`, `template`, `is_system_preset`, `sort_order`) VALUES
-- JumpServer 预设模板
(1, 'jumpserver_full_admin', 'JumpServer 完全管理员', 'preset', '拥有所有资产的所有操作权限', '{"mapping_type":"asset","actions":["*"],"assets":["*"],"protocols":["*"]}', 1, 1),
(1, 'jumpserver_readonly', 'JumpServer 只读用户', 'preset', '只能连接查看，不能上传下载', '{"mapping_type":"asset","actions":["connect"],"assets":["*"],"protocols":["ssh","rdp"]}', 1, 2),
(1, 'jumpserver_asset_only', 'JumpServer 资产访问', 'preset', '可以访问资产但不能修改节点', '{"mapping_type":"asset","actions":["connect","upload","download"],"assets":["*"],"protocols":["ssh"]}', 1, 3),

-- Jenkins 预设模板
(2, 'jenkins_admin', 'Jenkins 管理员', 'role', '拥有Jenkins完全管理权限', '{"mapping_type":"role","role_name":"admin","permissions":["*"]}', 1, 10),
(2, 'jenkins_developer', 'Jenkins 开发者', 'role', '可以构建和查看Job', '{"mapping_type":"role","role_name":"developer","permissions":["build","read","workspace"]}', 1, 11),

-- GitLab 预设模板
(3, 'gitlab_owner', 'GitLab 项目所有者', 'role', '拥有项目完全控制权限', '{"mapping_type":"project","access_level":50,"permissions":["*"]}', 1, 20),
(3, 'gitlab_developer', 'GitLab 开发者', 'role', '可以推送代码和管理问题', '{"mapping_type":"project","access_level":30,"permissions":["push","issue"]}', 1, 21);
```

### 1.4 单元测试

#### 测试用户身份映射模型
```go
// models/user_identity_mapping_test.go
package models_test

import (
	"oneops/backend/models"
	"testing"
	"time"
)

func TestUserIdentityMapping(t *testing.T) {
	// 设置测试数据库
	SetupTestDB()
	defer TearDownTestDB()

	// 创建测试用户
	user := &models.AuthUser{
		Username: "testuser",
		Nickname: "测试用户",
		Email:    "test@example.com",
	}
	db.Create(user)

	// 创建测试应用
	app := &models.Application{
		Name:    "JumpServer",
		Code:    "jms",
		Type:    "jumpserver",
		BaseURL: "http://localhost:8080",
	}
	db.Create(app)

	// 创建身份映射
	mapping := &models.UserIdentityMapping{
		AuthUserID:       user.ID,
		AppID:            app.ID,
		ExternalUsername: "testuser",
		MappingType:      "auto",
		MappingStatus:    "active",
		CreatedBy:        "admin",
	}

	// 测试创建
	err := db.Create(mapping).Error
	if err != nil {
		t.Errorf("创建身份映射失败: %v", err)
	}

	// 测试查询
	var result models.UserIdentityMapping
	err = db.Where("auth_user_id = ? AND app_id = ?", user.ID, app.ID).First(&result).Error
	if err != nil {
		t.Errorf("查询身份映射失败: %v", err)
	}

	// 验证结果
	if result.ExternalUsername != "testuser" {
		t.Errorf("外部用户名不正确，期望 testuser，实际 %s", result.ExternalUsername)
	}

	if result.MappingStatus != "active" {
		t.Errorf("映射状态不正确，期望 active，实际 %s", result.MappingStatus)
	}

	// 测试唯一约束
	duplicate := &models.UserIdentityMapping{
		AuthUserID:       user.ID,
		AppID:            app.ID,
		ExternalUsername: "testuser",
	}
	err = db.Create(duplicate).Error
	if err == nil {
		t.Errorf("应该阻止重复的身份映射")
	}
}
```

## 开发环境设置

### 1. 本地开发数据库
```bash
# 创建本地测试数据库
mysql -u root -p -e "CREATE DATABASE oneops_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 导入测试数据
mysql -u root -p oneops_test < backend/scripts/init_permission_templates.sql
```

### 2. 代码组织
```
backend/
├── models/
│   ├── user_identity_mapping.go          # 新增
│   ├── permission_mapping.go             # 优化
│   └── application_permission_template.go # 新增
├── services/
│   ├── user_identity_service.go           # 新增
│   ├── permission_mapping_service.go      # 新增
│   └── application_adapter.go            # 扩展
├── controllers/
│   ├── user_identity_controller.go        # 新增
│   └── permission_mapping_controller.go   # 新增
└── migrations/
    ├── 20240101_create_user_identity_mappings.sql
    ├── 20240102_create_permission_templates.sql
    └── 20240103_init_system_templates.sql
```

## 测试检查清单

### 第一阶段验收检查清单
```markdown
- [ ] user_identity_mappings 表创建成功
- [ ] application_permission_templates 表创建成功
- [ ] auth_group_permission_mappings 表优化完成
- [ ] 数据迁移脚本执行无错误
- [ ] 系统预设模板导入成功
- [ ] Go模型文件编译通过
- [ ] 单元测试通过
- [ ] 数据库约束测试通过
- [ ] 索引性能测试通过
- [ ] 数据备份和恢复测试通过
```

## 常见问题处理

### Q1: 数据库表创建失败
**问题**：外键约束创建失败
**解决**：
```sql
-- 检查关联表是否存在
SHOW TABLES LIKE 'auth_users';
SHOW TABLES LIKE 'applications';

-- 检查字段是否存在
DESCRIBE auth_users;
DESCRIBE applications;
```

### Q2: JSON字段测试失败
**问题**：JSON字段操作不支持
**解决**：
```bash
# 检查MySQL版本
SELECT VERSION();

-- 需要 MySQL 5.7+ 支持 JSON 类型
```

### Q3: 迁移脚本执行错误
**问题**：INSERT语句失败
**解决**：
```sql
-- 检查applications表中是否存在对应的应用
SELECT id, name, type FROM applications;

-- 如果不存在，先插入测试数据
INSERT INTO applications (name, code, type, base_url) VALUES
('JumpServer', 'jms', 'jumpserver', 'http://localhost:8080'),
('Jenkins', 'jk', 'jenkins', 'http://localhost:8080'),
('GitLab', 'gl', 'gitlab', 'http://localhost:8080');
```

## 开发工具配置

### 1. 数据库迁移工具
```bash
# 安装 golang-migrate
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/

# 创建迁移文件
migrate create -ext sql -dir backend/migrations -seq create_user_identity_mappings
```

### 2. API测试工具
```bash
# 安装 Postman 新 Collection
# 导入 API/PermissionMapping.postman_collection.json
```

### 3. 数据库监控
```bash
# 安装 MySQL 监控工具
docker run -d --name mysql-admin -p 8080:80 \
  -e MYSQL_ROOT_PASSWORD=root \
  phpmyadmin/phpmyadmin
```

## 下一步行动

1. **立即开始**：创建数据库迁移脚本
2. **准备测试**：设置本地测试环境
3. **代码准备**：创建Go模型文件
4. **文档更新**：更新API文档
5. **团队培训**：组织技术方案讲解

按照这个实施指南，可以有序地推进权限映射系统的改造工作！
