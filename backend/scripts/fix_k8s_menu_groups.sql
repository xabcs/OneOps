-- 修正 K8s 菜单分组的 parent_id
-- 确保子菜单指向正确的父分组

-- 修正网络分组的子菜单
UPDATE menus SET parent_id = 88, sort = 1 WHERE id = 83 AND path = '/k8s/resources/services';
UPDATE menus SET parent_id = 88, sort = 2 WHERE id = 94 AND path = '/k8s/resources/ingresses';

-- 修正配置管理分组的子菜单
UPDATE menus SET parent_id = 89, sort = 1 WHERE id = 84 AND path = '/k8s/resources/configmaps';
UPDATE menus SET parent_id = 89, sort = 2 WHERE id = 85 AND path = '/k8s/resources/secrets';

-- 重新排序 K8s 管理下的子菜单
UPDATE menus SET sort = 1 WHERE id = 80 AND name = '集群管理';
UPDATE menus SET sort = 2 WHERE id = 87 AND name = '工作负载';
UPDATE menus SET sort = 3 WHERE id = 88 AND name = '网络';
UPDATE menus SET sort = 4 WHERE id = 89 AND name = '配置管理';
UPDATE menus SET sort = 5 WHERE id = 86 AND name = '会话审计';

-- 验证修正结果
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
WHERE m.path LIKE '/k8s%'
ORDER BY
    CASE WHEN m.parent_id = 0 THEN m.id ELSE m.parent_id END,
    m.parent_id,
    m.sort;
