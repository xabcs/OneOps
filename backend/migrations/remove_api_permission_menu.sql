-- 删除废弃的 API权限管理 菜单
-- 原因：已废弃，使用层级权限代码管理

-- 删除菜单项
DELETE FROM menus WHERE id = 73 AND name = 'API权限管理';

-- 删除相关的角色菜单关联
DELETE FROM role_menus WHERE menu_id = 73;

-- 确认删除
SELECT COUNT(*) as remaining_count FROM menus WHERE name = 'API权限管理';
