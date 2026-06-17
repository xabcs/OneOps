-- 更新 K8s 菜单分组结构
-- 根据 PRD 需求，将菜单按功能分为三个分组：工作负载、网络、配置管理

-- 步骤1：添加新的分组菜单
INSERT INTO menus (name, icon, path, permission, parent_id, sort, status, menu_type, created_at, updated_at)
VALUES
('工作负载', 'mdi:cube-outline', '/k8s/workloads', '', 5, 2, 1, 'directory', NOW(), NOW()),
('网络', 'mdi:network-outline', '/k8s/network', '', 5, 3, 1, 'directory', NOW(), NOW()),
('配置管理', 'mdi:cog', '/k8s/config', '', 5, 4, 1, 'directory', NOW(), NOW());

-- 步骤2：获取新插入的分组菜单ID
SET @workloads_id = LAST_INSERT_ID();
SET @network_id = @workloads_id + 1;
SET @config_id = @workloads_id + 2;

-- 或者直接使用固定ID（根据 menus 表的 ID 分配）
SET @workloads_id = 87;
SET @network_id = 92;
SET @config_id = 94;

-- 步骤3：更新现有菜单的 parent_id，将其移到对应分组下
-- 工作负载分组
UPDATE menus SET parent_id = @workloads_id, sort = 1 WHERE id = 81 AND name = '工作负载';
UPDATE menus SET name = '无状态', parent_id = @workloads_id, sort = 2 WHERE id = 81 AND path = '/k8s/resources/deployments';
UPDATE menus SET parent_id = @workloads_id, sort = 3 WHERE id = 82 AND path = '/k8s/resources/pods';

-- 网络分组
UPDATE menus SET parent_id = @network_id, sort = 1 WHERE id = 83 AND path = '/k8s/resources/services';

-- 配置管理分组
UPDATE menus SET parent_id = @config_id, sort = 1 WHERE id = 84 AND path = '/k8s/resources/configmaps';
UPDATE menus SET parent_id = @config_id, sort = 2 WHERE id = 85 AND path = '/k8s/resources/secrets';

-- 步骤4：添加新的资源菜单（Phase 2）
-- 有状态
INSERT INTO menus (name, icon, path, permission, parent_id, sort, status, menu_type, created_at, updated_at)
VALUES ('有状态', 'mdi:database', '/k8s/resources/statefulsets', 'k8s:statefulset:query', @workloads_id, 4, 1, 'menu', NOW(), NOW());

-- 守护进程集
INSERT INTO menus (name, icon, path, permission, parent_id, sort, status, menu_type, created_at, updated_at)
VALUES ('守护进程集', 'mdi:scale-balanced', '/k8s/resources/daemonsets', 'k8s:daemonset:query', @workloads_id, 5, 1, 'menu', NOW(), NOW());

-- 任务
INSERT INTO menus (name, icon, path, permission, parent_id, sort, status, menu_type, created_at, updated_at)
VALUES ('任务', 'mdi:play-circle', '/k8s/resources/jobs', 'k8s:job:query', @workloads_id, 6, 1, 'menu', NOW(), NOW());

-- 定时任务
INSERT INTO menus (name, icon, path, permission, parent_id, sort, status, menu_type, created_at, updated_at)
VALUES ('定时任务', 'mdi:clock', '/k8s/resources/cronjobs', 'k8s:cronjob:query', @workloads_id, 7, 1, 'menu', NOW(), NOW());

-- 路由
INSERT INTO menus (name, icon, path, permission, parent_id, sort, status, menu_type, created_at, updated_at)
VALUES ('路由', 'mdi:router', '/k8s/resources/ingresses', 'k8s:ingress:query', @network_id, 2, 1, 'menu', NOW(), NOW());

-- 步骤5：调整会话审计的排序
UPDATE menus SET sort = 5 WHERE parent_id = 5 AND name = '会话审计';

-- 验证更新结果
SELECT id, name, path, parent_id, sort, menu_type FROM menus WHERE parent_id = 5 OR parent_id IN (@workloads_id, @network_id, @config_id) ORDER BY parent_id, sort;

-- 查看完整的菜单树
SELECT
    m1.id,
    m1.name,
    m1.path,
    m1.parent_id,
    m1.sort,
    m2.name as parent_name
FROM menus m1
LEFT JOIN menus m2 ON m1.parent_id = m2.id
WHERE m1.path LIKE '/k8s%'
ORDER BY m1.parent_id, m1.sort;
