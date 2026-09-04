<script setup lang="ts">
  import { computed, onMounted, onUnmounted, ref } from 'vue';
  import { useRouter } from 'vue-router';
  import { useFullscreen } from '@vueuse/core';
  import { ElMessage } from 'element-plus';
  import { GLOBAL_HEADER_MENU_ID } from '@/constants/app';
  import {
    fetchMyNotifySetting,
    fetchTicketMessages,
    fetchUnreadMessageCount,
    markTicketMessagesRead,
    updateMyNotifySetting
  } from '@/service/api';
  import { useAppStore } from '@/store/modules/app';
  import { useThemeStore } from '@/store/modules/theme';
  import GlobalLogo from '../global-logo/index.vue';
  import GlobalBreadcrumb from '../global-breadcrumb/index.vue';
  import GlobalSearch from '../global-search/index.vue';
  import ThemeButton from './components/theme-button.vue';
  import UserAvatar from './components/user-avatar.vue';

  defineOptions({ name: 'GlobalHeader' });

  interface Props {
    /** Whether to show the logo */
    showLogo?: App.Global.HeaderProps['showLogo'];
    /** Whether to show the menu toggler */
    showMenuToggler?: App.Global.HeaderProps['showMenuToggler'];
    /** Whether to show the menu */
    showMenu?: App.Global.HeaderProps['showMenu'];
  }

  defineProps<Props>();

  const appStore = useAppStore();
  const themeStore = useThemeStore();
  const router = useRouter();
  const { isFullscreen, toggle } = useFullscreen();

  // 计算 Header 的自定义背景样式
  const headerStyle = computed(() => {
    // 如果使用渐变
    if (themeStore.header.useHeaderGradient) {
      return {
        background: `linear-gradient(180deg, ${themeStore.header.headerGradientStart || 'rgba(232, 237, 255, 0.98)'} 0%, ${themeStore.header.headerGradientEnd || 'rgba(232, 240, 251, 0.94)'} 100%)`
      };
    }

    // 如果使用自定义单一颜色
    if (themeStore.header.useCustomColor && themeStore.header.customColor) {
      return {
        backgroundColor: themeStore.header.customColor
      };
    }

    // 默认情况，使用主题默认背景
    return {};
  });

  // 系统状态
  const currentTime = ref(new Date());
  const systemStatus = ref({
    api: true, // API连接状态
    db: true, // 数据库连接状态
    agents: 15 // 在线Agent数量
  });

  const systemInfo = computed(() => ({
    version: 'v1.0.0',
    environment: import.meta.env.MODE === 'production' ? '生产' : '开发'
  }));

  // 更新时间
  setInterval(() => {
    currentTime.value = new Date();
  }, 1000);

  // 格式化时间
  const formatTime = (date: Date) => {
    return date.toLocaleTimeString('zh-CN', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    });
  };

  // 快速链接
  const quickLinks = [
    {
      name: 'GitHub',
      icon: 'mdi:github-face',
      tooltip: '项目源码',
      action: () => window.open('https://github.com/your-org/oneops', '_blank')
    },
    {
      name: '文档',
      icon: 'mdi:file-document-outline',
      tooltip: '项目文档',
      action: () => window.open('https://github.com/your-org/oneops-docs', '_blank')
    }
  ];

  // 快速操作菜单
  const quickActions = [
    {
      name: '刷新数据',
      icon: 'mdi:refresh',
      action: () => location.reload()
    },
    {
      name: '清除缓存',
      icon: 'mdi:cached',
      action: () => {
        localStorage.clear();
        sessionStorage.clear();
        location.reload();
      }
    },
    {
      name: '系统设置',
      icon: 'mdi:cog-outline',
      path: '/system'
    }
  ];

  // 下拉菜单显示状态
  const showQuickMenu = ref(false);

  // 通知显示状态
  const showNotifications = ref(false);

  // ─────────────────────────────────────────────
  // 通知中心：工单站内消息（本人收件箱，后端兜底渠道）
  // ─────────────────────────────────────────────
  const unreadCount = ref(0);
  const notifications = ref<Api.Ticket.TicketMessage[]>([]);
  let unreadTimer: ReturnType<typeof setInterval> | null = null;

  async function loadUnread() {
    try {
      const { data } = await fetchUnreadMessageCount();
      unreadCount.value = typeof data === 'number' ? data : 0;
    } catch {
      // 静默降级：未登录/接口异常时徽标归零
    }
  }

  async function loadNotifications() {
    try {
      const { data } = await fetchTicketMessages({ page: 1, pageSize: 10 });
      notifications.value = Array.isArray(data?.list) ? data.list : [];
    } catch {
      notifications.value = [];
    }
  }

  // 事件类型 → 图标/语义
  const eventIconMap: Record<string, string> = {
    pending: 'mdi:file-document-edit-outline',
    result: 'mdi:check-circle-outline',
    reassign: 'mdi:account-switch-outline',
    cancel: 'mdi:file-cancel-outline',
    urge: 'mdi:alarm-light-outline',
    timeout: 'mdi:timer-alert-outline',
    escalation: 'mdi:arrow-up-bold-circle-outline'
  };

  function eventIcon(event: string) {
    return eventIconMap[event] || 'mdi:bell-outline';
  }

  // 相对时间展示
  function relativeTime(iso: string) {
    const ts = new Date(iso).getTime();
    if (Number.isNaN(ts)) return '';
    const diff = Date.now() - ts;
    if (diff < 60_000) return '刚刚';
    if (diff < 3_600_000) return `${Math.floor(diff / 60_000)}分钟前`;
    if (diff < 86_400_000) return `${Math.floor(diff / 3_600_000)}小时前`;
    return `${Math.floor(diff / 86_400_000)}天前`;
  }

  // 点击消息：标记已读并跳转工单中心
  async function handleNotificationClick(notif: Api.Ticket.TicketMessage) {
    showNotifications.value = false;
    if (!notif.isRead) {
      markTicketMessagesRead([notif.id]).then(loadUnread);
    }
    router.push('/ticket/center');
  }

  // 全部已读
  async function handleMarkAllRead() {
    if (unreadCount.value === 0) return;
    await markTicketMessagesRead([]);
    await loadUnread();
    await loadNotifications();
  }

  // ─────────────────────────────────────────────
  // 通知偏好（本人自助）：IM userid 绑定 + 事件/渠道开关
  // 语义：事件关=完全静音（含站内）；渠道关=不推送该渠道（站内保留）
  // ─────────────────────────────────────────────
  const showPrefDialog = ref(false);
  const prefLoading = ref(false);
  const prefSaving = ref(false);
  const prefForm = ref<Api.Ticket.UserNotifySetting>({
    dingtalkId: '',
    wechatId: '',
    mutedEvents: [],
    offChannels: []
  });

  // 事件与渠道文案（与后端 notifyEventMetas / 渠道池一致）
  const prefEvents: Array<{ value: string; label: string }> = [
    { value: 'pending', label: '待审批' },
    { value: 'result', label: '审批结果' },
    { value: 'reassign', label: '改派给我' },
    { value: 'cancel', label: '工单撤销' },
    { value: 'urge', label: '催办提醒' },
    { value: 'timeout', label: '审批超时提醒' },
    { value: 'escalation', label: '超时升级' }
  ];
  const prefChannels: Array<{ value: string; label: string }> = [
    { value: 'email', label: '邮件' },
    { value: 'wechat', label: '企业微信' },
    { value: 'dingtalk', label: '钉钉' }
  ];

  // 勾选语义取反：勾=接收/启用，未勾=屏蔽/停用（后端存的是屏蔽/停用列表）
  const receiveEvents = computed<string[]>({
    get: () => prefEvents.map(e => e.value).filter(v => !prefForm.value.mutedEvents.includes(v)),
    set: checked => {
      prefForm.value.mutedEvents = prefEvents.map(e => e.value).filter(v => !checked.includes(v));
    }
  });
  const pushChannels = computed<string[]>({
    get: () => prefChannels.map(c => c.value).filter(v => !prefForm.value.offChannels.includes(v)),
    set: checked => {
      prefForm.value.offChannels = prefChannels.map(c => c.value).filter(v => !checked.includes(v));
    }
  });

  async function openPrefDialog() {
    showNotifications.value = false;
    showPrefDialog.value = true;
    prefLoading.value = true;
    try {
      const { data } = await fetchMyNotifySetting();
      if (data) {
        prefForm.value = {
          dingtalkId: data.dingtalkId || '',
          wechatId: data.wechatId || '',
          mutedEvents: data.mutedEvents || [],
          offChannels: data.offChannels || []
        };
      }
    } catch {
      // 读取失败保留空默认（跟随全局）
    } finally {
      prefLoading.value = false;
    }
  }

  async function savePref() {
    prefSaving.value = true;
    try {
      const { error } = await updateMyNotifySetting(prefForm.value);
      if (!error) {
        ElMessage.success('通知偏好已保存');
        showPrefDialog.value = false;
      } else {
        ElMessage.error('保存失败');
      }
    } catch {
      ElMessage.error('保存失败');
    } finally {
      prefSaving.value = false;
    }
  }

  // 打开面板时刷新列表
  function toggleNotifications() {
    showNotifications.value = !showNotifications.value;
    if (showNotifications.value) {
      loadNotifications();
      loadUnread();
    }
  }

  onMounted(() => {
    loadUnread();
    unreadTimer = setInterval(loadUnread, 60_000); // 每分钟刷新未读徽标
  });

  onUnmounted(() => {
    if (unreadTimer) clearInterval(unreadTimer);
  });

  // 点击外部关闭下拉菜单
  const handleClickOutside = () => {
    showQuickMenu.value = false;
    showNotifications.value = false;
  };
</script>

<template>
  <DarkModeContainer
    class="global-header relative h-full flex-y-center px-12px shadow-header"
    :style="headerStyle"
    @click="handleClickOutside"
  >
    <GlobalLogo v-if="showLogo" class="h-full" :style="{ width: themeStore.sider.width + 'px' }" />
    <MenuToggler v-if="showMenuToggler" :collapsed="appStore.siderCollapse" @click="appStore.toggleSiderCollapse" />
    <div v-if="showMenu" :id="GLOBAL_HEADER_MENU_ID" class="h-full flex-y-center flex-1-hidden"></div>
    <div v-else class="h-full flex-y-center flex-1-hidden">
      <GlobalBreadcrumb v-if="!appStore.isMobile" class="ml-12px" />
    </div>

    <!-- 左侧系统状态区域 -->
    <div v-if="!appStore.isMobile" class="mx-16px h-full flex-y-center flex-1-hidden">
      <div class="flex items-center gap-12px text-xs">
        <!-- 状态指示器组 -->
        <div class="flex items-center gap-8px">
          <!-- API状态 -->
          <div class="status-item">
            <div class="status-dot" :class="systemStatus.api ? 'status-success' : 'status-error'"></div>
            <span class="status-label">API</span>
          </div>

          <!-- 数据库状态 -->
          <div class="status-item">
            <div class="status-dot" :class="systemStatus.db ? 'status-success' : 'status-error'"></div>
            <span class="status-label">数据库</span>
          </div>

          <!-- Agent状态 -->
          <div class="status-item">
            <Icon icon="mdi:robot" class="status-icon" />
            <span class="status-label">{{ systemStatus.agents }} Agent</span>
          </div>
        </div>

        <div class="separator">|</div>

        <!-- 环境信息 -->
        <div class="flex items-center gap-4px text-gray-500 dark:text-gray-400">
          <span class="rounded bg-gray-100 px-8px py-2px dark:bg-gray-800">{{ systemInfo.environment }}</span>
          <span class="text-gray-400">{{ systemInfo.version }}</span>
        </div>
      </div>
    </div>

    <!-- 右侧操作区域 -->
    <div class="h-full flex-y-center justify-end gap-6px">
      <!-- 通知中心：工单站内消息 -->
      <div v-if="!appStore.isMobile" class="header-action-item">
        <div class="relative" @click.stop="toggleNotifications">
          <ButtonIcon icon="mdi:bell-outline" tooltip-content="通知中心" />
          <span v-if="unreadCount > 0" class="notification-badge">{{ unreadCount > 99 ? '99+' : unreadCount }}</span>
          <!-- 通知下拉面板 -->
          <div v-if="showNotifications" class="dropdown-panel notification-dropdown">
            <div class="dropdown-header">
              <span>通知中心</span>
              <button
                type="button"
                class="text-xs text-blue-500 hover:text-blue-600"
                :disabled="unreadCount === 0"
                @click.stop="handleMarkAllRead"
              >
                全部已读
              </button>
            </div>
            <div class="dropdown-content">
              <div
                v-for="notif in notifications"
                :key="notif.id"
                class="notification-item"
                :class="notif.isRead === 0 ? 'notification-unread' : ''"
                @click="handleNotificationClick(notif)"
              >
                <div class="notification-icon">
                  <Icon :icon="eventIcon(notif.event)" />
                </div>
                <div class="notification-content">
                  <div class="notification-title">
                    {{ notif.title }}
                    <span v-if="notif.isRead === 0" class="unread-dot"></span>
                  </div>
                  <div class="notification-message">{{ notif.content }}</div>
                  <div class="notification-time">{{ relativeTime(notif.createdAt) }}</div>
                </div>
              </div>
              <div v-if="notifications.length === 0" class="py-12px text-center text-gray-500">暂无通知</div>
            </div>
            <div class="dropdown-footer" @click.stop="openPrefDialog">
              <Icon icon="mdi:cog-outline" />
              <span>通知偏好</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 通知偏好弹窗（本人自助） -->
      <ElDialog v-model="showPrefDialog" title="通知偏好" width="480px" append-to-body>
        <div v-loading="prefLoading" class="flex flex-col gap-16px">
          <div>
            <div class="pref-section-title">IM 账号绑定</div>
            <div class="mb-8px text-xs text-gray-400">配置后群消息按 userid @你，不再依赖手机号</div>
            <ElInput v-model="prefForm.dingtalkId" placeholder="钉钉 userid（选填）" class="mb-8px">
              <template #prefix><Icon icon="mdi:chat-outline" /></template>
            </ElInput>
            <ElInput v-model="prefForm.wechatId" placeholder="企业微信 userid（选填）">
              <template #prefix><Icon icon="mdi:wechat" /></template>
            </ElInput>
          </div>
          <div>
            <div class="pref-section-title">接收事件</div>
            <div class="mb-8px text-xs text-gray-400">取消勾选的事件将完全不通知（含站内消息）</div>
            <ElCheckboxGroup v-model="receiveEvents">
              <div class="flex flex-col gap-4px">
                <ElCheckbox v-for="ev in prefEvents" :key="ev.value" :label="ev.value">
                  {{ ev.label }}
                </ElCheckbox>
              </div>
            </ElCheckboxGroup>
          </div>
          <div>
            <div class="pref-section-title">推送渠道</div>
            <div class="mb-8px text-xs text-gray-400">取消勾选的渠道不再推送（站内消息保留）</div>
            <ElCheckboxGroup v-model="pushChannels">
              <div class="flex flex-col gap-4px">
                <ElCheckbox v-for="ch in prefChannels" :key="ch.value" :label="ch.value">
                  {{ ch.label }}
                </ElCheckbox>
              </div>
            </ElCheckboxGroup>
          </div>
        </div>
        <template #footer>
          <ElButton @click="showPrefDialog = false">取消</ElButton>
          <ElButton type="primary" :loading="prefSaving" @click="savePref">保存</ElButton>
        </template>
      </ElDialog>

      <!-- 快速操作菜单 -->
      <div v-if="!appStore.isMobile" class="header-action-item">
        <div class="relative" @click.stop="showQuickMenu = !showQuickMenu">
          <ButtonIcon icon="mdi:dots-horizontal" tooltip-content="快速操作" />
          <div v-if="showQuickMenu" class="dropdown-panel quick-menu-dropdown">
            <div v-for="action in quickActions" :key="action.name" class="dropdown-item" @click="action.action">
              <Icon :icon="action.icon" />
              <span>{{ action.name }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 快速链接按钮 -->
      <template v-if="!appStore.isMobile">
        <ButtonIcon
          v-for="link in quickLinks"
          :key="link.name"
          :icon="link.icon"
          :tooltip-content="link.tooltip"
          @click="link.action"
        />
      </template>

      <!-- 时间显示 -->
      <div v-if="!appStore.isMobile" class="time-display">
        <Icon icon="mdi:clock-outline" class="time-icon" />
        <span>{{ formatTime(currentTime) }}</span>
      </div>

      <GlobalSearch v-if="themeStore.header.globalSearch.visible" />
      <div>
        <FullScreen v-if="!appStore.isMobile" :full="isFullscreen" @click="toggle" />
      </div>
      <LangSwitch
        v-if="themeStore.header.multilingual.visible"
        :lang="appStore.locale"
        :lang-options="appStore.localeOptions"
        @change-lang="appStore.changeLocale"
      />
      <ThemeSchemaSwitch
        :theme-schema="themeStore.themeScheme"
        :is-dark="themeStore.darkMode"
        @switch="themeStore.toggleThemeScheme"
      />
      <div>
        <ThemeButton />
      </div>
      <UserAvatar />
    </div>
  </DarkModeContainer>
</template>

<style scoped>
  .status-item {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 4px 8px;
    border-radius: 6px;
    background: var(--el-fill-color-light);
    transition: all 0.2s;
  }

  .status-item:hover {
    background: var(--el-fill-color);
  }

  .status-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    transition: all 0.2s;
  }

  .status-success {
    background-color: #67c23a;
    box-shadow: 0 0 0 2px rgb(103 194 58 / 30%);
  }

  .status-error {
    background-color: #f56c6c;
    box-shadow: 0 0 0 2px rgb(245 108 108 / 30%);
    animation: pulse 2s infinite;
  }

  @keyframes pulse {
    0%,
    100% {
      opacity: 1;
    }

    50% {
      opacity: 0.5;
    }
  }

  .status-label {
    font-size: 12px;
    color: var(--el-text-color-regular);
  }

  .status-icon {
    font-size: 16px;
    color: var(--el-color-success);
  }

  .separator {
    color: var(--el-border-color);
    margin: 0 8px;
  }

  .header-action-item {
    position: relative;
  }

  .notification-badge {
    position: absolute;
    top: -4px;
    right: -4px;
    min-width: 16px;
    height: 16px;
    padding: 0 4px;
    border-radius: 8px;
    background: var(--el-color-danger);
    color: white;
    font-size: 11px;
    font-weight: 600;
    line-height: 16px;
    text-align: center;
    box-shadow: 0 2px 4px rgb(245 108 108 / 30%);
  }

  .dropdown-panel {
    position: absolute;
    top: 100%;
    right: 0;
    margin-top: 8px;
    min-width: 200px;
    background: var(--el-bg-color);
    border: 1px solid var(--el-border-color);
    border-radius: 8px;
    box-shadow: 0 4px 12px rgb(0 0 0 / 15%);
    z-index: 1000;
    overflow: hidden;
  }

  .notification-dropdown {
    width: 320px;
  }

  .dropdown-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 16px;
    border-bottom: 1px solid var(--el-border-color-light);
    font-weight: 600;
    font-size: 14px;
  }

  .dropdown-content {
    max-height: 400px;
    overflow-y: auto;
  }

  .dropdown-footer {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    padding: 10px 16px;
    border-top: 1px solid var(--el-border-color-light);
    font-size: 13px;
    color: var(--el-text-color-secondary);
    cursor: pointer;
    transition: all 0.2s;
  }

  .dropdown-footer:hover {
    color: var(--el-color-primary);
    background: var(--el-fill-color-light);
  }

  .pref-section-title {
    margin-bottom: 4px;
    font-weight: 600;
    font-size: 14px;
  }

  .pref-section-title + .el-checkbox-group .el-checkbox {
    display: flex;
  }

  .notification-item {
    display: flex;
    gap: 12px;
    padding: 12px 16px;
    border-bottom: 1px solid var(--el-border-color-light);
    cursor: pointer;
    transition: background 0.2s;
  }

  .notification-item:hover {
    background: var(--el-fill-color-light);
  }

  .notification-item:last-child {
    border-bottom: none;
  }

  .notification-icon {
    flex-shrink: 0;
    width: 32px;
    height: 32px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 16px;
  }

  .notification-warning .notification-icon {
    background: rgb(230 162 60 / 10%);
    color: var(--el-color-warning);
  }

  .notification-error .notification-icon {
    background: rgb(245 108 108 / 10%);
    color: var(--el-color-danger);
  }

  .notification-info .notification-icon {
    background: rgb(64 158 255 / 10%);
    color: var(--el-color-primary);
  }

  .notification-unread .notification-icon {
    background: rgb(64 158 255 / 12%);
    color: var(--el-color-primary);
  }

  .notification-unread .notification-title {
    color: var(--el-color-primary);
  }

  .unread-dot {
    display: inline-block;
    width: 6px;
    height: 6px;
    margin-left: 4px;
    border-radius: 50%;
    background: var(--el-color-danger);
    vertical-align: middle;
  }

  .notification-content {
    flex: 1;
    min-width: 0;
  }

  .notification-title {
    font-size: 13px;
    font-weight: 600;
    color: var(--el-text-color-primary);
    margin-bottom: 4px;
  }

  .notification-message {
    font-size: 12px;
    color: var(--el-text-color-regular);
    line-height: 1.4;
    margin-bottom: 4px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .notification-time {
    font-size: 11px;
    color: var(--el-text-color-secondary);
  }

  .quick-menu-dropdown {
    min-width: 140px;
  }

  .dropdown-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 16px;
    font-size: 13px;
    color: var(--el-text-color-regular);
    cursor: pointer;
    transition: all 0.2s;
  }

  .dropdown-item:hover {
    background: var(--el-fill-color-light);
    color: var(--el-color-primary);
  }

  .time-display {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 4px 8px;
    border-radius: 6px;
    font-family: Consolas, Monaco, monospace;
    font-size: 13px;
    color: var(--el-text-color-secondary);
    background: var(--el-fill-color-light);
  }

  .time-icon {
    font-size: 16px;
    color: var(--el-color-primary);
  }
</style>
