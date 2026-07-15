-- 立即解决 k8s_audit_sessions 路由错误
-- 删除数据库中的孤立菜单记录

USE ops;

-- 删除 K8s 会话审计菜单项
DELETE FROM menus WHERE path = '/k8s/audit/sessions';

-- 验证删除结果
SELECT id, name, path, parent_id, sort, status
FROM menus
WHERE parent_id = 5
ORDER BY sort;

-- 检查是否还有其他相关的孤立记录
SELECT id, name, path, parent_id
FROM menus
WHERE path LIKE '%audit%session%' OR path LIKE '%k8s%audit%';
