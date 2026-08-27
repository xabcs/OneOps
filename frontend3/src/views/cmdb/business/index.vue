<script setup lang="tsx">
  import { computed, onMounted, ref } from 'vue';
  import { Plus, Refresh } from '@element-plus/icons-vue';
  import { fetchDeleteBusinessUnit, fetchGetBusinessUnits } from '@/service/api';
  import BusinessOperateDrawer from './modules/business-operate-drawer.vue';

  defineOptions({ name: 'CmdbBusiness' });

  const loading = ref(false);
  const tableData = ref<CMDB.BusinessUnit[]>([]);

  const drawerVisible = ref(false);
  const operateType = ref<UI.TableOperateType>('add');
  const editingData = ref<CMDB.BusinessUnit | null>(null);

  const businessTreeOptions = computed(() => [{ id: 0, name: '根业务', children: tableData.value }]);

  async function getData() {
    loading.value = true;
    try {
      const { data, error } = await fetchGetBusinessUnits();
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

  function handleEdit(row: CMDB.BusinessUnit) {
    operateType.value = 'edit';
    editingData.value = row;
    drawerVisible.value = true;
  }

  async function handleDelete(row: CMDB.BusinessUnit) {
    if (row.children && row.children.length > 0) {
      ElMessage.warning('该业务下有子业务，无法删除');
      return;
    }
    try {
      await ElMessageBox.confirm(`确定要删除业务 "${row.name}" 吗？`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      });
      const { error } = await fetchDeleteBusinessUnit(row.id);
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
        <PermissionButton code="cmdb.business.create" type="primary" :icon="Plus" @click="handleAdd">新增业务</PermissionButton>
      </div>

      <ElTable v-loading="loading" :data="tableData" row-key="id" border stripe :tree-props="{ children: 'children' }">
        <ElTableColumn prop="id" label="ID" width="80" />
        <ElTableColumn prop="name" label="业务名称" min-width="160" show-overflow-tooltip />
        <ElTableColumn prop="code" label="业务代码" width="140" />
        <ElTableColumn prop="owner" label="负责人" width="120" />
        <ElTableColumn prop="phone" label="联系电话" width="140" />
        <ElTableColumn prop="sortOrder" label="排序" width="80" />
        <ElTableColumn label="状态" width="90">
          <template #default="{ row }">
            <ElTag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="remarks" label="备注" min-width="180" show-overflow-tooltip />
        <ElTableColumn label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <PermissionButton code="cmdb.business.update" type="primary" size="small" @click="handleEdit(row)">编辑</PermissionButton>
            <PermissionButton code="cmdb.business.delete" type="danger" size="small" @click="handleDelete(row)">删除</PermissionButton>
          </template>
        </ElTableColumn>
      </ElTable>

      <BusinessOperateDrawer
        v-model:visible="drawerVisible"
        :operate-type="operateType"
        :row-data="editingData"
        :tree-options="businessTreeOptions"
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
