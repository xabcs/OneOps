-- 重建 K8s 菜单结构 - Tab 切换方案
-- 删除分组下的子菜单，保留分组作为统一入口

-- 步骤1：删除作为子菜单添加的新资源类型（这些将通过Tab在页面内展示）
DELETE FROM menus WHERE id IN (90, 91, 92, 93, 94);

-- 步骤2：更新现有菜单的名称和路径
-- 工作负载：改为统一入口
UPDATE menus
SET name = '工作负载',
    path = '/k8s/workloads',
    menu_type = 'menu',
    sort = 2
WHERE id = 87 AND parent_id = 5;

-- 网络：改为统一入口
UPDATE menus
SET name = '网络',
    path = '/k8s/network',
    menu_type = 'menu',
    sort = 3
WHERE id = 88 AND parent_id = 5;

-- 配置管理：改为统一入口
UPDATE menus
SET name = '配置管理',
    path = '/k8s/config',
    menu_type = 'menu',
    sort = 4
WHERE id = 89 AND parent_id = 5;

-- 步骤3：调整集群管理和会话审计的排序
UPDATE menus SET sort = 1 WHERE id = 80 AND name = '集群管理';
UPDATE menus SET sort = 5 WHERE id = 86 AND name = '会话审计';

-- 步骤4：删除作为子菜单挂在分组下的现有资源菜单
DELETE FROM menus WHERE parent_id IN (87, 88, 89) AND id < 90;

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
