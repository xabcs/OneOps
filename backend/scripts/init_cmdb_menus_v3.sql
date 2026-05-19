-- ========================================
-- OneOps CMDB 菜单初始化 (修正版)
-- 创建时间: 2025-01-15
-- 版本: v1.0
-- ========================================

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

-- 为admin角色添加CMDB菜单权限
-- 获取admin角色ID
SET @admin_role_id = (SELECT id FROM roles WHERE code = 'admin' LIMIT 1);

-- 为admin角色添加所有CMDB菜单
INSERT INTO role_menus (role_id, menu_id)
SELECT @admin_role_id, id
FROM menus
WHERE id = @cmdb_id OR parent_id = @cmdb_id;

-- 查看插入的菜单
SELECT id, name, path, parent_id, sort, status
FROM menus
WHERE id = @cmdb_id OR parent_id = @cmdb_id
ORDER BY parent_id, sort;

-- 查看角色菜单关联
SELECT rm.role_id, r.name as role_name, m.id as menu_id, m.name as menu_name
FROM role_menus rm
JOIN roles r ON rm.role_id = r.id
JOIN menus m ON rm.menu_id = m.id
WHERE m.id = @cmdb_id OR m.parent_id = @cmdb_id;
