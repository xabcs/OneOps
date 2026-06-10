-- 添加服务器详情页菜单
-- 执行方式: mysql -u root -p密码 nexops < backend/scripts/add_server_detail_menu.sql

USE nexops;

-- 先获取服务器管理菜单的ID
SET @parent_id = (SELECT id FROM menus WHERE path = '/cmdb/servers' LIMIT 1);

-- 插入服务器详情页菜单（作为服务器管理的子菜单）
INSERT INTO menus (name, path, permission, menu_type, parent_id, sort, status, created_at, updated_at)
VALUES (
  '服务器详情',
  '/cmdb/server/detail',
  '',
  'menu',
  @parent_id,
  999,  -- 排在最后
  1,
  NOW(),
  NOW()
);

-- 获取新插入的菜单ID
SET @new_menu_id = LAST_INSERT_ID();

-- 更新超级管理员（使用 *:*:* 权限，无需手动添加）

-- 更新运维工程师角色，添加新菜单权限
UPDATE roles
SET menu_ids = JSON_ARRAY_APPEND(menu_ids, '$', @new_menu_id)
WHERE code = 'ops' AND JSON_SEARCH(menu_ids, 'one', @new_menu_id) IS NULL;

-- 更新审计员角色，添加新菜单权限
UPDATE roles
SET menu_ids = JSON_ARRAY_APPEND(menu_ids, '$', @new_menu_id)
WHERE code = 'auditor' AND JSON_SEARCH(menu_ids, 'one', @new_menu_id) IS NULL;

-- 更新普通用户角色，添加新菜单权限
UPDATE roles
SET menu_ids = JSON_ARRAY_APPEND(menu_ids, '$', @new_menu_id)
WHERE code = 'user' AND JSON_SEARCH(menu_ids, 'one', @new_menu_id) IS NULL;

-- 验证结果
SELECT id, name, path, parent_id, sort, menu_type, status
FROM menus
WHERE path = '/cmdb/server/detail';

SELECT '服务器详情页菜单添加完成！' AS message;
