<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import { ArrowDown, ArrowUp, Delete, Plus } from '@element-plus/icons-vue';
  import { createWorkflow, fetchRoleOptions, fetchTicketTypeOptions, fetchUserOptions, fetchWorkflowDetail, updateWorkflow } from '@/service/api';
  import { useForm, useFormRules } from '@/hooks/common/form';

  defineOptions({ name: 'WorkflowOperateDialog' });

  interface Props {
    /** operation type: add | edit */
    operateType: UI.TableOperateType;
    /** editing row data */
    rowData?: Api.Ticket.Workflow | null;
  }

  const props = defineProps<Props>();

  interface Emits {
    (e: 'submitted'): void;
  }

  const emit = defineEmits<Emits>();

  const visible = defineModel<boolean>('visible', { default: false });

  // @ts-expect-error vue-tsc noUnusedLocals: template ref
  const { formRef, validate, restoreValidation } = useForm();
  const { defaultRequiredRule } = useFormRules();

  /** 设计器节点模型（编辑态） */
  interface NodeModel {
    name: string;
    approverType: Api.Ticket.ApproverType;
    approverIds: number[];
    multiType: Api.Ticket.MultiType;
    conditions: Api.Ticket.ConditionItem[];
    conditionEnabled: boolean;
    /** 审批超时阈值（小时），0=不启用；超时提醒审批人，2倍阈值升级管理员 */
    timeoutHours: number;
  }

  interface BaseModel {
    typeId?: number;
    name: string;
    code: string;
    description: string;
    status: number;
  }

  const baseModel = ref<BaseModel>({ typeId: undefined, name: '', code: '', description: '', status: 1 });
  const nodes = ref<NodeModel[]>([]);

  const rules: Record<'typeId' | 'name' | 'code', App.Global.FormRule> = {
    typeId: { required: true, message: '请选择所属场景', trigger: 'change', type: 'number' },
    name: defaultRequiredRule,
    code: defaultRequiredRule
  };

  // 审批人选项 + 工单场景选项
  const userOptions = ref<{ id: number; username: string; nickname: string }[]>([]);
  const roleOptions = ref<{ id: number; name: string; code: string }[]>([]);
  const typeOptions = ref<Api.Ticket.TicketTypeOption[]>([]);

  async function loadOptions() {
    const [userRes, roleRes, typeRes] = await Promise.all([
      fetchUserOptions(),
      fetchRoleOptions(),
      fetchTicketTypeOptions()
    ]);
    if (!userRes.error && userRes.data) userOptions.value = userRes.data;
    if (!roleRes.error && roleRes.data) roleOptions.value = roleRes.data;
    if (!typeRes.error && typeRes.data) typeOptions.value = typeRes.data;
  }

  const approverTypeOptions: { label: string; value: Api.Ticket.ApproverType }[] = [
    { label: '指定用户', value: 'user' },
    { label: '指定角色', value: 'role' },
    { label: '发起人', value: 'initiator' }
  ];

  const multiTypeOptions: { label: string; value: Api.Ticket.MultiType }[] = [
    { label: '或签（任一人通过）', value: 'any' },
    { label: '会签（全部通过）', value: 'all' }
  ];

  const opOptions = [
    { label: '等于', value: 'eq' },
    { label: '不等于', value: 'ne' },
    { label: '包含于', value: 'in' },
    { label: '大于', value: 'gt' },
    { label: '小于', value: 'lt' }
  ];

  function createNode(): NodeModel {
    return {
      name: '',
      approverType: 'user',
      approverIds: [],
      multiType: 'any',
      conditions: [],
      conditionEnabled: false,
      timeoutHours: 0
    };
  }

  function handleInitModel() {
    baseModel.value = { name: '', code: '', description: '', status: 1 };
    nodes.value = [createNode()];
    loadOptions();

    if (props.operateType === 'edit' && props.rowData) {
      baseModel.value = {
        typeId: props.rowData.typeId,
        name: props.rowData.name,
        code: props.rowData.code,
        description: props.rowData.description,
        status: props.rowData.status
      };
      // 编辑时拉取含节点的详情
      fetchWorkflowDetail(props.rowData.id).then(({ data, error }) => {
        if (!error && data?.nodes?.length) {
          nodes.value = data.nodes
            .slice()
            .sort((a, b) => a.sortOrder - b.sortOrder)
            .map(node => {
              let conditions: Api.Ticket.ConditionItem[] = [];
              try {
                const parsed = node.condition ? JSON.parse(node.condition) : [];
                conditions = Array.isArray(parsed) ? parsed : [];
              } catch {
                conditions = [];
              }
              return {
                name: node.name,
                approverType: node.approverType,
                approverIds: node.approverIds
                  ? node.approverIds
                      .split(',')
                      .filter(Boolean)
                      .map(Number)
                  : [],
                multiType: (node.multiType || 'any') as Api.Ticket.MultiType,
                conditions,
                conditionEnabled: conditions.length > 0,
                timeoutHours: node.timeoutHours || 0
              };
            });
        }
      });
    }
  }

  function moveNode(index: number, dir: -1 | 1) {
    const target = index + dir;
    if (target < 0 || target >= nodes.value.length) return;
    const list = nodes.value;
    [list[index], list[target]] = [list[target], list[index]];
  }

  function removeNode(index: number) {
    if (nodes.value.length <= 1) {
      ElMessage.warning('至少保留一个审批节点');
      return;
    }
    nodes.value.splice(index, 1);
  }

  function addNode() {
    nodes.value.push(createNode());
  }

  function addCondition(node: NodeModel) {
    node.conditions.push({ field: 'priority', op: 'eq', value: '' });
  }

  /** 校验节点数据完整性 */
  function validateNodes(): string | null {
    for (let i = 0; i < nodes.value.length; i += 1) {
      const node = nodes.value[i];
      if (!node.name.trim()) return `第 ${i + 1} 个节点未填写名称`;
      if (node.approverType !== 'initiator' && node.approverIds.length === 0) {
        return `节点「${node.name}」未选择审批人`;
      }
      if (node.conditionEnabled) {
        for (const cond of node.conditions) {
          if (!cond.field || !cond.value) return `节点「${node.name}」的条件配置不完整`;
        }
      }
    }
    return null;
  }

  const submitLoading = ref(false);

  async function handleSubmit() {
    await validate();

    const nodeError = validateNodes();
    if (nodeError) {
      ElMessage.warning(nodeError);
      return;
    }

    const payload: Api.Ticket.WorkflowSaveRequest = {
      ...baseModel.value,
      typeId: baseModel.value.typeId!,
      nodes: nodes.value.map(node => ({
        name: node.name.trim(),
        approverType: node.approverType,
        approverIds: node.approverType === 'initiator' ? [] : node.approverIds,
        multiType: node.approverType === 'initiator' ? 'any' : node.multiType,
        condition: node.conditionEnabled ? node.conditions : [],
        timeoutHours: node.approverType === 'initiator' ? 0 : node.timeoutHours || 0
      }))
    };

    submitLoading.value = true;
    const { error } =
      props.operateType === 'edit'
        ? await updateWorkflow(props.rowData!.id, payload)
        : await createWorkflow(payload);
    submitLoading.value = false;

    if (!error) {
      ElMessage.success(props.operateType === 'edit' ? '更新成功' : '创建成功');
      visible.value = false;
      emit('submitted');
    }
  }

  const title = computed(() => (props.operateType === 'add' ? '新建流程' : '编辑流程'));

  watch(visible, val => {
    if (val) {
      handleInitModel();
      restoreValidation();
    }
  });
</script>

<template>
  <ElDialog v-model="visible" :title="title" :width="720" top="6vh" destroy-on-close>
    <ElForm ref="formRef" :model="baseModel" :rules="rules" label-width="90px">
      <ElFormItem label="工单类型" prop="typeId">
        <ElSelect v-model="baseModel.typeId" placeholder="选择该流程服务的工单场景" class="w-260px">
          <ElOption v-for="t in typeOptions" :key="t.id" :label="t.name" :value="t.id" />
        </ElSelect>
        <div class="w-full text-12px text-gray-400">同一场景可配置多个流程，发起工单时选用</div>
      </ElFormItem>
      <ElFormItem label="流程名称" prop="name">
        <ElInput v-model="baseModel.name" placeholder="如：SQL 审批流" maxlength="50" />
      </ElFormItem>
      <ElFormItem label="流程编码" prop="code">
        <ElInput
          v-model="baseModel.code"
          placeholder="如：sql-audit"
          :disabled="operateType === 'edit'"
          maxlength="50"
        />
      </ElFormItem>
      <ElFormItem label="描述" prop="description">
        <ElInput v-model="baseModel.description" type="textarea" :rows="2" placeholder="流程用途说明" />
      </ElFormItem>
      <ElFormItem label="状态" prop="status">
        <ElRadioGroup v-model="baseModel.status">
          <ElRadio :value="1">启用</ElRadio>
          <ElRadio :value="0">禁用</ElRadio>
        </ElRadioGroup>
      </ElFormItem>
    </ElForm>

    <ElDivider content-position="left">审批节点（按顺序流转）</ElDivider>

    <div class="flex flex-col gap-12px">
      <div
        v-for="(node, index) in nodes"
        :key="index"
        class="rounded-6px border border-gray-200 p-12px"
      >
        <div class="mb-8px flex items-center justify-between">
          <ElTag size="small" type="primary">节点 {{ index + 1 }}</ElTag>
          <div class="flex gap-4px">
            <ElButton size="small" text :icon="ArrowUp" :disabled="index === 0" @click="moveNode(index, -1)" />
            <ElButton
              size="small"
              text
              :icon="ArrowDown"
              :disabled="index === nodes.length - 1"
              @click="moveNode(index, 1)"
            />
            <ElButton size="small" text type="danger" :icon="Delete" @click="removeNode(index)" />
          </div>
        </div>

        <ElForm label-width="90px" label-position="left">
          <ElFormItem label="节点名称" required>
            <ElInput v-model="node.name" placeholder="如：DBA 审核" maxlength="50" class="w-260px" />
          </ElFormItem>
          <ElFormItem label="审批人类型">
            <ElSelect v-model="node.approverType" class="w-200px">
              <ElOption v-for="opt in approverTypeOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem v-if="node.approverType !== 'initiator'" label="审批人" required>
            <ElSelect
              v-model="node.approverIds"
              multiple
              filterable
              clearable
              :placeholder="node.approverType === 'user' ? '选择一个或多个用户' : '选择一个或多个角色'"
              class="w-full"
            >
              <template v-if="node.approverType === 'user'">
                <ElOption
                  v-for="u in userOptions"
                  :key="u.id"
                  :label="`${u.nickname}(${u.username})`"
                  :value="u.id"
                />
              </template>
              <template v-else>
                <ElOption v-for="r in roleOptions" :key="r.id" :label="r.name" :value="r.id" />
              </template>
            </ElSelect>
          </ElFormItem>
          <ElFormItem v-if="node.approverType !== 'initiator' && node.approverIds.length > 1" label="审批方式">
            <ElRadioGroup v-model="node.multiType">
              <ElRadio v-for="opt in multiTypeOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</ElRadio>
            </ElRadioGroup>
          </ElFormItem>

          <ElFormItem v-if="node.approverType !== 'initiator'" label="审批超时">
            <div class="w-full">
              <ElInputNumber
                v-model="node.timeoutHours"
                :min="0"
                :max="8760"
                :step="1"
                controls-position="right"
                class="w-160px"
              />
              <span class="ml-8px text-12px text-gray-400">
                小时（0=不启用）；超时自动提醒审批人，超过 2 倍阈值升级通知管理员
              </span>
            </div>
          </ElFormItem>

          <ElFormItem label="激活条件">
            <div class="w-full">
              <ElSwitch v-model="node.conditionEnabled" active-text="按表单字段条件进入该节点" />
              <div v-if="node.conditionEnabled" class="mt-8px flex flex-col gap-8px">
                <div v-for="(cond, ci) in node.conditions" :key="ci" class="flex items-center gap-8px">
                  <ElInput v-model="cond.field" placeholder="字段名 如 priority" class="w-160px" />
                  <ElSelect v-model="cond.op" class="w-100px">
                    <ElOption v-for="op in opOptions" :key="op.value" :label="op.label" :value="op.value" />
                  </ElSelect>
                  <ElInput v-model="cond.value" placeholder="值 如 urgent" class="flex-1" />
                  <ElButton
                    size="small"
                    text
                    type="danger"
                    :icon="Delete"
                    @click="node.conditions.splice(ci, 1)"
                  />
                </div>
                <ElButton size="small" text type="primary" :icon="Plus" @click="addCondition(node)">
                  添加条件
                </ElButton>
              </div>
            </div>
          </ElFormItem>
        </ElForm>
      </div>

      <ElButton type="primary" plain :icon="Plus" @click="addNode">添加审批节点</ElButton>
    </div>

    <template #footer>
      <ElButton @click="visible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitLoading" @click="handleSubmit">保存</ElButton>
    </template>
  </ElDialog>
</template>
