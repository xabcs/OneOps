<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import { fetchUserOptions, reassignTicket } from '@/service/api';

  defineOptions({ name: 'TicketReassignDialog' });

  const props = defineProps<{
    /** 工单 ID */
    ticketId: number;
    /** 当前审批节点名（展示用） */
    nodeName: string;
    /** 原审批人名（展示用） */
    currentApprovers: string;
  }>();

  const emit = defineEmits<{ success: [] }>();

  const visible = defineModel<boolean>('visible', { default: false });

  const submitLoading = ref(false);
  const userOptions = ref<{ id: number; username: string; nickname: string }[]>([]);
  const selectedIds = ref<number[]>([]);

  const selectedNames = computed(() =>
    userOptions.value.filter(u => selectedIds.value.includes(u.id)).map(u => u.nickname || u.username)
  );

  // 打开时加载用户选项
  watch(
    visible,
    async v => {
      if (!v) return;
      selectedIds.value = [];
      if (userOptions.value.length > 0) return;
      const { data, error } = await fetchUserOptions();
      if (!error && data) userOptions.value = data;
    },
    { immediate: true }
  );

  async function handleSubmit() {
    if (selectedIds.value.length === 0) {
      ElMessage.warning('请选择新的审批人');
      return;
    }
    submitLoading.value = true;
    const { error } = await reassignTicket(props.ticketId, selectedIds.value);
    submitLoading.value = false;
    if (!error) {
      ElMessage.success('已改派');
      visible.value = false;
      emit('success');
    }
  }
</script>

<template>
  <ElDialog v-model="visible" title="改派审批人" width="520px" destroy-on-close>
    <ElAlert
      type="warning"
      :closable="false"
      show-icon
      class="mb-12px"
      :title="`当前节点「${nodeName}」审批人：${currentApprovers || '-'}`"
      description="用于审批人离职/请假导致节点无法推进的场景；已产生的审批记录将保留。"
    />

    <div class="mb-8px text-14px">选择新审批人（可多选，会签节点需选多人）</div>
    <ElSelect
      v-model="selectedIds"
      multiple
      filterable
      placeholder="搜索并选择用户"
      class="w-full"
      :loading="userOptions.length === 0"
    >
      <ElOption
        v-for="u in userOptions"
        :key="u.id"
        :label="`${u.nickname || u.username}（${u.username}）`"
        :value="u.id"
      />
    </ElSelect>

    <div v-if="selectedNames.length > 0" class="mt-8px text-12px text-gray-400">
      将改派为：{{ selectedNames.join('、') }}
    </div>

    <template #footer>
      <ElButton @click="visible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitLoading" :disabled="selectedIds.length === 0" @click="handleSubmit">
        确认改派
      </ElButton>
    </template>
  </ElDialog>
</template>
