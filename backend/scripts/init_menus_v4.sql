-- ========================================
-- OneOps 完整菜单初始化 (V4 - 匹配新前端结构)
-- 创建时间: 2025-01-15
-- 版本: v4.0
-- 说明: 匹配新的前端模块化目录结构
-- 执行方式: mysql -u root -p ops < backend/scripts/init_menus_v4.sql
-- ========================================

USE ops;

-- 清空现有菜单（可选，如果需要重新初始化）
-- TRUNCATE TABLE menus;

-- ========== 一级菜单 ==========

-- 1. 首页
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('首页', 'mdi:monitor-dashboard', '/home', '', 'menu', 0, 1, 1, NOW(), NOW());

-- 2. 资产管理 (CMDB)
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('资产管理', 'mdi:server-network', '/cmdb', '', 'directory', 0, 2, 1, NOW(), NOW());
SET @cmdb_id = LAST_INSERT_ID();

-- 3. 监控中心
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('监控中心', 'mdi:chart-line', '/monitoring', '', 'directory', 0, 3, 1, NOW(), NOW());
SET @monitoring_id = LAST_INSERT_ID();

-- 4. 审计中心
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('审计中心', 'mdi:file-document', '/audit', '', 'directory', 0, 4, 1, NOW(), NOW());
SET @audit_id = LAST_INSERT_ID();

-- 5. 系统管理
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('系统管理', 'mdi:cog', '/manage', '', 'directory', 0, 6, 1, NOW(), NOW());
SET @manage_id = LAST_INSERT_ID();

-- ========== CMDB 二级菜单 ==========

-- 2.1 主机管理
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('主机管理', 'mdi:server', '/cmdb/servers', 'cmdb:server:query', 'menu', @cmdb_id, 1, 1, NOW(), NOW());

-- 2.2 业务管理
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('业务管理', 'mdi:sitemap', '/cmdb/business', 'cmdb:business:query', 'menu', @cmdb_id, 2, 1, NOW(), NOW());

-- 2.3 凭证管理 (包含访问凭证、SSH密钥)
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('凭证管理', 'mdi:key', '/cmdb/credentials', 'cmdb:credentials:query', 'directory', @cmdb_id, 3, 1, NOW(), NOW());
SET @cmdb_credentials_id = LAST_INSERT_ID();

-- 2.3.1 访问凭证
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('访问凭证', 'mdi:key-variant', '/cmdb/credentials/access', 'cmdb:credentials:access', 'menu', @cmdb_credentials_id, 1, 1, NOW(), NOW());

-- 2.3.2 SSH密钥
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('SSH密钥', 'mdi:ssh', '/cmdb/credentials/ssh', 'cmdb:credentials:ssh', 'menu', @cmdb_credentials_id, 2, 1, NOW(), NOW());

-- 2.4 访问策略
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('访问策略', 'mdi:shield-lock', '/cmdb/policies', 'cmdb:policies:query', 'menu', @cmdb_id, 4, 1, NOW(), NOW());

-- 2.5 配置管理 (目录类型，包含标签、机房、业务配置等)
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('配置管理', 'mdi:cog', '/cmdb/config', '', 'directory', @cmdb_id, 5, 1, NOW(), NOW());
SET @cmdb_config_id = LAST_INSERT_ID();

-- 2.5.1 业务系统配置
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('业务配置', 'mdi:sitemap', '/cmdb/config/business', 'cmdb:config:business', 'menu', @cmdb_config_id, 1, 1, NOW(), NOW());

-- 2.5.2 机房管理
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('机房管理', 'mdi:server', '/cmdb/config/rooms', 'cmdb:rooms:query', 'menu', @cmdb_config_id, 2, 1, NOW(), NOW());

-- 2.5.3 标签管理
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('标签管理', 'mdi:tag-multiple', '/cmdb/config/tags', 'cmdb:tags:query', 'menu', @cmdb_config_id, 3, 1, NOW(), NOW());

-- 2.5.4 监控代理配置
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('代理配置', 'mdi:robot', '/cmdb/config/agents', 'cmdb:agents:query', 'menu', @cmdb_config_id, 4, 1, NOW(), NOW());

-- 2.6 资产总览
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('资产总览', 'mdi:chart-pie', '/cmdb/dashboard', 'cmdb:dashboard:query', 'menu', @cmdb_id, 6, 1, NOW(), NOW());

-- 2.7 审计记录 (目录类型，包含变更记录、命令审计、会话审计)
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('审计记录', 'mdi:history', '/cmdb/audit', '', 'directory', @cmdb_id, 7, 1, NOW(), NOW());
SET @cmdb_audit_id = LAST_INSERT_ID();

-- 2.7.1 变更记录
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('变更记录', 'mdi:file-document', '/cmdb/audit/changes', 'cmdb:audit:changes', 'menu', @cmdb_audit_id, 1, 1, NOW(), NOW());

-- 2.7.2 命令审计 (目录类型)
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('命令审计', 'mdi:terminal', '/cmdb/audit/command', '', 'directory', @cmdb_audit_id, 2, 1, NOW(), NOW());
SET @cmdb_audit_command_id = LAST_INSERT_ID();

-- 2.7.2.1 命令历史
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('命令历史', 'mdi:history', '/cmdb/audit/command/history', 'cmdb:audit:command:history', 'menu', @cmdb_audit_command_id, 1, 1, NOW(), NOW());

-- 2.7.3 在线会话
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('在线会话', 'mdi:laptop', '/cmdb/audit/online', 'cmdb:audit:online', 'menu', @cmdb_audit_id, 3, 1, NOW(), NOW());

-- 2.7.4 历史会话
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('历史会话', 'mdi:history', '/cmdb/audit/sessions', 'cmdb:audit:sessions', 'menu', @cmdb_audit_id, 4, 1, NOW(), NOW());

-- 2.7.5 命令记录
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('命令记录', 'mdi:code-tags', '/cmdb/audit/commands', 'cmdb:audit:commands', 'menu', @cmdb_audit_id, 5, 1, NOW(), NOW());

-- ========== 监控中心 二级菜单 ==========

-- 3.1 监控概览
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('监控概览', 'mdi:chart-line', '/monitoring/overview', 'monitoring:overview:query', 'menu', @monitoring_id, 1, 1, NOW(), NOW());

-- 3.2 主机监控 (目录类型，包含详情页)
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('主机监控', 'mdi:server-network', '/monitoring/servers', 'monitoring:servers:query', 'directory', @monitoring_id, 2, 1, NOW(), NOW());

-- 3.3 告警管理
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('告警管理', 'mdi:alert-circle', '/monitoring/alerts', 'monitoring:alerts:query', 'menu', @monitoring_id, 3, 1, NOW(), NOW());

-- 3.4 趋势分析
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('趋势分析', 'mdi:chart-areaspline', '/monitoring/trends', 'monitoring:trends:query', 'menu', @monitoring_id, 4, 1, NOW(), NOW());

-- 3.5 巡检报告
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('巡检报告', 'mdi:file-document', '/monitoring/reports', 'monitoring:reports:query', 'menu', @monitoring_id, 5, 1, NOW(), NOW());

-- 3.6 监控设置
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('监控设置', 'mdi:cog', '/monitoring/settings', 'monitoring:settings:query', 'menu', @monitoring_id, 6, 1, NOW(), NOW());

-- ========== 审计中心 二级菜单 ==========

-- 4.1 登录审计
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('登录审计', 'mdi:login', '/audit/login', 'audit:login:query', 'menu', @audit_id, 1, 1, NOW(), NOW());

-- 4.2 操作审计
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('操作审计', 'mdi:account-edit', '/audit/operation', 'audit:operation:query', 'menu', @audit_id, 2, 1, NOW(), NOW());

-- 4.3 系统事件
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('系统事件', 'mdi:information', '/audit/system', 'audit:system:query', 'menu', @audit_id, 3, 1, NOW(), NOW());

-- ========== 系统管理 二级菜单 ==========

-- 6.1 用户管理
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('用户管理', 'mdi:account-multiple', '/manage/user', 'system:user:query', 'menu', @manage_id, 1, 1, NOW(), NOW());

-- 6.2 角色管理
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('角色管理', 'mdi:shield-account', '/manage/role', 'system:role:query', 'menu', @manage_id, 2, 1, NOW(), NOW());

-- 6.3 菜单管理
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES ('菜单管理', 'mdi:menu', '/manage/menu', 'system:menu:query', 'menu', @manage_id, 3, 1, NOW(), NOW());

-- ========== 更新超级管理员角色菜单权限 ==========

-- 获取超级管理员角色ID (假设为 admin 或 id=1)
SET @admin_role_id = (SELECT id FROM roles WHERE code = 'admin' LIMIT 1);

-- 如果超级管理员不存在，创建它
INSERT IGNORE INTO roles (name, code, description, status, created_at, updated_at)
VALUES ('超级管理员', 'admin', '系统超级管理员，拥有所有权限', 1, NOW(), NOW());

SET @admin_role_id = (SELECT id FROM roles WHERE code = 'admin' LIMIT 1);

-- 为超级管理员分配所有菜单
UPDATE roles
SET menu_ids = (
    SELECT JSON_ARRAYAGG(id)
    FROM menus
    WHERE status = 1
)
WHERE code = 'admin';

-- ========== 验证菜单结构 ==========

SELECT '=== 一级菜单 ===' AS '';
SELECT id, name, path, icon, sort
FROM menus
WHERE parent_id = 0 AND status = 1
ORDER BY sort;

SELECT '=== CMDB 子菜单 ===' AS '';
SELECT id, name, path, icon, sort
FROM menus
WHERE parent_id = @cmdb_id AND status = 1
ORDER BY sort;

SELECT '=== 监控中心子菜单 ===' AS '';
SELECT id, name, path, icon, sort
FROM menus
WHERE parent_id = @monitoring_id AND status = 1
ORDER BY sort;

SELECT '=== 审计中心子菜单 ===' AS '';
SELECT id, name, path, icon, sort
FROM menus
WHERE parent_id = @audit_id AND status = 1
ORDER BY sort;

SELECT '=== 系统管理子菜单 ===' AS '';
SELECT id, name, path, icon, sort
FROM menus
WHERE parent_id = @manage_id AND status = 1
ORDER BY sort;

SELECT '菜单初始化完成！' AS message;
