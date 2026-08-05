-- ======================================
-- OneOps 按钮级权限系统 - 数据库迁移脚本
-- 版本: 1.0.0
-- 创建日期: 2026-08-03
-- 描述: 从菜单级权限扩展到按钮级权限
-- ======================================

-- 备份提醒: 在执行此脚本前，请备份现有数据库!

-- ======================================
-- 1. 权限定义表
-- ======================================
CREATE TABLE IF NOT EXISTS `permissions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '权限ID',
  `code` varchar(100) NOT NULL COMMENT '权限编码，格式：module.resource.action',
  `name` varchar(50) NOT NULL COMMENT '权限名称',
  `description` varchar(200) DEFAULT NULL COMMENT '权限描述',
  `module` varchar(30) NOT NULL COMMENT '模块标识（system, monitor, cmdb, bastion等）',
  `resource` varchar(30) NOT NULL COMMENT '资源标识（user, role, server等）',
  `action` varchar(20) NOT NULL COMMENT '操作类型（view, create, update, delete, execute, export等）',
  `level` varchar(20) NOT NULL DEFAULT 'button' COMMENT '权限级别：menu(菜单权限), button(按钮权限), api(接口权限)',
  `parent_id` bigint unsigned DEFAULT NULL COMMENT '父权限ID（用于构建权限树结构）',
  `sort_order` int NOT NULL DEFAULT 0 COMMENT '排序顺序',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态：1=启用 0=禁用',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_code` (`code`),
  KEY `idx_module` (`module`),
  KEY `idx_level` (`level`),
  KEY `idx_parent_id` (`parent_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='权限定义表';

-- ======================================
-- 2. 角色-权限关联表
-- ======================================
CREATE TABLE IF NOT EXISTS `role_permissions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '关联ID',
  `role_id` bigint unsigned NOT NULL COMMENT '角色ID',
  `permission_id` bigint unsigned NOT NULL COMMENT '权限ID',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_role_permission` (`role_id`, `permission_id`),
  KEY `idx_role_id` (`role_id`),
  KEY `idx_permission_id` (`permission_id`),
  CONSTRAINT `fk_role_permissions_role` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_role_permissions_permission` FOREIGN KEY (`permission_id`) REFERENCES `permissions` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色权限关联表';

-- ======================================
-- 3. 用户-权限关联表 (可选，用于用户级权限覆盖)
-- ======================================
CREATE TABLE IF NOT EXISTS `user_permissions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '关联ID',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `permission_id` bigint unsigned NOT NULL COMMENT '权限ID',
  `granted_by` bigint unsigned DEFAULT NULL COMMENT '授权人ID',
  `expire_time` datetime DEFAULT NULL COMMENT '权限过期时间（NULL表示永久）',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_permission` (`user_id`, `permission_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_permission_id` (`permission_id`),
  KEY `idx_expire_time` (`expire_time`),
  CONSTRAINT `fk_user_permissions_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_user_permissions_permission` FOREIGN KEY (`permission_id`) REFERENCES `permissions` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户权限关联表（可选）';

-- ======================================
-- 4. 修改现有角色表（添加新字段，保留兼容性）
-- ======================================
-- 添加新字段用于权限管理（保留 menu_ids 字段以保持向后兼容）
ALTER TABLE `roles`
ADD COLUMN IF NOT EXISTS `permission_ids` TEXT DEFAULT NULL COMMENT '权限ID列表（JSON格式）' AFTER `menu_ids`;

-- ======================================
-- 5. 创建权限日志表（审计用途）
-- ======================================
CREATE TABLE IF NOT EXISTS `permission_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '日志ID',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `permission_code` varchar(100) NOT NULL COMMENT '权限编码',
  `resource_type` varchar(50) DEFAULT NULL COMMENT '资源类型',
  `resource_id` bigint unsigned DEFAULT NULL COMMENT '资源ID',
  `action` varchar(50) NOT NULL COMMENT '操作类型',
  `result` varchar(20) NOT NULL COMMENT '结果：allowed/denied',
  `ip_address` varchar(50) DEFAULT NULL COMMENT 'IP地址',
  `user_agent` varchar(500) DEFAULT NULL COMMENT '用户代理',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_permission_code` (`permission_code`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_result` (`result`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='权限操作日志表';

-- ======================================
-- 索引优化
-- ======================================
-- 为权限表的常用查询添加复合索引
CREATE INDEX IF NOT EXISTS `idx_permissions_lookup` ON `permissions` (`module`, `resource`, `action`, `status`);

-- 为角色权限关联表添加查询优化索引
CREATE INDEX IF NOT EXISTS `idx_role_permissions_query` ON `role_permissions` (`role_id`, `permission_id`);

-- ======================================
-- 兼容性说明
-- ======================================
-- 1. 保留现有的 menu_ids 字段，确保向后兼容
-- 2. 新旧权限系统可以并存，渐进式迁移
-- 3. 权限检查时优先使用新系统，回退到旧系统
-- 4. 建议在确认新系统稳定后，再清理旧字段

-- ======================================
-- 迁移完成后验证
-- ======================================
-- 验证表是否创建成功
-- SELECT COUNT(*) FROM permissions;
-- SELECT COUNT(*) FROM role_permissions;
-- SELECT COUNT(*) FROM user_permissions;
-- SELECT COUNT(*) FROM permission_logs;

-- 验证外键约束是否生效
-- SELECT * FROM information_schema.KEY_COLUMN_USAGE
-- WHERE TABLE_SCHEMA = 'msre'
-- AND TABLE_NAME IN ('role_permissions', 'user_permissions');