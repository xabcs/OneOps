-- ========================================
-- OneOps CMDB 资产管理数据库表
-- 创建时间: 2025-01-15
-- 版本: v1.0
-- ========================================

-- 1. 业务系统表
CREATE TABLE IF NOT EXISTS business_units (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '业务ID',
    name VARCHAR(100) NOT NULL COMMENT '业务名称',
    code VARCHAR(50) NOT NULL UNIQUE COMMENT '业务代码',
    parent_id BIGINT DEFAULT 0 COMMENT '父业务ID，0表示顶级',
    level INT DEFAULT 1 COMMENT '层级',
    owner VARCHAR(100) COMMENT '负责人',
    phone VARCHAR(20) COMMENT '联系电话',
    email VARCHAR(100) COMMENT '联系邮箱',
    sort_order INT DEFAULT 0 COMMENT '排序',
    status TINYINT DEFAULT 1 COMMENT '状态：1启用 0禁用',
    remarks TEXT COMMENT '备注',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_parent(parent_id),
    INDEX idx_code(code),
    INDEX idx_status(status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='业务系统表';

-- 插入默认业务数据
INSERT INTO business_units (name, code, parent_id, level, owner, sort_order) VALUES
('电商业务', 'ECOMMERCE', 0, 1, '张三', 1),
('支付系统', 'PAYMENT', 0, 1, '李四', 2),
('用户中心', 'USER_CENTER', 0, 1, '王五', 3),
('运维平台', 'ONEOPS', 0, 1, '运维组', 4);

-- 2. 机房表
CREATE TABLE IF NOT EXISTS server_rooms (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '机房ID',
    name VARCHAR(100) NOT NULL COMMENT '机房名称',
    code VARCHAR(50) NOT NULL UNIQUE COMMENT '机房代码',
    location VARCHAR(200) COMMENT '机房位置',
    address VARCHAR(500) COMMENT '详细地址',
    provider VARCHAR(100) COMMENT '服务商',
    contact VARCHAR(100) COMMENT '联系人',
    phone VARCHAR(20) COMMENT '联系电话',
    status TINYINT DEFAULT 1 COMMENT '状态：1启用 0禁用',
    remarks TEXT COMMENT '备注',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_code(code),
    INDEX idx_status(status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='机房表';

-- 插入默认机房数据
INSERT INTO server_rooms (name, code, location, provider) VALUES
('阿里云华东机房', 'ALIYU-HUADONG', '杭州', '阿里云'),
('腾讯云华南机房', 'TENCENT-HUANAN', '广州', '腾讯云'),
('自建机房A', 'SELF-A', '北京朝阳区', '自建');

-- 3. 机柜表
CREATE TABLE IF NOT EXISTS cabinets (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '机柜ID',
    name VARCHAR(100) NOT NULL COMMENT '机柜名称',
    code VARCHAR(50) NOT NULL UNIQUE COMMENT '机柜代码',
    room_id BIGINT COMMENT '机房ID',
    position VARCHAR(50) COMMENT '位置',
    capacity INT DEFAULT 42 COMMENT 'U数',
    used_u INT DEFAULT 0 COMMENT '已用U数',
    power_usage DECIMAL(10,2) DEFAULT 0.00 COMMENT '已用电力(KW)',
    power_capacity DECIMAL(10,2) DEFAULT 0.00 COMMENT '总电力(KW)',
    status TINYINT DEFAULT 1 COMMENT '状态：1启用 0禁用',
    remarks TEXT COMMENT '备注',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    FOREIGN KEY (room_id) REFERENCES server_rooms(id),
    INDEX idx_code(code),
    INDEX idx_room(room_id),
    INDEX idx_status(status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='机柜表';

-- 插入默认机柜数据
INSERT INTO cabinets (name, code, room_id, position, capacity, power_capacity) VALUES
('机柜A01', 'CAB-A01', 1, 'A区-01排', 42, 10.00),
('机柜A02', 'CAB-A02', 1, 'A区-02排', 42, 10.00),
('机柜B01', 'CAB-B01', 2, 'B区-01排', 42, 8.00);

-- 4. 服务器表
CREATE TABLE IF NOT EXISTS servers (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '服务器ID',
    hostname VARCHAR(100) NOT NULL UNIQUE COMMENT '主机名',
    ip VARCHAR(50) NOT NULL COMMENT '外网IP',
    inner_ip VARCHAR(50) COMMENT '内网IP',
    cpu INT DEFAULT 0 COMMENT 'CPU核心数',
    memory INT DEFAULT 0 COMMENT '内存(GB)',
    disk INT DEFAULT 0 COMMENT '磁盘(GB)',
    os VARCHAR(50) COMMENT '操作系统',
    os_version VARCHAR(50) COMMENT '系统版本',
    arch VARCHAR(20) COMMENT '系统架构：x86_64/arm64',
    env ENUM('prod','test','dev') DEFAULT 'test' COMMENT '环境：生产/测试/开发',
    status ENUM('online','offline','unknown') DEFAULT 'unknown' COMMENT '状态：在线/离线/未知',
    ssh_port INT DEFAULT 22 COMMENT 'SSH端口',
    ssh_user VARCHAR(50) DEFAULT 'root' COMMENT 'SSH用户',
    business_id BIGINT COMMENT '所属业务ID',
    cabinet_id BIGINT COMMENT '所在机柜ID',
    u_position INT COMMENT '机柜位置(U)',
    sn VARCHAR(100) COMMENT '序列号',
    manufacturer VARCHAR(100) COMMENT '厂商',
    model VARCHAR(100) COMMENT '型号',
    purchase_date DATE COMMENT '购买日期',
    expire_warranty DATE COMMENT '保修到期',
    asset_number VARCHAR(100) COMMENT '资产编号',
    instance_id VARCHAR(100) COMMENT '云主机实例ID',
    instance_type VARCHAR(50) COMMENT '云主机类型',
    region VARCHAR(50) COMMENT '区域',
    zone VARCHAR(50) COMMENT '可用区',
    provider VARCHAR(50) COMMENT '服务商：aliyun/tencent/aws/self',
    server_type ENUM('physical','vm','container') DEFAULT 'vm' COMMENT '类型：物理机/虚拟机/容器',
    remarks TEXT COMMENT '备注',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    last_check_time TIMESTAMP NULL COMMENT '最后检查时间',
    FOREIGN KEY (business_id) REFERENCES business_units(id),
    FOREIGN KEY (cabinet_id) REFERENCES cabinets(id),
    INDEX idx_ip(ip),
    INDEX idx_inner_ip(inner_ip),
    INDEX idx_hostname(hostname),
    INDEX idx_business(business_id),
    INDEX idx_cabinet(cabinet_id),
    INDEX idx_env(env),
    INDEX idx_status(status),
    INDEX idx_provider(provider)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务器表';

-- 插入示例服务器数据
INSERT INTO servers (hostname, ip, inner_ip, cpu, memory, disk, os, os_version, env, status, ssh_port, business_id, provider, server_type) VALUES
('web-server-01', '192.168.1.101', '10.0.1.101', 4, 16, 500, 'CentOS', '7.9', 'prod', 'online', 22, 1, 'aliyun', 'vm'),
('web-server-02', '192.168.1.102', '10.0.1.102', 4, 16, 500, 'CentOS', '7.9', 'prod', 'online', 22, 1, 'aliyun', 'vm'),
('api-server-01', '192.168.1.103', '10.0.1.103', 8, 32, 1000, 'Ubuntu', '20.04', 'prod', 'online', 22, 1, 'aliyun', 'vm'),
('db-server-01', '192.168.1.201', '10.0.2.201', 16, 64, 2000, 'CentOS', '7.9', 'prod', 'online', 22, 2, 'tencent', 'vm'),
('test-server-01', '192.168.1.50', '10.0.1.50', 2, 4, 100, 'Ubuntu', '18.04', 'test', 'offline', 22, 4, 'self', 'physical');

-- 5. 资产变更记录表
CREATE TABLE IF NOT EXISTS asset_changes (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '变更ID',
    asset_type VARCHAR(50) NOT NULL COMMENT '资产类型：server/cabinet/business',
    asset_id BIGINT NOT NULL COMMENT '资产ID',
    asset_name VARCHAR(100) COMMENT '资产名称（冗余字段）',
    field_name VARCHAR(50) NOT NULL COMMENT '变更字段',
    old_value TEXT COMMENT '旧值',
    new_value TEXT COMMENT '新值',
    change_type ENUM('create','update','delete') DEFAULT 'update' COMMENT '变更类型',
    operator VARCHAR(100) COMMENT '操作人',
    operator_id BIGINT COMMENT '操作人ID',
    operate_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '操作时间',
    remarks TEXT COMMENT '备注',
    INDEX idx_asset(asset_type, asset_id),
    INDEX idx_time(operate_time),
    INDEX idx_operator(operator_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='资产变更记录表';

-- 6. 服务器标签表（用于灵活分类）
CREATE TABLE IF NOT EXISTS server_tags (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '标签ID',
    name VARCHAR(50) NOT NULL UNIQUE COMMENT '标签名称',
    color VARCHAR(20) DEFAULT '#409EFF' COMMENT '标签颜色',
    description VARCHAR(200) COMMENT '标签描述',
    sort_order INT DEFAULT 0 COMMENT '排序',
    status TINYINT DEFAULT 1 COMMENT '状态：1启用 0禁用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_name(name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务器标签表';

-- 插入默认标签
INSERT INTO server_tags (name, color, description) VALUES
('重要', '#F56C6C', '重要生产服务器'),
('测试', '#E6A23C', '测试环境服务器'),
('备份', '#909399', '备份服务器'),
('Web', '#409EFF', 'Web应用服务器'),
('数据库', '#67C23A', '数据库服务器');

-- 7. 服务器标签关联表
CREATE TABLE IF NOT EXISTS server_tag_relations (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '关联ID',
    server_id BIGINT NOT NULL COMMENT '服务器ID',
    tag_id BIGINT NOT NULL COMMENT '标签ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES server_tags(id) ON DELETE CASCADE,
    UNIQUE KEY uk_server_tag(server_id, tag_id),
    INDEX idx_server(server_id),
    INDEX idx_tag(tag_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务器标签关联表';

-- 为示例服务器添加标签
INSERT INTO server_tag_relations (server_id, tag_id) VALUES
(1, 1), (1, 4),  -- web-server-01: 重要 + Web
(2, 4),          -- web-server-02: Web
(3, 4),          -- api-server-01: Web
(4, 5),          -- db-server-01: 数据库
(5, 2);          -- test-server-01: 测试
