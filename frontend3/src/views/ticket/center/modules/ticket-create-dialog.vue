<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import { ArrowLeft } from '@element-plus/icons-vue';
  import { createTicket, fetchWorkflowsByType } from '@/service/api';
  import { useForm } from '@/hooks/common/form';
  import { ticketPriorityOptions } from '../../constants';

  defineOptions({ name: 'TicketCreateDialog' });

  interface Props {
    /** 可用的工单类型（场景）选项 */
    typeOptions: Api.Ticket.TicketTypeOption[];
  }

  const props = defineProps<Props>();

  interface Emits {
    (e: 'created'): void;
  }

  const emit = defineEmits<Emits>();

  const visible = defineModel<boolean>('visible', { default: false });

  // @ts-expect-error vue-tsc noUnusedLocals: template ref
  const { formRef, validate, restoreValidation } = useForm();

  const selectedType = ref<Api.Ticket.TicketTypeOption | null>(null);
  const workflowOptions = ref<Api.Ticket.WorkflowOption[]>([]);
  const workflowLoading = ref(false);
  const selectedWorkflowId = ref<number>(0);

  interface CreateModel {
    title: string;
    priority: Api.Ticket.TicketPriority;
    formData: Record<string, unknown>;
  }

  function createDefaultModel(): CreateModel {
    return { title: '', priority: 'normal', formData: {} };
  }

  const model = ref<CreateModel>(createDefaultModel());
  const submitLoading = ref(false);

  /** 解析所选类型的表单 schema */
  const formFields = computed<Api.Ticket.FormField[]>(() => {
    if (!selectedType.value?.formSchema) return [];
    try {
      const parsed = JSON.parse(selectedType.value.formSchema) as unknown;
      return Array.isArray(parsed) ? (parsed as Api.Ticket.FormField[]) : [];
    } catch {
      return [];
    }
  });

  /** 动态必填规则（prop 为 formData.xxx） */
  const dynamicRules = computed(() => {
    const rules: Record<string, App.Global.FormRule> = {
      title: { required: true, message: '请输入工单标题', trigger: 'blur' }
    };
    formFields.value.forEach(field => {
      if (field.required) {
        const action = field.type === 'select' || field.type === 'date' ? '请选择' : '请输入';
        rules[`formData.${field.key}`] = {
          required: true,
          message: `${action}${field.label}`,
          trigger: field.type === 'select' || field.type === 'date' ? 'change' : 'blur'
        };
      }
    });
    return rules;
  });

  async function selectType(type: Api.Ticket.TicketTypeOption) {
    selectedType.value = type;
    model.value = createDefaultModel();
    restoreValidation();

    // 加载该场景下可用的审批流程
    workflowLoading.value = true;
    selectedWorkflowId.value = 0;
    const { data, error } = await fetchWorkflowsByType(type.id);
    workflowLoading.value = false;
    if (!error && data) {
      workflowOptions.value = data;
      if (data.length > 0) {
        selectedWorkflowId.value = data[0]!.id;
      }
    }
  }

  function backToSelect() {
    selectedType.value = null;
    workflowOptions.value = [];
    selectedWorkflowId.value = 0;
  }

  function handleClosed() {
    backToSelect();
    model.value = createDefaultModel();
  }

  async function handleSubmit() {
    await validate();
    if (!selectedType.value || !selectedWorkflowId.value) return;

    submitLoading.value = true;
    const { data, error } = await createTicket({
      typeId: selectedType.value.id,
      workflowId: selectedWorkflowId.value,
      title: model.value.title,
      priority: model.value.priority,
      formData: model.value.formData
    });
    submitLoading.value = false;

    if (!error && data) {
      ElMessage.success(`工单 ${data.ticketNo} 提交成功`);
      visible.value = false;
      emit('created');
    }
  }

  watch(visible, val => {
    if (val) {
      restoreValidation();
    }
  });
</script>

<template>
  <ElDialog
    v-model="visible"
    :title="selectedType ? '发起工单' : '选择工单场景'"
    :width="680"
    destroy-on-close
    @closed="handleClosed"
  >
    <!-- 第一步：选择场景 -->
    <template v-if="!selectedType">
      <ElEmpty v-if="!typeOptions.length" description="暂无可用的工单场景" />
      <div v-else class="grid grid-cols-2 gap-12px lt-sm:grid-cols-1">
        <div
          v-for="t in typeOptions"
          :key="t.id"
          class="cursor-pointer border border-gray-200 rounded-6px p-14px transition hover:border-primary"
          @click="selectType(t)"
        >
          <div class="flex items-center gap-8px">
            <SvgIcon v-if="t.icon" :icon="t.icon" class="h-22px w-22px text-primary" />
            <span class="text-16px font-medium">{{ t.name }}</span>
          </div>
          <div class="mt-6px text-13px text-gray-400">{{ t.description || '暂无描述' }}</div>
        </div>
      </div>
    </template>

    <!-- 第二步：选择流程 + 填写表单 -->
    <template v-else>
      <div class="mb-16px flex items-center justify-between rounded-4px bg-gray-100 px-12px py-8px">
        <div class="flex items-center gap-8px">
          <SvgIcon v-if="selectedType.icon" :icon="selectedType.icon" class="h-18px w-18px text-primary" />
          <span class="font-medium">{{ selectedType.name }}</span>
          <span v-if="selectedType.description" class="text-12px text-gray-400">{{ selectedType.description }}</span>
        </div>
        <ElButton size="small" text type="primary" :icon="ArrowLeft" @click="backToSelect">重选场景</ElButton>
      </div>

      <ElForm ref="formRef" :model="model" :rules="dynamicRules" label-width="100px">
        <!-- 审批流程（场景下可选，卡片式选择，样式与场景选择一致） -->
        <ElFormItem label="审批流程" required>
          <div v-loading="workflowLoading" class="w-full">
            <ElAlert
              v-if="!workflowLoading && !workflowOptions.length"
              type="warning"
              :closable="false"
              title="该场景暂无启用的审批流程，请联系管理员配置"
            />
            <div v-else class="w-full flex flex-col gap-8px">
              <div
                v-for="wf in workflowOptions"
                :key="wf.id"
                class="cursor-pointer border rounded-6px px-12px py-8px transition"
                :class="
                  selectedWorkflowId === wf.id ? 'border-primary bg-primary-50' : 'border-gray-200 hover:border-primary'
                "
                @click="selectedWorkflowId = wf.id"
              >
                <div class="flex items-center gap-8px">
                  <span
                    class="inline-block h-10px w-10px shrink-0 border-2px rounded-full"
                    :class="selectedWorkflowId === wf.id ? 'border-primary' : 'border-gray-300'"
                  />
                  <span class="font-medium">{{ wf.name }}</span>
                  <ElTag size="small" type="info">{{ wf.nodeCount }} 级审批</ElTag>
                </div>
                <div v-if="wf.description" class="mt-4px pl-18px text-12px text-gray-400">
                  {{ wf.description }}
                </div>
              </div>
            </div>
          </div>
        </ElFormItem>

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

        <!-- 动态表单 -->
        <ElFormItem v-for="field in formFields" :key="field.key" :label="field.label" :prop="`formData.${field.key}`">
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
    </template>

    <template #footer>
      <template v-if="selectedType">
        <ElButton @click="visible = false">取消</ElButton>
        <ElButton type="primary" :loading="submitLoading" :disabled="!selectedWorkflowId" @click="handleSubmit">
          提交工单
        </ElButton>
      </template>
    </template>
  </ElDialog>
</template>
