<script setup lang="tsx">
  import { onMounted, ref } from 'vue';
  import { Refresh, Search } from '@element-plus/icons-vue';
  import type { FlatResponseData } from '@sa/axios';
  import {
    deleteUserIdentityMapping,
    fetchApplicationOptions,
    fetchUserIdentityMappings
  } from '@/service/api/application-permission';
  import { defaultTransform, useUIPaginatedTable } from '@/hooks/common/table';
  import { createTagMap } from '@/utils/common';
  import PermissionButton from '@/components/common/PermissionButton.vue';

  defineOptions({ name: 'AuthUserIdentities' });

  interface SearchParams {
    page: number;
    pageSize: number;
    username: string;
    appId: number | null;
    status: string;
  }

  function getInitSearchParams(): SearchParams {
    return { page: 1, pageSize: 20, username: '', appId: null, status: '' };
  }

  const searchParams = ref<SearchParams>(getInitSearchParams());

  const applications = ref<{ id: number; name: string; code: string; type: string }[]>([]);

  async function getApplications() {
    const { data, error } = await fetchApplicationOptions();
    if (!error && data) {
      applications.value = data || [];
    }
  }

  const { columns, data, loading, mobilePagination, getData, getDataByPage } = useUIPaginatedTable({
    paginationProps: {
      currentPage: searchParams.value.page,
      pageSize: searchParams.value.pageSize,
      pageSizes: [10, 20, 50, 100]
    },
    api: () => fetchUserIdentityMappings(searchParams.value),
    transform: response =>
      defaultTransform<Api.ApplicationPermission.UserIdentityMapping>(
        response as unknown as FlatResponseData<
          unknown,
          Api.Common.PaginatingQueryRecord<Api.ApplicationPermission.UserIdentityMapping>
        >
      ),
    onPaginationParamsChange: params => {
      searchParams.value.page = params.currentPage ?? 1;
      searchParams.value.pageSize = params.pageSize ?? 20;
    },
    columns: () => [
      { prop: 'index', type: 'index', label: '序号', width: 60, align: 'center' },
      {
        prop: 'username',
        label: '授权中心用户',
        align: 'center',
        minWidth: 120,
        formatter: row => (
          <div>
            <div class="font-medium">{row.authUser?.username || '-'}</div>
            <div class="text-xs text-gray-500">{row.authUser?.nickname || ''}</div>
          </div>
        )
      },
      {
        prop: 'app',
        label: '应用',
        align: 'center',
        minWidth: 120,
        formatter: row => <span>{row.appIDField?.name || '-'}</span>
      },
      {
        prop: 'externalUsername',
        label: '外部用户名',
        align: 'center',
        minWidth: 120,
        formatter: row => <ElTag>{row.externalUsername}</ElTag>
      },
      {
        prop: 'externalUserId',
        label: '外部用户ID',
        align: 'center',
        minWidth: 120,
        formatter: row => <span>{row.externalUserId || '-'}</span>
      },
      {
        prop: 'mappingType',
        label: '映射类型',
        align: 'center',
        width: 110,
        formatter: row => {
          const t = getMappingTypeTag(row.mappingType);
          return <ElTag type={t.type}>{t.text}</ElTag>;
        }
      },
      {
        prop: 'mappingStatus',
        label: '映射状态',
        align: 'center',
        width: 100,
        formatter: row => {
          const t = getStatusTag(row.mappingStatus);
          return <ElTag type={t.type}>{t.text}</ElTag>;
        }
      },
      {
        prop: 'lastSyncTime',
        label: '最后同步时间',
        align: 'center',
        minWidth: 160,
        formatter: row => <span>{row.lastSyncTime || '-'}</span>
      },
      { prop: 'createdAt', label: '创建时间', align: 'center', minWidth: 160 },
      {
        prop: 'operate',
        label: '操作',
        align: 'center',
        className: 'msre-table-actions',
        width: 100,
        fixed: 'right',
        formatter: row => (
          <PermissionButton
            code="auth.user.delete"
            size="small"
            type="danger"
            link
            onClick={() => handleDelete(row.id)}
          >
            删除
          </PermissionButton>
        )
      }
    ]
  });

  /** 映射状态 → ElTag 标签映射 */
  const getStatusTag = createTagMap({
    active: { text: '激活', type: 'success' },
    inactive: { text: '禁用', type: 'info' },
    deleted: { text: '已删除', type: 'danger' }
  });

  /** 映射来源 → ElTag 标签映射 */
  const getMappingTypeTag = createTagMap({
    auto: { text: '自动创建', type: 'primary' },
    manual: { text: '手动创建', type: 'warning' }
  });

  function handleSearch() {
    getDataByPage(1);
  }

  function handleReset() {
    searchParams.value = getInitSearchParams();
    getDataByPage(1);
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
    } catch {
      // 用户取消
    }
  }

  onMounted(() => {
    getApplications();
  });
</script>

<template>
  <div class="table-page">
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
          <ElButton type="primary" :icon="Search" @click="handleSearch">查询</ElButton>
          <ElButton :icon="Refresh" @click="handleReset">重置</ElButton>
        </ElFormItem>
      </ElForm>
    </ElCard>

    <!-- 表格 -->
    <ElCard shadow="never" class="flex-1">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="text-lg font-medium">用户身份映射列表</span>
          <ElTag>共 {{ mobilePagination.total || 0 }} 条记录</ElTag>
        </div>
      </template>

      <div class="table-scroll-wrap">
        <ElTable v-loading="loading" :data="data" :border="false" height="100%">
          <ElTableColumn v-for="col in columns" :key="col.prop" v-bind="col" />
        </ElTable>
      </div>

      <div class="mt-4 flex justify-end">
        <ElPagination
          v-if="mobilePagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          v-bind="mobilePagination"
          @current-change="mobilePagination['current-change']"
          @size-change="mobilePagination['size-change']"
        />
      </div>
    </ElCard>
  </div>
</template>
