-- 更新角色菜单：移除"终端管理"(5)，添加"web终端"(60)
-- 执行方式: mysql -u root -p123456 nexops < backend/scripts/update_role_menus_terminal.sql

USE nexops;

-- 更新超级管理员（拥有所有菜单，无需手动更新，因为使用*:*:*权限）

-- 更新运维工程师：移除5，添加60
UPDATE roles
SET menu_ids = JSON_REMOVE(menu_ids, JSON_UNQUOTE(JSON_SEARCH(menu_ids, 'one', '5')))
WHERE id = 14;

UPDATE roles
SET menu_ids = JSON_ARRAY_APPEND(menu_ids, '$', 60)
WHERE id = 14 AND JSON_SEARCH(menu_ids, 'one', '60') IS NULL;

-- 更新测试角色：移除5
UPDATE roles
SET menu_ids = JSON_REMOVE(menu_ids, JSON_UNQUOTE(JSON_SEARCH(menu_ids, 'one', '5')))
WHERE id = 20;

UPDATE roles
SET menu_ids = JSON_ARRAY_APPEND(menu_ids, '$', 60)
WHERE id = 20 AND JSON_SEARCH(menu_ids, 'one', '60') IS NULL;

-- 验证结果
SELECT id, name, menu_ids FROM roles;

SELECT '角色菜单更新完成！' AS message;
