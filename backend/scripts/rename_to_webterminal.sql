-- 重命名 terminal_workbench 为 webterminal
-- 执行时间：2026-06-08

-- 1. 更新 menus 表：web终端 菜单
-- 将 /terminal/workbench 更新为 /webterminal
UPDATE menus
SET path = '/webterminal',
    name = 'web终端'
WHERE path = '/terminal/workbench';

-- 2. 验证结果
SELECT id, parent_id, name, path, icon, sort
FROM menus
WHERE path = '/webterminal' OR name = 'web终端';
