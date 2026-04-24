-- 清理操作审计日志中的"未知模块"
-- 执行此 SQL 可以修复历史数据中的 module 字段

-- 更新历史数据中的 module 字段
UPDATE operation_logs
SET module = CASE
    WHEN path LIKE '/api/monitoring%' THEN '监控中心'
    WHEN path LIKE '/api/audit%' THEN '审计管理'
    WHEN path LIKE '/api/system/menus%' THEN '菜单管理'
    WHEN path LIKE '/api/system/roles%' THEN '角色管理'
    WHEN path LIKE '/api/system/users%' THEN '用户管理'
    WHEN path LIKE '/api/user%' THEN '用户管理'
    WHEN path LIKE '/api/login%' THEN '认证管理'
    WHEN path LIKE '/api/logout%' THEN '认证管理'
    WHEN path LIKE '/api/tasks%' THEN '任务管理'
    WHEN path LIKE '/api/servers%' THEN '服务器管理'
    WHEN path LIKE '/api/containers%' THEN '容器管理'
    WHEN path LIKE '/api/certificates%' THEN '证书管理'
    WHEN path LIKE '/api/system%' THEN '系统管理'
    ELSE '其他'
END
WHERE module = '未知模块' OR module = '';

-- 查看修复结果
SELECT module, COUNT(*) as count
FROM operation_logs
GROUP BY module
ORDER BY count DESC;
