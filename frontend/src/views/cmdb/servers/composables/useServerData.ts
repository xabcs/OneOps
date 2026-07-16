/**
 * 服务器数据管理逻辑
 */

import { ref, computed } from 'vue'
import {
  fetchGetServers,
  fetchGetServerAttributes,
  fetchGetSSHCredentials,
  fetchGetServerTags,
  fetchUpdateServer,
  fetchCreateServer,
  fetchDeleteServer,
  fetchBatchDeployAgent,
  fetchBatchUninstallAgent,
  fetchSyncServerMetrics,
  type Server
} from '@/service/api'
import type { ServerFormData, ServerFilters, AgentStatus } from '../types/server.types'

export function useServerData() {
  // 数据状态
  const servers = ref<Server[]>([])
  const serverAttributes = ref<CMDB.Attribute[]>([])
  const userCredentials = ref<CMDB.SSHCredential[]>([])
  const systemCredentials = ref<CMDB.SSHCredential[]>([])
  const serverTags = ref<string[]>([])

  // 加载状态
  const loading = ref(false)
  const serverDetailLoading = ref(false)

  // 统计信息
  const serverStats = computed(() => {
    const stats = {
      total: servers.value.length,
      online: 0,
      offline: 0,
      warning: 0,
      healthy: 0
    }

    servers.value.forEach(server => {
      if (server.agentStatus === 'running') {
        stats.online++
      } else if (server.agentStatus === 'offline') {
        stats.offline++
      } else if (server.agentStatus === 'uninstalled') {
        stats.uninstalled++
      }

      // 简单的健康状态判断
      if (server.status === 1 && server.agentStatus === 'running') {
        stats.healthy++
      } else if (server.agentStatus !== 'running') {
        stats.warning++
      }
    })

    return stats
  })

  /**
   * 获取服务器列表
   */
  const getServers = async (filters: ServerFilters = {}) => {
    loading.value = true
    try {
      const params = {
        page: 1,
        pageSize: 1000,
        keyword: filters.keyword || '',
        env: filters.env || '',
        status: filters.status !== undefined ? filters.status : undefined,
        agentStatus: filters.agentStatus || '',
        groupId: filters.groupId !== undefined ? filters.groupId : undefined
      }

      const { data } = await fetchGetServers(params)
      if (data) {
        servers.value = data.list || []
      }
    } finally {
      loading.value = false
    }
  }

  /**
   * 获取服务器详情（含属性）
   */
  const getServerDetail = async (serverId: number) => {
    serverDetailLoading.value = true
    try {
      const [attrData, credData, tagData] = await Promise.all([
        fetchGetServerAttributes(),
        fetchGetSSHCredentials(),
        fetchGetServerTags()
      ])

      if (attrData?.data) {
        serverAttributes.value = attrData.data
      }
      if (credData?.data) {
        userCredentials.value = credData.data.filter(c => c.credentialType === 'user')
        systemCredentials.value = credData.data.filter(c => c.credentialType === 'system')
      }
      if (tagData?.data) {
        serverTags.value = tagData.data
      }
    } finally {
      serverDetailLoading.value = false
    }
  }

  /**
   * 创建服务器
   */
  const createServer = async (formData: ServerFormData) => {
    try {
      const response = await fetchCreateServer({
        hostname: formData.hostname || '',
        ip: formData.ip || '',
        innerIp: formData.innerIp || '',
        sshPort: formData.sshPort || 22,
        env: formData.env || '',
        provider: formData.provider || '',
        groupId: formData.groupId || [],
        credentialId: formData.credentialId || 0,
        systemCredentialId: formData.systemCredentialId || 0,
        cabinetId: formData.cabinetId || 0,
        remarks: formData.remarks || ''
      })

      if (response.data) {
        await getServers()
        return { success: true, data: response.data }
      }
      return { success: false, message: '创建失败' }
    } catch (error) {
      const message = error && typeof error === 'object' && 'message' in error
        ? (typeof error.message === 'string' ? error.message : '创建失败')
        : '创建失败'
      return { success: false, message }
    }
  }

  /**
   * 更新服务器
   */
  const updateServer = async (serverId: number, formData: ServerFormData) => {
    try {
      const response = await fetchUpdateServer(serverId, {
        hostname: formData.hostname,
        ip: formData.ip,
        innerIp: formData.innerIp,
        sshPort: formData.sshPort,
        env: formData.env,
        status: formData.status,
        groupId: formData.groupId || [],
        credentialId: formData.credentialId || 0,
        systemCredentialId: formData.systemCredentialId || 0,
        cabinetId: formData.cabinetId || 0,
        remarks: formData.remarks || ''
      })

      if (response.data) {
        await getServers()
        return { success: true, data: response.data }
      }
      return { success: false, message: '更新失败' }
    } catch (error) {
      const message = error && typeof error === 'object' && 'message' in error
        ? (typeof error.message === 'string' ? error.message : '更新失败')
        : '更新失败'
      return { success: false, message }
    }
  }

  /**
   * 删除服务器
   */
  const deleteServer = async (serverId: number) => {
    try {
      await fetchDeleteServer(serverId)
      await getServers()
      return { success: true }
    } catch (error) {
      const message = error && typeof error === 'object' && 'message' in error
        ? (typeof error.message === 'string' ? error.message : '删除失败')
        : '删除失败'
      return { success: false, message }
    }
  }

  /**
   * 批量部署Agent
   */
  const batchDeployAgent = async (serverIds: number[]) => {
    let successCount = 0
    let failCount = 0

    for (const serverId of serverIds) {
      try {
        await fetchBatchDeployAgent({ serverIds: [serverId] })
        successCount++
      } catch {
        failCount++
      }
    }

    await getServers()
    return { successCount, failCount }
  }

  /**
   * 批量卸载Agent
   */
  const batchUninstallAgent = async (serverIds: number[]) => {
    let successCount = 0
    let failCount = 0

    for (const serverId of serverIds) {
      try {
        await fetchBatchUninstallAgent({ serverIds: [serverId] })
        successCount++
      } catch {
        failCount++
      }
    }

    await getServers()
    return { successCount, failCount }
  }

  /**
   * 同步服务器指标
   */
  const syncServerMetrics = async (serverId: number) => {
    try {
      await fetchSyncServerMetrics(serverId)
      await getServers()
      return { success: true }
    } catch (error) {
      return { success: false, message: '同步失败' }
    }
  }

  return {
    // 状态
    servers,
    serverAttributes,
    userCredentials,
    systemCredentials,
    serverTags,
    loading,
    serverDetailLoading,
    serverStats,

    // 方法
    getServers,
    getServerDetail,
    createServer,
    updateServer,
    deleteServer,
    batchDeployAgent,
    batchUninstallAgent,
    syncServerMetrics
  }
}