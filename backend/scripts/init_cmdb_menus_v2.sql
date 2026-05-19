-- ========================================
-- OneOps CMDB 菜单初始化
-- 创建时间: 2025-01-15
-- 版本: v1.0
-- ========================================

-- 插入CMDB菜单
-- 注意：菜单sort值需要根据实际情况调整，确保在合适的位置显示

-- 1. CMDB资产管理主菜单 (一级菜单)
INSERT INTO menus (name, icon, path, sort, parent_id, status, created_at, updated_at)
VALUES ('资产管理', 'server', '/cmdb', 30, 0, 1, NOW(), NOW());

-- 获取刚插入的CMDB菜单ID
SET @cmdb_id = LAST_INSERT_ID();

-- 2. 服务器管理 (二级菜单)
INSERT INTO menus (name, icon, path, sort, parent_id, status, created_at, updated_at)
VALUES ('服务器管理', 'server', '/cmdb/servers', 1, @cmdb_id, 1, NOW(), NOW());

-- 3. 业务系统管理 (二级菜单)
INSERT INTO menus (name, icon, path, sort, parent_id, status, created_at, updated_at)
VALUES ('业务管理', 'tree', '/cmdb/business', 2, @cmdb_id, 1, NOW(), NOW());

-- 4. 机房机柜管理 (二级菜单)
INSERT INTO menus (name, icon, path, sort, parent_id, status, created_at, updated_at)
VALUES ('机房管理', 'location', '/cmdb/rooms', 3, @cmdb_id, 1, NOW(), NOW());

-- 5. 标签管理 (二级菜单)
INSERT INTO menus (name, icon, path, sort, parent_id, status, created_at, updated_at)
VALUES ('标签管理', 'tag', '/cmdb/tags', 4, @cmdb_id, 1, NOW(), NOW());

-- 6. 资产变更记录 (二级菜单)
INSERT INTO menus (name, icon, path, sort, parent_id, status, created_at, updated_at)
VALUES ('变更记录', 'document', '/cmdb/changes', 5, @cmdb_id, 1, NOW(), NOW());

-- 更新admin角色，添加CMDB菜单权限
-- 获取admin角色ID
SET @admin_role_id = (SELECT id FROM roles WHERE code = 'admin' LIMIT 1);

-- 获取所有CMDB相关菜单ID
SET @cmdb_menu_ids = (
    SELECT GROUP_CONCAT(id)
    FROM menus
    WHERE id = @cmdb_id OR parent_id = @cmdb_id
);

-- 更新admin角色的menu_ids (JSON格式)
UPDATE roles
SET menu_ids = (
    SELECT JSON_ARRAYAGG(id)
    FROM (
        SELECT id FROM menus WHERE id IN (@cmdb_menu_ids)
        UNION
        SELECT id FROM menus WHERE parent_id = @cmdb_id
    ) AS cmdb_menus
)
WHERE code = 'admin';

-- 查看插入的菜单
SELECT id, name, path, parent_id, sort, status
FROM menus
WHERE id = @cmdb_id OR parent_id = @cmdb_id
ORDER BY parent_id, sort;

-- 查看更新结果
SELECT code, menu_ids
FROM roles
WHERE code = 'admin';
