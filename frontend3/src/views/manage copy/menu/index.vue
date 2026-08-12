<script setup lang="tsx">
  import { computed, ref } from 'vue';
  import { ElMessageBox, ElNotification } from 'element-plus';
  import { jsonClone } from '@sa/utils';
  import { Menu as MenuIcon, Plus, Refresh, Search } from '@element-plus/icons-vue';
  import { fetchDeleteMenu, fetchGetMenuTree, fetchUpdateMenu } from '@/service/api';
  import { useAuthStore } from '@/store/modules/auth';
  import { useRouteStore } from '@/store/modules/route';
  import { useThemeStore } from '@/store/modules/theme';
  import { useTableOperate, useUIPaginatedTable } from '@/hooks/common/table';
  import { $t } from '@/locales';
  import MenuOperateDrawer from './modules/menu-operate-drawer.vue';
  import { cleanMenuTree, filterMenuTree, findSiblings, flattenMenuTree } from './modules/menu-tree-helper';
  import { createMenuColumns } from './modules/menu-columns';

  defineOptions({ name: 'MenuManage' });

  const authStore = useAuthStore();
  const routeStore = useRouteStore();
  const themeStore = useThemeStore();

  // Hero区域显示状态
  const heroVisible = computed(() => themeStore.contentTheme2.heroSection.visible !== false);

  const filterText = ref('');
  const flatMenuList = ref<Api.SystemManage.Menu[]>([]);
  const originalTreeData = ref<Api.SystemManage.Menu[]>([]);
  const tableKey = ref(0);

  // 移动菜单
  async function handleMove(row: Api.SystemManage.Menu, direction: 'up' | 'down') {
    const siblings = findSiblings(row.id, originalTreeData.value);
    if (!siblings || siblings.length < 2) {
      ElNotification({ title: '提示', message: '无法移动：同级菜单不足', type: 'warning', duration: 3000 });
      return;
    }

    const currentIndex = siblings.findIndex(item => item.id === row.id);
    const targetIndex = direction === 'up' ? currentIndex - 1 : currentIndex + 1;

    if (targetIndex < 0 || targetIndex >= siblings.length) {
      ElNotification({
        title: '提示',
        message: direction === 'up' ? '已经是第一个了' : '已经是最后一个了',
        type: 'warning',
        duration: 3000
      });
      return;
    }

    const targetRow = siblings[targetIndex];

    if (row.parentId !== targetRow.parentId) {
      ElNotification({ title: '错误', message: '只能在同级菜单之间排序', type: 'error', duration: 3000 });
      return;
    }

    try {
      await Promise.all([
        fetchUpdateMenu(row.id, { sort: targetRow.sort }),
        fetchUpdateMenu(targetRow.id, { sort: row.sort })
      ]);
      ElNotification({ title: '成功', message: '排序更新成功', type: 'success', duration: 3000 });
      await getData();
      tableKey.value++;
      await authStore.getUserInfo();
      routeStore.rebuildRoutes();
    } catch (error) {
      ElNotification({ title: '错误', message: '排序更新失败', type: 'error', duration: 3000 });
    }
  }

  // 状态变更
  async function handleStatusChange(row: Api.SystemManage.Menu, val: number) {
    try {
      await fetchUpdateMenu(row.id, { status: val });
      ElNotification({ title: '成功', message: `${val === 1 ? '启用' : '停用'}成功`, type: 'success', duration: 3000 });
      await getData();
      await authStore.getUserInfo();
      routeStore.rebuildRoutes();
    } catch (error) {
      ElNotification({ title: '错误', message: '操作失败', type: 'error', duration: 3000 });
      row.status = val === 1 ? 0 : 1;
    }
  }

  // 添加子菜单
  function handleAddChild(row: Api.SystemManage.Menu) {
    operateType.value = 'add';
    isAddingChild.value = true;
    parentMenuName.value = row.name;
    editingData.value = jsonClone(row);
    drawerVisible.value = true;
  }

  // 删除菜单
  async function handleDelete(id: number) {
    await ElMessageBox.confirm('确认删除吗？', '提示', {
      type: 'warning',
      confirmButtonText: '确定',
      cancelButtonText: '取消'
    });
    const { error } = await fetchDeleteMenu(id);
    if (!error) {
      ElNotification({ title: '成功', message: '删除成功', type: 'success', duration: 3000 });
      onDeleted();
      await authStore.getUserInfo();
      routeStore.rebuildRoutes();
    } else {
      ElNotification({ title: '错误', message: error?.response?.data?.message || error?.message || '删除失败', type: 'error', duration: 3000 });
    }
  }

  const { columns, data, loading, getData, getDataByPage } = useUIPaginatedTable({
    api: async () => {
      const { error, data } = await fetchGetMenuTree();
      if (!error && data && Array.isArray(data)) {
        const tree = data as Api.SystemManage.Menu[];
        const cleanedTree = cleanMenuTree(tree);
        originalTreeData.value = cleanedTree;
        flatMenuList.value = flattenMenuTree(cleanedTree);
        const displayData = filterText.value ? filterMenuTree(cleanedTree, filterText.value) : cleanedTree;
        return {
          data: displayData,
          list: cleanedTree,
          page: 1,
          pageSize: cleanedTree.length,
          total: cleanedTree.length
        };
      }
      return { data: [], list: [], page: 1, pageSize: 0, total: 0 };
    },
    transform: response => response,
    columns: createMenuColumns({
      handleMove,
      handleStatusChange,
      handleAddChild,
      handleEdit,
      handleDelete,
      originalTreeData: () => originalTreeData.value
    })
  });

  const {
    drawerVisible,
    operateType,
    editingData,
    handleAdd: _handleAdd,
    checkedRowKeys: _checkedRowKeys,
    onDeleted
  } = useTableOperate(data, 'id', getData);

  const isAddingChild = ref(false);
  const parentMenuName = ref('');

  function handleEdit(id: number) {
    isAddingChild.value = false;
    parentMenuName.value = '';
    operateType.value = 'edit';
    const foundItem = flatMenuList.value.find(item => item.id === id);
    editingData.value = foundItem ? jsonClone(foundItem) : null;
    drawerVisible.value = true;
  }

  function handleAdd() {
    isAddingChild.value = false;
    parentMenuName.value = '';
    _handleAdd();
  }

  async function handleMenuSubmitted() {
    await getDataByPage();
    await authStore.getUserInfo();
    routeStore.rebuildMenus();
  }

  async function handleFilterChange() {
    if (!filterText.value) {
      data.value = originalTreeData.value;
    } else {
      data.value = filterMenuTree(originalTreeData.value, filterText.value);
    }
  }

  async function refreshData() {
    await getDataByPage();
    await authStore.getUserInfo();
    routeStore.rebuildRoutes();
  }
</script>

<template>
  <div class="page-container menu-management-page">
    <!-- Hero 区域 -->
    <section
      v-if="heroVisible"
      class="hero-section"
      :style="{
        background: 'var(--sx-hero-bg)',
        border: '1px solid var(--sx-hero-border)',
        borderRadius: 'var(--sx-hero-radius)',
        boxShadow: 'var(--sx-hero-shadow)',
        padding: 'var(--sx-hero-padding)'
      }"
    >
      <div class="hero-content">
        <div class="hero-title-row">
          <span class="hero-icon">
            <ElIcon>
              <MenuIcon />
            </ElIcon>
          </span>
          <h2>{{ $t('page.manage.menu.title') }}</h2>
          <p class="hero-desc">管理系统菜单结构，支持树形层级展示与拖拽排序</p>
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
    <div
      class="content-card menu-content-card"
      :style="{
        background: 'var(--sx-content-card-bg)',
        border: '1px solid var(--sx-content-card-border)',
        borderRadius: 'var(--sx-content-card-radius)',
        boxShadow: 'var(--sx-content-card-shadow)',
        padding: 'var(--sx-content-card-padding)'
      }"
    >
      <!-- 工具栏 -->
      <div class="card-toolbar">
        <div class="toolbar-head">
          <span class="toolbar-title">菜单列表</span>
          <span class="toolbar-desc">管理菜单层级结构、路由配置与权限标识</span>
        </div>
        <div class="toolbar-actions">
          <ElButton type="primary" size="small" @click="handleAdd">
            <ElIcon>
              <Plus />
            </ElIcon>
            新增菜单
          </ElButton>
        </div>
      </div>

      <!-- 搜索工具栏 -->
      <div class="workbench-toolbar workbench-toolbar--history menus-toolbar">
        <div class="workbench-toolbar-left">
          <ElInput
            v-model="filterText"
            placeholder="搜索菜单名称 / 路径 / 权限"
            clearable
            style="width: 300px"
            @input="handleFilterChange"
            @clear="handleFilterChange"
          >
            <template #prefix>
              <ElIcon>
                <Search />
              </ElIcon>
            </template>
          </ElInput>
        </div>
      </div>

      <!-- 数据表格 -->
      <div class="table-section">
        <ElTable
          :key="tableKey"
          v-loading="loading"
          :data="data"
          border
          class="data-table"
          row-key="id"
          :tree-props="{ children: 'children', indent: 20 }"
        >
          <ElTableColumn v-for="col in columns" :key="col.prop" v-bind="col" />
        </ElTable>
      </div>
    </div>

    <!-- 抽屉 -->
    <MenuOperateDrawer
      v-model:visible="drawerVisible"
      :operate-type="operateType"
      :row-data="editingData"
      :is-adding-child="isAddingChild"
      :parent-menu-name="parentMenuName"
      @submitted="handleMenuSubmitted"
    />
  </div>
</template>

<style scoped lang="scss">
  @use './modules/menu-page.scss';
</style>
