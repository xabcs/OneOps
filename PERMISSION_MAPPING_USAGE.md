# 权限映射系统使用指南

## 快速开始

### 1. 权限分配流程示例

#### 场景：为"运维组"分配 JumpServer 资产访问权限

```bash
# 1. 创建授权中心用户组（如果还没有）
POST /api/auth-groups
{
  "name": "运维组",
  "code": "ops-team",
  "description": "负责系统运维的团队"
}

# 2. 为运维组分配 JumpServer 权限
POST /api/auth-groups/1/permissions
{
  "appId": 1,                    # JumpServer 应用ID
  "mappingType": "asset",        # 分配资产权限
  "externalId": "all-assets",    # 资产标识
  "externalName": "所有资产",     # 显示名称
  "permissionDetail": {
    "actions": ["connect", "upload", "download"],
    "assets": ["*"],
    "protocols": ["ssh", "rdp"]
  }
}
```

#### 响应示例
```json
{
  "code": 200,
  "message": "权限分配成功",
  "data": {
    "success": true,
    "message": "权限分配成功",
    "mappingId": 1,
    "externalId": "all-assets"
  }
}
```

### 2. 查询用户组权限

```bash
# 获取用户组的所有权限映射
GET /api/auth-groups/1/permissions

# 响应示例
{
  "code": 200,
  "message": "获取成功",
  "data": [
    {
      "id": 1,
      "authGroupId": 1,
      "authGroupName": "运维组",
      "appId": 1,
      "appName": "生产环境 JumpServer",
      "appType": "jumpserver",
      "mappingType": "asset",
      "externalId": "all-assets",
      "externalName": "所有资产",
      "permissionDetail": {
        "actions": ["connect", "upload", "download"],
        "assets": ["*"],
        "protocols": ["ssh", "rdp"]
      },
      "isEnabled": true
    },
    {
      "id": 2,
      "authGroupId": 1,
      "authGroupName": "运维组",
      "appId": 2,
      "appName": "Jenkins CI",
      "appType": "jenkins",
      "mappingType": "role",
      "externalId": "developer-role",
      "externalName": "开发者角色",
      "permissionDetail": {
        "jobs": ["*"],
        "permissions": ["build", "read", "workspace"]
      },
      "isEnabled": true
    }
  ]
}
```

### 3. 查询用户有效权限

```bash
# 获取用户通过用户组继承的所有权限
GET /api/users/5/effective-permissions

# 响应示例
{
  "code": 200,
  "message": "获取成功",
  "data": [
    {
      "authGroupName": "运维组",
      "appName": "生产环境 JumpServer",
      "appType": "jumpserver",
      "mappingType": "asset",
      "externalName": "所有资产",
      "permissionDetail": {
        "actions": ["connect", "upload", "download"]
      },
      "isEnabled": true
    },
    {
      "authGroupName": "开发组",
      "appName": "Jenkins CI",
      "appType": "jenkins",
      "mappingType": "role",
      "externalName": "开发者角色",
      "permissionDetail": {
        "jobs": ["frontend-*"],
        "permissions": ["build", "read"]
      },
      "isEnabled": true
    }
  ]
}
```

### 4. 撤销权限

```bash
# 撤销特定的权限映射
DELETE /api/permission-mappings/1

# 响应示例
{
  "code": 200,
  "message": "撤销权限成功"
}
```

## 权限模板管理

### 1. 创建权限模板

```bash
# 为 JumpServer 创建常用权限模板
POST /api/permission-templates
{
  "appId": 1,
  "permissionCode": "full-admin",
  "permissionName": "完全管理员权限",
  "permissionType": "preset",
  "description": "拥有所有资产的所有操作权限",
  "template": {
    "mappingType": "asset",
    "actions": ["*"],
    "assets": ["*"],
    "protocols": ["*"]
  },
  "isSystemPreset": false
}
```

### 2. 使用模板快速分配权限

```bash
# 前端实现：先获取模板列表，然后用户选择模板进行分配
GET /api/applications/1/permission-templates

# 响应示例
{
  "code": 200,
  "message": "获取成功",
  "data": [
    {
      "id": 1,
      "permissionCode": "full-admin",
      "permissionName": "完全管理员权限",
      "permissionType": "preset",
      "description": "拥有所有资产的所有操作权限",
      "template": {
        "mappingType": "asset",
        "actions": ["*"],
        "assets": ["*"],
        "protocols": ["*"]
      }
    },
    {
      "id": 2,
      "permissionCode": "readonly-user",
      "permissionName": "只读用户",
      "permissionType": "preset",
      "description": "只能连接查看，不能上传下载",
      "template": {
        "mappingType": "asset",
        "actions": ["connect"],
        "assets": ["*"],
        "protocols": ["ssh", "rdp"]
      }
    }
  ]
}
```

## 不同应用类型的权限分配示例

### JumpServer 权限分配

#### 1. 资产权限
```json
POST /api/auth-groups/1/permissions
{
  "appId": 1,
  "mappingType": "asset",
  "externalId": "server-group-1",
  "externalName": "生产服务器组",
  "permissionDetail": {
    "actions": ["connect", "upload", "download"],
    "assets": ["server-1", "server-2", "server-3"],
    "protocols": ["ssh"]
  }
}
```

#### 2. 节点权限
```json
{
  "appId": 1,
  "mappingType": "node",
  "externalId": "node-org-1",
  "externalName": "北京机房节点",
  "permissionDetail": {
    "actions": ["connect"],
    "nodes": ["node-1"],
    "protocols": ["ssh", "rdp"]
  }
}
```

#### 3. 授权规则权限
```json
{
  "appId": 1,
  "mappingType": "rule",
  "externalId": "rule-123",
  "externalName": "运维权限规则",
  "permissionDetail": {
    "ruleId": "123",
    "priority": 50,
    "actions": ["connect", "upload"],
    "assets": ["asset-*"],
    "nodes": ["node-*"]
  }
}
```

### Jenkins 权限分配

#### 1. 角色权限
```json
POST /api/auth-groups/1/permissions
{
  "appId": 2,
  "mappingType": "role",
  "externalId": "developer-role",
  "externalName": "开发者角色",
  "permissionDetail": {
    "roleName": "developer",
    "permissions": ["build", "read", "workspace"],
    "jobs": ["frontend-*", "backend-*"]
  }
}
```

### GitLab 权限分配

#### 1. 项目权限
```json
POST /api/auth-groups/1/permissions
{
  "appId": 3,
  "mappingType": "project",
  "externalId": "project-123",
  "externalName": "前端项目",
  "permissionDetail": {
    "accessLevel": 30,  // Developer
    "permissions": ["push", "issue", "wiki"]
  }
}
```

## 权限同步与监控

### 1. 手动触发权限同步
```bash
POST /api/permission-mappings/sync
```

### 2. 定时任务配置（建议）
```go
// 每小时同步一次失败的权限映射
cron.New(cron.Option{
    "0 * * * *": "SyncAllPendingPermissions",
})
```

## 数据库迁移

### 创建权限映射相关表
```sql
-- 权限映射表
CREATE TABLE `auth_group_permission_mappings` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `auth_group_id` bigint unsigned NOT NULL COMMENT '授权中心用户组ID',
  `app_id` bigint unsigned NOT NULL COMMENT '外部应用ID',
  `mapping_type` varchar(20) NOT NULL COMMENT '权限对象类型',
  `external_id` varchar(100) DEFAULT NULL COMMENT '外部系统中的对象ID',
  `external_name` varchar(200) DEFAULT NULL COMMENT '外部系统中的对象名称',
  `permission_detail` json COMMENT '权限详情(JSON格式)',
  `is_enabled` tinyint(1) DEFAULT '1' COMMENT '是否启用',
  `priority` int DEFAULT '0' COMMENT '优先级',
  `granted_by` varchar(50) DEFAULT NULL COMMENT '授权人',
  `granted_at` datetime DEFAULT NULL COMMENT '授权时间',
  `expire_time` datetime DEFAULT NULL COMMENT '过期时间',
  `last_synced_at` datetime DEFAULT NULL COMMENT '最后同步时间',
  `sync_status` varchar(20) DEFAULT 'pending' COMMENT '同步状态',
  `sync_error_message` text COMMENT '同步错误信息',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_auth_app` (`auth_group_id`, `app_id`),
  KEY `idx_app` (`app_id`),
  CONSTRAINT `fk_mapping_group` FOREIGN KEY (`auth_group_id`) REFERENCES `auth_groups` (`id`),
  CONSTRAINT `fk_mapping_app` FOREIGN KEY (`app_id`) REFERENCES `applications` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='授权中心用户组权限映射表';

-- 权限模板表
CREATE TABLE `application_permission_templates` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `app_id` bigint unsigned NOT NULL COMMENT '外部应用ID',
  `permission_code` varchar(100) NOT NULL COMMENT '权限代码',
  `permission_name` varchar(100) NOT NULL COMMENT '权限名称',
  `permission_type` varchar(50) NOT NULL COMMENT '权限类型',
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='外部应用权限定义模板表';
```

## 前端集成建议

### 权限分配页面组件

```vue
<template>
  <div class="permission-assignment">
    <!-- 用户组选择 -->
    <el-select v-model="selectedAuthGroup" placeholder="选择用户组">
      <el-option
        v-for="group in authGroups"
        :key="group.id"
        :label="group.name"
        :value="group.id"
      />
    </el-select>

    <!-- 外部应用选择 -->
    <el-select v-model="selectedApp" placeholder="选择应用">
      <el-option
        v-for="app in applications"
        :key="app.id"
        :label="app.name"
        :value="app.id"
      />
    </el-select>

    <!-- 权限模板选择 -->
    <el-select v-model="selectedTemplate" placeholder="选择权限模板">
      <el-option
        v-for="template in templates"
        :key="template.id"
        :label="template.permissionName"
        :value="template.id"
      />
    </el-select>

    <!-- 权限详情展示 -->
    <div v-if="selectedTemplate" class="permission-detail">
      <h4>权限详情</h4>
      <json-viewer :value="getPermissionDetail()" />
    </div>

    <!-- 分配按钮 -->
    <el-button @click="assignPermission" type="primary">
      分配权限
    </el-button>
  </div>
</template>
```

## 常见问题解答

### Q1: 权限分配失败怎么办？
A: 检查以下几点：
1. 外部应用的认证配置是否正确
2. 外部应用是否支持API操作权限
3. externalId 是否在外部应用中存在
4. 查看同步错误信息获取详细失败原因

### Q2: 权限同步失败会自动重试吗？
A: 是的，系统会记录同步失败状态，可以通过定时任务或手动触发重新同步。

### Q3: 如何批量分配权限？
A: 建议使用权限模板功能，预先定义常用的权限组合，然后通过模板快速分配。

### Q4: 权限映射会过期吗？
A: 支持设置过期时间，过期后权限映射仍然保留但不会生效。可以手动更新或撤销。

### Q5: 如何审计权限变更？
A: 系统记录了授权人、授权时间、同步状态等信息，可以查询权限映射表获取完整的审计信息。

## 总结

这个权限映射系统的核心优势：

1. **统一管理**：在一个地方管理所有外部应用的权限
2. **灵活配置**：支持多种权限类型和自定义权限详情
3. **自动化**：自动同步权限到外部应用，减少手动操作
4. **可视化**：清晰展示用户组和权限的关系
5. **安全性**：权限变更有审计记录，支持过期时间控制

通过这个系统，你可以轻松实现"授权中心用户组 → 外部应用权限"的映射管理，大大简化了多系统权限管理的复杂度。
