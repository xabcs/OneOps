<script setup lang="tsx">
  import { onMounted, ref } from 'vue';
  import { Plus, Refresh } from '@element-plus/icons-vue';
  import { fetchDeleteServerRoom, fetchGetServerRooms } from '@/service/api';
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

  function getStatusTag(status: number): { text: string; type: 'success' | 'danger' } {
    return status === 1 ? { text: '启用', type: 'success' } : { text: '禁用', type: 'danger' };
  }

  onMounted(() => {
    getData();
  });
</script>

<template>
  <div class="table-page">
    <ElCard class="card-wrapper">
      <div class="mb-16px flex justify-between">
        <ElSpace>
          <ElInput
            v-model="searchKeyword"
            placeholder="搜索机房名称或代码"
            clearable
            style="width: 200px"
            @input="getData"
          >
            <template #prefix>
              <icon-mdi-magnify class="align-sub text-icon" />
            </template>
          </ElInput>
          <ElButton :icon="Refresh" @click="getData">刷新</ElButton>
        </ElSpace>
        <PermissionButton code="cmdb.rooms.create" type="primary" :icon="Plus" @click="handleAdd">
          新增机房
        </PermissionButton>
      </div>

      <div class="table-scroll-wrap">
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
      </div>

      <RoomOperateDrawer
        v-model:visible="drawerVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="getData"
      />
    </ElCard>
  </div>
</template>

<style scoped lang="scss">
  .card-wrapper {
    @apply flex-col-stretch p-16px;
  }
</style>
