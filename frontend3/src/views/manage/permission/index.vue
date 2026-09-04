<script setup lang="tsx">
  import { onUnmounted, ref } from 'vue';
  import { ElMessage, ElMessageBox } from 'element-plus';
  import { Key, Link, Plus, Refresh, Search } from '@element-plus/icons-vue';
  import { fetchDeletePermission, fetchGetPermissionList } from '@/service/api';
  import { defaultTransform, useTableOperate, useUIPaginatedTable } from '@/hooks/common/table';
  import PermissionOperateDrawer from './modules/permission-operate-drawer.vue';
  import PermissionRouteModal from './modules/permission-route-modal.vue';

  defineOptions({ name: 'PermissionManage' });

  function getInitSearchParams(): Api.SystemManage.PermissionSearchParams {
    return {
      page: 1,
      pageSize: 10,
      name: undefined,
      code: undefined,
      module: undefined,
      level: undefined,
      status: undefined
    };
  }

  const searchParams = ref(getInitSearchParams());

  /** 模块标签颜色 */
  const moduleColors: Record<string, 'primary' | 'success' | 'warning' | 'info' | 'danger'> = {
    system: 'primary',
    cmdb: 'success',
    k8s: 'warning',
    monitor: 'info',
    audit: 'danger',
    auth: 'warning'
  };

  /** 级别标签 */
  const levelLabels: Record<number, string> = {
    1: '模块',
    2: '资源',
    3: '操作'
  };

  const { columns, data, getData, getDataByPage, loading, mobilePagination } = useUIPaginatedTable({
    paginationProps: {
      currentPage: searchParams.value.page,
      pageSize: searchParams.value.pageSize
    },
    api: () => fetchGetPermissionList(searchParams.value),
    transform: response => defaultTransform(response),
    onPaginationParamsChange: params => {
      searchParams.value.page = params.currentPage;
      searchParams.value.pageSize = params.pageSize;
    },
    columns: () => [
      { type: 'selection', width: 50 },
      { prop: 'id', label: 'ID', width: 60 },
      { prop: 'name', label: '权限名称', minWidth: 120 },
      {
        prop: 'code',
        label: '权限编码',
        minWidth: 160,
        formatter: (row: Api.SystemManage.Permission) => (
          <span style="font-family: monospace; font-size: 12px; color: var(--el-color-primary)">{row.code}</span>
        )
      },
      {
        prop: 'module',
        label: '模块',
        width: 90,
        formatter: (row: Api.SystemManage.Permission) => (
          <ElTag size="small" type={moduleColors[row.module] || 'info'}>
            {row.module}
          </ElTag>
        )
      },
      { prop: 'resource', label: '资源', width: 100 },
      { prop: 'action', label: '操作', width: 90 },
      {
        prop: 'level',
        label: '级别',
        width: 70,
        formatter: (row: Api.SystemManage.Permission) => (
          <ElTag size="small" type="info">
            {levelLabels[row.level] || row.level}
          </ElTag>
        )
      },
      {
        prop: 'routeCount',
        label: '路由映射',
        width: 90,
        formatter: (row: Api.SystemManage.Permission) => (
          <ElButton text type="primary" size="small" onClick={() => openRouteModal(row.code)}>
            查看
          </ElButton>
        )
      },
      {
        prop: 'status',
        label: '状态',
        width: 70,
        formatter: (row: Api.SystemManage.Permission) => (
          <ElTag size="small" type={row.status === 1 ? 'success' : 'danger'}>
            {row.status === 1 ? '启用' : '禁用'}
          </ElTag>
        )
      },
      {
        prop: 'operate',
        label: '操作',
        width: 120,
        fixed: 'right',
        formatter: (row: Api.SystemManage.Permission) => (
          <ElSpace size="small">
            <ElButton text type="primary" size="small" onClick={() => handleEdit(row.id)}>
              编辑
            </ElButton>
            <ElButton text type="danger" size="small" onClick={() => handleDelete(row.id)}>
              删除
            </ElButton>
          </ElSpace>
        )
      }
    ]
  });

  const { drawerVisible, operateType, editingData, handleAdd, handleEdit, onDeleted } = useTableOperate(
    data,
    'id',
    getData
  );

  /** 路由映射管理弹窗 */
  const routeModalVisible = ref(false);
  const routeModalCode = ref('');

  function openRouteModal(code?: string) {
    routeModalCode.value = code || '';
    routeModalVisible.value = true;
  }

  async function handleDelete(id: number) {
    await ElMessageBox.confirm('确认删除此权限吗？', '提示', {
      type: 'warning',
      confirmButtonText: '确定',
      cancelButtonText: '取消'
    });
    const { error } = await fetchDeletePermission(id);
    if (!error) {
      ElMessage.success('删除成功');
      onDeleted();
    }
  }

  function resetSearchParams() {
    searchParams.value = getInitSearchParams();
  }

  async function refreshData() {
    await getDataByPage();
  }

  // 搜索防抖
  let searchTimeout: ReturnType<typeof setTimeout> | null = null;
  function handleSearchInput() {
    if (searchTimeout) clearTimeout(searchTimeout);
    searchTimeout = setTimeout(() => {
      searchParams.value.page = 1;
      getDataByPage();
    }, 300);
  }

  function handleSearch() {
    searchParams.value.page = 1;
    getDataByPage();
  }

  onUnmounted(() => {
    if (searchTimeout) {
      clearTimeout(searchTimeout);
      searchTimeout = null;
    }
  });
</script>

<template>
  <div class="flex flex-col gap-16px">
    <!-- Hero -->
    <ElCard shadow="hover">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-12px">
          <ElIcon :size="24">
            <Key />
          </ElIcon>
          <div class="flex flex-col gap-2px">
            <h2 class="m-0 text-18px font-bold">权限管理</h2>
            <p class="m-0 text-13px opacity-70">管理系统权限码与路由映射，控制 API 访问权限</p>
          </div>
        </div>
        <ElButton size="small" :loading="loading" @click="refreshData">
          <ElIcon>
            <Refresh />
          </ElIcon>
          刷新
        </ElButton>
      </div>
    </ElCard>

    <!-- 列表卡片 -->
    <ElCard shadow="hover">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="text-16px font-bold">权限列表</span>
          <ElSpace size="small">
            <ElButton size="small" @click="openRouteModal()">
              <ElIcon><Link /></ElIcon>
              路由映射
            </ElButton>
            <PermissionButton code="system.permission.create" type="primary" size="small" @click="handleAdd">
              <ElIcon>
                <Plus />
              </ElIcon>
              新增权限
            </PermissionButton>
          </ElSpace>
        </div>
      </template>

      <!-- 搜索栏 -->
      <ElSpace wrap class="mb-16px">
        <ElInput
          v-model="searchParams.name"
          placeholder="搜索权限名称"
          clearable
          style="width: 180px"
          :prefix-icon="Search"
          @input="handleSearchInput"
        />
        <ElInput
          v-model="searchParams.code"
          placeholder="搜索权限编码"
          clearable
          style="width: 200px"
          :prefix-icon="Search"
          @input="handleSearchInput"
        />
        <ElInput
          v-model="searchParams.module"
          placeholder="搜索模块"
          clearable
          style="width: 140px"
          :prefix-icon="Search"
          @input="handleSearchInput"
        />
        <ElButton @click="resetSearchParams">
          <ElIcon><Refresh /></ElIcon>
          重置
        </ElButton>
        <ElButton type="primary" @click="handleSearch">
          <ElIcon><Search /></ElIcon>
          搜索
        </ElButton>
      </ElSpace>

      <!-- 表格 -->
      <ElTable v-loading="loading" :data="data" border stripe row-key="id">
        <ElTableColumn v-for="col in columns" :key="col.prop" v-bind="col" />
      </ElTable>

      <!-- 分页 -->
      <div v-if="mobilePagination.total" class="mt-16px flex justify-end">
        <ElPagination
          layout="total, sizes, prev, pager, next, jumper"
          v-bind="mobilePagination"
          @current-change="mobilePagination['current-change']"
          @size-change="mobilePagination['size-change']"
        />
      </div>
    </ElCard>

    <!-- 新增/编辑抽屉 -->
    <PermissionOperateDrawer
      v-model:visible="drawerVisible"
      :operate-type="operateType"
      :row-data="editingData"
      @submitted="getData"
    />

    <!-- 路由映射管理弹窗 -->
    <PermissionRouteModal v-model:visible="routeModalVisible" :initial-permission-code="routeModalCode" />
  </div>
</template>
