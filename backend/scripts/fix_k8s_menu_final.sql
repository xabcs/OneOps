-- 修复 K8s 菜单结构 - Tab 切换方案（最终版）
-- 清理混乱的菜单，重新建立简洁的4个入口

-- 步骤1：删除所有K8s相关的菜单（除了K8s管理本身）
DELETE FROM menus WHERE parent_id = 5;
DELETE FROM menus WHERE id IN (80, 86, 87, 88, 89);

-- 步骤2：重新创建简洁的菜单结构
-- 集群管理
INSERT INTO menus (id, name, icon, path, permission, parent_id, sort, status, menu_type, created_at, updated_at)
VALUES (80, '集群管理', 'mdi:server-network', '/k8s/clusters', 'k8s:cluster:query', 5, 1, 1, 'menu', NOW(), NOW());

-- 工作负载（统一入口，页面内Tab切换）
INSERT INTO menus (id, name, icon, path, permission, parent_id, sort, status, menu_type, created_at, updated_at)
VALUES (87, '工作负载', 'mdi:cube-outline', '/k8s/workloads', 'k8s:workload:query', 5, 2, 1, 'menu', NOW(), NOW());

-- 网络（统一入口，页面内Tab切换）
INSERT INTO menus (id, name, icon, path, permission, parent_id, sort, status, menu_type, created_at, updated_at)
VALUES (88, '网络', 'mdi:network-outline', '/k8s/network', 'k8s:network:query', 5, 3, 1, 'menu', NOW(), NOW());

-- 配置管理（统一入口，页面内Tab切换）
INSERT INTO menus (id, name, icon, path, permission, parent_id, sort, status, menu_type, created_at, updated_at)
VALUES (89, '配置管理', 'mdi:cog', '/k8s/config', 'k8s:config:query', 5, 4, 1, 'menu', NOW(), NOW());

-- 注意：会话审计功能已移除，不再在 K8s 管理中显示
-- 会话审计应使用堡垒机模块的功能

-- 步骤3：删除所有挂在不存在父菜单下的子菜单
DELETE FROM menus WHERE parent_id IN (87, 88, 89, 92, 94) AND parent_id NOT IN (SELECT id FROM menus WHERE id IN (87, 88, 89));

-- 验证结果
SELECT
    m.id,
    m.name,
    m.path,
    m.parent_id,
    m.sort,
    m.menu_type,
    p.name as parent_name
FROM menus m
LEFT JOIN menus p ON m.parent_id = p.id
WHERE m.path LIKE '/k8s%' OR m.parent_id = 5
ORDER BY m.parent_id, m.sort;

-- 查看完整的K8s菜单树
SELECT id, name, path, parent_id, sort, menu_type
FROM menus
WHERE id = 5 OR parent_id = 5
ORDER BY parent_id, sort;
