<script setup lang="ts">
  import { onMounted, ref } from 'vue';
  import type { Component } from 'vue';
  import { ElMessage, ElMessageBox } from 'element-plus';
  import { Bell, ChatDotRound, Message, Plus } from '@element-plus/icons-vue';
  import {
    deleteAlertRule,
    deleteNotificationChannel,
    fetchAlertRules,
    fetchNotificationChannels,
    testNotificationChannel,
    updateAlertRuleStatus
  } from '@/service/api';
  import AlertRuleDialog from './modules/AlertRuleDialog.vue';
  import NotificationChannelDialog from './modules/NotificationChannelDialog.vue';

  defineOptions({
    name: 'MonitoringSettings'
  });

  const activeTab = ref('rules');

  // ============================================
  // 告警规则管理
  // ============================================

  const alertRules = ref<Monitoring.AlertRule[]>([]);
  const alertRulesLoading = ref(false);
  const showRuleDialog = ref(false);
  const ruleDialogMode = ref<'create' | 'edit'>('create');
  const currentRule = ref<Monitoring.AlertRule | null>(null);

  // 监控指标选项
  const metricOptions = [
    { label: 'CPU 使用率', value: 'cpu_usage' },
    { label: '内存使用率', value: 'memory_usage' },
    { label: '磁盘使用率', value: 'disk_usage' },
    { label: '1分钟负载', value: 'load1' },
    { label: '5分钟负载', value: 'load5' },
    { label: '15分钟负载', value: 'load15' }
  ];

  // 加载告警规则
  async function loadAlertRules() {
    alertRulesLoading.value = true;
    try {
      const { data } = await fetchAlertRules();
      alertRules.value = data || [];
    } catch (error) {
      ElMessage.error('加载告警规则失败');
      alertRules.value = [];
    } finally {
      alertRulesLoading.value = false;
    }
  }

  // 新增告警规则
  function handleCreateRule() {
    ruleDialogMode.value = 'create';
    currentRule.value = null;
    showRuleDialog.value = true;
  }

  // 编辑告警规则
  function handleEditRule(rule: Monitoring.AlertRule) {
    ruleDialogMode.value = 'edit';
    currentRule.value = rule;
    showRuleDialog.value = true;
  }

  // 删除告警规则
  async function handleDeleteRule(rule: Monitoring.AlertRule) {
    try {
      await ElMessageBox.confirm(`确定要删除告警规则 "${rule.name}" 吗？`, '确认删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      });

      const { error } = await deleteAlertRule(rule.id);
      if (!error) {
        ElMessage.success('删除成功');
        await loadAlertRules();
      } else {
        ElMessage.error('删除失败');
      }
    } catch (error: unknown) {
      if ((error as string) !== 'cancel') {
        ElMessage.error('删除失败');
      }
    }
  }

  // 切换告警规则状态
  async function handleToggleRuleStatus(rule: Monitoring.AlertRule) {
    try {
      const { error } = await updateAlertRuleStatus(rule.id, !rule.enabled);
      if (!error) {
        ElMessage.success(rule.enabled ? '已禁用' : '已启用');
        await loadAlertRules();
      } else {
        ElMessage.error('操作失败');
      }
    } catch (error) {
      ElMessage.error('操作失败');
    }
  }

  // 获取级别标签类型
  type TagType = 'primary' | 'info' | 'success' | 'warning' | 'danger';
  function getLevelTagType(level: string): TagType | undefined {
    const map: Record<string, TagType> = {
      critical: 'danger',
      high: 'warning',
      medium: 'info',
      low: 'primary',
      info: 'success'
    };
    return map[level];
  }

  // 获取级别文本
  function getLevelText(level: string) {
    const map: Record<string, string> = {
      critical: '严重',
      high: '高',
      medium: '中',
      low: '低',
      info: '信息'
    };
    return map[level] || level;
  }

  // 获取指标文本
  function getMetricText(metric: string) {
    const option = metricOptions.find(opt => opt.value === metric);
    return option?.label || metric;
  }

  // ============================================
  // 通知渠道管理
  // ============================================

  const notificationChannelsData = ref<Monitoring.NotificationChannel[]>([]);
  const notificationLoading = ref(false);
  const showChannelDialog = ref(false);
  const channelDialogMode = ref<'create' | 'edit'>('create');
  const currentChannel = ref<Monitoring.NotificationChannel | null>(null);

  // 通知渠道类型选项
  const channelTypeOptions = [
    { label: '邮件', value: 'email' },
    { label: '企业微信', value: 'wechat' },
    { label: '钉钉', value: 'dingtalk' },
    { label: '飞书', value: 'feishu' }
  ];

  // 加载通知渠道
  async function loadNotificationChannels() {
    notificationLoading.value = true;
    try {
      const { data } = await fetchNotificationChannels();
      notificationChannelsData.value = data || [];
    } catch (error) {
      ElMessage.error('加载通知渠道失败');
      notificationChannelsData.value = [];
    } finally {
      notificationLoading.value = false;
    }
  }

  // 新增通知渠道
  function handleCreateChannel() {
    channelDialogMode.value = 'create';
    currentChannel.value = null;
    showChannelDialog.value = true;
  }

  // 编辑通知渠道
  function handleEditChannel(channel: Monitoring.NotificationChannel) {
    channelDialogMode.value = 'edit';
    currentChannel.value = channel;
    showChannelDialog.value = true;
  }

  // 删除通知渠道
  async function handleDeleteChannel(channel: Monitoring.NotificationChannel) {
    try {
      await ElMessageBox.confirm(`确定要删除通知渠道 "${channel.channelName}" 吗？`, '确认删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      });

      const { error } = await deleteNotificationChannel(channel.id);
      if (!error) {
        ElMessage.success('删除成功');
        await loadNotificationChannels();
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
  async function handleTestChannel(channel: Monitoring.NotificationChannel) {
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
  function getChannelTypeText(type: string): string {
    const option = channelTypeOptions.find(opt => opt.value === type);
    return option?.label || type;
  }

  // 页面加载时初始化
  onMounted(() => {
    loadAlertRules();
    loadNotificationChannels();
  });
</script>

<template>
  <div class="p-4 space-y-4">
    <!-- 标题栏 -->
    <ElCard shadow="never">
      <div class="flex items-center justify-between">
        <span class="text-lg font-semibold">监控配置</span>
      </div>
    </ElCard>

    <!-- 配置选项卡 -->
    <ElCard shadow="never">
      <ElTabs v-model="activeTab">
        <!-- 告警规则 -->
        <ElTabPane label="告警规则" name="rules">
          <div class="max-w-6xl">
            <div class="mb-4 flex items-center justify-between">
              <ElText type="info">配置触发告警的规则条件</ElText>
              <ElButton type="primary" @click="handleCreateRule">
                <ElIcon :size="16">
                  <Plus />
                </ElIcon>
                新增规则
              </ElButton>
            </div>

            <ElTable v-loading="alertRulesLoading" :data="alertRules" border stripe>
              <ElTableColumn label="规则名称" prop="name" min-width="150" />
              <ElTableColumn label="监控指标" width="130">
                <template #default="{ row }">
                  {{ getMetricText(row.metric) }}
                </template>
              </ElTableColumn>
              <ElTableColumn label="条件" width="70" align="center">
                <template #default="{ row }">
                  {{ row.condition }}
                </template>
              </ElTableColumn>
              <ElTableColumn label="阈值" width="100" align="center">
                <template #default="{ row }">{{ row.threshold }}{{ row.metric.includes('usage') ? '%' : '' }}</template>
              </ElTableColumn>
              <ElTableColumn label="持续时间" width="100" align="center">
                <template #default="{ row }">{{ row.duration }}秒</template>
              </ElTableColumn>
              <ElTableColumn label="级别" width="80" align="center">
                <template #default="{ row }">
                  <ElTag :type="getLevelTagType(row.level)" size="small">
                    {{ getLevelText(row.level) }}
                  </ElTag>
                </template>
              </ElTableColumn>
              <ElTableColumn label="状态" width="90" align="center">
                <template #default="{ row }">
                  <ElSwitch :model-value="row.enabled" @change="handleToggleRuleStatus(row)" />
                </template>
              </ElTableColumn>
              <ElTableColumn label="操作" width="150" align="center" fixed="right">
                <template #default="{ row }">
                  <ElButton type="primary" link size="small" @click="handleEditRule(row)">编辑</ElButton>
                  <ElButton type="danger" link size="small" @click="handleDeleteRule(row)">删除</ElButton>
                </template>
              </ElTableColumn>
            </ElTable>

            <ElEmpty v-if="!alertRulesLoading && alertRules.length === 0" description="暂无告警规则" />
          </div>
        </ElTabPane>

        <!-- 通知配置 -->
        <ElTabPane label="通知配置" name="notification">
          <div class="max-w-4xl">
            <div class="mb-4 flex items-center justify-between">
              <ElText type="info">配置告警通知渠道</ElText>
              <ElButton type="primary" @click="handleCreateChannel">
                <ElIcon :size="16">
                  <Plus />
                </ElIcon>
                新增渠道
              </ElButton>
            </div>

            <ElTable v-loading="notificationLoading" :data="notificationChannelsData" border stripe>
              <ElTableColumn label="类型" width="100" align="center">
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
                  <ElSwitch :model-value="row.enabled" @change="row.enabled = !row.enabled" />
                </template>
              </ElTableColumn>
              <ElTableColumn label="操作" width="200" align="center" fixed="right">
                <template #default="{ row }">
                  <ElButton type="primary" link size="small" @click="handleEditChannel(row)">编辑</ElButton>
                  <ElButton type="success" link size="small" :disabled="!row.enabled" @click="handleTestChannel(row)">
                    测试
                  </ElButton>
                  <ElButton type="danger" link size="small" @click="handleDeleteChannel(row)">删除</ElButton>
                </template>
              </ElTableColumn>
            </ElTable>

            <ElEmpty v-if="!notificationLoading && notificationChannelsData.length === 0" description="暂无通知渠道" />
          </div>
        </ElTabPane>
      </ElTabs>
    </ElCard>

    <!-- 告警规则对话框 -->
    <AlertRuleDialog
      v-model:visible="showRuleDialog"
      :mode="ruleDialogMode"
      :rule="currentRule"
      @submitted="loadAlertRules"
    />

    <!-- 通知渠道对话框 -->
    <NotificationChannelDialog
      v-model:visible="showChannelDialog"
      :mode="channelDialogMode"
      :channel="currentChannel"
      @submitted="loadNotificationChannels"
    />
  </div>
</template>

<style scoped>
  .time-picker {
    width: 100%;
    max-width: 350px;
  }
</style>
