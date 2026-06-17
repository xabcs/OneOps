-- 移除 K8s 菜单下的会话审计项
-- 说明：会话审计功能应属于堡垒机模块，不应在 K8s 资源管理中

USE ops;

-- 删除 K8s 菜单下的会话审计菜单项
DELETE FROM menus WHERE path = '/k8s/audit/sessions';

-- 验证删除结果
SELECT id, name, path, parent_id, sort
FROM menus
WHERE parent_id = 5
ORDER BY sort;
