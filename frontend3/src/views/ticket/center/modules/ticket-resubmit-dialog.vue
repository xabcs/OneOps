<script setup lang="ts">
  import { reactive, ref, watch } from 'vue';
  import type { FormInstance, FormItemRule, FormRules } from 'element-plus';
  import { resubmitTicket } from '@/service/api';
  import { ticketPriorityOptions } from '../../constants';

  defineOptions({ name: 'TicketResubmitDialog' });

  const props = defineProps<{
    /** 工单 ID */
    ticketId: number;
    /** 原标题（回填） */
    title: string;
    /** 原优先级（回填） */
    priority: Api.Ticket.TicketPriority;
    /** 类型表单 schema（动态渲染） */
    formSchema: Api.Ticket.FormField[];
    /** 原表单数据（回填） */
    formData: Record<string, unknown>;
    /** 驳回意见（提示用） */
    rejectComment?: string;
  }>();

  const emit = defineEmits<{ success: [] }>();

  const visible = defineModel<boolean>('visible', { default: false });

  const formRef = ref<FormInstance>();
  const submitLoading = ref(false);

  const model = reactive({
    title: '',
    priority: 'normal' as Api.Ticket.TicketPriority,
    formData: {} as Record<string, unknown>
  });

  /** 动态必填规则（select/date 用 change 触发） */
  const dynamicRules: Record<string, FormItemRule[]> = {};
  props.formSchema.forEach(f => {
    if (f.required) {
      const isPick = f.type === 'select' || f.type === 'date';
      dynamicRules[`formData.${f.key}`] = [
        {
          required: true,
          message: isPick ? `请选择${f.label}` : `请输入${f.label}`,
          trigger: isPick ? 'change' : 'blur'
        }
      ];
    }
  });
  const rules: FormRules = {
    title: [{ required: true, message: '请输入工单标题', trigger: 'blur' }],
    ...dynamicRules
  };

  // 打开时回填原数据（深拷贝，避免直接改动详情数据）
  watch(visible, v => {
    if (v) {
      model.title = props.title;
      model.priority = props.priority;
      model.formData = JSON.parse(JSON.stringify(props.formData ?? {}));
    }
  });

  async function handleSubmit() {
    await formRef.value?.validate();
    submitLoading.value = true;
    const { error } = await resubmitTicket(props.ticketId, {
      title: model.title,
      priority: model.priority,
      formData: model.formData
    });
    submitLoading.value = false;
    if (!error) {
      ElMessage.success('已重新提交');
      visible.value = false;
      emit('success');
    }
  }
</script>

<template>
  <ElDialog v-model="visible" title="修改后重新提交" width="560px" top="6vh" destroy-on-close>
    <ElAlert
      v-if="rejectComment"
      type="error"
      :closable="false"
      show-icon
      class="mb-12px"
      :title="`驳回意见：${rejectComment}`"
    />

    <ElForm ref="formRef" :model="model" :rules="rules" label-width="100px">
      <ElFormItem label="工单标题" prop="title">
        <ElInput v-model="model.title" placeholder="请简要描述申请内容" maxlength="100" show-word-limit />
      </ElFormItem>
      <ElFormItem label="优先级" prop="priority">
        <ElRadioGroup v-model="model.priority">
          <ElRadioButton v-for="opt in ticketPriorityOptions" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </ElRadioButton>
        </ElRadioGroup>
      </ElFormItem>

      <ElFormItem
        v-for="field in formSchema"
        :key="field.key"
        :label="field.label"
        :prop="`formData.${field.key}`"
      >
        <ElInput
          v-if="field.type === 'input'"
          v-model="model.formData[field.key] as string"
          :placeholder="field.placeholder || `请输入${field.label}`"
        />
        <ElInput
          v-else-if="field.type === 'textarea'"
          v-model="model.formData[field.key] as string"
          type="textarea"
          :rows="4"
          :placeholder="field.placeholder || `请输入${field.label}`"
        />
        <ElInputNumber
          v-else-if="field.type === 'number'"
          v-model="model.formData[field.key] as number"
          :placeholder="field.placeholder || `请输入${field.label}`"
          class="w-200px"
        />
        <ElSelect
          v-else-if="field.type === 'select'"
          v-model="model.formData[field.key] as string"
          :placeholder="field.placeholder || `请选择${field.label}`"
          clearable
          class="w-260px"
        >
          <ElOption v-for="opt in field.options || []" :key="opt" :label="opt" :value="opt" />
        </ElSelect>
        <ElDatePicker
          v-else-if="field.type === 'date'"
          v-model="model.formData[field.key] as string"
          type="datetime"
          value-format="YYYY-MM-DD HH:mm:ss"
          :placeholder="field.placeholder || `请选择${field.label}`"
          class="w-260px"
        />
      </ElFormItem>
    </ElForm>

    <template #footer>
      <ElButton @click="visible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitLoading" @click="handleSubmit">重新提交</ElButton>
    </template>
  </ElDialog>
</template>
