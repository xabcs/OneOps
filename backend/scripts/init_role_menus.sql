-- 角色菜单关联初始化脚本
-- 为不同角色分配菜单权限

-- 获取菜单ID
SET @home_id = (SELECT id FROM menus WHERE path = '/home' LIMIT 1);
SET @manage_id = (SELECT id FROM menus WHERE path = '/manage' LIMIT 1);
SET @about_id = (SELECT id FROM menus WHERE path = '/about' LIMIT 1);
SET @user_center_id = (SELECT id FROM menus WHERE path = '/user-center' LIMIT 1);
SET @function_id = (SELECT id FROM menus WHERE path = '/function' LIMIT 1);
SET @plugin_id = (SELECT id FROM menus WHERE path = '/plugin' LIMIT 1);

-- 普通用户 (user)：拥有基础菜单权限
-- 分配菜单：首页、功能示例、关于、用户中心
UPDATE roles
SET menu_ids = JSON_ARRAY(@home_id, @function_id, @about_id, @user_center_id)
WHERE code = 'user';

-- 运维工程师 (ops)：拥有系统管理权限
-- 分配菜单：首页、系统管理、功能示例、插件示例、关于、用户中心
UPDATE roles
SET menu_ids = JSON_ARRAY(@home_id, @manage_id, @function_id, @plugin_id, @about_id, @user_center_id)
WHERE code = 'ops';

-- 审计员 (auditor)：只拥有基础权限
-- 分配菜单：首页、关于
UPDATE roles
SET menu_ids = JSON_ARRAY(@home_id, @about_id)
WHERE code = 'auditor';

-- 测试角色 (test)：拥有测试权限
-- 分配菜单：首页、功能示例、关于
UPDATE roles
SET menu_ids = JSON_ARRAY(@home_id, @function_id, @about_id)
WHERE code = 'test';

-- 查询验证
SELECT id, name, code, menu_ids FROM roles;
