<script setup lang="ts">
  import { onMounted, ref } from 'vue';
  import { ElMessage, ElMessageBox } from 'element-plus';
  import { Plus } from '@element-plus/icons-vue';
  import { deleteAlertRule, fetchAlertRules, updateAlertRuleStatus } from '@/service/api';
  import { useAuthStore } from '@/store/modules/auth';
  import AlertRuleDialog from './modules/AlertRuleDialog.vue';

  defineOptions({
    name: 'MonitoringSettings'
  });

  const authStore = useAuthStore();
  const activeTab = ref('rules');

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
  // 告警规则管理（通知渠道已迁至：系统管理 → 通知渠道）
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

  // 页面加载时初始化
  onMounted(() => {
    loadAlertRules();
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

    <!-- 告警规则 -->
    <ElCard shadow="never">
      <div class="max-w-6xl">
        <div class="mb-4 flex items-center justify-between">
          <ElText type="info">配置触发告警的规则条件；通知渠道在「系统管理 → 通知渠道」维护</ElText>
          <PermissionButton code="monitor.alert_rule.create" type="primary" @click="handleCreateRule">
            <ElIcon :size="16">
              <Plus />
            </ElIcon>
            新增规则
          </PermissionButton>
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
              <ElSwitch
                :model-value="row.enabled"
                :disabled="!authStore.hasPermission('monitor.alert_rule.update')"
                title="缺少权限：monitor.alert_rule.update"
                @click="handleDisabledSwitchClick('monitor.alert_rule.update')"
                @change="handleToggleRuleStatus(row)"
              />
            </template>
          </ElTableColumn>
          <ElTableColumn label="操作" width="150" align="center" fixed="right">
            <template #default="{ row }">
              <PermissionButton
                code="monitor.alert_rule.update"
                type="primary"
                link
                size="small"
                @click="handleEditRule(row)"
              >
                编辑
              </PermissionButton>
              <PermissionButton
                code="monitor.alert_rule.delete"
                type="danger"
                link
                size="small"
                @click="handleDeleteRule(row)"
              >
                删除
              </PermissionButton>
            </template>
          </ElTableColumn>
        </ElTable>

        <ElEmpty v-if="!alertRulesLoading && alertRules.length === 0" description="暂无告警规则" />
      </div>
    </ElCard>

    <!-- 告警规则对话框 -->
    <AlertRuleDialog
      v-model:visible="showRuleDialog"
      :mode="ruleDialogMode"
      :rule="currentRule"
      @submitted="loadAlertRules"
    />
  </div>
</template>
