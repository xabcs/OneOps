/**
 * CMDB 模块 API 服务
 * 统一管理所有 CMDB 相关的 API 调用
 */

import { request } from '@/service/request'

/**
 * 服务器相关 API
 */

// 获取服务器列表
export function fetchGetServers(params?: {
  page?: number
  pageSize?: number
  keyword?: string
  env?: string
  status?: number
  agentStatus?: string
  groupId?: number
}) {
  return request<Server.PageResult>({
    url: '/api/v1/cmdb/servers',
    method: 'get',
    params
  })
}

// 获取服务器详情
export function fetchGetServerById(id: number) {
  return request<CMDB.Server>({
    url: `/api/v1/cmdb/servers/${id}`,
    method: 'get'
  })
}

// 创建服务器
export function fetchCreateServer(data: CMDB.ServerFormData) {
  return request<{ id: number }>({
    url: '/api/v1/cmdb/servers',
    method: 'post',
    data
  })
}

// 更新服务器
export function fetchUpdateServer(id: number, data: Partial<CMDB.ServerFormData>) {
  return request({
    url: `/api/v1/cmdb/servers/${id}`,
    method: 'put',
    data
  })
}

// 删除服务器
export function fetchDeleteServer(id: number) {
  return request({
    url: `/api/v1/cmdb/servers/${id}`,
    method: 'delete'
  })
}

// 批量部署 Agent
export function fetchBatchDeployAgent(data: { serverIds: number[] }) {
  return request({
    url: '/api/v1/cmdb/servers/batch/deploy-agent',
    method: 'post',
    data
  })
}

// 批量卸载 Agent
export function fetchBatchUninstallAgent(data: { serverIds: number[] }) {
  return request({
    url: '/api/v1/cmdb/servers/batch/uninstall-agent',
    method: 'post',
    data
  })
}

// 同步服务器指标
export function fetchSyncServerMetrics(id: number) {
  return request({
    url: `/api/v1/cmdb/servers/${id}/sync-metrics`,
    method: 'post'
  })
}

/**
 * 分组相关 API
 */

// 获取服务器分组列表
export function fetchGetServerGroups() {
  return request<CMDB.ServerGroup[]>({
    url: '/api/v1/cmdb/groups',
    method: 'get'
  })
}

// 创建服务器分组
export function fetchCreateServerGroup(data: CMDB.GroupFormData) {
  return request<{ id: number }>({
    url: '/api/v1/cmdb/groups',
    method: 'post',
    data
  })
}

// 更新服务器分组
export function fetchUpdateServerGroup(id: number, data: Partial<CMDB.GroupFormData>) {
  return request({
    url: `/api/v1/cmdb/groups/${id}`,
    method: 'put',
    data
  })
}

// 删除服务器分组
export function fetchDeleteServerGroup(id: number) {
  return request({
    url: `/api/v1/cmdb/groups/${id}`,
    method: 'delete'
  })
}

// 分配服务器到分组
export function fetchAssignServerToGroups(data: {
  serverIds: number[]
  groupIds: number[]
}) {
  return request({
    url: '/api/v1/cmdb/servers/assign-groups',
    method: 'post',
    data
  })
}

/**
 * 属性相关 API
 */

// 获取服务器属性定义
export function fetchGetServerAttributes() {
  return request<CMDB.Attribute[]>({
    url: '/api/v1/cmdb/attributes',
    method: 'get'
  })
}

// 保存服务器属性
export function fetchSaveServerAttributes(data: {
  serverId: number
  attributes: Record<string, any>
}) {
  return request({
    url: `/api/v1/cmdb/servers/${data.serverId}/attributes`,
    method: 'post',
    data: { attributes: data.attributes }
  })
}

/**
 * 凭证相关 API
 */

// 获取 SSH 凭证列表
export function fetchGetSSHCredentials() {
  return request<CMDB.SSHCredential[]>({
    url: '/api/v1/cmdb/credentials',
    method: 'get'
  })
}

// 测试 SSH 连接
export function fetchTestSSHConnection(data: {
  serverId: number
  credentialId: number
}) {
  return request({
    url: '/api/v1/cmbd/servers/test-connection',
    method: 'post',
    data
  })
}

/**
 * 资产相关 API
 */

// 获取机房列表
export function fetchGetServerRooms() {
  return request<CMDB.Room[]>({
    url: '/api/v1/cmdb/rooms',
    method: 'get'
  })
}

// 获取机柜列表
export function fetchGetCabinets(roomId?: number) {
  return request<CMDB.Cabinet[]>({
    url: '/api/v1/cmdb/cabinets',
    method: 'get',
    params: roomId ? { roomId } : undefined
  })
}

// 获取业务单元列表
export function fetchGetBusinessUnits() {
  return request<CMDB.BusinessUnit[]>({
    url: '/api/v1/cmdb/business-units',
    method: 'get'
  })
}

// 获取服务器标签
export function fetchGetServerTags() {
  return request<string[]>({
    url: '/api/v1/cmdb/tags',
    method: 'get'
  })
}

// 导出索引
export default {
  // 服务器
  fetchGetServers,
  fetchGetServerById,
  fetchCreateServer,
  fetchUpdateServer,
  fetchDeleteServer,
  fetchBatchDeployAgent,
  fetchBatchUninstallAgent,
  fetchSyncServerMetrics,

  // 分组
  fetchGetServerGroups,
  fetchCreateServerGroup,
  fetchUpdateServerGroup,
  fetchDeleteServerGroup,
  fetchAssignServerToGroups,

  // 属性
  fetchGetServerAttributes,
  fetchSaveServerAttributes,

  // 凭证
  fetchGetSSHCredentials,
  fetchTestSSHConnection,

  // 资产
  fetchGetServerRooms,
  fetchGetCabinets,
  fetchGetBusinessUnits,
  fetchGetServerTags
}