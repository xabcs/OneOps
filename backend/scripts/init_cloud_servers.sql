-- ========================================
-- OneOps 云主机表
-- 创建时间: 2026-05-19
-- 版本: v1.0
-- ========================================

-- 云主机表（继承自servers表，存储云平台特有信息）
CREATE TABLE IF NOT EXISTS cloud_servers (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '云主机ID',
    server_id BIGINT NOT NULL COMMENT '关联服务器ID',
    provider VARCHAR(50) NOT NULL COMMENT '云服务商：aliyun/tencent/aws/huawei',
    instance_id VARCHAR(100) COMMENT '云主机实例ID',
    instance_type VARCHAR(50) COMMENT '实例规格',
    instance_name VARCHAR(100) COMMENT '实例名称',
    region VARCHAR(50) COMMENT '地域',
    zone VARCHAR(50) COMMENT '可用区',
    vpc_id VARCHAR(100) COMMENT 'VPC ID',
    subnet_id VARCHAR(100) COMMENT '子网ID',
    public_ip VARCHAR(50) COMMENT '公网IP',
    private_ip VARCHAR(50) COMMENT '内网IP',
    security_groups TEXT COMMENT '安全组（JSON格式）',
    charge_type VARCHAR(20) COMMENT '计费类型：postpay按量付费/prepay包年包月',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_server_id (server_id),
    UNIQUE KEY uk_provider_instance (provider, instance_id),
    INDEX idx_provider(provider),
    INDEX idx_region(region),
    FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='云主机表';

-- 查看创建的表结构
DESCRIBE cloud_servers;
