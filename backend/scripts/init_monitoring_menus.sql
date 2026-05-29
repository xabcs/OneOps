-- 监控中心菜单初始化脚本
-- 执行方式: mysql -u root -p ops < backend/scripts/init_monitoring_menus.sql

USE ops;

-- 1. 创建监控中心一级菜单
INSERT INTO menus (name, code, route_key, parent_id, icon, sort_order, status, created_at, updated_at)
VALUES (
    '监控中心',
    'monitoring',
    'monitoring',
    0,
    'mdi:monitor-dashboard',
    50,
    1,
    NOW(),
    NOW()
);

-- 获取监控中心菜单的ID（刚插入的ID）
SET @monitoring_id = LAST_INSERT_ID();

-- 2. 创建监控中心子菜单
INSERT INTO menus (name, code, route_key, parent_id, icon, sort_order, status, created_at, updated_at) VALUES
('监控概览', 'monitoring_overview', 'monitoring_overview', @monitoring_id, 'mdi:chart-line', 1, 1, NOW(), NOW()),
('主机监控', 'monitoring_servers', 'monitoring_servers', @monitoring_id, 'mdi:server-network', 2, 1, NOW(), NOW()),
('主机监控详情', 'monitoring_servers-detail', 'monitoring_servers-detail', @monitoring_id, 'mdi:information', 3, 1, NOW(), NOW()),
('趋势分析', 'monitoring_trends', 'monitoring_trends', @monitoring_id, 'mdi:chart-areaspline', 4, 1, NOW(), NOW()),
('告警管理', 'monitoring_alerts', 'monitoring_alerts', @monitoring_id, 'mdi:alert-circle', 5, 1, NOW(), NOW()),
('监控设置', 'monitoring_settings', 'monitoring_settings', @monitoring_id, 'mdi:cog', 6, 1, NOW(), NOW()),
('巡检报告', 'monitoring_reports', 'monitoring_reports', @monitoring_id, 'mdi:file-document', 7, 1, NOW(), NOW());

-- 3. 将监控中心菜单分配给超级管理员角色（假设角色ID为1）
-- 首先获取所有监控相关菜单的ID
SELECT @overview_id := id FROM menus WHERE code = 'monitoring_overview';
SELECT @servers_id := id FROM menus WHERE code = 'monitoring_servers';
SELECT @servers_detail_id := id FROM menus WHERE code = 'monitoring_servers-detail';
SELECT @trends_id := id FROM menus WHERE code = 'monitoring_trends';
SELECT @alerts_id := id FROM menus WHERE code = 'monitoring_alerts';
SELECT @settings_id := id FROM menus WHERE code = 'monitoring_settings';
SELECT @reports_id := id FROM menus WHERE code = 'monitoring_reports';

-- 更新超级管理员角色的菜单列表（角色ID=1）
UPDATE roles
SET menu_ids = JSON_ARRAY(
    @monitoring_id, @overview_id, @servers_id, @servers_detail_id,
    @trends_id, @alerts_id, @settings_id, @reports_id
)
WHERE id = 1;

-- 如果需要给其他角色也分配监控菜单，可以执行以下SQL：
-- UPDATE roles SET menu_ids = JSON_ARRAY(...) WHERE id = ?;

-- 4. 验证菜单是否创建成功
SELECT id, name, code, route_key, parent_id, sort_order, status
FROM menus
WHERE code LIKE 'monitoring%'
ORDER BY parent_id, sort_order;

SELECT '监控中心菜单初始化完成！' AS message;
