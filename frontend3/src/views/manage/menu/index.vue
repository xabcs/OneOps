<script setup lang="tsx">
  import { computed, ref } from 'vue';
  import { ElMessage, ElMessageBox } from 'element-plus';
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
      ElMessage.warning('无法移动：同级菜单不足');
      return;
    }

    const currentIndex = siblings.findIndex(item => item.id === row.id);
    const targetIndex = direction === 'up' ? currentIndex - 1 : currentIndex + 1;

    if (targetIndex < 0 || targetIndex >= siblings.length) {
      ElMessage.warning(direction === 'up' ? '已经是第一个了' : '已经是最后一个了');
      return;
    }

    const targetRow = siblings[targetIndex];

    if (row.parentId !== targetRow.parentId) {
      ElMessage.error('只能在同级菜单之间排序');
      return;
    }

    await Promise.all([
      fetchUpdateMenu(row.id, { sort: targetRow.sort }),
      fetchUpdateMenu(targetRow.id, { sort: row.sort })
    ]);
    ElMessage.success('排序更新成功');
    await getData();
    tableKey.value++;
    await authStore.getUserInfo();
    routeStore.rebuildRoutes();
  }

  // 状态变更
  async function handleStatusChange(row: Api.SystemManage.Menu, val: number) {
    try {
      await fetchUpdateMenu(row.id, { status: val });
      ElMessage.success(`${val === 1 ? '启用' : '停用'}成功`);
      await getData();
      await authStore.getUserInfo();
      routeStore.rebuildRoutes();
    } catch (error) {
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
      ElMessage.success('删除成功');
      onDeleted();
      await authStore.getUserInfo();
      routeStore.rebuildRoutes();
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
  <div class="flex flex-col gap-16px">
    <!-- Hero 区域 -->
    <ElCard v-if="heroVisible" shadow="hover">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-12px">
          <ElIcon :size="24">
            <MenuIcon />
          </ElIcon>
          <div class="flex flex-col gap-2px">
            <h2 class="text-18px font-bold m-0">{{ $t('page.manage.menu.title') }}</h2>
            <p class="text-13px opacity-70 m-0">管理系统菜单结构，支持树形层级展示与拖拽排序</p>
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

    <!-- 内容卡片 -->
    <ElCard shadow="hover">
      <template #header>
        <div class="flex items-center justify-between">
          <div class="flex flex-col gap-2px">
            <span class="text-16px font-bold">菜单列表</span>
            <span class="text-13px opacity-70">管理菜单层级结构、路由配置与权限标识</span>
          </div>
          <PermissionButton code="system.menu.create" type="primary" size="small" @click="handleAdd">
            <ElIcon>
              <Plus />
            </ElIcon>
            新增菜单
          </PermissionButton>
        </div>
      </template>

      <!-- 搜索工具栏 -->
      <ElSpace wrap class="mb-16px">
        <ElInput
          v-model="filterText"
          placeholder="搜索菜单名称 / 路径 / 权限"
          clearable
          style="width: 300px"
          :prefix-icon="Search"
          @input="handleFilterChange"
          @clear="handleFilterChange"
        />
        <ElButton @click="filterText = ''; handleFilterChange()">
          <ElIcon>
            <Refresh />
          </ElIcon>
          重置
        </ElButton>
        <ElButton type="primary" @click="handleFilterChange">
          <ElIcon>
            <Search />
          </ElIcon>
          搜索
        </ElButton>
      </ElSpace>

      <!-- 数据表格 -->
      <ElTable
        :key="tableKey"
        v-loading="loading"
        :data="data"
        border
        row-key="id"
        :tree-props="{ children: 'children', indent: 20 }"
      >
        <ElTableColumn v-for="col in columns" :key="col.prop" v-bind="col" />
      </ElTable>
    </ElCard>

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
