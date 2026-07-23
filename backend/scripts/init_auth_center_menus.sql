-- 初始化授权中心菜单
-- 执行时间：2026-07-22
-- 说明：创建授权中心的完整菜单结构

-- 1. 添加授权中心主菜单
INSERT INTO menus (name, title, icon, path, component, redirect, order_num, parent_id, is_hidden, is_hidden_children, status, create_time, update_time)
VALUES ('auth', '授权中心', 'mdi:shield-account', '/auth', 'layout.base', '/auth/users', 3, 0, 0, 0, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE title='授权中心', icon='mdi:shield-account', path='/auth', component='layout.base', redirect='/auth/users', order_num=3;

-- 获取授权中心菜单ID
SET @auth_menu_id = (SELECT id FROM menus WHERE name = 'auth' LIMIT 1);

-- 2. 添加用户管理菜单
INSERT INTO menus (name, title, icon, path, component, redirect, order_num, parent_id, is_hidden, is_hidden_children, status, create_time, update_time)
VALUES ('auth_users', '用户管理', 'mdi:account-multiple', '/auth/users', 'view.auth_users', NULL, 1, @auth_menu_id, 0, 0, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE title='用户管理', icon='mdi:account-multiple', path='/auth/users', component='view.auth_users', parent_id=@auth_menu_id, order_num=1;

-- 3. 添加角色管理菜单
INSERT INTO menus (name, title, icon, path, component, redirect, order_num, parent_id, is_hidden, is_hidden_children, status, create_time, update_time)
VALUES ('auth_roles', '角色管理', 'mdi:account-star', '/auth/roles', 'view.auth_roles', NULL, 2, @auth_menu_id, 0, 0, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE title='角色管理', icon='mdi:account-star', path='/auth/roles', component='view.auth_roles', parent_id=@auth_menu_id, order_num=2;

-- 4. 添加外部应用菜单
INSERT INTO menus (name, title, icon, path, component, redirect, order_num, parent_id, is_hidden, is_hidden_children, status, create_time, update_time)
VALUES ('auth_applications', '外部应用', 'mdi:application-cog', '/auth/applications', 'view.auth_applications', NULL, 3, @auth_menu_id, 0, 0, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE title='外部应用', icon='mdi:application-cog', path='/auth/applications', component='view.auth_applications', parent_id=@auth_menu_id, order_num=3;

-- 5. 添加角色绑定菜单（重要！）
INSERT INTO menus (name, title, icon, path, component, redirect, order_num, parent_id, is_hidden, is_hidden_children, status, create_time, update_time)
VALUES ('auth_rolebindings', '角色绑定', 'mdi:link-variant', '/auth/rolebindings', 'view.auth_rolebindings', NULL, 4, @auth_menu_id, 0, 0, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE title='角色绑定', icon='mdi:link-variant', path='/auth/rolebindings', component='view.auth_rolebindings', parent_id=@auth_menu_id, order_num=4;

-- 6. 添加用户授权菜单（重要！）
INSERT INTO menus (name, title, icon, path, component, redirect, order_num, parent_id, is_hidden, is_hidden_children, status, create_time, update_time)
VALUES ('auth_userauthorization', '用户授权', 'mdi:account-key', '/auth/userauthorization', 'view.auth_userauthorization', NULL, 5, @auth_menu_id, 0, 0, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE title='用户授权', icon='mdi:account-key', path='/auth/userauthorization', component='view.auth_userauthorization', parent_id=@auth_menu_id, order_num=5;

-- 7. 添加操作日志菜单
INSERT INTO menus (name, title, icon, path, component, redirect, order_num, parent_id, is_hidden, is_hidden_children, status, create_time, update_time)
VALUES ('auth_operationlogs', '操作日志', 'mdi:file-document-outline', '/auth/operationlogs', 'view.auth_operationlogs', NULL, 6, @auth_menu_id, 0, 0, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE title='操作日志', icon='mdi:file-document-outline', path='/auth/operationlogs', component='view.auth_operationlogs', parent_id=@auth_menu_id, order_num=6;

-- 查看插入结果
SELECT id, name, title, path, component, parent_id, order_num, status
FROM menus
WHERE name LIKE 'auth%'
ORDER BY parent_id, order_num;

-- 说明：
-- 1. 使用 INSERT ... ON DUPLICATE KEY UPDATE 确保可以重复执行
-- 2. component 必须与路由配置中的 component 名称匹配
--    - view.auth_users 对应 imports.ts 中的 auth_users
--    - view.auth_rolebindings 对应 imports.ts 中的 auth_rolebindings
--    - view.auth_userauthorization 对应 imports.ts 中的 auth_userauthorization
-- 3. parent_id 应该指向授权中心主菜单的ID
-- 4. order_num 控制菜单显示顺序
-- 5. status=1 表示启用，0 表示禁用
-- 6. is_hidden=0 表示菜单可见
