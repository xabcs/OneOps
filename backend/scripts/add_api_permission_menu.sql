-- API权限管理菜单配置脚本
-- 为系统管理模块添加API权限管理菜单项

-- 插入API权限管理菜单（系统管理 -> API权限管理）
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, resource)
VALUES (
    'API权限管理',
    'key',
    '/manage/api-permission',
    'system.api-permission.view',
    2, -- 菜单类型：2表示子菜单
    (SELECT id FROM menus WHERE path = '/manage' LIMIT 1), -- 父菜单ID（系统管理）
    50, -- 排序（在角色管理之后）
    1, -- 启用状态
    'api-permission' -- 对应的资源标识
) ON DUPLICATE KEY UPDATE
  name = 'API权限管理',
  path = '/manage/api-permission',
  permission = 'system.api-permission.view',
  sort = 50,
  status = 1,
  resource = 'api-permission';

-- 验证插入结果
SELECT id, name, path, permission, menu_type, parent_id, sort, status, resource
FROM menus
WHERE path = '/manage/api-permission';