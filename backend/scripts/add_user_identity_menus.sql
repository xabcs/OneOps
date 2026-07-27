-- 添加用户身份映射和用户有效权限菜单
-- 执行时间：2026-07-24
-- 说明：为授权中心添加用户身份映射和用户有效权限管理菜单

-- 1. 添加用户身份映射菜单
INSERT INTO menus (name, icon, path, parent_id, sort, status, created_at, updated_at)
VALUES ('用户身份映射', 'mdi:account-switch', '/auth/user-identities', 7, 7, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE icon='mdi:account-switch', path='/auth/user-identities', parent_id=7, sort=7, updated_at=NOW();

-- 2. 添加用户有效权限菜单
INSERT INTO menus (name, icon, path, parent_id, sort, status, created_at, updated_at)
VALUES ('用户有效权限', 'mdi:shield-check', '/auth/user-permissions', 7, 8, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE icon='mdi:shield-check', path='/auth/user-permissions', parent_id=7, sort=8, updated_at=NOW();

-- 查看插入结果
SELECT id, name, icon, path, parent_id, sort, status
FROM menus
WHERE parent_id = 7
ORDER BY sort;
