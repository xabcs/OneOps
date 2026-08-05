-- 修复API权限管理菜单的路由名称
-- 问题：菜单中的routeName是 manage_apipermission，但前端路由实际是 manage_api-permission

-- 查看当前的API权限菜单配置
SELECT id, name, route_name, path FROM menus WHERE name LIKE '%API权限%' OR name LIKE '%api-permission%';

-- 修复路由名称（添加连字符）
UPDATE menus
SET route_name = 'manage_api-permission'
WHERE route_name = 'manage_apipermission';

-- 如果没有记录，插入新的菜单项
INSERT INTO menus (name, icon, path, route_name, menu_type, parent_id, sort, status, resource)
SELECT 'API权限管理', 'api', '/manage/api-permission', 'manage_api-permission', 'menu',
    (SELECT id FROM menus WHERE name = '系统管理' LIMIT 1),
    40, 1, 'api-permission'
WHERE NOT EXISTS (
    SELECT 1 FROM menus WHERE route_name = 'manage_api-permission'
);

-- 验证修复结果
SELECT id, name, route_name, path, parent_id, sort FROM menus WHERE route_name = 'manage_api-permission';
