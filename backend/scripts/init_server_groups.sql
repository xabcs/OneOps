-- ========================================
-- OneOps 主机分组表
-- 创建时间: 2026-05-19
-- 版本: v1.0
-- ========================================

-- 主机分组表
CREATE TABLE IF NOT EXISTS server_groups (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '分组ID',
    name VARCHAR(100) NOT NULL COMMENT '分组名称',
    code VARCHAR(50) NOT NULL COMMENT '分组代码',
    parent_id BIGINT DEFAULT 0 COMMENT '父分组ID，0表示顶级',
    level INT DEFAULT 1 COMMENT '层级',
    description TEXT COMMENT '分组描述',
    color VARCHAR(20) DEFAULT '#409EFF' COMMENT '分组颜色',
    icon VARCHAR(50) DEFAULT 'mdi:folder' COMMENT '分组图标',
    sort_order INT DEFAULT 0 COMMENT '排序',
    status TINYINT DEFAULT 1 COMMENT '状态：1启用 0禁用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_parent(parent_id),
    INDEX idx_code(code),
    INDEX idx_status(status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='主机分组表';

-- 服务器分组关联表
CREATE TABLE IF NOT EXISTS server_group_relations (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '关联ID',
    server_id BIGINT NOT NULL COMMENT '服务器ID',
    group_id BIGINT NOT NULL COMMENT '分组ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    UNIQUE KEY uk_server_group (server_id, group_id),
    INDEX idx_server(server_id),
    INDEX idx_group(group_id),
    FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES server_groups(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务器分组关联表';

-- 插入示例分组数据
INSERT INTO server_groups (name, code, parent_id, level, description, color, icon, sort_order) VALUES
('生产环境', 'PROD', 0, 1, '生产环境服务器组', '#f56c6c', 'mdi:server', 1),
('测试环境', 'TEST', 0, 1, '测试环境服务器组', '#e6a23c', 'mdi:test-tube', 2),
('开发环境', 'DEV', 0, 1, '开发环境服务器组', '#409eff', 'mdi:code-tags', 3),
('电商业务', 'ECOMMERCE', 0, 1, '电商业务相关服务器', '#67c23a', 'mdi:shopping', 4),
('支付系统', 'PAYMENT', 0, 1, '支付系统相关服务器', '#e6a23c', 'mdi:credit-card', 5),
('Web服务器', 'WEB', 1, 2, 'Web应用服务器', '#409eff', 'mdi:web', 1),
('数据库服务器', 'DATABASE', 1, 2, '数据库服务器', '#f56c6c', 'mdi:database', 2),
('应用服务器', 'APP', 1, 2, '应用服务器', '#67c23a', 'mdi:application', 3);

-- 将现有服务器分配到分组
INSERT INTO server_group_relations (server_id, group_id)
SELECT id, 1 FROM servers WHERE env = 'prod'
UNION ALL
SELECT id, 2 FROM servers WHERE env = 'test'
UNION ALL
SELECT id, 3 FROM servers WHERE env = 'dev';

-- 查看创建的分组
SELECT id, name, code, parent_id, level, color, icon
FROM server_groups
ORDER BY parent_id, sort_order;
