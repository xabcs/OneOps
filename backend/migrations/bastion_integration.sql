-- 堡垒机能力集成 - 数据库迁移脚本
-- 执行方式: mysql -u root -p nexops < backend/migrations/bastion_integration.sql

-- 1. 创建访问策略表
CREATE TABLE IF NOT EXISTS asset_access_policies (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL COMMENT '策略名称',
    subject_type ENUM('user', 'role', 'user_group') NOT NULL COMMENT '授权对象类型',
    subject_id BIGINT UNSIGNED NOT NULL COMMENT '授权对象ID',
    asset_scope_type ENUM('server', 'group', 'business', 'tag', 'all') NOT NULL COMMENT '资产范围类型',
    asset_scope_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '资产范围ID，0表示全部',
    login_accounts JSON COMMENT '允许使用的登录账号 ["root", "appuser"]',
    protocols JSON COMMENT '允许的协议 ["ssh", "sftp"]',
    allow_file_transfer BOOLEAN DEFAULT TRUE COMMENT '是否允许文件传输',
    allow_sudo BOOLEAN DEFAULT FALSE COMMENT '是否允许sudo',
    require_approval BOOLEAN DEFAULT FALSE COMMENT '是否需要审批',
    time_window JSON COMMENT '时间窗口 {"start": "09:00", "end": "18:00", "days": [1,2,3,4,5]}',
    high_risk_commands JSON COMMENT '高危命令列表 ["rm -rf", "mkfs", "shutdown"]',
    status TINYINT DEFAULT 1 COMMENT '状态 1启用 0禁用',
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    INDEX idx_subject (subject_type, subject_id),
    INDEX idx_asset_scope (asset_scope_type, asset_scope_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='资产访问策略表';

-- 2. 创建会话表
CREATE TABLE IF NOT EXISTS bastion_sessions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    server_id BIGINT UNSIGNED NOT NULL COMMENT '服务器ID',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    username VARCHAR(50) NOT NULL COMMENT '用户名',
    login_account VARCHAR(50) NOT NULL COMMENT '登录账号',
    client_ip VARCHAR(50) COMMENT '客户端IP',
    protocol ENUM('ssh', 'sftp') DEFAULT 'ssh' COMMENT '连接协议',
    ssh_credential_id BIGINT UNSIGNED COMMENT '使用的SSH凭证ID',
    started_at DATETIME(3) COMMENT '开始时间',
    ended_at DATETIME(3) COMMENT '结束时间',
    duration INT DEFAULT 0 COMMENT '持续时长(秒)',
    status ENUM('active', 'closed', 'error', 'terminated') DEFAULT 'active' COMMENT '会话状态',
    close_reason VARCHAR(200) COMMENT '关闭原因',
    created_at DATETIME(3) NULL,
    INDEX idx_server_user (server_id, user_id),
    INDEX idx_user (user_id),
    INDEX idx_status (status),
    INDEX idx_started_at (started_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='堡垒机会话表';

-- 3. 创建命令审计表
CREATE TABLE IF NOT EXISTS bastion_commands (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    session_id BIGINT UNSIGNED NOT NULL COMMENT '会话ID',
    command TEXT NOT NULL COMMENT '执行的命令',
    executed_at DATETIME(3) COMMENT '执行时间',
    exit_code INT COMMENT '退出码',
    risk_level ENUM('safe', 'low', 'medium', 'high', 'critical') DEFAULT 'safe' COMMENT '风险等级',
    blocked BOOLEAN DEFAULT FALSE COMMENT '是否被拦截',
    output_summary TEXT COMMENT '输出摘要',
    created_at DATETIME(3) NULL,
    INDEX idx_session (session_id),
    INDEX idx_risk_level (risk_level),
    INDEX idx_executed_at (executed_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='命令审计表';

-- 4. 创建文件传输审计表
CREATE TABLE IF NOT EXISTS bastion_file_transfers (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    session_id BIGINT UNSIGNED NOT NULL COMMENT '会话ID',
    direction ENUM('upload', 'download') NOT NULL COMMENT '传输方向',
    remote_path VARCHAR(500) NOT NULL COMMENT '远程路径',
    local_path VARCHAR(500) COMMENT '本地路径',
    file_size BIGINT DEFAULT 0 COMMENT '文件大小(字节)',
    status ENUM('pending', 'transferring', 'success', 'failed') DEFAULT 'pending' COMMENT '传输状态',
    error_message VARCHAR(500) COMMENT '错误信息',
    started_at DATETIME(3) COMMENT '开始时间',
    completed_at DATETIME(3) COMMENT '完成时间',
    created_at DATETIME(3) NULL,
    INDEX idx_session (session_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文件传输审计表';

-- 5. 创建审批表
CREATE TABLE IF NOT EXISTS bastion_approvals (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    requester_id BIGINT UNSIGNED NOT NULL COMMENT '申请人ID',
    approver_id BIGINT UNSIGNED COMMENT '审批人ID',
    server_id BIGINT UNSIGNED NOT NULL COMMENT '服务器ID',
    login_account VARCHAR(50) NOT NULL COMMENT '登录账号',
    reason TEXT COMMENT '申请原因',
    status ENUM('pending', 'approved', 'rejected', 'expired') DEFAULT 'pending' COMMENT '审批状态',
    requested_at DATETIME(3) COMMENT '申请时间',
    approved_at DATETIME(3) COMMENT '审批时间',
    expired_at DATETIME(3) COMMENT '过期时间',
    comment TEXT COMMENT '审批意见',
    created_at DATETIME(3) NULL,
    INDEX idx_requester (requester_id),
    INDEX idx_approver (approver_id),
    INDEX idx_status (status),
    INDEX idx_server (server_id),
    INDEX idx_expired_at (expired_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='连接审批表';

-- 6. 扩展服务器表
-- 检查列是否存在，不存在则添加
SET @dbname = DATABASE();
SET @tablename = 'servers';
SET @columnname1 = 'ssh_credential_id';
SET @columnname2 = 'last_connect_time';
SET @columnname3 = 'connectivity_status';
SET @columnname4 = 'last_check_time';

SET @preparedStatement1 = (SELECT IF(
  (
    SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = @dbname
    AND TABLE_NAME = @tablename
    AND COLUMN_NAME = @columnname1
  ) > 0,
  'SELECT 1',
  CONCAT('ALTER TABLE ', @tablename, ' ADD COLUMN ssh_credential_id BIGINT UNSIGNED NULL COMMENT ''绑定的SSH凭证ID'' AFTER provider, ADD INDEX idx_ssh_credential (ssh_credential_id)')
));
PREPARE alterIfNotExists1 FROM @preparedStatement1;
EXECUTE alterIfNotExists1;
DEALLOCATE PREPARE alterIfNotExists1;

SET @preparedStatement2 = (SELECT IF(
  (
    SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = @dbname
    AND TABLE_NAME = @tablename
    AND COLUMN_NAME = @columnname2
  ) > 0,
  'SELECT 1',
  CONCAT('ALTER TABLE ', @tablename, ' ADD COLUMN last_connect_time DATETIME(3) COMMENT ''最后连接时间'' AFTER ssh_credential_id')
));
PREPARE alterIfNotExists2 FROM @preparedStatement2;
EXECUTE alterIfNotExists2;
DEALLOCATE PREPARE alterIfNotExists2;

SET @preparedStatement3 = (SELECT IF(
  (
    SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = @dbname
    AND TABLE_NAME = @tablename
    AND COLUMN_NAME = @columnname3
  ) > 0,
  'SELECT 1',
  CONCAT('ALTER TABLE ', @tablename, ' ADD COLUMN connectivity_status ENUM(''online'', ''offline'', ''unknown'') DEFAULT ''unknown'' COMMENT ''连通性状态'' AFTER last_connect_time')
));
PREPARE alterIfNotExists3 FROM @preparedStatement3;
EXECUTE alterIfNotExists3;
DEALLOCATE PREPARE alterIfNotExists3;

SET @preparedStatement4 = (SELECT IF(
  (
    SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = @dbname
    AND TABLE_NAME = @tablename
    AND COLUMN_NAME = @columnname4
  ) > 0,
  'SELECT 1',
  CONCAT('ALTER TABLE ', @tablename, ' ADD COLUMN last_check_time DATETIME(3) COMMENT ''最后连通性检查时间'' AFTER connectivity_status')
));
PREPARE alterIfNotExists4 FROM @preparedStatement4;
EXECUTE alterIfNotExists4;
DEALLOCATE PREPARE alterIfNotExists4;

-- 7. 插入默认权限数据
-- 检查权限是否已存在
INSERT IGNORE INTO permissions (code, name, description, created_at, updated_at)
VALUES
('cmdb:server:connect', '连接服务器', '允许通过堡垒机连接服务器', NOW(), NOW()),
('cmdb:session:query', '查看会话', '允许查看会话列表和详情', NOW(), NOW()),
('cmdb:session:terminate', '断开会话', '允许强制断开会话', NOW(), NOW()),
('cmdb:session:replay', '回放会话', '允许回放会话记录', NOW(), NOW()),
('cmdb:command:query', '查看命令', '允许查看命令审计', NOW(), NOW()),
('cmdb:file-transfer:query', '查看文件传输', '允许查看文件传输记录', NOW(), NOW()),
('cmdb:access-policy:query', '查看访问策略', '允许查看访问策略', NOW(), NOW()),
('cmdb:access-policy:create', '创建访问策略', '允许创建访问策略', NOW(), NOW()),
('cmdb:access-policy:update', '更新访问策略', '允许更新访问策略', NOW(), NOW()),
('cmdb:access-policy:delete', '删除访问策略', '允许删除访问策略', NOW(), NOW()),
('cmdb:approval:query', '查看审批', '允许查看审批申请', NOW(), NOW()),
('cmdb:approval:approve', '审批申请', '允许审批连接申请', NOW(), NOW());

-- 8. 为超级管理员角色添加新权限
-- 获取超级管理员角色的ID
SET @super_role_id = (SELECT id FROM roles WHERE code = 'admin' LIMIT 1);

-- 如果存在超级管理员角色，为其添加所有新权限
-- 这部分需要根据实际的 role_permissions 表结构来调整
-- 如果是 JSON 格式存储在 permissions 字段中：
UPDATE roles
SET permissions = JSON_MERGE_PRESERVE(
    COALESCE(permissions, JSON_ARRAY()),
    (SELECT JSON_ARRAYAGG(code) FROM permissions WHERE code LIKE 'cmdb:%' AND code NOT IN (SELECT JSON_UNQUOTE(JSON_EXTRACT(permissions, CONCAT('$[', seq, ']'))) FROM roles, (SELECT 0 AS seq UNION SELECT 1 UNION SELECT 2 UNION SELECT 3 UNION SELECT 4 UNION SELECT 5) seq WHERE JSON_EXTRACT(permissions, CONCAT('$[', seq, ']')) IS NOT NULL))
)
WHERE id = @super_role_id;

-- 9. 创建默认的访问策略示例（可选）
INSERT INTO asset_access_policies (
    name, subject_type, subject_id,
    asset_scope_type, asset_scope_id,
    login_accounts, protocols,
    allow_file_transfer, allow_sudo, require_approval,
    status, created_at, updated_at
) VALUES (
    '管理员全部权限', 'role', @super_role_id,
    'all', 0,
    JSON_ARRAY('root', 'admin'),
    JSON_ARRAY('ssh', 'sftp'),
    TRUE, TRUE, FALSE,
    1, NOW(), NOW()
) ON DUPLICATE KEY UPDATE updated_at = NOW();

-- 10. 创建用于会话清理的存储过程（可选）
DELIMITER $$

CREATE PROCEDURE IF NOT EXISTS CleanOldSessions()
BEGIN
    DECLARE cleanup_days INT DEFAULT 90;
    DECLARE cleanup_date DATETIME;

    SET cleanup_date = DATE_SUB(NOW(), INTERVAL cleanup_days DAY);

    -- 删除旧的已关闭会话的命令记录
    DELETE FROM bastion_commands
    WHERE session_id IN (
        SELECT id FROM bastion_sessions
        WHERE status IN ('closed', 'error', 'terminated')
        AND ended_at < cleanup_date
    );

    -- 删除旧的已关闭会话的文件传输记录
    DELETE FROM bastion_file_transfers
    WHERE session_id IN (
        SELECT id FROM bastion_sessions
        WHERE status IN ('closed', 'error', 'terminated')
        AND ended_at < cleanup_date
    );

    -- 删除旧的已关闭会话
    DELETE FROM bastion_sessions
    WHERE status IN ('closed', 'error', 'terminated')
    AND ended_at < cleanup_date;

    -- 删除过期的审批记录
    DELETE FROM bastion_approvals
    WHERE expired_at < NOW()
    OR (status = 'approved' AND approved_at < DATE_SUB(NOW(), INTERVAL 30 DAY));

END$$

DELIMITER ;

-- 执行迁移完成提示
SELECT '堡垒机能力集成数据库迁移完成！' AS message;
SELECT '新创建的表:' AS info;
SHOW TABLES LIKE 'bastion_%';
SHOW TABLES LIKE 'asset_access_policies';
