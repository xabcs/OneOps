-- 重新初始化凭证冗余字段（包含ID、名称和类型）
-- 执行方式: mysql -u root -p123456 nexops < backend/scripts/reinit_credential_type.sql

USE nexops;

-- 更新凭证冗余字段，包含credential_type
UPDATE servers s
SET s.credential_names = (
    SELECT CONCAT('[',
           GROUP_CONCAT(
               JSON_OBJECT(
                   'id', c.id,
                   'name', c.name,
                   'credential_type', c.credential_type
               )
               ORDER BY c.id
           ),
           ']')
    FROM server_credentials sc
    JOIN ssh_credentials c ON sc.credential_id = c.id
    WHERE sc.server_id = s.id
);

-- 验证数据
SELECT id, hostname, credential_names FROM servers LIMIT 3;

SELECT '凭证类型字段重新初始化完成！' AS message;
