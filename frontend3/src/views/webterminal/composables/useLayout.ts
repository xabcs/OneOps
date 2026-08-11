import { onUnmounted, ref } from 'vue';
import { onKeyStroke } from '@vueuse/core';

/**
 * 布局管理 composable
 * 管理侧边栏、全屏、活动栏等布局状态
 */
export function useLayout() {
  // 布局状态
  const showActivityBar = ref(true);
  const showSidebar = ref(true);
  const isFullscreen = ref(false);
  const sidebarWidth = ref(300);
  const isResizing = ref(false);
  const activeActivityItem = ref('assets');

  // 服务器菜单折叠状态
  const serverMenuExpanded = ref(false);
  const serverMenuPosition = ref({ top: 0 });

  // 主菜单折叠状态
  const mainMenuExpanded = ref(false);
  const mainMenuPosition = ref({ top: 0 });

  // 监听 ESC 键退出全屏
  onKeyStroke('Escape', () => {
    if (isFullscreen.value) {
      toggleFullscreen();
    }
  });

  // 开始调整侧边栏大小
  function startResize(e: MouseEvent) {
    isResizing.value = true;
    const startX = e.clientX;
    const startWidth = sidebarWidth.value;

    const onMouseMove = (e: MouseEvent) => {
      if (!isResizing.value) return;
      const diff = e.clientX - startX;
      const newWidth = Math.max(200, Math.min(600, startWidth + diff));
      sidebarWidth.value = newWidth;
    };

    const onMouseUp = () => {
      isResizing.value = false;
      window.removeEventListener('mousemove', onMouseMove);
      window.removeEventListener('mouseup', onMouseUp);
      document.body.style.cursor = '';
      document.body.style.userSelect = '';
    };

    window.addEventListener('mousemove', onMouseMove);
    window.addEventListener('mouseup', onMouseUp);
    document.body.style.cursor = 'col-resize';
    document.body.style.userSelect = 'none';
  }

  // 切换侧边栏显示
  function toggleSidebar() {
    showSidebar.value = !showSidebar.value;
  }

  // 切换全屏
  function toggleFullscreen() {
    isFullscreen.value = !isFullscreen.value;
  }

  // 切换服务器菜单展开/折叠
  function toggleServerMenu(event: MouseEvent) {
    const target = event.currentTarget as HTMLElement;
    const rect = target.getBoundingClientRect();
    serverMenuPosition.value = { top: rect.top };
    serverMenuExpanded.value = !serverMenuExpanded.value;
    mainMenuExpanded.value = false;
  }

  // 切换主菜单展开/折叠
  function toggleMainMenu(event: MouseEvent) {
    const target = event.currentTarget as HTMLElement;
    const rect = target.getBoundingClientRect();
    mainMenuPosition.value = { top: rect.top };
    mainMenuExpanded.value = !mainMenuExpanded.value;
    serverMenuExpanded.value = false;
  }

  // 点击外部关闭下拉菜单
  function handleClickOutside(event: MouseEvent) {
    const target = event.target as HTMLElement;
    const activityBar = document.querySelector('.wb-activity-bar');
    const serverDropdown = document.querySelector('.wb-server-menu-dropdown');
    const mainDropdown = document.querySelector('.wb-main-menu-dropdown');

    if (activityBar && !activityBar.contains(target)) {
      if (serverDropdown && !serverDropdown.contains(target)) {
        serverMenuExpanded.value = false;
      }
      if (mainDropdown && !mainDropdown.contains(target)) {
        mainMenuExpanded.value = false;
      }
    }
  }

  // 清理
  onUnmounted(() => {
    if (isResizing.value) {
      isResizing.value = false;
      document.body.style.cursor = '';
      document.body.style.userSelect = '';
    }
    document.removeEventListener('click', handleClickOutside);
  });

  return {
    showActivityBar,
    showSidebar,
    isFullscreen,
    sidebarWidth,
    isResizing,
    activeActivityItem,
    serverMenuExpanded,
    serverMenuPosition,
    mainMenuExpanded,
    mainMenuPosition,
    startResize,
    toggleSidebar,
    toggleFullscreen,
    toggleServerMenu,
    toggleMainMenu,
    handleClickOutside
  };
}
