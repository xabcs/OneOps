-- 将"终端工作台"移到"资产管理"下，并改名为"web终端"
-- 执行方式: mysql -u root -p123456 nexops < backend/scripts/update_terminal_menu.sql

USE nexops;

-- 1. 将"终端工作台"菜单移到"资产管理"下（parent_id: 5 -> 2）
UPDATE menus
SET parent_id = 2,
    name = 'web终端',
    sort = 99  -- 放在最后
WHERE id = 60;

-- 2. 删除"终端管理"目录（因为下面没有子菜单了）
DELETE FROM menus WHERE id = 5;

-- 3. 验证结果
SELECT id, name, path, parent_id, sort, menu_type
FROM menus
WHERE parent_id = 2 OR name LIKE '%终端%'
ORDER BY parent_id, sort;

SELECT '菜单更新完成！' AS message;
