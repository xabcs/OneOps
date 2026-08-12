<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import { assignUserToGroup, fetchAuthGroups } from '@/service/api/application-permission';
  import { useForm, useFormRules } from '@/hooks/common/form';

  defineOptions({ name: 'AuthUserAuthorizationOperateDrawer' });

  interface Props {
    /** selected user id */
    userId: number | null;
  }

  const props = defineProps<Props>();

  interface Emits {
    (e: 'submitted'): void;
    (e: 'showResults', results: Api.ApplicationPermission.AssignmentResultItem[]): void;
  }

  const emit = defineEmits<Emits>();

  // P1a: defineModel('visible')
  const visible = defineModel<boolean>('visible', { default: false });

  // P0a: drawer 内部 useForm()
  // @ts-expect-error vue-tsc noUnusedLocals: template ref
  const { formRef, validate, restoreValidation } = useForm();
  const { defaultRequiredRule } = useFormRules();

  // 加载用户组列表
  const groups = ref<Api.ApplicationPermission.AuthGroup[]>([]);

  async function loadGroups() {
    const { data, error } = await fetchAuthGroupOptions();
    if (!error && data) {
      groups.value = data || [];
    }
  }

  type Model = {
    groupId: number | null;
  };

  function createDefaultModel(): Model {
    return { groupId: null };
  }

  const model = ref<Model>(createDefaultModel());

  const title = computed(() => '分配用户组');

  function handleInitModel() {
    model.value = createDefaultModel();
  }

  function closeDrawer() {
    visible.value = false;
  }

  async function handleSubmit() {
    await validate();

    if (!props.userId || !model.value.groupId) {
      ElMessage.warning('请填写完整信息');
      return;
    }

    const { data, error } = await assignUserToGroup({
      userId: props.userId,
      groupId: model.value.groupId
    });

    if (!error && data) {
      closeDrawer();
      emit('submitted');

      // 显示授权结果（drawer 内部副作用：判断是否需要弹结果对话框）
      if (data.results && data.results.length > 0) {
        const failCount = data.results.filter(r => !r.success).length;
        if (failCount === 0) {
          ElMessage.success(`授权成功！已在 ${data.results.length} 个外部系统中授权`);
        } else {
          emit('showResults', data.results);
        }
      } else {
        ElMessage.success('分配用户组成功');
      }
    }
  }

  watch(visible, val => {
    if (val) {
      handleInitModel();
      restoreValidation();
      loadGroups();
    }
  });
</script>

<template>
  <ElDrawer v-model="visible" :title="title" :width="400">
    <ElForm ref="formRef" :model="model" label-width="80px">
      <ElFormItem label="用户组" prop="groupId" :rules="[defaultRequiredRule]">
        <ElSelect v-model="model.groupId" placeholder="请选择用户组" class="w-full">
          <ElOption v-for="group in groups" :key="group.id" :label="group.name" :value="group.id" />
        </ElSelect>
      </ElFormItem>
    </ElForm>

    <template #footer>
      <ElButton @click="closeDrawer">取消</ElButton>
      <ElButton type="primary" @click="handleSubmit">确定</ElButton>
    </template>
  </ElDrawer>
</template>
