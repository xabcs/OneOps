<script setup lang="ts">
  import { onMounted, ref } from 'vue';
  import type { Component } from 'vue';
  import { ElMessage, ElMessageBox } from 'element-plus';
  import { Bell, ChatDotRound, Message, Plus } from '@element-plus/icons-vue';
  import {
    deleteNotificationChannel,
    fetchNotificationChannels,
    testNotificationChannel,
    updateNotificationChannel
  } from '@/service/api';
  import { useAuthStore } from '@/store/modules/auth';
  import NotificationChannelDialog from './modules/NotificationChannelDialog.vue';

  defineOptions({
    name: 'NotificationChannels'
  });

  const authStore = useAuthStore();

  // 无权限时点击置灰开关的提示（disabled 的 ElSwitch 不拦截原生 click 冒泡）
  function handleDisabledSwitchClick(code: string) {
    if (!authStore.hasPermission(code)) {
      ElMessage({
        type: 'warning',
        message: `缺少权限：${code}，请联系管理员在角色管理中开通`,
        grouping: true,
        showClose: true
      });
    }
  }

  // ============================================
  // 通知渠道管理（平台级渠道池：监控告警 / 工单通知共同引用）
  // ============================================

  const channels = ref<Monitoring.NotificationChannel[]>([]);
  const loading = ref(false);
  const showDialog = ref(false);
  const dialogMode = ref<'create' | 'edit'>('create');
  const currentChannel = ref<Monitoring.NotificationChannel | null>(null);

  // 通知渠道类型选项
  const channelTypeOptions = [
    { label: '邮件', value: 'email' },
    { label: '企业微信', value: 'wechat' },
    { label: '钉钉', value: 'dingtalk' },
    { label: '飞书', value: 'feishu' }
  ];

  // 加载通知渠道
  async function loadChannels() {
    loading.value = true;
    try {
      const { data } = await fetchNotificationChannels();
      channels.value = data || [];
    } catch (error) {
      ElMessage.error('加载通知渠道失败');
      channels.value = [];
    } finally {
      loading.value = false;
    }
  }

  // 新增通知渠道
  function handleCreate() {
    dialogMode.value = 'create';
    currentChannel.value = null;
    showDialog.value = true;
  }

  // 编辑通知渠道
  function handleEdit(channel: Monitoring.NotificationChannel) {
    dialogMode.value = 'edit';
    currentChannel.value = channel;
    showDialog.value = true;
  }

  // 删除通知渠道
  async function handleDelete(channel: Monitoring.NotificationChannel) {
    try {
      await ElMessageBox.confirm(
        `删除后，引用该渠道的工单通知事件将不再通过它发送。确定删除 "${channel.channelName}" 吗？`,
        '确认删除',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }
      );

      const { error } = await deleteNotificationChannel(channel.id);
      if (!error) {
        ElMessage.success('删除成功');
        await loadChannels();
      } else {
        ElMessage.error('删除失败');
      }
    } catch (error: unknown) {
      if ((error as string) !== 'cancel') {
        ElMessage.error('删除失败');
      }
    }
  }

  // 测试通知渠道
  async function handleTest(channel: Monitoring.NotificationChannel) {
    try {
      await ElMessageBox.confirm(`确定要测试发送通知到 "${channel.channelName}" 吗？`, '测试通知', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'info'
      });

      const { error } = await testNotificationChannel(channel.id);
      if (!error) {
        ElMessage.success('测试通知发送成功');
      } else {
        ElMessage.error('测试通知发送失败');
      }
    } catch (error: unknown) {
      if ((error as string) !== 'cancel') {
        ElMessage.error('测试通知发送失败');
      }
    }
  }

  // 切换渠道启用状态（复用全量更新接口，保持其余配置不变）
  async function handleToggleStatus(channel: Monitoring.NotificationChannel) {
    const { error } = await updateNotificationChannel(channel.id, {
      channelName: channel.channelName,
      channelType: channel.channelType,
      config: channel.config,
      enabled: !channel.enabled
    } as Monitoring.NotificationChannelForm);
    if (!error) {
      ElMessage.success(channel.enabled ? '已禁用' : '已启用');
      await loadChannels();
    } else {
      ElMessage.error('操作失败');
    }
  }

  // 获取渠道类型图标
  function getChannelIcon(type: string): Component {
    const icons: Record<string, Component> = {
      email: Message,
      wechat: ChatDotRound,
      dingtalk: Bell,
      feishu: Bell
    };
    return icons[type] || Message;
  }

  // 获取渠道类型文本
  function getChannelTypeText(type: string) {
    const option = channelTypeOptions.find(opt => opt.value === type);
    return option?.label || type;
  }

  // 页面加载时初始化
  onMounted(() => {
    loadChannels();
  });
</script>

<template>
  <div class="table-page">
    <!-- 标题栏 -->
    <ElCard shadow="never" class="card-static">
      <div class="flex items-center justify-between">
        <div>
          <span class="text-lg font-semibold">通知渠道</span>
          <div class="mt-1 text-xs text-gray-400">
            平台级渠道池：监控告警与工单通知（工单中心-通知设置）共同引用此处的渠道
          </div>
        </div>
        <PermissionButton code="monitor.notification.create" type="primary" @click="handleCreate">
          <ElIcon :size="16">
            <Plus />
          </ElIcon>
          新增渠道
        </PermissionButton>
      </div>
    </ElCard>

    <!-- 渠道列表 -->
    <ElCard shadow="never">
      <div class="table-scroll-wrap">
        <ElTable v-loading="loading" :data="channels" border stripe height="100%">
          <ElTableColumn label="类型" width="110" align="center">
            <template #default="{ row }">
              <ElIcon :size="20">
                <component :is="getChannelIcon(row.channelType)" />
              </ElIcon>
              <span class="ml-2">{{ getChannelTypeText(row.channelType) }}</span>
            </template>
          </ElTableColumn>
          <ElTableColumn label="名称" prop="channelName" min-width="150" />
          <ElTableColumn label="配置信息" min-width="250">
            <template #default="{ row }">
              <div v-if="row.channelType === 'email'" class="text-sm">
                <div class="text-gray-600">
                  SMTP: {{ (row.config as Monitoring.EmailConfig).smtpHost }}:{{
                    (row.config as Monitoring.EmailConfig).smtpPort
                  }}
                </div>
                <div class="text-gray-600">发件人: {{ (row.config as Monitoring.EmailConfig).from }}</div>
              </div>
              <div v-else-if="row.channelType === 'wechat'" class="text-sm">
                <div class="truncate text-gray-600" :title="(row.config as Monitoring.WeChatConfig).webhookUrl">
                  Webhook: {{ (row.config as Monitoring.WeChatConfig).webhookUrl }}
                </div>
              </div>
              <div v-else class="text-sm text-gray-400">点击"编辑"查看配置</div>
            </template>
          </ElTableColumn>
          <ElTableColumn label="状态" width="90" align="center">
            <template #default="{ row }">
              <ElSwitch
                :model-value="row.enabled"
                :disabled="!authStore.hasPermission('monitor.notification.update')"
                title="缺少权限：monitor.notification.update"
                @click="handleDisabledSwitchClick('monitor.notification.update')"
                @change="handleToggleStatus(row)"
              />
            </template>
          </ElTableColumn>
          <ElTableColumn label="操作" width="200" align="center" fixed="right" class-name="msre-table-actions">
            <template #default="{ row }">
              <PermissionButton
                link
                type="primary"
                size="small"
                code="monitor.notification.update"
                @click="handleEdit(row)"
              >
                编辑
              </PermissionButton>
              <PermissionButton
                link
                type="primary"
                size="small"
                code="monitor.notification.test"
                :disabled="!row.enabled"
                @click="handleTest(row)"
              >
                测试
              </PermissionButton>
              <PermissionButton
                link
                type="danger"
                size="small"
                code="monitor.notification.delete"
                @click="handleDelete(row)"
              >
                删除
              </PermissionButton>
            </template>
          </ElTableColumn>
        </ElTable>
      </div>
    </ElCard>

    <!-- 通知渠道对话框 -->
    <NotificationChannelDialog
      v-model:visible="showDialog"
      :mode="dialogMode"
      :channel="currentChannel"
      @submitted="loadChannels"
    />
  </div>
</template>
