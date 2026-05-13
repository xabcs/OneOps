-- 动态路由菜单初始化脚本
-- 用于 SoybeanAdmin 动态路由模式

-- 清空现有菜单（可选，如果需要重新初始化）
TRUNCATE TABLE menus;

-- 一级菜单：首页
INSERT INTO menus (name, icon, path, permission, parent_id, sort, status, created_at, updated_at)
VALUES ('首页', 'mdi:monitor-dashboard', '/home', '', 0, 1, 1, NOW(), NOW());

-- 一级菜单：系统管理
INSERT INTO menus (name, icon, path, permission, parent_id, sort, status, created_at, updated_at)
VALUES ('系统管理', 'carbon:cloud-service-management', '/manage', '', 0, 2, 1, NOW(), NOW());

-- 获取系统管理菜单的ID（用于子菜单）
SET @manage_id = LAST_INSERT_ID();

-- 系统管理子菜单：用户管理
INSERT INTO menus (name, icon, path, permission, parent_id, sort, status, created_at, updated_at)
VALUES ('用户管理', 'ic:round-manage-accounts', '/manage/user', 'system:user:query', @manage_id, 1, 1, NOW(), NOW());

-- 系统管理子菜单：角色管理
INSERT INTO menus (name, icon, path, permission, parent_id, sort, status, created_at, updated_at)
VALUES ('角色管理', 'carbon:user-role', '/manage/role', 'system:role:query', @manage_id, 2, 1, NOW(), NOW());

-- 系统管理子菜单：菜单管理
INSERT INTO menus (name, icon, path, permission, parent_id, sort, status, created_at, updated_at)
VALUES ('菜单管理', 'material-symbols:route', '/manage/menu', 'system:menu:query', @manage_id, 3, 1, NOW(), NOW());

-- 一级菜单：功能示例
INSERT INTO menus (name, icon, path, permission, parent_id, sort, status, created_at, updated_at)
VALUES ('功能示例', 'icon-park-outline:all-application', '/function', '', 0, 3, 1, NOW(), NOW());

SET @function_id = LAST_INSERT_ID();

-- 功能示例子菜单：多标签页
INSERT INTO menus (name, icon, path, permission, parent_id, sort, status, created_at, updated_at)
VALUES ('多标签页', 'ic:round-tab', '/function/tab', '', @function_id, 1, 1, NOW(), NOW());

-- 功能示例子菜单：请求
INSERT INTO menus (name, icon, path, permission, parent_id, sort, status, created_at, updated_at)
VALUES ('请求', 'carbon:network-overlay', '/function/request', '', @function_id, 2, 1, NOW(), NOW());

-- 一级菜单：插件示例
INSERT INTO menus (name, icon, path, permission, parent_id, sort, status, created_at, updated_at)
VALUES ('插件示例', 'clarity:plugin-line', '/plugin', '', 0, 4, 1, NOW(), NOW());

SET @plugin_id = LAST_INSERT_ID();

-- 插件示例子菜单：图表
INSERT INTO menus (name, icon, path, permission, parent_id, sort, status, created_at, updated_at)
VALUES ('图表', 'mdi:chart-areaspline', '/plugin/charts', '', @plugin_id, 1, 1, NOW(), NOW());

-- 插件示例子菜单：编辑器
INSERT INTO menus (name, icon, path, permission, parent_id, sort, status, created_at, updated_at)
VALUES ('编辑器', 'icon-park-outline:editor', '/plugin/editor', '', @plugin_id, 2, 1, NOW(), NOW());

-- 插件示例子菜单：表格
INSERT INTO menus (name, icon, path, permission, parent_id, sort, status, created_at, updated_at)
VALUES ('表格', 'icon-park-outline:table', '/plugin/tables', '', @plugin_id, 3, 1, NOW(), NOW());

-- 一级菜单：关于
INSERT INTO menus (name, icon, path, permission, parent_id, sort, status, created_at, updated_at)
VALUES ('关于', 'fluent:book-information-24-regular', '/about', '', 0, 5, 1, NOW(), NOW());

-- 一级菜单：用户中心
INSERT INTO menus (name, icon, path, permission, parent_id, sort, status, created_at, updated_at)
VALUES ('用户中心', 'mdi:user-circle-outline', '/user-center', '', 0, 6, 1, NOW(), NOW());

-- 验证插入结果
SELECT id, name, path, parent_id, sort FROM menus ORDER BY sort;
