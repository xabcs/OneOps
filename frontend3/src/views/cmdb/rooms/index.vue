<script setup lang="tsx">
  import { onMounted, ref } from 'vue';
  import { useDebounceFn } from '@vueuse/core';
  import { Plus, Refresh } from '@element-plus/icons-vue';
  import { fetchDeleteServerRoom, fetchGetServerRooms } from '@/service/api';
  import { createTagMap } from '@/utils/common';
  import RoomOperateDrawer from './modules/room-operate-drawer.vue';

  defineOptions({ name: 'CmdbRooms' });

  const loading = ref(false);
  const tableData = ref<CMDB.ServerRoom[]>([]);
  const searchKeyword = ref('');

  const drawerVisible = ref(false);
  const operateType = ref<UI.TableOperateType>('add');
  const editingData = ref<CMDB.ServerRoom | null>(null);

  async function getData() {
    loading.value = true;
    try {
      const { data, error } = await fetchGetServerRooms();
      if (!error && data) {
        const keyword = searchKeyword.value.trim().toLowerCase();
        tableData.value = keyword
          ? (data || []).filter(item =>
              [item.name, item.code, item.location, item.provider].some(value => value?.toLowerCase().includes(keyword))
            )
          : data || [];
      }
    } finally {
      loading.value = false;
    }
  }

  function handleAdd() {
    operateType.value = 'add';
    editingData.value = null;
    drawerVisible.value = true;
  }

  function handleEdit(row: CMDB.ServerRoom) {
    operateType.value = 'edit';
    editingData.value = row;
    drawerVisible.value = true;
  }

  async function handleDelete(row: CMDB.ServerRoom) {
    try {
      await ElMessageBox.confirm(`确定要删除机房 "${row.name}" 吗？`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      });
      const { error } = await fetchDeleteServerRoom(row.id);
      if (!error) {
        ElMessage.success('删除成功');
        getData();
      }
    } catch {
      // 用户取消
    }
  }

  /** 机房状态（1 启用 / 0 禁用）→ ElTag 标签映射 */
  const roomStatusTag = createTagMap({
    '1': { text: '启用', type: 'success' },
    '0': { text: '禁用', type: 'danger' }
  });

  function getStatusTag(status: number) {
    return roomStatusTag(String(status));
  }

  // 搜索输入处理（防抖 300ms，避免每敲一字就发起全量请求）
  const handleSearchInput = useDebounceFn(() => {
    getData();
  }, 300);

  // 重置搜索关键字并刷新
  function handleReset() {
    searchKeyword.value = '';
    getData();
  }

  onMounted(() => {
    getData();
  });
</script>

<template>
  <ListPageLayout
    title="机房管理"
    description="维护机房基础信息、位置与服务商联系方式"
    @search="getData"
    @reset="handleReset"
  >
    <!-- 搜索筛选 -->
    <template #search>
      <ElInput
        v-model="searchKeyword"
        placeholder="搜索机房名称或代码"
        clearable
        class="w-200px"
        @input="handleSearchInput"
      >
        <template #prefix>
          <icon-mdi-magnify class="align-sub text-icon" />
        </template>
      </ElInput>
    </template>

    <!-- 工具栏 -->
    <template #toolbar>
      <ElButton :icon="Refresh" @click="getData">刷新</ElButton>
      <PermissionButton code="cmdb.rooms.create" type="primary" :icon="Plus" @click="handleAdd">
        新增机房
      </PermissionButton>
    </template>

    <!-- 表格 -->
    <ElTable v-loading="loading" :data="tableData" border stripe height="100%">
      <ElTableColumn prop="id" label="ID" width="70" align="center" />
      <ElTableColumn prop="name" label="机房名称" min-width="120" align="center" />
      <ElTableColumn prop="code" label="机房代码" width="120" align="center" />
      <ElTableColumn prop="location" label="位置" min-width="150" align="center" show-overflow-tooltip />
      <ElTableColumn prop="address" label="详细地址" min-width="200" align="center" show-overflow-tooltip />
      <ElTableColumn prop="provider" label="服务商" width="100" align="center" />
      <ElTableColumn prop="contact" label="联系人" width="100" align="center" />
      <ElTableColumn prop="phone" label="联系电话" width="120" align="center" />
      <ElTableColumn label="状态" width="80" align="center">
        <template #default="{ row }">
          <ElTag :type="getStatusTag(row.status).type" size="small">
            {{ getStatusTag(row.status).text }}
          </ElTag>
        </template>
      </ElTableColumn>
      <ElTableColumn prop="remarks" label="备注" min-width="150" align="center" show-overflow-tooltip />
      <ElTableColumn label="操作" width="180" align="center" fixed="right" class-name="msre-table-actions">
        <template #default="{ row }">
          <PermissionButton link type="primary" size="small" code="cmdb.rooms.update" @click="handleEdit(row)">
            编辑
          </PermissionButton>
          <PermissionButton link type="danger" size="small" code="cmdb.rooms.delete" @click="handleDelete(row)">
            删除
          </PermissionButton>
        </template>
      </ElTableColumn>
    </ElTable>

    <RoomOperateDrawer
      v-model:visible="drawerVisible"
      :operate-type="operateType"
      :row-data="editingData"
      @submitted="getData"
    />
  </ListPageLayout>
</template>
