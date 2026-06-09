-- 重新初始化冗余字段数据（包含ID和名称）
-- 执行方式: mysql -u root -p123456 nexops < backend/scripts/reinit_redundant_fields.sql

USE nexops;

-- 1. 清空现有数据
UPDATE servers SET group_names = '', credential_names = '';

-- 2. 重新初始化（包含ID和名称）
UPDATE servers s
SET s.group_names = (
    SELECT CONCAT('[',
           GROUP_CONCAT(JSON_OBJECT('id', g.id, 'name', g.name) ORDER BY g.id),
           ']')
    FROM server_group_relations sgr
    JOIN server_groups g ON sgr.group_id = g.id
    WHERE sgr.server_id = s.id
);

UPDATE servers s
SET s.credential_names = (
    SELECT CONCAT('[',
           GROUP_CONCAT(JSON_OBJECT('id', c.id, 'name', c.name) ORDER BY c.id),
           ']')
    FROM server_credentials sc
    JOIN ssh_credentials c ON sc.credential_id = c.id
    WHERE sc.server_id = s.id
);

-- 3. 验证数据
SELECT id, hostname, group_names, credential_names FROM servers LIMIT 5;

SELECT '冗余字段重新初始化完成！' AS message;
