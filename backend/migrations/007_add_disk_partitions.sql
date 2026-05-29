-- 添加磁盘分区信息字段
-- 执行日期: 2026-05-25
-- 描述: 在 servers 表中添加 disk_partitions JSON 字段，用于存储磁盘分区信息

ALTER TABLE servers
ADD COLUMN disk_partitions JSON NULL COMMENT '磁盘分区信息 [{"mount":"/","usage":80.5},{"mount":"/var","usage":90.2}]' AFTER disk_usage;

-- 验证字段是否添加成功
SELECT
    COLUMN_NAME,
    DATA_TYPE,
    COLUMN_TYPE,
    IS_NULLABLE,
    COLUMN_DEFAULT,
    COLUMN_COMMENT
FROM INFORMATION_SCHEMA.COLUMNS
WHERE TABLE_SCHEMA = 'ops'
  AND TABLE_NAME = 'servers'
  AND COLUMN_NAME = 'disk_partitions';
