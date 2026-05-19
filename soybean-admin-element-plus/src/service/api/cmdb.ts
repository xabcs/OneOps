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
export function fetchGetServerConfig(data: {
  hostname: string;
  ip: string;
  sshUser?: string;
  sshPort?: number;
}) {
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
 * 获取指定分组下的服务器列表
 */
export function fetchGetServersByGroup(groupId: number, params?: {
  page?: number;
  pageSize?: number;
}) {
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
 */
export function fetchGetSSHCredentials() {
  return request<CMDB.SSHCredential[]>({
    url: '/cmdb/ssh-credentials',
    method: 'get'
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
