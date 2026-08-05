# 前端路由命名问题诊断与解决

## 问题现状

前端报错：
```
Error: No match for {"name":"manage_apipermission","params":{}}
```

实际前端路由名称：`manage_api-permission`（带连字符）✅

## 根本原因分析

### 1. 前端路由配置（正确）

`frontend/src/router/elegant/transform.ts`:
```typescript
"manage_api-permission": "/manage/api-permission",  // ✅ 正确
```

`frontend/src/router/elegant/routes.ts`:
```typescript
{
  name: 'manage_api-permission',  // ✅ 正确
  path: '/manage/api-permission',
}
```

### 2. 后端数据库（正确）

```sql
SELECT * FROM menus WHERE name = 'API权限管理';
-- path: /manage/api-permission  ✅ 正确
```

### 3. 问题来源

`manage_apipermission`（缺少连字符）这个路由名称来源：

**可能原因1：浏览器缓存**
- 用户浏览器缓存了旧的菜单数据或路由状态
- 前端路由之前可能使用过错误的命名

**可能原因2：动态路由生成错误**
- 前端可能从 `/manage/api-permission` 路径生成路由名称时错误地移除了连字符

## 解决方案

### 方案1：清除浏览器缓存（最简单）

```bash
# 用户操作
1. 清除浏览器缓存
2. 或者使用无痕模式打开
3. 或者重新登录
```

### 方案2：检查是否有其他地方生成了错误的路由名称<tool_call>Write<arg_key>content</arg_key><arg_value>-- 完整的数据库检查和修复SQL

-- 1. 检查所有管理模块菜单
SELECT
    id,
    name,
    path,
    permission,
    resource,
    CASE
        WHEN path LIKE '%api-permission%' THEN '✅ 正确路径'
        WHEN path LIKE '%apipermission%' THEN '❌ 错误路径（缺少连字符）'
        ELSE '普通菜单'
    END AS status
FROM menus
WHERE path LIKE '/manage/%'
ORDER BY sort;

-- 2. 检查是否有重复的API权限管理菜单
SELECT
    id,
    name,
    path,
    permission,
    resource,
    created_at
FROM menus
WHERE name = 'API权限管理'
ORDER BY id;

-- 3. 如果有重复，删除重复的菜单（保留最新的或正确的）
-- 注意：执行前请备份数据

-- 先查看要删除的记录
SELECT * FROM menus
WHERE name = 'API权限管理'
  AND path != '/manage/api-permission';

-- 删除错误的重复记录（谨慎执行！）
-- DELETE FROM menus
-- WHERE name = 'API权限管理'
--   AND path != '/manage/api-permission';

-- 4. 确保只有一个正确的API权限管理菜单
-- 如果没有记录，插入新的
INSERT INTO menus (name, icon, path, permission, menu_type, parent_id, sort, status, resource, created_at, updated_at)
SELECT
    'API权限管理',
    'api',
    '/manage/api-permission',
    'system.api-permission.view',
    'menu',
    (SELECT id FROM (SELECT id FROM menus WHERE name = '系统管理' LIMIT 1) AS tmp),
    40,
    1,
    'api-permission',
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM menus
    WHERE name = 'API权限管理'
      AND path = '/manage/api-permission'
);

-- 5. 最终验证
SELECT
    '✅ API权限管理菜单配置正确' AS result,
    id,
    name,
    path,
    permission,
    resource
FROM menus
WHERE name = 'API权限管理'
  AND path = '/manage/api-permission';
