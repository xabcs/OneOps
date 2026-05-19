-- ========================================
-- OneOps SSH认证凭证表
-- 创建时间: 2026-05-19
-- 版本: v1.0
-- ========================================

-- SSH认证凭证表
CREATE TABLE IF NOT EXISTS ssh_credentials (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '凭证ID',
    name VARCHAR(100) NOT NULL COMMENT '凭证名称',
    description TEXT COMMENT '凭证描述',
    username VARCHAR(50) NOT NULL COMMENT 'SSH用户名',
    auth_type VARCHAR(20) NOT NULL DEFAULT 'password' COMMENT '认证类型：password密钥/key',
    password VARCHAR(255) COMMENT '密码（加密存储）',
    private_key TEXT COMMENT '私钥（加密存储）',
    passphrase VARCHAR(255) COMMENT '私钥密码（加密存储）',
    port INT DEFAULT 22 COMMENT 'SSH端口',
    sort_order INT DEFAULT 0 COMMENT '排序',
    status TINYINT DEFAULT 1 COMMENT '状态：1启用 0禁用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_name(name),
    INDEX idx_status(status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='SSH认证凭证表';

-- 插入示例凭证数据（密码为加密后的值，这里仅作示例）
-- 实际使用时应该通过后端API创建加密的凭证
INSERT INTO ssh_credentials (name, description, username, auth_type, password, port, sort_order) VALUES
('默认root凭证', '系统默认的root登录凭证', 'root', 'password', 'encrypted_password_here', 22, 1),
('应用服务器凭证', '应用服务器的专用登录凭证', 'appuser', 'password', 'encrypted_password_here', 22, 2);

-- 查看创建的凭证
SELECT id, name, username, auth_type, port, status
FROM ssh_credentials
ORDER BY sort_order;
