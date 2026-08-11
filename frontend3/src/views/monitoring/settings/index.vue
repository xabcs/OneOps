<script setup lang="ts">
    import { computed, onMounted, reactive, ref } from 'vue';
    import { ElMessage, ElMessageBox } from 'element-plus';
    import { Bell, ChatDotRound, Delete, Edit, Message, Plus } from '@element-plus/icons-vue';
    import type { Component } from 'vue';
    import {
      createAlertRule,
      createNotificationChannel,
      deleteAlertRule,
      deleteNotificationChannel,
      fetchAlertRules,
      fetchNotificationChannels,
      testNotificationChannel,
      updateAlertRule,
      updateAlertRuleStatus,
      updateNotificationChannel
    } from '@/service/api';

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
    const currentRule = ref<Monitoring.AlertRuleForm>({
      name: '',
      level: 'high',
      metric: 'cpu_usage',
      condition: '>',
      threshold: 80,
      duration: 300,
      description: ''
    });

    // 告警级别选项
    const alertLevelOptions = [
      { label: '严重', value: 'critical' },
      { label: '高', value: 'high' },
      { label: '中', value: 'medium' },
      { label: '低', value: 'low' },
      { label: '信息', value: 'info' }
    ];

    // 监控指标选项
    const metricOptions = [
      { label: 'CPU 使用率', value: 'cpu_usage' },
      { label: '内存使用率', value: 'memory_usage' },
      { label: '磁盘使用率', value: 'disk_usage' },
      { label: '1分钟负载', value: 'load1' },
      { label: '5分钟负载', value: 'load5' },
      { label: '15分钟负载', value: 'load15' }
    ];

    // 告警条件选项
    const conditionOptions = [
      { label: '大于', value: '>' },
      { label: '小于', value: '<' },
      { label: '等于', value: '==' },
      { label: '不等于', value: '!=' }
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
      currentRule.value = {
        name: '',
        level: 'high',
        metric: 'cpu_usage',
        condition: '>',
        threshold: 80,
        duration: 300,
        description: ''
      };
      showRuleDialog.value = true;
    }

    // 编辑告警规则
    function handleEditRule(rule: Monitoring.AlertRule) {
      ruleDialogMode.value = 'edit';
      currentRule.value = {
        id: rule.id,
        name: rule.name,
        level: rule.level,
        metric: rule.metric,
        condition: rule.condition,
        threshold: rule.threshold,
        duration: rule.duration,
        description: rule.description,
        enabled: rule.enabled
      };
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

        await deleteAlertRule(rule.id);
        ElMessage.success('删除成功');
        await loadAlertRules();
      } catch (error: unknown) {
        if ((error as string) !== 'cancel') {
          ElMessage.error('删除失败');
        }
      }
    }

    // 切换告警规则状态
    async function handleToggleRuleStatus(rule: Monitoring.AlertRule) {
      try {
        await updateAlertRuleStatus(rule.id, !rule.enabled);
        ElMessage.success(rule.enabled ? '已禁用' : '已启用');
        await loadAlertRules();
      } catch (error) {
        ElMessage.error('操作失败');
      }
    }

    // 保存告警规则
    async function handleSaveRule() {
      // 验证表单
      if (!currentRule.value.name) {
        ElMessage.warning('请输入规则名称');
        return;
      }
      if (currentRule.value.threshold <= 0) {
        ElMessage.warning('阈值必须大于0');
        return;
      }

      try {
        if (ruleDialogMode.value === 'create') {
          await createAlertRule(currentRule.value);
          ElMessage.success('创建成功');
        } else {
          await updateAlertRule(currentRule.value.id!, currentRule.value);
          ElMessage.success('更新成功');
        }
        showRuleDialog.value = false;
        await loadAlertRules();
      } catch (error) {
        ElMessage.error(ruleDialogMode.value === 'create' ? '创建失败' : '更新失败');
      }
    }

    // 获取级别标签类型
    function getLevelTagType(level: string) {
      const map: Record<string, string> = {
        critical: 'danger',
        high: 'warning',
        medium: 'info',
        low: 'primary',
        info: 'success'
      };
      return map[level] || '';
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
    const currentChannel = ref<Monitoring.NotificationChannelForm>({
      channelType: 'email',
      channelName: '',
      config: {} as Record<string, unknown>,
      enabled: true
    });

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
      currentChannel.value = {
        channelType: 'email',
        channelName: '',
        config: {
          smtpHost: '',
          smtpPort: 587,
          username: '',
          password: '',
          from: '',
          fromName: '',
          useTLS: true,
          recipients: []
        },
        enabled: true
      };
      showChannelDialog.value = true;
    }

    // 编辑通知渠道
    function handleEditChannel(channel: Monitoring.NotificationChannel) {
      channelDialogMode.value = 'edit';
      currentChannel.value = {
        id: channel.id,
        channelType: channel.channelType,
        channelName: channel.channelName,
        config: channel.config as Monitoring.EmailConfig,
        enabled: channel.enabled
      };
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

        await deleteNotificationChannel(channel.id);
        ElMessage.success('删除成功');
        await loadNotificationChannels();
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

        await testNotificationChannel(channel.id);
        ElMessage.success('测试通知发送成功');
      } catch (error: unknown) {
        if ((error as string) !== 'cancel') {
          ElMessage.error('测试通知发送失败');
        }
      }
    }

    // 保存通知渠道
    async function handleSaveChannel() {
      // 验证表单
      if (!currentChannel.value.channelName) {
        ElMessage.warning('请输入渠道名称');
        return;
      }

      // 根据类型验证配置
      if (currentChannel.value.channelType === 'email') {
        const config = currentChannel.value.config as Monitoring.EmailConfig;
        if (!config.smtpHost || !config.from) {
          ElMessage.warning('请完善邮件配置');
          return;
        }
      } else if (currentChannel.value.channelType === 'wechat') {
        const config = currentChannel.value.config as Monitoring.WeChatConfig;
        if (!config.webhookUrl) {
          ElMessage.warning('请输入企业微信 Webhook URL');
          return;
        }
      }

      try {
        if (channelDialogMode.value === 'create') {
          await createNotificationChannel(currentChannel.value);
          ElMessage.success('创建成功');
        } else {
          await updateNotificationChannel(currentChannel.value.id!, currentChannel.value);
          ElMessage.success('更新成功');
        }
        showChannelDialog.value = false;
        await loadNotificationChannels();
      } catch (error) {
        ElMessage.error(channelDialogMode.value === 'create' ? '创建失败' : '更新失败');
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
                                            SMTP: {{ (row.config as Monitoring.EmailConfig).smtpHost }}:{{ (row.config as Monitoring.EmailConfig).smtpPort }}
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
        <ElDialog v-model="showRuleDialog" :title="ruleDialogMode === 'create' ? '新增告警规则' : '编辑告警规则'" width="600px" :close-on-click-modal="false">
            <ElForm :model="currentRule" label-width="120px">
                <ElFormItem label="规则名称" required>
                    <ElInput v-model="currentRule.name" placeholder="请输入规则名称" />
                </ElFormItem>

                <ElFormItem label="告警级别" required>
                    <ElSelect v-model="currentRule.level" placeholder="请选择告警级别">
                        <ElOption v-for="level in alertLevelOptions" :key="level.value" :label="level.label" :value="level.value" />
                    </ElSelect>
                </ElFormItem>

                <ElFormItem label="监控指标" required>
                    <ElSelect v-model="currentRule.metric" placeholder="请选择监控指标">
                        <ElOption v-for="metric in metricOptions" :key="metric.value" :label="metric.label" :value="metric.value" />
                    </ElSelect>
                </ElFormItem>

                <ElFormItem label="判断条件" required>
                    <ElSelect v-model="currentRule.condition" placeholder="请选择条件">
                        <ElOption v-for="cond in conditionOptions" :key="cond.value" :label="cond.label" :value="cond.value" />
                    </ElSelect>
                </ElFormItem>

                <ElFormItem label="阈值" required>
                    <ElInputNumber v-model="currentRule.threshold" :min="0" :max="100" :precision="2" controls-position="right" />
                    <span class="ml-2 text-gray-500">
                        {{ currentRule.metric.includes('usage') ? '%' : '' }}
                    </span>
                </ElFormItem>

                <ElFormItem label="持续时间" required>
                    <ElInputNumber v-model="currentRule.duration" :min="0" :max="86400" :step="60" controls-position="right" />
                    <span class="ml-2 text-gray-500">秒（指标条件需持续多久才触发告警）</span>
                </ElFormItem>

                <ElFormItem label="描述">
                    <ElInput v-model="currentRule.description" type="textarea" :rows="3" placeholder="请输入规则描述" />
                </ElFormItem>
            </ElForm>

            <template #footer>
                <ElButton @click="showRuleDialog = false">取消</ElButton>
                <ElButton type="primary" @click="handleSaveRule">确定</ElButton>
            </template>
        </ElDialog>

        <!-- 通知渠道对话框 -->
        <ElDialog v-model="showChannelDialog" :title="channelDialogMode === 'create' ? '新增通知渠道' : '编辑通知渠道'" width="650px" :close-on-click-modal="false">
            <ElForm :model="currentChannel" label-width="120px">
                <ElFormItem label="渠道类型" required>
                    <ElSelect v-model="currentChannel.channelType" placeholder="请选择渠道类型" :disabled="channelDialogMode === 'edit'">
                        <ElOption v-for="type in channelTypeOptions" :key="type.value" :label="type.label" :value="type.value" />
                    </ElSelect>
                </ElFormItem>

                <ElFormItem label="渠道名称" required>
                    <ElInput v-model="currentChannel.channelName" placeholder="请输入渠道名称" />
                </ElFormItem>

                <!-- 邮件配置 -->
                <template v-if="currentChannel.channelType === 'email'">
                    <ElDivider content-position="left">邮件服务器配置</ElDivider>

                    <ElFormItem label="SMTP 服务器" required>
                        <ElInput v-model="(currentChannel.config as Monitoring.EmailConfig).smtpHost" placeholder="smtp.example.com" />
                        <span class="ml-2 text-gray-500">服务器地址</span>
                    </ElFormItem>

                    <ElFormItem label="SMTP 端口" required>
                        <ElInputNumber v-model="(currentChannel.config as Monitoring.EmailConfig).smtpPort" :min="1" :max="65535" controls-position="right" />
                    </ElFormItem>

                    <ElFormItem label="用户名">
                        <ElInput v-model="(currentChannel.config as Monitoring.EmailConfig).username" placeholder="认证用户名" />
                    </ElFormItem>

                    <ElFormItem label="密码">
                        <ElInput v-model="(currentChannel.config as Monitoring.EmailConfig).password" type="password" placeholder="认证密码" show-password />
                    </ElFormItem>

                    <ElFormItem label="发件人邮箱" required>
                        <ElInput v-model="(currentChannel.config as Monitoring.EmailConfig).from" placeholder="noreply@example.com" />
                    </ElFormItem>

                    <ElFormItem label="发件人名称">
                        <ElInput v-model="(currentChannel.config as Monitoring.EmailConfig).fromName" placeholder="OneOps 监控" />
                    </ElFormItem>

                    <ElFormItem label="启用 TLS">
                        <ElSwitch v-model="(currentChannel.config as Monitoring.EmailConfig).useTLS" />
                        <span class="ml-2 text-gray-500">使用加密连接</span>
                    </ElFormItem>
                </template>

                <!-- 企业微信配置 -->
                <template v-if="currentChannel.channelType === 'wechat'">
                    <ElDivider content-position="left">企业微信机器人配置</ElDivider>

                    <ElFormItem label="Webhook URL" required>
                        <ElInput v-model="(currentChannel.config as Monitoring.WeChatConfig).webhookUrl" type="textarea" :rows="3" placeholder="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx" />
                    </ElFormItem>

                    <ElAlert title="提示" type="info" :closable="false" show-icon class="mb-4">
                        <template #default>
                            <div>在企业微信群聊中添加机器人，获取 Webhook 地址。</div>
                            <div>支持 @ 提醒指定人。</div>
                        </template>
                    </ElAlert>
                </template>

                <!-- 钉钉/飞书配置 -->
                <template v-if="currentChannel.channelType === 'dingtalk' || currentChannel.channelType === 'feishu'">
                    <ElDivider content-position="left">
                        {{ currentChannel.channelType === 'dingtalk' ? '钉钉' : '飞书' }} 机器人配置
                    </ElDivider>

                    <ElFormItem label="Webhook URL" required>
                        <ElInput v-model="(currentChannel.config as Monitoring.WeChatConfig).webhookUrl" type="textarea" :rows="3" placeholder="请输入 Webhook 地址" />
                    </ElFormItem>

                    <ElAlert title="提示" type="info" :closable="false" show-icon>
                        <template #default>
                            <div>在对应的平台群聊中添加自定义机器人，获取 Webhook 地址。</div>
                        </template>
                    </ElAlert>
                </template>

                <ElFormItem label="启用">
                    <ElSwitch v-model="currentChannel.enabled" />
                    <span class="ml-2 text-gray-500">启用后该渠道将接收告警通知</span>
                </ElFormItem>
            </ElForm>

            <template #footer>
                <ElButton @click="showChannelDialog = false">取消</ElButton>
                <ElButton type="primary" @click="handleSaveChannel">确定</ElButton>
            </template>
        </ElDialog>
    </div>
</template>

<style scoped>
    .time-picker {
      width: 100%;
      max-width: 350px;
    }
</style>
