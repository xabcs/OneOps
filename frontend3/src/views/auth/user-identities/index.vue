<script setup lang="tsx">
  import { onMounted, ref } from 'vue';
  import { Delete } from '@element-plus/icons-vue';
  import {
    deleteUserIdentityMapping,
    fetchApplications,
    fetchUserIdentityMappings
  } from '@/service/api/application-permission';

  defineOptions({ name: 'AuthUserIdentities' });

  const loading = ref(false);
  const tableData = ref<Api.ApplicationPermission.UserIdentityMapping[]>([]);
  const applications = ref<Api.ApplicationPermission.Application[]>([]);
  const searchParams = ref({
    username: '',
    appId: null as number | null,
    status: ''
  });

  const pagination = ref({
    page: 1,
    pageSize: 20,
    total: 0
  });

  async function getApplications() {
    const { data, error } = await fetchApplications({ page: 1, pageSize: 1000 });
    if (!error && data) {
      applications.value = data.list || [];
    }
  }

  async function getData() {
    loading.value = true;
    try {
      const { data, error } = await fetchUserIdentityMappings({
        page: pagination.value.page,
        pageSize: pagination.value.pageSize,
        ...searchParams.value
      });
      if (!error && data) {
        tableData.value = data.list || [];
        pagination.value.total = data.total || 0;
      }
    } finally {
      loading.value = false;
    }
  }

  function handleSearch() {
    pagination.value.page = 1;
    getData();
  }

  function handleReset() {
    searchParams.value = {
      username: '',
      appId: null,
      status: ''
    };
    handleSearch();
  }

  function handlePageChange(page: number) {
    pagination.value.page = page;
    getData();
  }

  function handleSizeChange(size: number) {
    pagination.value.pageSize = size;
    pagination.value.page = 1;
    getData();
  }

  async function handleDelete(id: number) {
    try {
      await ElMessageBox.confirm('确定要删除此身份映射吗？删除后用户将无法访问对应的应用。', '确认删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      });

      const { error } = await deleteUserIdentityMapping(id);
      if (!error) {
        ElMessage.success('删除身份映射成功');
        getData();
      }
    } catch (error) {
      // 用户取消
    }
  }

  function getStatusTag(status: string) {
    const statusMap: Record<string, { type: '' | 'success' | 'warning' | 'info' | 'danger'; label: string }> = {
      active: { type: 'success', label: '激活' },
      inactive: { type: 'info', label: '禁用' },
      deleted: { type: 'danger', label: '已删除' }
    };
    return statusMap[status] || { type: '', label: status };
  }

  function getMappingTypeTag(type: string) {
    const typeMap: Record<string, { type: '' | 'success' | 'warning' | 'info' | 'danger'; label: string }> = {
      auto: { type: 'primary', label: '自动创建' },
      manual: { type: 'warning', label: '手动创建' }
    };
    return typeMap[type] || { type: '', label: type };
  }

  onMounted(() => {
    getApplications();
    getData();
  });
</script>

<template>
  <div class="min-h-500px flex-col gap-4">
    <!-- 搜索栏 -->
    <ElCard shadow="never">
      <ElForm :model="searchParams" inline>
        <ElFormItem label="用户名">
          <ElInput v-model="searchParams.username" placeholder="请输入用户名" class="w-200px" clearable />
        </ElFormItem>
        <ElFormItem label="应用">
          <ElSelect v-model="searchParams.appId" placeholder="全部应用" class="w-200px" clearable>
            <ElOption v-for="app in applications" :key="app.id" :label="app.name" :value="app.id" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="状态">
          <ElSelect v-model="searchParams.status" placeholder="全部状态" class="w-150px" clearable>
            <ElOption label="激活" value="active" />
            <ElOption label="禁用" value="inactive" />
            <ElOption label="已删除" value="deleted" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem>
          <ElButton type="primary" @click="handleSearch">查询</ElButton>
          <ElButton @click="handleReset">重置</ElButton>
        </ElFormItem>
      </ElForm>
    </ElCard>

    <!-- 表格 -->
    <ElCard shadow="never" class="flex-1">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="text-lg font-medium">用户身份映射列表</span>
          <ElTag>共 {{ pagination.total }} 条记录</ElTag>
        </div>
      </template>

      <ElTable v-loading="loading" :data="tableData" :border="false">
        <ElTableColumn type="index" label="序号" width="60" align="center" />
        <ElTableColumn label="授权中心用户" align="center" min-width="120">
          <template #default="{ row }">
            <div>
              <div class="font-medium">{{ row.authUser?.username || '-' }}</div>
              <div class="text-xs text-gray-500">{{ row.authUser?.nickname || '' }}</div>
            </div>
          </template>
        </ElTableColumn>
        <ElTableColumn label="应用" align="center" min-width="120">
          <template #default="{ row }">
            {{ row.appIDField?.name || '-' }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="外部用户名" align="center" min-width="120">
          <template #default="{ row }">
            <ElTag>{{ row.externalUsername }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="外部用户ID" align="center" min-width="120">
          <template #default="{ row }">
            {{ row.externalUserId || '-' }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="映射类型" align="center" width="110">
          <template #default="{ row }">
            <ElTag :type="getMappingTypeTag(row.mappingType).type">
              {{ getMappingTypeTag(row.mappingType).label }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="映射状态" align="center" width="100">
          <template #default="{ row }">
            <ElTag :type="getStatusTag(row.mappingStatus).type">
              {{ getStatusTag(row.mappingStatus).label }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="最后同步时间" align="center" min-width="160">
          <template #default="{ row }">
            {{ row.lastSyncAt || '-' }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="创建时间" align="center" min-width="160">
          <template #default="{ row }">
            {{ row.createdAt }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="操作" align="center" width="100" fixed="right">
          <template #default="{ row }">
            <ElButton size="small" type="danger" :icon="Delete" @click="handleDelete(row.id)">删除</ElButton>
          </template>
        </ElTableColumn>
      </ElTable>

      <div class="mt-4 flex justify-end">
        <ElPagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="handlePageChange"
          @size-change="handleSizeChange"
        />
      </div>
    </ElCard>
  </div>
</template>

<style scoped></style>
