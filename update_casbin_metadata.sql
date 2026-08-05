-- 为 casbin_rule 表中的策略添加业务信息
-- v3: API名称
-- v4: API描述
-- v5: 模块名称

USE msre;

-- 查看当前数据
SELECT * FROM casbin_rule WHERE p_type = 'p' LIMIT 10;

-- 更新用户管理相关API的业务信息
UPDATE casbin_rule SET
    v3 = '用户列表',
    v4 = '获取用户列表',
    v5 = 'user'
WHERE v1 = '/api/system/users' AND v2 = 'GET';

UPDATE casbin_rule SET
    v3 = '用户详情',
    v4 = '获取单个用户详情',
    v5 = 'user'
WHERE v1 = '/api/system/users/:id' AND v2 = 'GET';

UPDATE casbin_rule SET
    v3 = '创建用户',
    v4 = '创建新用户',
    v5 = 'user'
WHERE v1 = '/api/system/users' AND v2 = 'POST';

UPDATE casbin_rule SET
    v3 = '更新用户',
    v4 = '更新用户信息',
    v5 = 'user'
WHERE v1 = '/api/system/users/:id' AND v2 = 'PUT';

UPDATE casbin_rule SET
    v3 = '删除用户',
    v4 = '删除用户',
    v5 = 'user'
WHERE v1 = '/api/system/users/:id' AND v2 = 'DELETE';

UPDATE casbin_rule SET
    v3 = '重置密码',
    v4 = '重置用户密码',
    v5 = 'user'
WHERE v1 = '/api/system/users/:id/password' AND v2 = 'PUT';

-- 更新角色管理相关API的业务信息
UPDATE casbin_rule SET
    v3 = '角色列表',
    v4 = '获取角色列表',
    v5 = 'role'
WHERE v1 = '/api/system/roles' AND v2 = 'GET';

UPDATE casbin_rule SET
    v3 = '角色详情',
    v4 = '获取单个角色详情',
    v5 = 'role'
WHERE v1 = '/api/system/roles/:id' AND v2 = 'GET';

UPDATE casbin_rule SET
    v3 = '创建角色',
    v4 = '创建新角色',
    v5 = 'role'
WHERE v1 = '/api/system/roles' AND v2 = 'POST';

UPDATE casbin_rule SET
    v3 = '更新角色',
    v4 = '更新角色信息',
    v5 = 'role'
WHERE v1 = '/api/system/roles/:id' AND v2 = 'PUT';

UPDATE casbin_rule SET
    v3 = '删除角色',
    v4 = '删除角色',
    v5 = 'role'
WHERE v1 = '/api/system/roles/:id' AND v2 = 'DELETE';

-- 更新菜单管理相关API的业务信息
UPDATE casbin_rule SET
    v3 = '菜单列表',
    v4 = '获取菜单列表',
    v5 = 'menu'
WHERE v1 = '/api/system/menus' AND v2 = 'GET';

UPDATE casbin_rule SET
    v3 = '菜单树',
    v4 = '获取菜单树结构',
    v5 = 'menu'
WHERE v1 = '/api/system/menus/tree' AND v2 = 'GET';

UPDATE casbin_rule SET
    v3 = '创建菜单',
    v4 = '创建新菜单',
    v5 = 'menu'
WHERE v1 = '/api/system/menus' AND v2 = 'POST';

UPDATE casbin_rule SET
    v3 = '更新菜单',
    v4 = '更新菜单信息',
    v5 = 'menu'
WHERE v1 = '/api/system/menus/:id' AND v2 = 'PUT';

UPDATE casbin_rule SET
    v3 = '删除菜单',
    v4 = '删除菜单',
    v5 = 'menu'
WHERE v1 = '/api/system/menus/:id' AND v2 = 'DELETE';

-- 更新权限管理相关API的业务信息
UPDATE casbin_rule SET
    v3 = '权限列表',
    v4 = '获取权限列表',
    v5 = 'permission'
WHERE v1 = '/api/system/permissions' AND v2 = 'GET';

UPDATE casbin_rule SET
    v3 = '权限树',
    v4 = '获取权限树结构',
    v5 = 'permission'
WHERE v1 = '/api/system/permissions/tree' AND v2 = 'GET';

UPDATE casbin_rule SET
    v3 = '创建权限',
    v4 = '创建新权限',
    v5 = 'permission'
WHERE v1 = '/api/system/permissions' AND v2 = 'POST';

UPDATE casbin_rule SET
    v3 = '更新权限',
    v4 = '更新权限信息',
    v5 = 'permission'
WHERE v1 = '/api/system/permissions/:id' AND v2 = 'PUT';

UPDATE casbin_rule SET
    v3 = '删除权限',
    v4 = '删除权限',
    v5 = 'permission'
WHERE v1 = '/api/system/permissions/:id' AND v2 = 'DELETE';

-- 验证更新结果
SELECT
    v0 as role,
    v1 as path,
    v2 as method,
    v3 as name,
    v4 as description,
    v5 as module
FROM casbin_rule
WHERE p_type = 'p'
ORDER BY v5, v1, v2
LIMIT 20;
