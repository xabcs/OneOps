-- =============================================
-- 后端路由命名规范统一修复脚本
-- 规范：多个单词使用连字符连接（kebab-case）
-- =============================================

USE msre;

-- 查看当前所有管理模块的菜单
SELECT id, name, route_name, path
FROM menus
WHERE path LIKE '/manage/%'
ORDER BY sort;

-- 修复API权限管理路由名称
-- 后端路径：/api/system/casbin/*
-- 前端路由名称应为：manage_api-permission（使用连字符）
UPDATE menus
SET route_name = 'manage_api-permission'
WHERE route_name = 'manage_apipermission'
   OR (name = 'API权限管理' AND route_name != 'manage_api-permission');

-- 验证修复结果
SELECT '修复结果：' AS info;
SELECT id, name, route_name, path
FROM menus
WHERE name LIKE '%API权限%' OR route_name LIKE '%api%permission%';

-- 显示所有manage开头的路由名称
SELECT '当前所有管理模块路由：' AS info;
SELECT id, name, route_name, path, sort
FROM menus
WHERE path LIKE '/manage/%'
ORDER BY sort;

-- 确认路由命名规范一致性
SELECT
    CASE
        WHEN route_name REGEXP '[a-z]-[a-z]' THEN '✅ 符合规范（使用连字符）'
        WHEN route_name REGEXP '^[a-z]+_[a-z]+$' THEN '✅ 符合规范（单个单词）'
        ELSE '⚠️  需要检查'
    END AS status,
    name,
    route_name,
    path
FROM menus
WHERE path LIKE '/manage/%';
