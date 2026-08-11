import { request } from '../request';

/**
 * 获取服务器列表
 */
export function fetchGetServers(params?: CMDB.ServerQuery) {
  return request<CMDB.PageResponse<CMDB.Server>>({
    url: '/cmdb/servers',
    method: 'get',
    params
  });
}

/**
 * 获取服务器详情
 */
export function fetchGetServerById(id: number) {
  return request<CMDB.Server>({
    url: `/cmdb/servers/${id}`,
    method: 'get'
  });
}

/**
 * 获取连接所需的服务器信息（轻量级，只包含基本信息和凭证）
 */
export function fetchGetServerForConnect(id: number) {
  return request<CMDB.Server>({
    url: `/cmdb/servers/${id}/connect`,
    method: 'get'
  });
}

/**
 * 创建服务器
 */
export function fetchCreateServer(data: CMDB.ServerForm) {
  return request<CMDB.Server>({
    url: '/cmdb/servers',
    method: 'post',
    data
  });
}

/**
 * 更新服务器
 */
export function fetchUpdateServer(id: number, data: Partial<CMDB.ServerForm>) {
  return request({
    url: `/cmdb/servers/${id}`,
    method: 'put',
    data
  });
}

/**
 * 删除服务器
 */
export function fetchDeleteServer(id: number) {
  return request({
    url: `/cmdb/servers/${id}`,
    method: 'delete'
  });
}

/**
 * 获取服务器统计
 */
export function fetchGetServerStats() {
  return request<CMDB.ServerStats>({
    url: '/cmdb/servers/stats',
    method: 'get'
  });
}

/**
 * 获取服务器配置（通过SSH）
 */
export function fetchGetServerConfig(data: { hostname: string; ip: string; sshPort?: number }) {
  return request<{
    cpu: number;
    memory: number;
    disk: number;
    os: string;
    osVersion: string;
    arch: string;
    hostname: string;
  }>({
    url: '/cmdb/servers/config',
    method: 'post',
    data
  });
}

/**
 * 获取主机分组列表（树形结构）
 */
export function fetchGetServerGroups() {
  return request<CMDB.ServerGroup[]>({
    url: '/cmdb/groups',
    method: 'get'
  });
}

/**
 * 获取完整的资产树（分组+服务器），一次性返回所有数据
 */
export function fetchGetAssetTree() {
  return request<{
    groups: CMDB.ServerGroup[];
    ungroupedServers: CMDB.Server[];
  }>({
    url: '/cmdb/asset-tree',
    method: 'get'
  });
}

/**
 * 创建主机分组
 */
export function fetchCreateServerGroup(data: CMDB.ServerGroupForm) {
  return request({
    url: '/cmdb/groups',
    method: 'post',
    data
  });
}

/**
 * 更新主机分组
 */
export function fetchUpdateServerGroup(id: number, data: Partial<CMDB.ServerGroupForm>) {
  return request({
    url: `/cmdb/groups/${id}`,
    method: 'put',
    data
  });
}

/**
 * 删除主机分组
 */
export function fetchDeleteServerGroup(id: number) {
  return request({
    url: `/cmdb/groups/${id}`,
    method: 'delete'
  });
}

/**
 * 将服务器分配到分组
 */
export function fetchAssignServerToGroup(serverId: number, groupId: number) {
  return request({
    url: '/cmdb/groups/assign',
    method: 'post',
    data: { serverId, groupId }
  });
}

/**
 * 将服务器分配到多个分组
 */
export function fetchAssignServerToGroups(serverId: number, groupIds: number[]) {
  return request({
    url: '/cmdb/groups/assign-multi',
    method: 'post',
    data: { serverId, groupIds }
  });
}

/**
 * 获取指定分组下的服务器列表
 */
export function fetchGetServersByGroup(
  groupId: number,
  params?: {
    page?: number;
    pageSize?: number;
  }
) {
  return request<CMDB.PageResponse<CMDB.Server>>({
    url: `/cmdb/group-servers/${groupId}`,
    method: 'get',
    params
  });
}

/**
 * 获取业务系统列表（树形结构）
 */
export function fetchGetBusinessUnits() {
  return request<CMDB.BusinessUnit[]>({
    url: '/cmdb/business-units',
    method: 'get'
  });
}

/**
 * 创建业务系统
 */
export function fetchCreateBusinessUnit(data: CMDB.BusinessUnitForm) {
  return request<CMDB.BusinessUnit>({
    url: '/cmdb/business-units',
    method: 'post',
    data
  });
}

/**
 * 更新业务系统
 */
export function fetchUpdateBusinessUnit(id: number, data: Partial<CMDB.BusinessUnitForm>) {
  return request({
    url: `/cmdb/business-units/${id}`,
    method: 'put',
    data
  });
}

/**
 * 删除业务系统
 */
export function fetchDeleteBusinessUnit(id: number) {
  return request({
    url: `/cmdb/business-units/${id}`,
    method: 'delete'
  });
}

/**
 * 获取机房列表
 */
export function fetchGetServerRooms() {
  return request<CMDB.ServerRoom[]>({
    url: '/cmdb/rooms',
    method: 'get'
  });
}

/**
 * 获取机柜列表
 */
export function fetchGetCabinets(roomId?: number) {
  return request<CMDB.Cabinet[]>({
    url: '/cmdb/cabinets',
    method: 'get',
    params: { roomId }
  });
}

/**
 * 创建机房
 */
export function fetchCreateServerRoom(data: CMDB.ServerRoomForm) {
  return request({
    url: '/cmdb/rooms',
    method: 'post',
    data
  });
}

/**
 * 更新机房
 */
export function fetchUpdateServerRoom(id: number, data: Partial<CMDB.ServerRoomForm>) {
  return request({
    url: `/cmdb/rooms/${id}`,
    method: 'put',
    data
  });
}

/**
 * 删除机房
 */
export function fetchDeleteServerRoom(id: number) {
  return request({
    url: `/cmdb/rooms/${id}`,
    method: 'delete'
  });
}

/**
 * 获取服务器标签列表
 */
export function fetchGetServerTags() {
  return request<CMDB.ServerTag[]>({
    url: '/cmdb/tags',
    method: 'get'
  });
}

/**
 * 创建服务器标签
 */
export function fetchCreateServerTag(data: CMDB.ServerTagForm) {
  return request<CMDB.ServerTag>({
    url: '/cmdb/tags',
    method: 'post',
    data
  });
}

/**
 * 更新服务器标签
 */
export function fetchUpdateServerTag(id: number, data: Partial<CMDB.ServerTagForm>) {
  return request({
    url: `/cmdb/tags/${id}`,
    method: 'put',
    data
  });
}

/**
 * 删除服务器标签
 */
export function fetchDeleteServerTag(id: number) {
  return request({
    url: `/cmdb/tags/${id}`,
    method: 'delete'
  });
}

/**
 * 为服务器分配标签
 */
export function fetchAssignServerTag(serverId: number, tagId: number) {
  return request({
    url: '/cmdb/tags/assign',
    method: 'post',
    data: { serverId, tagId }
  });
}

/**
 * 移除服务器标签
 */
export function fetchRemoveServerTag(serverId: number, tagId: number) {
  return request({
    url: `/cmdb/server-tags/${serverId}/${tagId}`,
    method: 'delete'
  });
}

/**
 * 获取资产变更记录
 */
export function fetchGetAssetChanges(params?: {
  assetType?: string;
  assetId?: number;
  page?: number;
  pageSize?: number;
}) {
  return request<CMDB.PageResponse<CMDB.AssetChange>>({
    url: '/cmdb/asset-changes',
    method: 'get',
    params
  });
}

/**
 * 获取SSH凭证列表
 * @param type 凭证类型筛选：'user'=用户连接凭证，'system'=系统运维凭证，不传=全部
 */
export function fetchGetSSHCredentials(type?: CMDB.CredentialType) {
  return request<CMDB.SSHCredential[]>({
    url: '/cmdb/ssh-credentials',
    method: 'get',
    params: type ? { type } : undefined
  });
}

/**
 * 创建SSH凭证
 */
export function fetchCreateSSHCredential(data: CMDB.SSHCredentialForm) {
  return request({
    url: '/cmdb/ssh-credentials',
    method: 'post',
    data
  });
}

/**
 * 更新SSH凭证
 */
export function fetchUpdateSSHCredential(id: number, data: Partial<CMDB.SSHCredentialForm>) {
  return request({
    url: `/cmdb/ssh-credentials/${id}`,
    method: 'put',
    data
  });
}

/**
 * 删除SSH凭证
 */
export function fetchDeleteSSHCredential(id: number) {
  return request({
    url: `/cmdb/ssh-credentials/${id}`,
    method: 'delete'
  });
}

/**
 * 测试SSH凭证连接
 */
export function fetchTestSSHCredential(id: number, testIp: string, testPort: number) {
  return request({
    url: `/cmdb/ssh-credentials/${id}/test`,
    method: 'post',
    data: { testIp, testPort }
  });
}

// ========== 堡垒机功能 ==========

/**
 * 连接服务器
 */
export function fetchConnectServer(
  serverId: number,
  data: {
    protocol: 'ssh' | 'sftp';
    credentialId: number;
  }
) {
  return request<{
    sessionId: number;
    websocketUrl: string;
    serverName: string;
    serverIp: string;
  }>({
    url: `/cmdb/servers/${serverId}/connect`,
    method: 'post',
    data
  });
}

/**
 * 检查连接权限（返回可用凭证列表）
 */
export function fetchSyncServerMetrics(serverId: number) {
  return request<null>({
    url: `/cmdb/servers/${serverId}/sync-metrics`,
    method: 'post'
  });
}

export function fetchCheckConnectPermission(serverId: number) {
  return request<{
    hasPermission: boolean;
    credentials: CMDB.SSHCredential[];
  }>({
    url: `/cmdb/servers/${serverId}/permission`,
    method: 'get'
  });
}

/**
 * 获取服务器的会话列表
 */
export function fetchGetServerSessions(
  serverId: number,
  params?: {
    page?: number;
    pageSize?: number;
  }
) {
  return request<CMDB.PageResponse<Bastion.BastionSession>>({
    url: `/cmdb/sessions`,
    method: 'get',
    params: {
      ...params,
      serverId // 使用 serverId 参数过滤
    }
  });
}

/**
 * 获取会话列表
 */
export function fetchGetSessions(params?: {
  serverId?: number;
  userId?: number;
  status?: string;
  protocol?: string;
  clientIp?: string;
  loginAccount?: string;
  startDate?: string;
  endDate?: string;
  page?: number;
  pageSize?: number;
}) {
  return request<CMDB.PageResponse<Bastion.BastionSession>>({
    url: '/cmdb/sessions',
    method: 'get',
    params
  });
}

/**
 * 获取会话列表（轻量级，只返回列表展示需要的字段）
 * 比完整接口快 5-6 倍，不返回敏感信息和冗余数据
 */
export function fetchGetSessionsList(params?: {
  serverId?: number;
  userId?: number;
  status?: string;
  protocol?: string;
  clientIp?: string;
  loginAccount?: string;
  startDate?: string;
  endDate?: string;
  page?: number;
  pageSize?: number;
}) {
  return request<CMDB.PageResponse<Bastion.BastionSession>>({
    url: '/cmdb/sessions/list',
    method: 'get',
    params
  });
}

/**
 * 获取活跃会话
 */
export function fetchGetActiveSessions() {
  return request<Bastion.BastionSession[]>({
    url: '/cmdb/sessions/active',
    method: 'get'
  });
}

/**
 * 从内存获取真正的活跃会话（WebSocket 连接仍然存在的会话）
 * 比查询数据库更准确，不会包含已断开但状态未更新的会话
 */
export function fetchGetActiveSessionsFromMemory() {
  return request<Bastion.BastionSession[]>({
    url: '/cmdb/sessions/active-memory',
    method: 'get'
  });
}

/**
 * 获取会话统计
 */
export function fetchGetSessionStats() {
  return request<{
    active: number;
    today: number;
  }>({
    url: '/cmdb/sessions/stats',
    method: 'get'
  });
}

/**
 * 获取会话详情
 */
export function fetchGetSessionById(sessionId: number) {
  return request<Bastion.BastionSession>({
    url: `/cmdb/sessions/${sessionId}`,
    method: 'get'
  });
}

/**
 * 终止会话
 */
export function fetchTerminateSession(sessionId: number) {
  return request({
    url: `/cmdb/sessions/${sessionId}/terminate`,
    method: 'post'
  });
}

/**
 * 获取会话命令列表
 */
export function fetchGetSessionCommands(sessionId: number) {
  return request<Bastion.BastionCommand[]>({
    url: `/cmdb/sessions/${sessionId}/commands`,
    method: 'get'
  });
}

/**
 * 获取命令列表
 */
export function fetchGetCommands(params?: {
  sessionId?: number;
  riskLevel?: string;
  blocked?: boolean;
  command?: string;
  startDate?: string;
  endDate?: string;
  page?: number;
  pageSize?: number;
}) {
  return request<CMDB.PageResponse<Bastion.BastionCommand>>({
    url: '/cmdb/commands',
    method: 'get',
    params
  });
}

/**
 * 获取会话文件传输记录
 */
export function fetchGetSessionFileTransfers(sessionId: number) {
  return request<Bastion.BastionFileTransfer[]>({
    url: `/cmdb/sessions/${sessionId}/file-transfers`,
    method: 'get'
  });
}

/**
 * 获取文件传输列表
 */
export function fetchGetFileTransfers(params?: {
  sessionId?: number;
  direction?: 'upload' | 'download';
  status?: string;
  startDate?: string;
  endDate?: string;
  page?: number;
  pageSize?: number;
}) {
  return request<CMDB.PageResponse<Bastion.BastionFileTransfer>>({
    url: '/cmdb/file-transfers',
    method: 'get',
    params
  });
}

/**
 * 调整终端大小
 */
export function fetchResizeTerminal(
  sessionId: number,
  data: {
    rows: number;
    cols: number;
  }
) {
  return request({
    url: `/cmdb/sessions/${sessionId}/resize`,
    method: 'post',
    data
  });
}

/**
 * 获取访问策略列表
 */
export function fetchGetAccessPolicies(params?: { page?: number; pageSize?: number }) {
  return request<CMDB.PageResponse<Bastion.AccessPolicy>>({
    url: '/cmdb/access-policies',
    method: 'get',
    params
  });
}

/**
 * 获取访问策略详情
 */
export function fetchGetAccessPolicyById(id: number) {
  return request<Bastion.AccessPolicy>({
    url: `/cmdb/access-policies/${id}`,
    method: 'get'
  });
}

/**
 * 创建访问策略
 */
export function fetchCreateAccessPolicy(data: Bastion.AccessPolicyForm) {
  return request({
    url: '/cmdb/access-policies',
    method: 'post',
    data
  });
}

/**
 * 更新访问策略
 */
export function fetchUpdateAccessPolicy(id: number, data: Partial<Bastion.AccessPolicyForm>) {
  return request({
    url: `/cmdb/access-policies/${id}`,
    method: 'put',
    data
  });
}

/**
 * 删除访问策略
 */
export function fetchDeleteAccessPolicy(id: number) {
  return request({
    url: `/cmdb/access-policies/${id}`,
    method: 'delete'
  });
}

/**
 * 部署 Agent
 */
export function fetchDeployAgent(serverId: number) {
  return request<null>({
    url: `/cmdb/servers/${serverId}/agent/deploy`,
    method: 'post'
  });
}

/**
 * 重启 Agent
 */
export function fetchRestartAgent(serverId: number) {
  return request<null>({
    url: `/cmdb/servers/${serverId}/agent/restart`,
    method: 'post'
  });
}

/**
 * 卸载 Agent
 */
export function fetchUninstallAgent(serverId: number) {
  return request<null>({
    url: `/cmdb/servers/${serverId}/agent/uninstall`,
    method: 'post'
  });
}

/**
 * 查询 Agent 状态
 */
export function fetchGetAgentStatus(serverId: number) {
  return request<{ agentStatus: string; agentPort: number; agentVersion: string; lastHeartbeatAt: string }>({
    url: `/cmdb/servers/${serverId}/agent/status`,
    method: 'get'
  });
}

/**
 * 获取 Agent 管理列表
 */
export function fetchGetAgentList(params?: {
  hostname?: string;
  ip?: string;
  agentStatus?: string;
  page?: number;
  pageSize?: number;
}) {
  return request<CMDB.PageResponse<CMDB.Server>>({
    url: '/cmdb/agents',
    method: 'get',
    params
  });
}

/**
 * 批量部署 Agent
 */
export function fetchBatchDeployAgent(serverIds: number[]) {
  return request<null>({
    url: '/cmdb/agents/batch-deploy',
    method: 'post',
    data: { serverIds }
  });
}

/**
 * 批量卸载 Agent
 */
export function fetchBatchUninstallAgent(serverIds: number[]) {
  return request<null>({
    url: '/cmdb/agents/batch-uninstall',
    method: 'post',
    data: { serverIds }
  });
}

/**
 * 删除 Agent 记录（仅清空数据库，不 SSH）
 */
export function fetchDeleteAgentRecord(serverId: number) {
  return request<null>({
    url: `/cmdb/agents/${serverId}`,
    method: 'delete'
  });
}

/**
 * 获取主机属性列表
 */
export function fetchGetServerAttributes(serverId: number) {
  return request<System.ServerAttribute[]>({
    url: `/system/server-attributes/${serverId}`,
    method: 'get'
  });
}

/**
 * 保存主机属性列表
 */
export function fetchSaveServerAttributes(serverId: number, data: System.ServerAttribute[]) {
  return request({
    url: `/system/server-attributes/${serverId}`,
    method: 'post',
    data
  });
}

/**
 * 测试SSH连接
 */
export function fetchTestSSHConnection(serverId: number) {
  return request<{
    success: boolean;
    message: string;
    host?: string;
    port?: number;
    user?: string;
    latency?: string;
  }>({
    url: `/cmdb/servers/${serverId}/test-connection`,
    method: 'post'
  });
}

// ========================================
// Agent 版本管理 API
// ========================================

/**
 * 获取 Agent 版本列表
 */
export function fetchGetAgentVersions() {
  return request<CMDB.AgentVersion[]>({
    url: '/cmdb/agent-versions',
    method: 'get'
  });
}

/**
 * 获取最新 Agent 版本
 */
export function fetchGetLatestAgentVersion() {
  return request<CMDB.AgentVersion>({
    url: '/cmdb/agent-versions/latest',
    method: 'get'
  });
}

/**
 * 获取指定版本详情
 */
export function fetchGetAgentVersionByID(id: number) {
  return request<CMDB.AgentVersion>({
    url: `/cmdb/agent-versions/${id}`,
    method: 'get'
  });
}

/**
 * 创建 Agent 版本
 */
export function fetchCreateAgentVersion(data: CMDB.AgentVersionForm) {
  return request({
    url: '/cmdb/agent-versions',
    method: 'post',
    data
  });
}

/**
 * 更新 Agent 版本
 */
export function fetchUpdateAgentVersion(id: number, data: Partial<CMDB.AgentVersionForm>) {
  return request({
    url: `/cmdb/agent-versions/${id}`,
    method: 'put',
    data
  });
}

/**
 * 删除 Agent 版本
 */
export function fetchDeleteAgentVersion(id: number) {
  return request({
    url: `/cmdb/agent-versions/${id}`,
    method: 'delete'
  });
}

// ========================================
// Agent 升级管理 API
// ========================================

/**
 * 升级单台主机的 Agent
 */
export function fetchUpgradeAgent(serverId: number, targetVersion: string) {
  return request({
    url: `/cmdb/servers/${serverId}/agent/upgrade`,
    method: 'post',
    data: { targetVersion }
  });
}

/**
 * 获取升级任务列表
 */
export function fetchGetUpgradeTasks(params?: { page?: number; pageSize?: number; status?: string }) {
  return request<{
    list: CMDB.AgentUpgradeTask[];
    total: number;
  }>({
    url: '/cmdb/agent-upgrade-tasks',
    method: 'get',
    params
  });
}

/**
 * 获取升级任务详情
 */
export function fetchGetUpgradeTaskByID(taskId: number) {
  return request<CMDB.AgentUpgradeTask>({
    url: `/cmdb/agent-upgrade-tasks/${taskId}`,
    method: 'get'
  });
}
