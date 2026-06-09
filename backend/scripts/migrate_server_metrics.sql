-- 分离监控指标到独立表
-- 执行方式: mysql -u root -p123456 nexops < backend/scripts/migrate_server_metrics.sql

USE nexops;

-- 1. 创建监控指标表（不使用外键约束，避免迁移问题）
CREATE TABLE IF NOT EXISTS server_metrics (
    server_id INT UNSIGNED PRIMARY KEY COMMENT '服务器ID',
    cpu_usage FLOAT DEFAULT 0 COMMENT 'CPU使用率',
    memory_usage FLOAT DEFAULT 0 COMMENT '内存使用率',
    disk_usage FLOAT DEFAULT 0 COMMENT '磁盘使用率',
    load1 FLOAT DEFAULT 0 COMMENT '1分钟负载',
    load5 FLOAT DEFAULT 0 COMMENT '5分钟负载',
    load15 FLOAT DEFAULT 0 COMMENT '15分钟负载',
    disk_partitions JSON COMMENT '磁盘分区信息',
    metrics_updated_at TIMESTAMP NULL COMMENT '指标更新时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务器监控指标表';

-- 2. 迁移现有监控数据
INSERT IGNORE INTO server_metrics (server_id, cpu_usage, memory_usage, disk_usage, load1, load5, load15, disk_partitions, metrics_updated_at)
SELECT
    id,
    cpu_usage,
    memory_usage,
    disk_usage,
    load1,
    load5,
    load15,
    disk_partitions,
    metrics_updated_at
FROM servers
WHERE cpu_usage > 0 OR memory_usage > 0 OR disk_usage > 0;

-- 3. 创建索引优化查询
CREATE INDEX idx_metrics_updated ON server_metrics(metrics_updated_at);

-- 4. 验证数据
SELECT COUNT(*) as '迁移的监控记录数' FROM server_metrics;

SELECT '监控表分离完成！' AS message;
