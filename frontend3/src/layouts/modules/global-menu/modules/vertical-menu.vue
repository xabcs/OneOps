<script setup lang="ts">
  import { nextTick, ref, watch } from 'vue';
  import { useRoute } from 'vue-router';
  import type { MenuInstance } from 'element-plus';
  import type { RouteKey } from '@elegant-router/types';
  import { SimpleScrollbar } from '@sa/materials';
  import { GLOBAL_SIDER_MENU_ID } from '@/constants/app';
  import { useAppStore } from '@/store/modules/app';
  import { useRouteStore } from '@/store/modules/route';
  import { useMenu } from '../../../context';
  import MenuItem from '../components/menu-item.vue';

  defineOptions({ name: 'VerticalMenu' });

  const route = useRoute();
  const appStore = useAppStore();
  const routeStore = useRouteStore();
  const { selectedKey, selectedKeyDummy, handleSelect } = useMenu();

  const menuRef = ref<MenuInstance>();

  const expandedKeys = ref<string[]>([]);

  // 当前已展开的子菜单 keys（default-openeds 为非受控属性，路由切换后不会自动同步，需跟踪用户操作）
  const openedKeys = ref<string[]>([]);

  function updateExpandedKeys() {
    if (appStore.siderCollapse || !selectedKey.value) {
      expandedKeys.value = [];
      return;
    }
    expandedKeys.value = routeStore.getSelectedMenuKeyPath(selectedKey.value);
  }

  function handleMenuOpen(index: string) {
    if (!openedKeys.value.includes(index)) {
      openedKeys.value.push(index);
    }
  }

  function handleMenuClose(index: string) {
    openedKeys.value = openedKeys.value.filter(key => key !== index);
  }

  /** 路由切换后同步菜单展开状态：收起不属于当前路径的菜单，展开当前路由的父级 */
  function syncOpenedMenus() {
    const menu = menuRef.value;
    if (!menu) return;
    openedKeys.value.filter(key => !expandedKeys.value.includes(key)).forEach(key => menu.close(key));
    expandedKeys.value.slice(0, -1).forEach(key => menu.open(key));
  }

  watch(
    () => route.name,
    () => {
      updateExpandedKeys();
      nextTick(syncOpenedMenus);
    },
    { immediate: true }
  );

  // 侧边栏折叠时 ElMenu 会关闭全部子菜单，同步清空跟踪状态
  watch(
    () => appStore.siderCollapse,
    collapsed => {
      if (collapsed) {
        openedKeys.value = [];
      } else {
        nextTick(syncOpenedMenus);
      }
    }
  );
</script>

<template>
  <Teleport :to="`#${GLOBAL_SIDER_MENU_ID}`">
    <SimpleScrollbar>
      <ElMenu
        ref="menuRef"
        mode="vertical"
        unique-opened
        :default-active="selectedKeyDummy"
        :default-openeds="expandedKeys"
        :collapse="appStore.siderCollapse"
        :collapse-transition="false"
        @select="val => handleSelect(val as RouteKey)"
        @open="handleMenuOpen"
        @close="handleMenuClose"
      >
        <MenuItem v-for="item in routeStore.menus" :key="item.key" :item="item" :index="item.key" />
      </ElMenu>
    </SimpleScrollbar>
  </Teleport>
</template>

<style scoped></style>
