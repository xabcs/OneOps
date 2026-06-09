-- 添加冗余字段优化服务器列表查询性能
-- 执行方式: mysql -u root -p123456 nexops < backend/scripts/migrate_add_redundant_fields.sql

USE nexops;

-- 1. 添加冗余字段到 servers 表（如果已存在则忽略）
SET @column_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
                   WHERE TABLE_SCHEMA = 'nexops' AND TABLE_NAME = 'servers' AND COLUMN_NAME = 'group_names');

SET @sql = IF(@column_exists = 0,
    'ALTER TABLE servers ADD COLUMN group_names VARCHAR(500) DEFAULT '''' COMMENT ''分组JSON数组(ID+名称)'''',
    'SELECT ''''group_names already exists''''');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @column_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
                   WHERE TABLE_SCHEMA = 'nexops' AND TABLE_NAME = 'servers' AND COLUMN_NAME = 'credential_names');

SET @sql = IF(@column_exists = 0,
    'ALTER TABLE servers ADD COLUMN credential_names VARCHAR(500) DEFAULT '''' COMMENT ''凭证JSON数组(ID+名称)'''',
    'SELECT ''''credential_names already exists''''');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 2. 初始化现有数据（包含ID和名称）
UPDATE servers s
SET s.group_names = (
    SELECT CONCAT('[',
           GROUP_CONCAT(JSON_OBJECT('id', g.id, 'name', g.name)),
           ']')
    FROM server_group_relations sgr
    JOIN server_groups g ON sgr.group_id = g.id
    WHERE sgr.server_id = s.id
);

UPDATE servers s
SET s.credential_names = (
    SELECT CONCAT('[',
           GROUP_CONCAT(JSON_OBJECT('id', c.id, 'name', c.name)),
           ']')
    FROM server_credentials sc
    JOIN ssh_credentials c ON sc.credential_id = c.id
    WHERE sc.server_id = s.id
);

-- 3. 验证数据
SELECT id, hostname, group_names, credential_names FROM servers LIMIT 5;

SELECT '冗余字段添加完成！' AS message;
