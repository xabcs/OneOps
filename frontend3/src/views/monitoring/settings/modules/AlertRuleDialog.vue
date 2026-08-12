<script setup lang="ts">
  import { ref, watch } from 'vue';
  import { ElMessage } from 'element-plus';
  import { createAlertRule, updateAlertRule } from '@/service/api';

  defineOptions({ name: 'AlertRuleDialog' });

  const props = defineProps<{
    mode: 'create' | 'edit';
    rule?: Monitoring.AlertRule | null;
  }>();

  const emit = defineEmits<{
    (e: 'submitted'): void;
  }>();

  const visible = defineModel<boolean>('visible', { default: false });

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

  const form = ref<Monitoring.AlertRuleForm>({
    name: '',
    level: 'high',
    metric: 'cpu_usage',
    condition: '>',
    threshold: 80,
    duration: 300,
    description: ''
  });

  const submitting = ref(false);

  // 同步 props 到表单
  watch(
    visible,
    val => {
      if (val) {
        if (props.mode === 'edit' && props.rule) {
          form.value = {
            id: props.rule.id,
            name: props.rule.name,
            level: props.rule.level,
            metric: props.rule.metric,
            condition: props.rule.condition,
            threshold: props.rule.threshold,
            duration: props.rule.duration,
            description: props.rule.description,
            enabled: props.rule.enabled
          };
        } else {
          form.value = {
            name: '',
            level: 'high',
            metric: 'cpu_usage',
            condition: '>',
            threshold: 80,
            duration: 300,
            description: ''
          };
        }
      }
    },
    { immediate: true }
  );

  async function handleSubmit() {
    if (!form.value.name) {
      ElMessage.warning('请输入规则名称');
      return;
    }
    if (form.value.threshold <= 0) {
      ElMessage.warning('阈值必须大于0');
      return;
    }

    submitting.value = true;
    try {
      if (props.mode === 'create') {
        const { error } = await createAlertRule(form.value);
        if (!error) {
          ElMessage.success('创建成功');
        } else {
          ElMessage.error('创建失败');
          return;
        }
      } else {
        const { error } = await updateAlertRule(form.value.id!, form.value);
        if (!error) {
          ElMessage.success('更新成功');
        } else {
          ElMessage.error('更新失败');
          return;
        }
      }
      visible.value = false;
      emit('submitted');
    } catch {
      ElMessage.error(props.mode === 'create' ? '创建失败' : '更新失败');
    } finally {
      submitting.value = false;
    }
  }
</script>

<template>
  <ElDialog
    v-model="visible"
    :title="mode === 'create' ? '新增告警规则' : '编辑告警规则'"
    width="600px"
    :close-on-click-modal="false"
  >
    <ElForm :model="form" label-width="120px">
      <ElFormItem label="规则名称" required>
        <ElInput v-model="form.name" placeholder="请输入规则名称" />
      </ElFormItem>

      <ElFormItem label="告警级别" required>
        <ElSelect v-model="form.level" placeholder="请选择告警级别">
          <ElOption v-for="level in alertLevelOptions" :key="level.value" :label="level.label" :value="level.value" />
        </ElSelect>
      </ElFormItem>

      <ElFormItem label="监控指标" required>
        <ElSelect v-model="form.metric" placeholder="请选择监控指标">
          <ElOption v-for="metric in metricOptions" :key="metric.value" :label="metric.label" :value="metric.value" />
        </ElSelect>
      </ElFormItem>

      <ElFormItem label="判断条件" required>
        <ElSelect v-model="form.condition" placeholder="请选择条件">
          <ElOption v-for="cond in conditionOptions" :key="cond.value" :label="cond.label" :value="cond.value" />
        </ElSelect>
      </ElFormItem>

      <ElFormItem label="阈值" required>
        <ElInputNumber v-model="form.threshold" :min="0" :max="100" :precision="2" controls-position="right" />
        <span class="ml-2 text-gray-500">
          {{ form.metric.includes('usage') ? '%' : '' }}
        </span>
      </ElFormItem>

      <ElFormItem label="持续时间" required>
        <ElInputNumber v-model="form.duration" :min="0" :max="86400" :step="60" controls-position="right" />
        <span class="ml-2 text-gray-500">秒（指标条件需持续多久才触发告警）</span>
      </ElFormItem>

      <ElFormItem label="描述">
        <ElInput v-model="form.description" type="textarea" :rows="3" placeholder="请输入规则描述" />
      </ElFormItem>
    </ElForm>

    <template #footer>
      <ElButton @click="visible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitting" @click="handleSubmit">确定</ElButton>
    </template>
  </ElDialog>
</template>
