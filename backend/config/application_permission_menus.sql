-- 添加外部权限管理菜单
-- 注意：执行前请根据实际情况调整 parent_id 和 order_num

-- 1. 添加主菜单：外部权限管理
INSERT INTO menus (name, title, icon, path, component, redirect, order_num, parent_id, is_hidden, is_hidden_children, status, create_time, update_time)
VALUES ('application-permission', '外部权限管理', 'mdi:application-cog', '/application-permission', 'layout.base$view.application-permission', NULL, 5, 0, 0, 0, 1, NOW(), NOW());

-- 获取刚插入的菜单ID（用于后续子菜单的parent_id）
SET @main_menu_id = LAST_INSERT_ID();

-- 2. 添加子菜单：应用管理
INSERT INTO menus (name, title, icon, path, component, redirect, order_num, parent_id, is_hidden, is_hidden_children, status, create_time, update_time)
VALUES ('application-permission-apps', '应用管理', 'mdi:application', '/application-permission/apps', 'view.application_permission', NULL, 1, @main_menu_id, 0, 0, 1, NOW(), NOW());

-- 3. 添加子菜单：角色绑定
INSERT INTO menus (name, title, icon, path, component, redirect, order_num, parent_id, is_hidden, is_hidden_children, status, create_time, update_time)
VALUES ('application-permission-bindings', '角色绑定', 'mdi:link', '/application-permission/bindings', 'view.application_permission', NULL, 2, @main_menu_id, 0, 0, 1, NOW(), NOW());

-- 4. 添加子菜单：用户授权
INSERT INTO menus (name, title, icon, path, component, redirect, order_num, parent_id, is_hidden, is_hidden_children, status, create_time, update_time)
VALUES ('application-permission-users', '用户授权', 'mdi:account-key', '/application-permission/users', 'view.application_permission', NULL, 3, @main_menu_id, 0, 0, 1, NOW(), NOW());

-- 5. 添加子菜单：操作日志
INSERT INTO menus (name, title, icon, path, component, redirect, order_num, parent_id, is_hidden, is_hidden_children, status, create_time, update_time)
VALUES ('application-permission-logs', '操作日志', 'mdi:file-document', '/application-permission/logs', 'view.application_permission', NULL, 4, @main_menu_id, 0, 0, 1, NOW(), NOW());

-- 查看插入的菜单
SELECT * FROM menus WHERE name LIKE 'application-permission%' ORDER BY order_num;
