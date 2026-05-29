-- ========================================
-- Agent 版本管理系统数据库迁移脚本
-- 创建时间: 2026-05-28
-- 版本: v1.0
-- ========================================

-- Agent 版本表
CREATE TABLE IF NOT EXISTS agent_versions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '版本ID',
    version VARCHAR(50) NOT NULL UNIQUE COMMENT '版本号 (如: 1.0.0, 1.1.0)',
    release_notes TEXT COMMENT '发布说明',
    changelog TEXT COMMENT '更新日志',
    released_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '发布时间',

    -- 二进制文件信息
    amd64_binary_path VARCHAR(255) COMMENT 'AMD64 二进制文件路径',
    amd64_binary_hash VARCHAR(64) COMMENT 'AMD64 文件 SHA256 哈希',
    amd64_binary_size BIGINT COMMENT 'AMD64 文件大小(字节)',
    arm64_binary_path VARCHAR(255) COMMENT 'ARM64 二进制文件路径',
    arm64_binary_hash VARCHAR(64) COMMENT 'ARM64 文件 SHA256 哈希',
    arm64_binary_size BIGINT COMMENT 'ARM64 文件大小(字节)',

    -- 版本状态
    is_latest TINYINT(1) DEFAULT 0 COMMENT '是否为最新版本',
    is_deprecated TINYINT(1) DEFAULT 0 COMMENT '是否已弃用',

    -- 功能支持标记
    features JSON COMMENT '支持的功能列表 {"extended_metrics":true,"custom_configs":false}',
    min_compatible_version VARCHAR(50) COMMENT '最小兼容版本',
    max_compatible_version VARCHAR(50) COMMENT '最大兼容版本',

    -- 统计信息
    download_count INT DEFAULT 0 COMMENT '下载次数',
    deploy_count INT DEFAULT 0 COMMENT '部署次数',

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    INDEX idx_version(version),
    INDEX idx_is_latest(is_latest),
    INDEX idx_released_at(released_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Agent版本表';

-- Agent 升级任务表
CREATE TABLE IF NOT EXISTS agent_upgrade_tasks (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '任务ID',
    task_name VARCHAR(100) COMMENT '任务名称',

    -- 升级目标信息
    target_version VARCHAR(50) NOT NULL COMMENT '目标版本',
    target_server_ids JSON COMMENT '目标主机ID列表 [1,2,3]',

    -- 任务状态
    status ENUM('pending','running','completed','failed','cancelled') DEFAULT 'pending' COMMENT '任务状态',
    current_step INT DEFAULT 0 COMMENT '当前步骤',
    total_steps INT DEFAULT 0 COMMENT '总步骤数',

    -- 进度统计
    total_count INT DEFAULT 0 COMMENT '总主机数',
    success_count INT DEFAULT 0 COMMENT '成功数量',
    failed_count INT DEFAULT 0 COMMENT '失败数量',
    skipped_count INT DEFAULT 0 COMMENT '跳过数量（已是最新版本）',

    -- 时间记录
    started_at TIMESTAMP NULL COMMENT '开始时间',
    completed_at TIMESTAMP NULL COMMENT '完成时间',

    -- 详细日志
    error_message TEXT COMMENT '错误信息',
    operation_log JSON COMMENT '操作日志 [{"serverId":1,"status":"success","time":"2026-05-28 10:00:00"}]',

    created_by VARCHAR(50) COMMENT '创建人',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    INDEX idx_status(status),
    INDEX idx_created_at(created_at),
    INDEX idx_created_by(created_by)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Agent升级任务表';

-- 插入默认版本 1.0.0
INSERT INTO agent_versions (
    version,
    release_notes,
    changelog,
    is_latest,
    features,
    download_count,
    deploy_count
) VALUES (
    '1.0.0',
    'Agent 初始版本，支持基础监控指标采集',
    '- 支持基础监控指标采集\n- 支持心跳上报\n- 支持 /metrics 端点',
    1,
    '{"extended_metrics":false,"custom_configs":false}',
    0,
    0
) ON DUPLICATE KEY UPDATE updated_at = NOW();
