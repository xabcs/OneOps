-- 授权中心"角色"改名为"用户组"数据库迁移脚本
-- 执行时间：2026-07-22
-- 说明：将 auth_roles、auth_user_roles、role_bindings 表重命名为用户组相关命名

-- ============================================
-- 步骤 1: 重命名表
-- ============================================

-- 1.1 重命名 auth_roles → auth_groups
RENAME TABLE auth_roles TO auth_groups;

-- 1.2 重命名 auth_user_roles → auth_user_groups
RENAME TABLE auth_user_roles TO auth_user_groups;

-- 1.3 重命名 role_bindings → group_bindings
RENAME TABLE role_bindings TO group_bindings;

-- ============================================
-- 步骤 2: 重命名字段
-- ============================================

-- 2.1 修改 auth_groups 表字段名
ALTER TABLE auth_groups
  CHANGE COLUMN `code` `group_code` VARCHAR(50) NOT NULL COMMENT '用户组代码',
  CHANGE COLUMN `name` `group_name` VARCHAR(50) NOT NULL COMMENT '用户组名称';

-- 2.2 修改 auth_user_groups 表字段名
ALTER TABLE auth_user_groups
  CHANGE COLUMN `role_id` `group_id` INT UNSIGNED NOT NULL COMMENT '用户组ID';

-- 2.3 修改 group_bindings 表字段名
ALTER TABLE group_bindings
  CHANGE COLUMN `role_id` `group_id` INT UNSIGNED NOT NULL COMMENT '授权中心用户组ID',
  CHANGE COLUMN `application_role_id` `application_permission_id` INT UNSIGNED NOT NULL COMMENT '外部应用权限ID';

-- ============================================
-- 步骤 3: 更新外键约束
-- ============================================

-- 注意：由于使用了 RENAME TABLE，外键约束会自动更新
-- 但需要确保外键名称也更新（可选，建议更新以保持一致性）

-- 查看当前外键约束
SELECT
  CONSTRAINT_NAME,
  TABLE_NAME,
  COLUMN_NAME,
  REFERENCED_TABLE_NAME,
  REFERENCED_COLUMN_NAME
FROM information_schema.KEY_COLUMN_USAGE
WHERE TABLE_SCHEMA = DATABASE()
  AND TABLE_NAME IN ('auth_user_groups', 'group_bindings')
  AND REFERENCED_TABLE_NAME IS NOT NULL;

-- ============================================
-- 步骤 4: 更新索引名称（可选）
-- ============================================

-- 如果需要更新索引名称以保持一致性，可以执行以下语句：

-- 4.1 auth_user_groups 表索引
ALTER TABLE auth_user_groups
  DROP INDEX idx_role_id,
  ADD INDEX idx_group_id (group_id);

-- 4.2 group_bindings 表索引
ALTER TABLE group_bindings
  DROP INDEX idx_role_id,
  ADD INDEX idx_group_id (group_id);

-- ============================================
-- 步骤 5: 验证数据完整性
-- ============================================

-- 5.1 验证用户组数据
SELECT COUNT(*) as '用户组数量' FROM auth_groups;

-- 5.2 验证用户组成员数据
SELECT COUNT(*) as '用户组成员数量' FROM auth_user_groups;

-- 5.3 验证权限绑定数据
SELECT COUNT(*) as '权限绑定数量' FROM group_bindings;

-- 5.4 验证数据关系完整性
SELECT
  '用户组成员关系' as '类型',
  COUNT(*) as '记录数',
  COUNT(DISTINCT user_id) as '用户数',
  COUNT(DISTINCT group_id) as '用户组数'
FROM auth_user_groups

UNION ALL

SELECT
  '权限绑定关系' as '类型',
  COUNT(*) as '记录数',
  COUNT(DISTINCT group_id) as '用户组数',
  COUNT(DISTINCT app_id) as '应用数'
FROM group_bindings;

-- ============================================
-- 步骤 6: 更新表注释
-- ============================================

ALTER TABLE auth_groups COMMENT '授权中心用户组表';
ALTER TABLE auth_user_groups COMMENT '用户组成员关系表';
ALTER TABLE group_bindings COMMENT '用户组权限绑定表';

-- ============================================
-- 步骤 7: 最终验证
-- ============================================

-- 查看所有相关表的结构
SHOW CREATE TABLE auth_groups;
SHOW CREATE TABLE auth_user_groups;
SHOW CREATE TABLE group_bindings;

-- 确认所有数据都正确迁移
SELECT
  'auth_groups' as table_name,
  COUNT(*) as row_count
FROM auth_groups

UNION ALL

SELECT
  'auth_user_groups' as table_name,
  COUNT(*) as row_count
FROM auth_user_groups

UNION ALL

SELECT
  'group_bindings' as table_name,
  COUNT(*) as row_count
FROM group_bindings;

-- ============================================
-- 回滚脚本（如果需要）
-- ============================================

/*
-- 如果需要回滚，执行以下脚本：

RENAME TABLE auth_groups TO auth_roles;
RENAME TABLE auth_user_groups TO auth_user_roles;
RENAME TABLE group_bindings TO role_bindings;

ALTER TABLE auth_roles
  CHANGE COLUMN `group_code` `code` VARCHAR(50) NOT NULL COMMENT '角色代码',
  CHANGE COLUMN `group_name` `name` VARCHAR(50) NOT NULL COMMENT '角色名称';

ALTER TABLE auth_user_roles
  CHANGE COLUMN `group_id` `role_id` INT UNSIGNED NOT NULL COMMENT '角色ID';

ALTER TABLE role_bindings
  CHANGE COLUMN `group_id` `role_id` INT UNSIGNED NOT NULL COMMENT '授权中心角色ID',
  CHANGE COLUMN `application_permission_id` `application_role_id` INT UNSIGNED NOT NULL COMMENT '外部应用角色ID';

ALTER TABLE auth_roles COMMENT '授权中心角色表';
ALTER TABLE auth_user_roles COMMENT '用户角色分配表';
ALTER TABLE role_bindings COMMENT '角色绑定表';
*/

-- ============================================
-- 执行完成提示
-- ============================================

SELECT '数据库迁移完成！' as '状态',
       '请继续执行后端和前端代码重构' as '下一步';
