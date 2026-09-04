<script setup lang="ts">
  import { ref, watch } from 'vue';
  import { ElMessage, ElMessageBox } from 'element-plus';
  import { fetchAddGroupMembers, fetchGetGroupMembers, fetchRemoveGroupMember, fetchUserOptions } from '@/service/api';
  import { executeWithPermission } from '@/hooks/business/auth';

  defineOptions({ name: 'GroupMemberModal' });

  interface Props {
    visible: boolean;
    groupData: Api.SystemManage.UserGroup | null;
  }

  const props = defineProps<Props>();

  interface Emits {
    (e: 'update:visible', value: boolean): void;
    (e: 'submitted'): void;
  }

  const emit = defineEmits<Emits>();

  const loading = ref(false);
  const members = ref<Api.SystemManage.GroupMember[]>([]);
  const userOptions = ref<{ id: number; username: string; nickname: string }[]>([]);
  const selectedUserIds = ref<number[]>([]);
  const submitting = ref(false);

  async function loadMembers() {
    if (!props.groupData) return;
    loading.value = true;
    try {
      const { data, error } = await fetchGetGroupMembers(props.groupData.id);
      if (!error && data) {
        members.value = data || [];
      }
    } finally {
      loading.value = false;
    }
  }

  async function loadUserOptions() {
    const { data, error } = await fetchUserOptions();
    if (!error && data) {
      userOptions.value = data || [];
    }
  }

  function handleClose() {
    emit('update:visible', false);
  }

  async function handleAddMembers() {
    if (!props.groupData || selectedUserIds.value.length === 0) {
      ElMessage.warning('请选择要添加的用户');
      return;
    }
    await executeWithPermission('system.user.update', async () => {
      submitting.value = true;
      try {
        const { error } = await fetchAddGroupMembers(props.groupData!.id, selectedUserIds.value);
        if (!error) {
          ElMessage.success('添加成员成功');
          selectedUserIds.value = [];
          await loadMembers();
          emit('submitted');
        }
      } finally {
        submitting.value = false;
      }
    });
  }

  async function handleRemoveMember(row: Api.SystemManage.GroupMember) {
    if (!props.groupData) return;
    await ElMessageBox.confirm(`确定要将用户 "${row.username}" 移出该用户组吗？`, '提示', {
      type: 'warning'
    });
    await executeWithPermission('system.user.update', async () => {
      const { error } = await fetchRemoveGroupMember(props.groupData!.id, row.userId);
      if (!error) {
        ElMessage.success('移除成功');
        await loadMembers();
        emit('submitted');
      }
    });
  }

  watch(
    () => props.visible,
    val => {
      if (val) {
        selectedUserIds.value = [];
        loadMembers();
        loadUserOptions();
      }
    }
  );
</script>

<template>
  <ElDialog
    :model-value="props.visible"
    :title="`成员管理 - ${props.groupData?.name ?? ''}`"
    width="640px"
    @update:model-value="handleClose"
    @closed="handleClose"
  >
    <div class="mb-16px flex items-center gap-8px">
      <ElSelect
        v-model="selectedUserIds"
        multiple
        filterable
        clearable
        placeholder="选择要添加的用户"
        style="width: 420px"
      >
        <ElOption
          v-for="u in userOptions"
          :key="u.id"
          :value="u.id"
          :label="`${u.username}（${u.nickname || u.username}）`"
        />
      </ElSelect>
      <PermissionButton code="system.user.update" type="primary" :loading="submitting" @click="handleAddMembers">
        添加成员
      </PermissionButton>
    </div>
    <ElTable v-loading="loading" :data="members" stripe max-height="420">
      <ElTableColumn prop="userId" label="用户ID" width="80" />
      <ElTableColumn prop="username" label="用户名" />
      <ElTableColumn prop="nickname" label="昵称" />
      <ElTableColumn prop="createdAt" label="加入时间" width="170" />
      <ElTableColumn label="操作" width="100">
        <template #default="{ row }">
          <PermissionButton code="system.user.update" type="danger" size="small" link @click="handleRemoveMember(row)">
            移除
          </PermissionButton>
        </template>
      </ElTableColumn>
    </ElTable>
    <template #footer>
      <ElButton @click="handleClose">关闭</ElButton>
    </template>
  </ElDialog>
</template>

<style scoped></style>
