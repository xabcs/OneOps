/**
 * 服务器表单逻辑（创建/编辑）
 * 重构：消除 defineExpose({ serverFormRef }) 反模式
 * - 表单 ref 由子组件内部 useForm() 管理
 * - 验证在子组件内完成
 * - 父组件通过 emit('submitted') 通知保存成功
 */

import { reactive, ref } from 'vue';
import { ElNotification } from 'element-plus';
import type { FormRules } from 'element-plus';
import { fetchAssignServerToGroups, fetchCreateServer, fetchUpdateServer, fetchAssignServerTag, fetchRemoveServerTag } from '@/service/api';

export function useServerForm() {
  // ===== 对话框状态 =====
  const dialogVisible = ref(false);
  const dialogTitle = ref('');
  const serverType = ref<'normal' | 'cloud'>('normal');
  const submitError = ref('');
  const activeCollapse = ref<string[]>([]);

  // 编辑抽屉
  const editDrawerVisible = ref(false);
  const editDrawerActiveTab = ref('basic');

  // 旧标签 ID 缓存（用于编辑时计算标签差异）
  const oldTagIdsCache: number[] = [];

  // 主机表单
  const serverForm = reactive<
    CMDB.ServerForm & { groupIds?: number[]; tagIds?: number[]; roomId?: number; cabinetId?: number }
  >({
    hostname: '',
    ip: '',
    innerIp: '',
    credentialIds: [],
    systemCredentialId: undefined,
    serverType: 'vm',
    groupIds: [],
    tagIds: [],
    roomId: undefined,
    cabinetId: undefined,
    sshPort: 22,
    remarks: '',
    cpu: 0,
    memory: 0,
    disk: 0,
    os: '',
    businessId: undefined as unknown as number
  });

  // 云主机表单
  const cloudForm = reactive<CMDB.CloudServerForm>({
    provider: 'aliyun',
    instanceName: '',
    instanceType: '',
    region: '',
    zone: '',
    chargeType: 'postpay'
  });

  // 表单验证规则（传递给子组件）
  const serverFormRules: FormRules = {
    hostname: [
      { required: true, message: '请输入主机名', trigger: 'blur' },
      { min: 2, max: 100, message: '主机名长度在 2 到 100 个字符', trigger: 'blur' },
      { pattern: /^[a-zA-Z0-9.-]+$/, message: '主机名只能包含字母、数字、点和连字符', trigger: 'blur' }
    ],
    ip: [
      { required: true, message: '请输入连接IP', trigger: 'blur' },
      { pattern: /^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$/, message: '请输入有效的IP地址', trigger: 'blur' }
    ],
    credentialIds: [{ required: true, type: 'array', min: 1, message: '请至少选择一个SSH凭证', trigger: 'change' }]
  };

  // ===== 辅助 =====
  function focusFormInput(selector: string) {
    setTimeout(() => {
      const input = document.querySelector(selector);
      if (input instanceof HTMLInputElement) input.focus();
    }, 100);
  }

  function handleSaveError(error: unknown) {
    const errObj = error && typeof error === 'object' ? (error as Record<string, unknown>) : {};
    const response = errObj.response as { data?: { code?: number; message?: string } } | undefined;
    const errorCode = response?.data?.code;
    const errorMessage = response?.data?.message || (typeof errObj.message === 'string' ? errObj.message : undefined);

    if (errorCode === 40001) {
      submitError.value = errorMessage || '主机名已存在，请使用其他主机名';
      focusFormInput('.hostname-input input');
      return;
    }
    if (errorCode === 40002) {
      submitError.value = errorMessage || 'IP地址已存在，请使用其他IP地址';
      focusFormInput('.ip-input input');
      return;
    }
    ElNotification.error(errorMessage || '操作失败，请稍后重试');
  }

  // ===== 打开对话框 =====
  function openAddDialog(selectedGroupId?: number) {
    dialogTitle.value = '创建主机';
    serverType.value = 'normal';
    submitError.value = '';
    activeCollapse.value = [];

    Object.assign(serverForm, {
      hostname: '',
      ip: '',
      innerIp: '',
      credentialIds: [],
      systemCredentialId: undefined,
      serverType: 'vm',
      groupIds: selectedGroupId ? [selectedGroupId] : [],
      tagIds: [],
      roomId: undefined,
      cabinetId: undefined,
      sshPort: 22,
      remarks: '',
      cpu: 0,
      memory: 0,
      disk: 0,
      os: '',
      businessId: undefined as unknown as number
    });

    Object.assign(cloudForm, {
      provider: 'aliyun',
      instanceId: '',
      instanceName: '',
      instanceType: '',
      region: '',
      zone: '',
      chargeType: 'postpay'
    });

    dialogVisible.value = true;
  }

  function openEditDrawer(row: CMDB.Server) {
    dialogTitle.value = '编辑主机';
    serverType.value = row.cloudInfo ? 'cloud' : 'normal';
    submitError.value = '';
    editDrawerActiveTab.value = 'basic';

    const groupIds = row.groups?.map(g => g.id) || [];
    const tagIds = row.tags?.map(t => t.id) || [];

    // 缓存旧标签 ID，用于保存时计算差异
    oldTagIdsCache.length = 0;
    oldTagIdsCache.push(...tagIds);

    Object.assign(serverForm, {
      id: row.id,
      hostname: row.hostname,
      ip: row.ip,
      innerIp: row.innerIp,
      credentialIds:
        row.credentials?.filter(c => c.credentialType === 'user').map(c => c.id) ||
        (row.sshCredentialId ? [row.sshCredentialId] : []),
      systemCredentialId: row.systemCredentialId || row.systemCredential?.id || undefined,
      serverType: row.serverType,
      groupIds,
      tagIds: row.tags?.map(t => t.id) || [],
      roomId: row.cabinet?.roomId,
      cabinetId: row.cabinetId,
      sshPort: row.sshPort,
      remarks: row.remarks,
      cpu: row.cpu || 0,
      memory: row.memory || 0,
      disk: row.disk || 0,
      os: row.os || '',
      businessId: row.businessId || (undefined as unknown as number)
    });

    if (row.cloudInfo) {
      Object.assign(cloudForm, {
        provider: row.provider as CMDB.CloudProvider,
        instanceId: row.cloudInfo.instanceId,
        instanceName: row.cloudInfo.instanceName,
        instanceType: row.cloudInfo.instanceType,
        region: row.cloudInfo.region,
        zone: row.cloudInfo.zone,
        chargeType: row.cloudInfo.chargeType
      });
    } else {
      Object.assign(cloudForm, {
        provider: 'aliyun',
        instanceId: '',
        instanceName: '',
        instanceType: '',
        region: '',
        zone: '',
        chargeType: 'postpay'
      });
    }

    editDrawerVisible.value = true;
  }

  // ===== 保存（不再调用 validate，验证由子组件完成）=====
  async function submitForm(saveAttributeFn: (serverId: number) => Promise<void>) {
    submitError.value = '';

    try {
      const groupIds = (serverForm.groupIds || []).map((id: number) => Number(id));
      const tagIds = (serverForm.tagIds || []).map((id: number) => Number(id));

      const formData: CMDB.ServerForm = {
        ...serverForm,
        serverType: serverType.value === 'cloud' ? 'vm' : serverForm.serverType || 'vm',
        cloudInfo:
          serverType.value === 'cloud'
            ? { ...cloudForm, provider: cloudForm.provider, publicIp: serverForm.ip, privateIp: serverForm.innerIp }
            : null
      };

      delete formData.tagIds;
      delete formData.groupIds;
      delete formData.roomId;

      let savedServerId: number;

      if (serverForm.id) {
        // 编辑：先获取当前标签，计算差异
        await fetchUpdateServer(serverForm.id, formData);
        savedServerId = serverForm.id;
        await saveAttributeFn(serverForm.id);

        // 同步标签
        await syncServerTags(serverForm.id, tagIds);

        ElNotification.success('更新成功');
      } else {
        const result = await fetchCreateServer(formData);
        const newServerId = result.data?.id;
        savedServerId = newServerId!;
        if (newServerId) await saveAttributeFn(newServerId);

        // 新增：分配标签
        for (const tagId of tagIds) {
          await fetchAssignServerTag(savedServerId, tagId);
        }

        ElNotification.success('创建成功');
      }

      if (groupIds.length > 0) {
        await fetchAssignServerToGroups(savedServerId, groupIds);
      }

      dialogVisible.value = false;
      editDrawerVisible.value = false;
      return true;
    } catch (error: unknown) {
      handleSaveError(error);
      return false;
    }
  }

  /** 同步主机标签：计算新旧差异，增量分配/移除 */
  async function syncServerTags(serverId: number, newTagIds: number[]) {
    // 获取当前标签列表（从 serverForm 编辑前的值，或从 API 获取）
    // 编辑时 openEditDrawer 已回填 tagIds，但用户可能修改了
    // 这里通过对比 row.tags 和新 tagIds 来计算差异
    // 由于当前作用域无法直接访问原 row，使用简单策略：先移除全部再重新分配
    // 但更高效的方式是传旧标签进来，这里采用直接 diff 的方式
    // 简化实现：直接全量同步
    // 注意：fetchRemoveServerTag 需要逐个调用
    // 此处依赖外部传入旧标签，如果没有则跳过移除逻辑
    // 实际实现：在 openEditDrawer 时保存旧标签 ID
    for (const tagId of newTagIds) {
      await fetchAssignServerTag(serverId, tagId);
    }
    // 移除不再选中的标签
    if (oldTagIdsCache.length > 0) {
      const toRemove = oldTagIdsCache.filter(id => !newTagIds.includes(id));
      for (const tagId of toRemove) {
        await fetchRemoveServerTag(serverId, tagId);
      }
    }
  }

  function resetSubmitError() {
    submitError.value = '';
  }

  return {
    dialogVisible,
    dialogTitle,
    serverType,
    submitError,
    activeCollapse,
    editDrawerVisible,
    editDrawerActiveTab,
    serverForm,
    cloudForm,
    serverFormRules,
    openAddDialog,
    openEditDrawer,
    submitForm,
    resetSubmitError
  };
}
