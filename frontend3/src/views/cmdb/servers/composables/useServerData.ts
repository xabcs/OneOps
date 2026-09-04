/**
 * 服务器数据获取与列表管理
 * 从 index.vue 拆分：主机列表加载、分页、选择、关联数据（凭证/机房/机柜/标签/业务系统/属性）
 */

import { reactive, ref } from 'vue';
import { ElNotification } from 'element-plus';
import {
  fetchGetAttributes,
  fetchGetBusinessUnits,
  fetchGetCabinets,
  fetchGetSSHCredentials,
  fetchGetServerAttributes,
  fetchGetServerRooms,
  fetchGetServerTags,
  fetchGetServers,
  fetchSaveServerAttributes,
  fetchSyncServerMetrics
} from '@/service/api';
import type { SearchType } from '../types/server.types';

export function useServerData() {
  // ===== 列表数据 =====
  const tableData = ref<CMDB.Server[]>([]);
  const loading = ref(false);
  const total = ref(0);
  const selectedIds = ref<number[]>([]);

  // 分页
  const pagination = reactive({ page: 1, pageSize: 20 });

  // ===== 关联数据 =====
  const userCredentials = ref<CMDB.SSHCredential[]>([]);
  const systemCredentials = ref<CMDB.SSHCredential[]>([]);
  const businessUnits = ref<CMDB.BusinessUnit[]>([]);
  const serverRooms = ref<CMDB.ServerRoom[]>([]);
  const cabinets = ref<CMDB.Cabinet[]>([]);
  const serverTags = ref<CMDB.ServerTag[]>([]);
  const attributeDefinitions = ref<Api.SystemManage.AttributeDefinition[]>([]);

  // ===== 数据获取 =====
  async function getServers(searchParams?: {
    searchType: SearchType;
    searchKeyword: string;
    groupId?: number;
    ungrouped?: boolean;
    tagId?: number;
  }) {
    loading.value = true;
    try {
      const params: CMDB.ServerQuery = { page: pagination.page, pageSize: pagination.pageSize };

      if (searchParams?.groupId) {
        params.groupId = searchParams.groupId;
      }

      if (searchParams?.ungrouped) {
        params.ungrouped = true;
      }

      if (searchParams?.tagId) {
        params.tagId = searchParams.tagId;
      }

      if (searchParams?.searchKeyword) {
        switch (searchParams.searchType) {
          case 'hostname':
            params.hostname = searchParams.searchKeyword;
            break;
          case 'ip':
            params.ip = searchParams.searchKeyword;
            break;
        }
      }

      const { data } = await fetchGetServers(params);
      tableData.value = data?.list || [];
      total.value = data?.total || 0;
    } catch (error) {
      console.error('获取主机列表失败:', error);
      ElNotification.error('获取主机列表失败');
    } finally {
      loading.value = false;
    }
  }

  async function getSSHCredentials() {
    try {
      const [userRes, systemRes] = await Promise.all([
        fetchGetSSHCredentials('user'),
        fetchGetSSHCredentials('system')
      ]);
      userCredentials.value = userRes.data || [];
      systemCredentials.value = systemRes.data || [];
    } catch (error) {
      console.error('获取SSH凭证失败:', error);
    }
  }

  async function getServerRooms() {
    try {
      const { data } = await fetchGetServerRooms();
      serverRooms.value = data || [];
    } catch (error) {
      console.error('获取机房列表失败:', error);
    }
  }

  async function getCabinets(roomId?: number) {
    try {
      const { data } = await fetchGetCabinets(roomId);
      cabinets.value = data || [];
    } catch (error) {
      console.error('获取机柜列表失败:', error);
    }
  }

  async function getServerTags() {
    try {
      const { data } = await fetchGetServerTags();
      serverTags.value = data || [];
    } catch (error) {
      console.error('获取标签列表失败:', error);
    }
  }

  async function getBusinessUnits() {
    try {
      const res = await fetchGetBusinessUnits();
      businessUnits.value = res.data || [];
    } catch {
      /* ignore */
    }
  }

  async function getSystemOptions() {
    try {
      const { data } = await fetchGetAttributes();
      const sortedAttrs = (data || []).sort(
        (a: Api.SystemManage.AttributeDefinition, b: Api.SystemManage.AttributeDefinition) => a.sortOrder - b.sortOrder
      );
      attributeDefinitions.value = sortedAttrs;
    } catch (error) {
      console.error('加载系统选项失败:', error);
      attributeDefinitions.value = [];
    }
  }

  // ===== 选择 =====
  function handleSelectAll(selection: CMDB.Server[]) {
    selectedIds.value = selection.map(s => s.id);
  }

  function handleSelectionChange(selection: CMDB.Server[]) {
    selectedIds.value = selection.map(s => s.id);
  }

  // ===== 分页 =====
  function handlePageChange(page: number) {
    pagination.page = page;
  }

  function handlePageSizeChange(pageSize: number) {
    pagination.pageSize = pageSize;
    pagination.page = 1;
  }

  // ===== 属性辅助 =====
  async function loadAttributeDefinitions() {
    try {
      const { data } = await fetchGetAttributes();
      attributeDefinitions.value = data || [];
    } catch (error) {
      console.error('加载属性定义失败:', error);
    }
  }

  async function loadServerAttributes(serverId: number) {
    try {
      const { data } = await fetchGetServerAttributes(serverId);
      return data || [];
    } catch (error) {
      console.error('加载主机属性失败:', error);
      return [];
    }
  }

  async function saveServerAttributes(serverId: number, serverAttributes: Api.SystemManage.ServerAttribute[]) {
    try {
      const attributesToSave = serverAttributes
        .filter(attr => attr.attributeValue && attr.attributeValue.trim() !== '')
        .map(attr => ({
          attributeId: attr.attributeId,
          attributeKey: attr.attributeKey,
          attributeValue: attr.attributeValue
        }));
      if (attributesToSave.length > 0) {
        await fetchSaveServerAttributes(serverId, attributesToSave);
      }
    } catch (error) {
      console.error('保存主机属性失败:', error);
    }
  }

  // ===== 同步指标 =====
  async function handleSyncMetrics(row: CMDB.Server) {
    try {
      await fetchSyncServerMetrics(row.id);
      ElNotification.info('采集中，3秒后自动刷新...');
    } catch {
      ElNotification.error('提交采集任务失败');
    }
  }

  // ===== 辅助 =====
  function handleRoomChange(roomId: number, serverForm: CMDB.ServerForm) {
    serverForm.cabinetId = undefined;
    if (roomId) {
      getCabinets(roomId);
    } else {
      cabinets.value = [];
    }
  }

  return {
    // 数据
    tableData,
    loading,
    total,
    selectedIds,
    pagination,
    userCredentials,
    systemCredentials,
    businessUnits,
    serverRooms,
    cabinets,
    serverTags,
    attributeDefinitions,
    // 方法
    getServers,
    getSSHCredentials,
    getServerRooms,
    getCabinets,
    getServerTags,
    getBusinessUnits,
    getSystemOptions,
    handleSelectAll,
    handleSelectionChange,
    handlePageChange,
    handlePageSizeChange,
    loadAttributeDefinitions,
    loadServerAttributes,
    saveServerAttributes,
    handleSyncMetrics,
    handleRoomChange
  };
}
