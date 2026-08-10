-- 诊断功能权限配置SQL脚本
-- 为OneOps系统添加K8s诊断相关权限

-- 1. 添加诊断权限定义
INSERT INTO sys_permissions (name, code, category, description, risk_level, created_at)
VALUES
('执行K8s诊断', 'k8s:diagnostic:execute', 'K8s诊断', '对K8s集群中的Java应用执行Arthas诊断', 'high', NOW()),
('查看诊断结果', 'k8s:diagnostic:view', 'K8s诊断', '查看K8s诊断结果和历史记录', 'medium', NOW())
ON DUPLICATE KEY UPDATE
  description = VALUES(description),
  risk_level = VALUES(risk_level),
  updated_at = NOW();

-- 2. 为管理员角色分配诊断权限
INSERT INTO sys_role_permissions (role_id, permission_id, created_at)
SELECT
  (SELECT id FROM sys_roles WHERE code = 'admin'),
  (SELECT id FROM sys_permissions WHERE code = 'k8s:diagnostic:execute'),
  NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM sys_role_permissions rp
  JOIN sys_roles r ON rp.role_id = r.id
  JOIN sys_permissions p ON rp.permission_id = p.id
  WHERE r.code = 'admin' AND p.code = 'k8s:diagnostic:execute'
);

INSERT INTO sys_role_permissions (role_id, permission_id, created_at)
SELECT
  (SELECT id FROM sys_roles WHERE code = 'admin'),
  (SELECT id FROM sys_permissions WHERE code = 'k8s:diagnostic:view'),
  NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM sys_role_permissions rp
  JOIN sys_roles r ON rp.role_id = r.id
  JOIN sys_permissions p ON rp.permission_id = p.id
  WHERE r.code = 'admin' AND p.code = 'k8s:diagnostic:view'
);

-- 3. 为开发者角色分配查看权限（如果需要）
INSERT INTO sys_role_permissions (role_id, permission_id, created_at)
SELECT
  (SELECT id FROM sys_roles WHERE code = 'developer'),
  (SELECT id FROM sys_permissions WHERE code = 'k8s:diagnostic:view'),
  NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM sys_role_permissions rp
  JOIN sys_roles r ON rp.role_id = r.id
  JOIN sys_permissions p ON rp.permission_id = p.id
  WHERE r.code = 'developer' AND p.code = 'k8s:diagnostic:view'
);

-- 4. 为运维人员角色分配完整诊断权限（如果需要）
INSERT INTO sys_role_permissions (role_id, permission_id, created_at)
SELECT
  (SELECT id FROM sys_roles WHERE code = 'ops'),
  (SELECT id FROM sys_permissions WHERE code = 'k8s:diagnostic:execute'),
  NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM sys_role_permissions rp
  JOIN sys_roles r ON rp.role_id = r.id
  JOIN sys_permissions p ON rp.permission_id = p.id
  WHERE r.code = 'ops' AND p.code = 'k8s:diagnostic:execute'
);

INSERT INTO sys_role_permissions (role_id, permission_id, created_at)
SELECT
  (SELECT id FROM sys_roles WHERE code = 'ops'),
  (SELECT id FROM sys_permissions WHERE code = 'k8s:diagnostic:view'),
  NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM sys_role_permissions rp
  JOIN sys_roles r ON rp.role_id = r.id
  JOIN sys_permissions p ON rp.permission_id = p.id
  WHERE r.code = 'ops' AND p.code = 'k8s:diagnostic:view'
);

-- 5. 创建诊断历史表
CREATE TABLE IF NOT EXISTS k8s_diagnostic_history (
  id INT AUTO_INCREMENT PRIMARY KEY,
  cluster_id VARCHAR(50) NOT NULL COMMENT '集群ID',
  cluster_name VARCHAR(100) NOT NULL COMMENT '集群名称',
  namespace VARCHAR(100) NOT NULL COMMENT '命名空间',
  pod_name VARCHAR(100) NOT NULL COMMENT 'Pod名称',
  container_name VARCHAR(100) COMMENT '容器名称',
  command VARCHAR(50) NOT NULL COMMENT '诊断命令',
  args TEXT COMMENT '命令参数',
  result_status VARCHAR(20) NOT NULL COMMENT '结果状态: success/error',
  result_output LONGTEXT COMMENT '结果输出',
  result_error TEXT COMMENT '错误信息',
  result_method VARCHAR(20) COMMENT '诊断方式: sidecar/daemonset',
  result_duration INT COMMENT '执行耗时(ms)',
  user_id INT NOT NULL COMMENT '执行用户ID',
  username VARCHAR(50) NOT NULL COMMENT '执行用户名',
  timestamp DATETIME NOT NULL COMMENT '执行时间',
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_cluster (cluster_id),
  INDEX idx_pod (namespace, pod_name),
  INDEX idx_user (user_id),
  INDEX idx_timestamp (timestamp),
  INDEX idx_command (command)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='K8s诊断历史记录';

-- 6. 创建诊断配置表（可选）
CREATE TABLE IF NOT EXISTS k8s_diagnostic_config (
  id INT AUTO_INCREMENT PRIMARY KEY,
  cluster_id VARCHAR(50) NOT NULL COMMENT '集群ID',
  namespace VARCHAR(100) NOT NULL COMMENT '命名空间',
  config_key VARCHAR(100) NOT NULL COMMENT '配置键',
  config_value TEXT NOT NULL COMMENT '配置值',
  description VARCHAR(255) COMMENT '配置说明',
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY unique_config (cluster_id, namespace, config_key),
  INDEX idx_cluster (cluster_id, namespace)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='K8s诊断配置';

-- 7. 插入默认诊断配置
INSERT INTO k8s_diagnostic_config (cluster_id, namespace, config_key, config_value, description)
VALUES
('*', '*', 'default_timeout', '60', '默认诊断超时时间(秒)'),
('*', '*', 'max_output_size', '10485760', '最大输出大小(字节)'),
('*', '*', 'enable_auto_cleanup', 'true', '是否自动清理临时文件'),
('*', '*', 'history_retention_days', '30', '历史记录保留天数')
ON DUPLICATE KEY UPDATE
  config_value = VALUES(config_value),
  description = VALUES(description),
  updated_at = NOW();

-- 8. 添加菜单项到K8s管理
INSERT INTO sys_menus (parent_id, name, code, path, component, icon, sort, permission_code, visible, created_at)
SELECT
  (SELECT id FROM sys_menus WHERE code = 'k8s'),
  '诊断中心',
  'k8s_diagnostic',
  '/k8s/diagnostic',
  'k8s/diagnostic/index',
  'Operation',
  50,
  'k8s:diagnostic:execute',
  1,
  NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM sys_menus WHERE code = 'k8s_diagnostic'
);

-- 9. 验证权限配置
SELECT
  p.name AS permission_name,
  p.code AS permission_code,
  p.description,
  r.name AS role_name,
  rp.created_at AS granted_at
FROM sys_permissions p
JOIN sys_role_permissions rp ON p.id = rp.permission_id
JOIN sys_roles r ON rp.role_id = r.id
WHERE p.code LIKE 'k8s:diagnostic%'
ORDER BY p.code, r.name;

-- 10. 验证表创建
SHOW TABLES LIKE '%diagnostic%';

-- 11. 查看菜单配置
SELECT
  id,
  name,
  code,
  path,
  icon,
  sort,
  permission_code
FROM sys_menus
WHERE code LIKE '%diagnostic%'
ORDER BY sort;

-- 执行完成提示
SELECT '诊断权限配置完成！' AS status;