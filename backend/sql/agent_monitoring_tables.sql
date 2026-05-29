-- OneOps Agent 监控增强 - 数据库表结构
-- 版本：v1.0
-- 创建日期：2026-05-26
-- 说明：支持 Agent 扩展指标、告警管理、通知渠道等功能

-- ============================================
-- 1. Agent 指标存储表（热数据）
-- ============================================
CREATE TABLE IF NOT EXISTS agent_metrics (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    server_id INT UNSIGNED NOT NULL COMMENT '主机ID',
    metric_type VARCHAR(50) NOT NULL COMMENT '指标类型: performance/system/asset/security/process/network',
    metric_data JSON NOT NULL COMMENT '指标数据(JSON格式)',
    report_time DATETIME NOT NULL COMMENT '指标上报时间',
    received_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '接收时间',

    INDEX idx_server_time (server_id, report_time),
    INDEX idx_type_time (metric_type, report_time),
    INDEX idx_report_time (report_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Agent指标表（热数据-30天）';

-- 分区定义（按月分区）- 暂时禁用，需要在生产环境根据数据量启用
-- ALTER TABLE agent_metrics
-- PARTITION BY RANGE (TO_DAYS(report_time)) (
--     PARTITION p_2024_05 VALUES LESS THAN (TO_DAYS('2024-06-01')),
--     PARTITION p_2024_06 VALUES LESS THAN (TO_DAYS('2024-07-01')),
--     PARTITION p_2024_07 VALUES LESS THAN (TO_DAYS('2024-08-01')),
--     PARTITION p_2024_08 VALUES LESS THAN (TO_DAYS('2024-09-01')),
--     PARTITION p_2024_09 VALUES LESS THAN (TO_DAYS('2024-10-01')),
--     PARTITION p_2024_10 VALUES LESS THAN (TO_DAYS('2024-11-01')),
--     PARTITION p_2024_11 VALUES LESS THAN (TO_DAYS('2024-12-01')),
--     PARTITION p_2024_12 VALUES LESS THAN (TO_DAYS('2025-01-01')),
--     PARTITION p_future VALUES LESS THAN MAXVALUE
-- );

-- ============================================
-- 2. Agent 指标归档表（冷数据）
-- ============================================
CREATE TABLE IF NOT EXISTS agent_metrics_archive (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    server_id INT UNSIGNED NOT NULL COMMENT '主机ID',
    metric_type VARCHAR(50) NOT NULL COMMENT '指标类型',
    metric_data JSON NOT NULL COMMENT '指标数据(JSON格式)',
    report_time DATETIME NOT NULL COMMENT '指标上报时间',
    archived_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '归档时间',

    INDEX idx_server_time (server_id, report_time),
    INDEX idx_archived_at (archived_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Agent指标归档表（温数据-90天）';

-- 分区定义（按年分区）- 暂时禁用，需要在生产环境根据数据量启用
-- ALTER TABLE agent_metrics_archive
-- PARTITION BY RANGE (TO_DAYS(report_time)) (
--     PARTITION p_2024 VALUES LESS THAN (TO_DAYS('2025-01-01')),
--     PARTITION p_2025 VALUES LESS THAN (TO_DAYS('2026-01-01')),
--     PARTITION p_future VALUES LESS THAN MAXVALUE
-- );

-- ============================================
-- 3. 告警表
-- ============================================
CREATE TABLE IF NOT EXISTS agent_alerts (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    server_id INT UNSIGNED NOT NULL COMMENT '主机ID',
    rule_id VARCHAR(50) NOT NULL COMMENT '规则ID',
    level VARCHAR(20) NOT NULL COMMENT '告警级别: critical/high/medium/low/info',
    message VARCHAR(500) NOT NULL COMMENT '告警消息',
    metric_value DECIMAL(10,2) COMMENT '当前指标值',
    threshold DECIMAL(10,2) COMMENT '告警阈值',
    first_seen DATETIME NOT NULL COMMENT '首次发生时间',
    last_seen DATETIME NOT NULL COMMENT '最后发生时间',
    acknowledged BOOLEAN DEFAULT FALSE COMMENT '是否已确认',
    acknowledged_by VARCHAR(100) COMMENT '确认人',
    acknowledged_at DATETIME COMMENT '确认时间',
    resolved_at DATETIME COMMENT '解决时间',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',

    INDEX idx_server_level (server_id, level),
    INDEX idx_acknowledged (acknowledged, first_seen),
    INDEX idx_first_seen (first_seen),
    INDEX idx_level_status (level, acknowledged, resolved_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Agent告警表';

-- ============================================
-- 4. 告警规则配置表
-- ============================================
CREATE TABLE IF NOT EXISTS agent_alert_rules (
    id VARCHAR(50) PRIMARY KEY COMMENT '规则ID（唯一标识）',
    name VARCHAR(100) NOT NULL COMMENT '规则名称',
    level VARCHAR(20) NOT NULL COMMENT '告警级别: critical/high/medium/low/info',
    metric VARCHAR(50) NOT NULL COMMENT '监控指标',
    `condition` VARCHAR(10) NOT NULL COMMENT '判断条件: >/<>/==/!=',
    threshold DECIMAL(10,2) NOT NULL COMMENT '阈值',
    duration INT NOT NULL DEFAULT 0 COMMENT '持续时间（秒）',
    description TEXT COMMENT '告警描述',
    enabled BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    INDEX idx_level_enabled (level, enabled),
    INDEX idx_enabled (enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Agent告警规则配置表';

-- 插入默认告警规则
INSERT INTO agent_alert_rules (id, name, level, metric, `condition`, threshold, duration, description) VALUES
('cpu_critical', 'CPU严重告警', 'critical', 'cpu_usage', '>', 95.0, 600, 'CPU使用率持续超过95%，可能影响业务性能'),
('cpu_high', 'CPU高负载告警', 'high', 'cpu_usage', '>', 80.0, 1800, 'CPU使用率超过80%，需关注'),
('memory_critical', '内存严重告警', 'critical', 'memory_usage', '>', 95.0, 300, '内存使用率超过95%，可能导致OOM'),
('memory_high', '内存使用率过高', 'high', 'memory_usage', '>', 85.0, 1800, '内存使用率超过85%，需关注'),
('disk_full', '磁盘空间告警', 'critical', 'disk_usage', '>', 95.0, 0, '磁盘空间不足95%，可能导致服务异常'),
('disk_high', '磁盘空间不足', 'high', 'disk_usage', '>', 85.0, 0, '磁盘空间超过85%，需关注'),
('load_high', '系统负载过高', 'critical', 'load5', '>', 4.0, 900, '系统负载过高，响应可能变慢')
ON DUPLICATE KEY UPDATE updated_at = CURRENT_TIMESTAMP;

-- ============================================
-- 5. 通知渠道配置表
-- ============================================
CREATE TABLE IF NOT EXISTS notification_channels (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    channel_type VARCHAR(20) NOT NULL COMMENT '渠道类型: wechat/dingtalk/email/sms',
    channel_name VARCHAR(100) NOT NULL COMMENT '渠道名称',
    config JSON NOT NULL COMMENT '渠道配置(JSON格式)',
    enabled BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    INDEX idx_type_enabled (channel_type, enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='通知渠道配置表';

-- ============================================
-- 6. 通知历史表
-- ============================================
CREATE TABLE IF NOT EXISTS notification_history (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    alert_id BIGINT UNSIGNED NOT NULL COMMENT '告警ID',
    channel_id INT UNSIGNED NOT NULL COMMENT '通知渠道ID',
    status VARCHAR(20) NOT NULL COMMENT '发送状态: success/failed/pending',
    message TEXT COMMENT '通知内容',
    error_msg TEXT COMMENT '错误信息',
    sent_at DATETIME COMMENT '发送时间',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',

    INDEX idx_alert_id (alert_id),
    INDEX idx_channel_status (channel_id, status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='通知历史表';

-- ============================================
-- 7. 巡检报告表
-- ============================================
CREATE TABLE IF NOT EXISTS inspection_reports (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    report_type VARCHAR(50) NOT NULL COMMENT '报告类型: host/security/capacity',
    title VARCHAR(200) NOT NULL COMMENT '报告标题',
    server_ids JSON COMMENT '主机ID列表(JSON数组)',
    report_data JSON NOT NULL COMMENT '报告数据(JSON格式)',
    status VARCHAR(20) DEFAULT 'pending' COMMENT '状态: pending/generating/completed/failed',
    created_by VARCHAR(100) COMMENT '创建人',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    completed_at DATETIME COMMENT '完成时间',

    INDEX idx_type_status (report_type, status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='巡检报告表';

-- ============================================
-- 8. 配置变更记录表
-- ============================================
CREATE TABLE IF NOT EXISTS config_changes (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    server_id INT UNSIGNED NOT NULL COMMENT '主机ID',
    file_path VARCHAR(500) NOT NULL COMMENT '配置文件路径',
    old_checksum VARCHAR(64) COMMENT '旧校验和',
    new_checksum VARCHAR(64) COMMENT '新校验和',
    change_type VARCHAR(20) NOT NULL COMMENT '变更类型: created/modified/deleted',
    detected_at DATETIME NOT NULL COMMENT '检测时间',

    INDEX idx_server_time (server_id, detected_at),
    INDEX idx_file_path (file_path)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='配置变更记录表';

-- ============================================
-- 9. 数据归档存储过程
-- ============================================
-- 注意：存储过程和事件调度器需要手动创建，或者通过后端代码实现归档逻辑
-- 以下代码需要在 MySQL 客户端中直接执行，不能通过管道执行

-- DELIMITER $$
--
-- -- 归档30天前的指标数据到归档表
-- CREATE PROCEDURE IF NOT EXISTS archive_old_metrics()
-- BEGIN
--     DECLARE cutoff_date DATETIME;
--     SET cutoff_date = DATE_SUB(NOW(), INTERVAL 30 DAY);
--
--     -- 将热数据迁移到归档表
--     INSERT INTO agent_metrics_archive
--         (server_id, metric_type, metric_data, report_time)
--     SELECT server_id, metric_type, metric_data, report_time
--     FROM agent_metrics
--     WHERE report_time < cutoff_date;
--
--     -- 删除已归档的热数据
--     DELETE FROM agent_metrics
--     WHERE report_time < cutoff_date;
--
--     -- 记录归档数量
--     SELECT ROW_COUNT() AS archived_count;
-- END$$
--
-- -- 清理365天前的归档数据
-- CREATE PROCEDURE IF NOT EXISTS cleanup_old_archived_metrics()
-- BEGIN
--     DECLARE cutoff_date DATETIME;
--     SET cutoff_date = DATE_SUB(NOW(), INTERVAL 365 DAY);
--
--     DELETE FROM agent_metrics_archive
--     WHERE report_time < cutoff_date;
--
--     SELECT ROW_COUNT() AS deleted_count;
-- END$$
--
-- -- 创建定时归档事件（需手动启用）
-- CREATE EVENT IF NOT EXISTS archive_metrics_event
-- ON SCHEDULE EVERY 1 DAY
-- STARTS '2026-05-26 02:00:00'
-- DO
--   CALL archive_old_metrics();
-- END$$
--
-- CREATE EVENT IF NOT EXISTS cleanup_archived_metrics_event
-- ON SCHEDULE EVERY 7 DAY
-- STARTS '2026-05-26 03:00:00'
-- DO
--   CALL cleanup_old_archived_metrics();
-- END$$
--
-- DELIMITER ;

-- ============================================
-- 10. 视图定义（可选，用于复杂查询）
-- ============================================

-- 主机最新指标视图
CREATE OR REPLACE VIEW v_server_latest_metrics AS
SELECT
    server_id,
    MAX(report_time) as latest_report_time
FROM agent_metrics
GROUP BY server_id;

-- 活跃告警视图
CREATE OR REPLACE VIEW v_active_alerts AS
SELECT
    a.*,
    s.hostname,
    s.ip
FROM agent_alerts a
JOIN servers s ON a.server_id = s.id
WHERE a.acknowledged = FALSE
  AND a.resolved_at IS NULL
ORDER BY a.level DESC, a.first_seen DESC;

-- ============================================
-- 11. 索引优化说明
-- ============================================
-- agent_metrics 表索引使用说明：
-- idx_server_time: 查询特定主机的指标（按时间倒序）
-- idx_type_time: 按类型查询指标（按时间倒序）
-- idx_report_time: 时间范围查询

-- agent_alerts 表索引使用说明：
-- idx_server_level: 查询特定主机的告警（按级别排序）
-- idx_acknowledged: 查询未确认的告警
-- idx_level_status: 多条件查询（级别+确认状态+解决状态）

-- ============================================
-- 12. 数据字典
-- ============================================

-- metric_type 枚举值说明
-- performance: 性能指标（CPU/内存/磁盘/网络IO详情）
-- system: 系统信息（OS/内核/主机名/运行时间）
-- asset: 资产信息（CPU型号/内存插槽/磁盘型号）
-- service: 服务状态（systemd服务/监听端口）
-- process: 进程信息（Top 20 进程列表）
-- network: 网络配置（网卡/IP/路由/DNS）
-- security: 安全信息（SSH/防火墙/用户/登录）
-- config: 配置变更记录

-- level 枚举值说明
-- critical: 严重（立即处理）
-- high: 高级（1小时内）
-- medium: 中级（当天处理）
-- low: 低级（知晓即可）
-- info: 信息（仅供参考）

-- channel_type 枚举值说明
-- wechat: 企业微信
-- dingtalk: 钉钉
-- email: 邮件
-- sms: 短信

-- ============================================
-- 13. 数据清理策略
-- ============================================
-- 热数据保留期：30天
-- 温数据保留期：90天
-- 冷数据保留期：365天
-- 清理频率：热数据每天归档，冷数据每周清理

-- ============================================
-- 14. 备注说明
-- ============================================
-- 所有 JSON 字段使用 MySQL 5.7+ 的 JSON 类型支持
-- 分区表自动管理，需定期添加新分区
-- 通知历史表保留90天，超期自动清理
-- 配置变更记录保留180天
