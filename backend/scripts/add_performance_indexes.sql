-- 性能优化索引脚本
-- 执行方式: mysql -u root -p123456 nexops < backend/scripts/add_performance_indexes.sql

USE nexops;

-- 1. 服务器列表查询优化索引
CREATE INDEX idx_servers_hostname ON servers(hostname);
CREATE INDEX idx_servers_ip ON servers(ip);
CREATE INDEX idx_servers_status ON servers(status);
CREATE INDEX idx_servers_agent_status ON servers(agent_status);
CREATE INDEX idx_servers_env ON servers(env);
CREATE INDEX idx_servers_provider ON servers(provider);

-- 2. SSH凭证查询优化
CREATE INDEX idx_ssh_credentials_type ON ssh_credentials(credential_type);
CREATE INDEX idx_ssh_credentials_sort ON ssh_credentials(sort_order, id);

-- 3. 业务系统查询优化
CREATE INDEX idx_business_units_parent ON business_units(parent_id);
CREATE INDEX idx_business_units_sort ON business_units(sort_order, id);

-- 4. 服务器属性查询优化
CREATE INDEX idx_server_attributes_server ON server_attributes(server_id);
CREATE INDEX idx_server_attributes_attribute ON server_attributes(attribute_id);

-- 5. 分组关系查询优化
CREATE INDEX idx_server_group_relations_server ON server_group_relations(server_id);
CREATE INDEX idx_server_group_relations_group ON server_group_relations(group_id);

-- 6. 凭证管理查询优化
CREATE INDEX idx_credentials_server ON server_credentials(server_id);
CREATE INDEX idx_credentials_credential ON server_credentials(credential_id);

-- 7. 审计日志查询优化
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_module ON audit_logs(module);
CREATE INDEX idx_audit_logs_username ON audit_logs(username);

-- 验证索引创建
SHOW INDEX FROM servers;
SHOW INDEX FROM ssh_credentials;
SHOW INDEX FROM business_units;

SELECT '性能索引创建完成！' AS message;
