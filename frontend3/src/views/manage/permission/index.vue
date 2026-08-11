<script setup lang="tsx">
    import { computed, onMounted, ref } from 'vue';
    import { ElNotification, ElMessage } from 'element-plus';
    import { Delete, Management, Plus, Refresh, Search } from '@element-plus/icons-vue';
    import { fetchGetPermissionList, fetchDeletePermission, fetchUpdatePermission } from '@/service/api';
    import { useThemeStore } from '@/store/modules/theme';
    import { defaultTransform, useTableOperate, useUIPaginatedTable } from '@/hooks/common/table';
    import { $t } from '@/locales';
    import { executeWithPermission } from '@/utils/permission';
    import PermissionSearch from './modules/permission-search.vue';
    import PermissionOperateDrawer from './modules/permission-operate-drawer.vue';

    defineOptions({ name: 'PermissionManage' });

    const themeStore = useThemeStore();

    // Hero区域显示状态
    const heroVisible = computed(() => themeStore.contentTheme2.heroSection.visible !== false);

    const searchParams = ref(getInitSearchParams());

    function getInitSearchParams(): Api.SystemManage.PermissionSearchParams {
      return {
        page: 1,
        pageSize: 10,
        status: undefined,
        name: undefined,
        code: undefined,
        module: undefined
      };
    }

    const { columns, columnChecks, data, getData, getDataByPage, loading, mobilePagination } = useUIPaginatedTable({
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
        { prop: 'selection', type: 'selection', width: 48 },
        { prop: 'index', type: 'index', label: $t('common.index'), width: 64 },
        { prop: 'name', label: $t('page.manage.permission.permissionName'), minWidth: 120 },
        { prop: 'code', label: $t('page.manage.permission.permissionCode'), minWidth: 150 },
        { prop: 'module', label: $t('page.manage.permission.module'), minWidth: 100 },
        { prop: 'resource', label: $t('page.manage.permission.resource'), minWidth: 100 },
        { prop: 'action', label: $t('page.manage.permission.action'), minWidth: 100 },
        {
          prop: 'level',
          label: $t('page.manage.permission.level'),
          align: 'center',
          width: 100,
          formatter: row => {
            const levelMap: Record<number, { text: string, type: UI.ThemeColor }> = {
              1: { text: '模块', type: 'primary' },
              2: { text: '页面', type: 'success' },
              3: { text: '按钮', type: 'warning' },
              4: { text: 'API', type: 'info' }
            };
            const level = levelMap[row.level] || { text: '未知', type: 'default' };
            return (
              <ElTag size="small" type={level.type}>
                {level.text}
              </ElTag>
            );
          }
        },
        {
          prop: 'status',
          label: $t('page.manage.permission.status'),
          align: 'center',
          width: 100,
          formatter: row => {
            if (row.status === undefined) {
              return '';
            }

            return (
              <ElSwitch
                v-model={row.status}
                activeValue={1}
                inactiveValue={0}
                onChange={(val: number) => handleStatusChange(row, val)}
              />
            );
          }
        },
        {
          prop: 'operate',
          label: $t('common.operate'),
          align: 'center',
          width: 180,
          formatter: row => (
            <div class="flex-center gap-8px">
              <ElButton type="primary" plain size="small" onClick={() => edit(row.id)}>
                {$t('common.edit')}
              </ElButton>
              <ElPopconfirm title={$t('common.confirmDelete')} onConfirm={() => handleDelete(row.id)}>
                {{
                  reference: () => (
                    <ElButton type="danger" plain size="small">
                      {$t('common.delete')}
                    </ElButton>
                  )
                }}
              </ElPopconfirm>
            </div>
          )
        }
      ]
    });

    const { drawerVisible, operateType, editingData, handleAdd, handleEdit, checkedRowKeys, onBatchDeleted, onDeleted } =
      useTableOperate(data, 'id', getData);

    // 权限操作完成后的数据刷新
    async function handlePermissionOperateSubmitted() {
      await getData();
    }

    async function handleBatchDelete() {
      await executeBatchDelete(async () => {
        if (checkedRowKeys.value.length === 0) {
          ElMessage.warning('请选择要删除的权限');
          return;
        }

        let successCount = 0;
        let failCount = 0;

        for (const id of checkedRowKeys.value) {
          const { error } = await fetchDeletePermission(id);
          if (!error) {
            successCount++;
          } else {
            failCount++;
          }
        }

        if (successCount > 0) {
          ElNotification({
            title: '批量删除完成',
            message: `成功删除 ${successCount} 个权限`,
            type: 'success',
            duration: 3000,
            position: 'top-right'
          });
        }

        if (failCount > 0) {
          ElNotification({
            title: '部分删除失败',
            message: `${failCount} 个权限删除失败`,
            type: 'error',
            duration: 5000,
            position: 'top-right'
          });
        }

        onBatchDeleted();
      });
    }

    async function handleDelete(id: number) {
      await executeWithPermission('system.permission.delete', async () => {
        const permission = data.value.find(p => p.id === id);
        if (!permission) {
          ElMessage.error('权限不存在');
          return;
        }

        const { error } = await fetchDeletePermission(id);

        if (!error) {
          ElNotification({
            title: '删除成功',
            message: `权限 "${permission.name}" 已成功删除`,
            type: 'success',
            duration: 3000,
            position: 'top-right'
          });
          onDeleted();
        } else {
          ElNotification({
            title: '删除失败',
            message: error.msg || '删除权限失败',
            type: 'error',
            duration: 3000,
            position: 'top-right'
          });
        }
      });
    }

    async function handleStatusChange(row: Api.SystemManage.Permission, val: number) {
      await executeWithPermission('system.permission.update', async () => {
        const { error } = await fetchUpdatePermission(row.id, { status: val });

        if (!error) {
          ElMessage.success(`${val === 1 ? '启用' : '禁用'}成功`);
        } else {
          row.status = val === 1 ? 0 : 1;
          ElMessage.error('状态更新失败');
        }
      });
    }

    function resetSearchParams() {
      searchParams.value = getInitSearchParams();
    }

    function edit(id: number) {
      handleEdit(id);
    }

    // 刷新数据
    async function refreshData() {
      await getDataByPage();
    }

    // 搜索输入处理（防抖）
    let searchTimeout: ReturnType<typeof setTimeout> | null = null;
    function handleSearchInput() {
      if (searchTimeout) {
        clearTimeout(searchTimeout);
      }
      searchTimeout = setTimeout(() => {
        searchParams.value.page = 1;
        getDataByPage();
      }, 300);
    }

    // 手动搜索（点击搜索按钮）
    function handleSearch() {
      searchParams.value.page = 1;
      getDataByPage();
    }
</script>

<template>
    <div class="page-container permission-management-page">
        <!-- Hero 区域 -->
        <section v-if="heroVisible" class="hero-section" :style="{
        background: 'var(--sx-hero-bg)',
        border: '1px solid var(--sx-hero-border)',
        borderRadius: 'var(--sx-hero-radius)',
        boxShadow: 'var(--sx-hero-shadow)',
        padding: 'var(--sx-hero-padding)'
      }">
            <div class="hero-content">
                <div class="hero-title-row">
                    <span class="hero-icon">
                        <ElIcon>
                            <Management />
                        </ElIcon>
                    </span>
                    <h2>{{ $t('page.manage.permission.title') }}</h2>
                    <p class="hero-desc">管理系统权限，支持模块、页面、按钮、API级别的权限控制</p>
                </div>
            </div>
            <div class="hero-actions">
                <ElButton size="small" :loading="loading" @click="refreshData">
                    <ElIcon>
                        <Refresh />
                    </ElIcon>
                    刷新
                </ElButton>
            </div>
        </section>

        <!-- 内容卡片 -->
        <div class="content-card permission-content-card" :style="{
        background: 'var(--sx-content-card-bg)',
        border: '1px solid var(--sx-content-card-border)',
        borderRadius: 'var(--sx-content-card-radius)',
        boxShadow: 'var(--sx-content-card-shadow)',
        padding: 'var(--sx-content-card-padding)'
      }">
            <!-- 工具栏 -->
            <div class="card-toolbar">
                <div class="toolbar-head">
                    <span class="toolbar-title">权限列表</span>
                    <span class="toolbar-desc">管理系统的所有权限，包括模块、页面、按钮和API权限</span>
                </div>
                <div class="toolbar-actions">
                    <ElButton type="primary" size="small" @click="handleAdd">
                        <ElIcon>
                            <Plus />
                        </ElIcon>
                        新增权限
                    </ElButton>
                    <ElButton type="danger" size="small" :disabled="checkedRowKeys.length === 0" @click="handleBatchDelete">
                        <ElIcon>
                            <Delete />
                        </ElIcon>
                        批量删除
                    </ElButton>
                </div>
            </div>

            <!-- 搜索工具栏 -->
            <div class="workbench-toolbar workbench-toolbar--history permissions-toolbar">
                <div class="workbench-toolbar-left">
                    <ElInput v-model="searchParams.name" placeholder="搜索权限名称" clearable style="width: 200px" @input="handleSearchInput" />
                    <ElInput v-model="searchParams.code" placeholder="搜索权限编码" clearable style="width: 200px" @input="handleSearchInput" />
                    <ElSelect v-model="searchParams.module" placeholder="选择模块" clearable style="width: 150px" @change="handleSearch">
                        <ElOption label="系统管理" value="system" />
                        <ElOption label="授权中心" value="auth" />
                        <ElOption label="资产管理" value="cmdb" />
                        <ElOption label="监控中心" value="monitoring" />
                        <ElOption label="K8s管理" value="k8s" />
                    </ElSelect>
                </div>
                <div class="workbench-toolbar-right">
                    <ElButton class="filter-refresh-btn" @click="resetSearchParams">
                        <ElIcon>
                            <Refresh />
                        </ElIcon>
                        重置
                    </ElButton>
                    <ElButton class="filter-refresh-btn" type="primary" @click="handleSearch">
                        <ElIcon>
                            <Search />
                        </ElIcon>
                        搜索
                    </ElButton>
                </div>
            </div>

            <!-- 数据表格 -->
            <div class="table-section">
                <ElTable v-loading="loading" :data="data" border stripe class="data-table" row-key="id" @selection-change="checkedRowKeys = $event">
                    <ElTableColumn v-for="col in columns" :key="col.prop" v-bind="col" />
                </ElTable>
            </div>

            <!-- 分页 -->
            <div class="table-pagination">
                <ElPagination v-if="mobilePagination.total" layout="total, sizes, prev, pager, next" v-bind="mobilePagination" @current-change="mobilePagination['current-change']" @size-change="mobilePagination['size-change']" />
            </div>
        </div>

        <!-- 抽屉和模态框 -->
        <PermissionOperateDrawer v-model:visible="drawerVisible" :operate-type="operateType" :row-data="editingData" @submitted="handlePermissionOperateSubmitted" />
    </div>
</template>

<style scoped lang="scss">
    .permission-management-page {
      padding: 16px 20px;
      background: var(--el-bg-color-page);
      min-height: 100vh;
      display: flex;
      flex-direction: column;
      gap: 6px;
    }

    .hero-section {
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 8px;
    }

    .hero-content {
      display: flex;
      flex-direction: column;
      gap: 4px;
    }

    .hero-title-row {
      display: flex;
      align-items: center;
      gap: 4px;
      flex-wrap: wrap;
    }

    .hero-icon {
      width: 42px;
      height: 42px;
      border-radius: 14px;
      display: inline-flex;
      align-items: center;
      justify-content: center;
      font-size: 20px;
      color: var(--sx-hero-icon-color);
      background: var(--sx-hero-icon-bg);
      border: 1px solid var(--sx-hero-icon-border);
      box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.8);
    }

    .hero-title-row h2 {
      color: var(--sx-text-primary);
      font-size: 23px;
      font-weight: 700;
      margin: 0;
    }

    .hero-desc {
      color: var(--sx-text-secondary);
      font-size: 13px;
      line-height: 1.45;
      margin: 0;
      max-width: 600px;
    }

    .hero-actions {
      display: flex;
      gap: 4px;
    }

    .content-card {
      display: flex;
      flex-direction: column;
      gap: 8px;
    }

    .card-toolbar {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      gap: 12px;
      padding-bottom: 12px;
      border-bottom: 1px solid var(--sx-border-soft);
    }

    .toolbar-head {
      display: flex;
      flex-direction: column;
      gap: 4px;
    }

    .toolbar-title {
      color: var(--sx-text-primary);
      font-size: 16px;
      font-weight: 700;
    }

    .toolbar-desc {
      color: var(--sx-text-secondary);
      font-size: 12px;
    }

    .toolbar-actions {
      display: flex;
      gap: 4px;
      align-items: center;
    }

    .workbench-toolbar {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 12px;
      flex-wrap: wrap;
      margin: 6px 0 8px;
    }

    .workbench-toolbar-left,
    .workbench-toolbar-right {
      display: flex;
      align-items: center;
      gap: 10px;
      flex-wrap: wrap;
    }

    .workbench-toolbar.workbench-toolbar--history {
      background: var(--sx-toolbar-bg);
      border: 1px solid var(--sx-toolbar-border);
      border-radius: var(--sx-toolbar-radius);
      padding: var(--sx-toolbar-padding);
      box-shadow: var(--sx-toolbar-shadow);
      gap: 8px;
      margin: 6px 0;

      .workbench-toolbar-left,
      .workbench-toolbar-right {
        gap: 5px;
      }

      .el-input__wrapper {
        background: var(--sx-search-input-bg);
        border-radius: var(--sx-search-input-radius);
        box-shadow: 0 0 0 1px var(--sx-search-input-border) inset;
        transition: all 0.3s ease;

        &:hover {
          box-shadow: 0 0 0 1px var(--sx-search-input-hover-border) inset;
        }
      }

      .el-input.is-focus .el-input__wrapper {
        box-shadow: 0 0 0 1px var(--sx-search-input-focus-border) inset;
      }

      .el-button:not(.is-link) {
        background: var(--sx-search-button-bg);
        color: var(--sx-search-button-text);
        transition: all 0.3s ease;

        &:hover {
          background: var(--sx-search-button-hover);
        }
      }
    }

    .table-section {
      flex: 1;
      display: flex;
      flex-direction: column;
    }

    .data-table {
      width: 100%;
      border-radius: var(--sx-table-radius);
      overflow: hidden;
      border: 1px solid var(--sx-table-border);
      border-collapse: separate;
      border-spacing: 0;

      --el-table-border-color: var(--sx-table-border);
      --el-table-header-bg-color: var(--sx-table-header-bg);
      --el-table-row-hover-bg-color: var(--sx-table-row-hover);

      :deep(th.el-table__cell) {
        background-color: var(--sx-table-header-bg);
        color: var(--sx-table-header-text);
        font-weight: 600;
        border-radius: 0;
      }

      :deep(td.el-table__cell) {
        border-radius: 0;
      }

      :deep(.el-checkbox__inner),
      :deep(.el-checkbox__inner::before),
      :deep(.el-checkbox__inner::after) {
        border-radius: 0;
      }

      &--striped :deep(.el-table__body tr.el-table__row--striped td) {
        background-color: var(--sx-table-striped-bg);
      }
    }

    .table-pagination {
      display: flex;
      justify-content: flex-end;
      padding-top: 12px;

      .el-pagination {
        .btn-next,
        .btn-prev,
        .el-pager li {
          background: var(--sx-pagination-button-bg);
          color: var(--sx-pagination-button-text);
          border-radius: var(--sx-pagination-radius);
          transition: all 0.3s ease;

          &:hover {
            background: var(--sx-pagination-button-hover);
          }

          &.is-active {
            background: var(--sx-pagination-active-bg);
            color: var(--sx-pagination-active-text);
          }
        }
      }
    }

    @media (max-width: 768px) {
      .hero-section {
        flex-direction: column;
        align-items: flex-start;
      }

      .card-toolbar {
        flex-direction: column;
        align-items: stretch;
        gap: 8px;
      }
    }
</style>
