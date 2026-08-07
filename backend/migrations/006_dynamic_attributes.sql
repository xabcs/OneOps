-- 动态属性系统数据库迁移脚本
-- 执行方式：mysql -u root -p ops < migrations/006_dynamic_attributes.sql

-- 1. 创建属性定义表
CREATE TABLE IF NOT EXISTS sys_attribute_definitions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '属性ID',
    name VARCHAR(100) NOT NULL COMMENT '属性名称（显示用）',
    `key` VARCHAR(50) NOT NULL COMMENT '属性键（唯一标识）',
    category VARCHAR(50) NOT NULL COMMENT '分类：system/location/environment/hardware/custom',
    type VARCHAR(20) NOT NULL DEFAULT 'text' COMMENT '类型：text/select/multiselect/number/date/boolean',
    options TEXT COMMENT '选项配置（JSON格式，select/multiselect类型使用）',
    required BOOLEAN DEFAULT FALSE COMMENT '是否必填',
    default_value VARCHAR(255) COMMENT '默认值',
    sort_order INT DEFAULT 0 COMMENT '排序（数字越小越靠前）',
    status TINYINT DEFAULT 1 COMMENT '状态：1启用 0禁用',
    description TEXT COMMENT '属性说明',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_key (`key`),
    KEY idx_category (category),
    KEY idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='属性定义表';

-- 2. 创建主机属性值表
CREATE TABLE IF NOT EXISTS cmdb_server_attributes (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '属性ID',
    server_id BIGINT UNSIGNED NOT NULL COMMENT '主机ID',
    attribute_id BIGINT UNSIGNED NOT NULL COMMENT '属性定义ID',
    attribute_key VARCHAR(50) NOT NULL COMMENT '属性键（冗余字段）',
    attribute_value TEXT COMMENT '属性值',
    value_type VARCHAR(20) DEFAULT 'string' COMMENT '值类型：string/number/boolean/date/array',
    category VARCHAR(50) COMMENT '分类（冗余字段）',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_server_attribute (server_id, attribute_id),
    KEY idx_server_id (server_id),
    KEY idx_attribute_id (attribute_id),
    KEY idx_key_value (attribute_key, attribute_value(255)),

    FOREIGN KEY (server_id) REFERENCES cmdb_servers(id) ON DELETE CASCADE,
    FOREIGN KEY (attribute_id) REFERENCES sys_attribute_definitions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='主机属性值表';

-- 3. 插入预置属性定义

-- 业务系统
INSERT INTO sys_attribute_definitions (name, `key`, category, type, options, sort_order, description) VALUES
('业务系统', 'business_system', 'system', 'select', '[{"value":"ecommerce","label":"电商系统"},{"value":"crm","label":"CRM系统"},{"value":"erp","label":"ERP系统"},{"value":"monitor","label":"监控系统"}]', 1, '主机所属的业务系统');

-- 机房
INSERT INTO sys_attribute_definitions (name, `key`, category, type, options, sort_order, description) VALUES
('机房', 'room', 'location', 'select', '[{"value":"hz","label":"杭州机房"},{"value":"bj","label":"北京机房"},{"value":"sh","label":"上海机房"},{"value":"sz","label":"深圳机房"}]', 2, '主机所在的机房');

-- 机柜
INSERT INTO sys_attribute_definitions (name, `key`, category, type, sort_order, description) VALUES
('机柜', 'cabinet', 'location', 'text', NULL, 3, '主机所在的机柜');

-- 环境
INSERT INTO sys_attribute_definitions (name, `key`, category, type, options, default_value, required, sort_order, description) VALUES
('环境', 'env', 'environment', 'select', '[{"value":"prod","label":"生产"},{"value":"test","label":"测试"},{"value":"dev","label":"开发"}]', 'test', FALSE, 4, '主机运行环境');

-- 标签
INSERT INTO sys_attribute_definitions (name, `key`, category, type, options, sort_order, description) VALUES
('标签', 'tags', 'system', 'multiselect', '[{"value":"important","label":"重要"},{"value":"backup","label":"备份节点"},{"value":"monitor","label":"监控节点"},{"value":"web","label":"Web服务器"},{"value":"db","label":"数据库服务器"}]', 5, '主机的标签分类');

-- 所属项目
INSERT INTO sys_attribute_definitions (name, `key`, category, type, sort_order, description) VALUES
('所属项目', 'project', 'system', 'text', NULL, 6, '主机所属的项目');

-- 购买日期
INSERT INTO sys_attribute_definitions (name, `key`, category, type, sort_order, description) VALUES
('购买日期', 'purchase_date', 'hardware', 'date', NULL, 7, '主机购买日期');

-- 过保日期
INSERT INTO sys_attribute_definitions (name, `key`, category, type, sort_order, description) VALUES
('过保日期', 'warranty_date', 'hardware', 'date', NULL, 8, '主机过保日期');

-- 责任人
INSERT INTO sys_attribute_definitions (name, `key`, category, type, sort_order, description) VALUES
('责任人', 'owner', 'system', 'text', NULL, 9, '主机责任人');

-- 联系方式
INSERT INTO sys_attribute_definitions (name, `key`, category, type, sort_order, description) VALUES
('联系方式', 'contact', 'system', 'text', NULL, 10, '责任人联系方式');

-- 备注
INSERT INTO sys_attribute_definitions (name, `key`, category, type, sort_order, description) VALUES
('备注', 'remark', 'custom', 'text', NULL, 11, '主机备注信息');

-- 4. 迁移现有数据到动态属性表（可选）

-- 迁移业务系统
INSERT INTO cmdb_server_attributes (server_id, attribute_id, attribute_key, attribute_value, category)
SELECT
    s.id,
    (SELECT id FROM sys_attribute_definitions WHERE `key` = 'business_system'),
    'business_system',
    COALESCE(bu.name, '未分类'),
    'system'
FROM cmdb_servers s
LEFT JOIN cmdb_business_units bu ON s.business_id = bu.id
WHERE s.business_id IS NOT NULL;

-- 迁移机房信息（如果有）
-- INSERT INTO cmdb_server_attributes (server_id, attribute_id, attribute_key, attribute_value, category)
-- SELECT
--     s.id,
--     (SELECT id FROM sys_attribute_definitions WHERE `key` = 'room'),
--     'room',
--     c.code,
--     'location'
-- FROM cmdb_servers s
-- LEFT JOIN cmdb_cabinets c ON s.cabinet_id = c.id
-- WHERE s.cabinet_id IS NOT NULL;

-- 5. 验证迁移结果
SELECT
    '属性定义数量' AS item,
    COUNT(*) AS count
FROM sys_attribute_definitions
UNION ALL
SELECT
    '主机属性数量' AS item,
    COUNT(*) AS count
FROM cmdb_server_attributes;
