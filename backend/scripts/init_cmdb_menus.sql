-- ========================================
-- OneOps CMDB 菜单初始化
-- 创建时间: 2025-01-15
-- 版本: v1.0
-- ========================================

-- 插入CMDB菜单
-- 注意：菜单order值需要根据实际情况调整，确保在合适的位置显示

-- 1. CMDB资产管理主菜单 (一级菜单)
INSERT INTO menus (name, title, icon, path, component, sort, parent_id, status, created_at, updated_at)
VALUES ('cmdb', '资产管理', 'mdi:server-network', '/cmdb', NULL, 30, 0, 1, NOW(), NOW());

-- 获取刚插入的CMDB菜单ID (SET @cmdb_id = LAST_INSERT_ID())
SET @cmdb_id = LAST_INSERT_ID();

-- 2. 服务器管理 (二级菜单)
INSERT INTO menus (name, title, icon, path, component, sort, parent_id, status, created_at, updated_at)
VALUES ('cmdb_servers', '服务器管理', 'mdi:server', '/cmdb/servers', 'cmdb_servers', 1, @cmdb_id, 1, NOW(), NOW());

-- 3. 业务系统管理 (二级菜单)
INSERT INTO menus (name, title, icon, path, component, sort, parent_id, status, created_at, updated_at)
VALUES ('cmdb_business', '业务管理', 'mdi:sitemap', '/cmdb/business', 'self', 2, @cmdb_id, 1, NOW(), NOW());

-- 4. 机房机柜管理 (二级菜单)
INSERT INTO menus (name, title, icon, path, component, sort, parent_id, status, created_at, updated_at)
VALUES ('cmdb_rooms', '机房管理', 'mdi:server', '/cmdb/rooms', 'self', 3, @cmdb_id, 1, NOW(), NOW());

-- 5. 标签管理 (二级菜单)
INSERT INTO menus (name, title, icon, path, component, sort, parent_id, status, created_at, updated_at)
VALUES ('cmdb_tags', '标签管理', 'mdi:tag-multiple', '/cmdb/tags', 'self', 4, @cmdb_id, 1, NOW(), NOW());

-- 6. 资产变更记录 (二级菜单)
INSERT INTO menus (name, title, icon, path, component, sort, parent_id, status, created_at, updated_at)
VALUES ('cmdb_changes', '变更记录', 'mdi:history', '/cmdb/changes', 'self', 5, @cmdb_id, 1, NOW(), NOW());

-- 更新管理员角色的菜单权限
-- 假设管理员角色代码为 'admin' 或 'R_SUPER'
UPDATE roles
SET menu_ids = JSON_ARRAY_APPEND(
    COALESCE(menu_ids, JSON_ARRAY()),
    '$',
    (SELECT id FROM menus WHERE name = 'cmdb' ORDER BY id DESC LIMIT 1)
)
WHERE code = 'admin';

-- 为admin角色添加所有CMDB子菜单
UPDATE roles
SET menu_ids = (
    SELECT JSON_ARRAYAGG(id)
    FROM menus
    WHERE name IN ('cmdb', 'cmdb_servers', 'cmdb_business', 'cmdb_rooms', 'cmdb_tags', 'cmdb_changes')
    OR parent_id IN (SELECT id FROM menus WHERE name = 'cmdb')
)
WHERE code = 'admin';

-- 查看插入的菜单
SELECT id, name, title, parent_id, sort, status
FROM menus
WHERE name LIKE 'cmdb%'
ORDER BY parent_id, sort;

-- 查看admin角色的菜单权限
SELECT code, menu_ids
FROM roles
WHERE code = 'admin';
