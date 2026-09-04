<script setup lang="tsx">
  import { onMounted, ref } from 'vue';
  import { Plus, Refresh } from '@element-plus/icons-vue';
  import { fetchDeleteServerTag, fetchGetServerTags } from '@/service/api';
  import TagOperateDrawer from './modules/tag-operate-drawer.vue';

  defineOptions({ name: 'CmdbTags' });

  // 非分页接口：fetchGetServerTags 返回数组
  const loading = ref(false);
  const tableData = ref<CMDB.ServerTag[]>([]);

  // drawer 状态（手动，因为非 useUIPaginatedTable）
  const drawerVisible = ref(false);
  const operateType = ref<UI.TableOperateType>('add');
  const editingData = ref<CMDB.ServerTag | null>(null);

  async function getData() {
    loading.value = true;
    try {
      const { data, error } = await fetchGetServerTags();
      if (!error) {
        tableData.value = data || [];
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

  function handleEdit(row: CMDB.ServerTag) {
    operateType.value = 'edit';
    editingData.value = row;
    drawerVisible.value = true;
  }

  async function handleDelete(row: CMDB.ServerTag) {
    try {
      await ElMessageBox.confirm(`确定要删除标签 "${row.name}" 吗？`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      });
      const { error } = await fetchDeleteServerTag(row.id);
      if (!error) {
        ElMessage.success('删除成功');
        getData();
      }
    } catch {
      // 用户取消
    }
  }

  onMounted(() => {
    getData();
  });
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <ElCard class="card-wrapper">
      <div class="mb-16px flex justify-between">
        <ElButton :icon="Refresh" @click="getData">刷新</ElButton>
        <PermissionButton code="cmdb.tags.create" type="primary" :icon="Plus" @click="handleAdd">
          新增标签
        </PermissionButton>
      </div>

      <ElTable v-loading="loading" :data="tableData" border stripe>
        <ElTableColumn prop="id" label="ID" width="80" />
        <ElTableColumn prop="name" label="标签名称" min-width="150" />
        <ElTableColumn label="颜色" width="100">
          <template #default="{ row }">
            <span
              v-if="row.color"
              :style="{ backgroundColor: row.color }"
              class="inline-block h-22px w-40px rounded-4px align-middle"
            />
          </template>
        </ElTableColumn>
        <ElTableColumn prop="description" label="描述" min-width="220" show-overflow-tooltip />
        <ElTableColumn prop="sortOrder" label="排序" width="80" />
        <ElTableColumn label="状态" width="90">
          <template #default="{ row }">
            <ElTag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <PermissionButton code="cmdb.tags.update" type="primary" size="small" @click="handleEdit(row)">
              编辑
            </PermissionButton>
            <PermissionButton code="cmdb.tags.delete" type="danger" size="small" @click="handleDelete(row)">
              删除
            </PermissionButton>
          </template>
        </ElTableColumn>
      </ElTable>

      <TagOperateDrawer
        v-model:visible="drawerVisible"
        :operate-type="operateType"
        :row-data="editingData"
        @submitted="getData"
      />
    </ElCard>
  </div>
</template>
