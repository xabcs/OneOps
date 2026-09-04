{"code":200,"success":true,"message":"扩缩容 Deployment 成功"}
<script setup lang="ts">
  import { computed, onMounted, ref, watch } from 'vue';
  import {
    ElAlert,
    ElButton,
    ElDivider,
    ElDrawer,
    ElInput,
    ElMessage,
    ElMessageBox,
    ElOption,
    ElSelect,
    ElTabPane,
    ElTable,
    ElTableColumn,
    ElTabs,
    ElTag,
    ElText
  } from 'element-plus';
  import { Plus, Refresh, Search } from '@element-plus/icons-vue';
  import {
    assignK8sNativeBinding,
    fetchK8sAllNativeBindings,
    fetchK8sClusterNamespaces,
    fetchK8sClusterOptions,
    fetchK8sRoleOptions,
    fetchK8sSubjectOptions,
    revokeK8sNativeBinding
  } from '@/service/api/k8s';
  import { useAuthStore } from '@/store/modules/auth';
  import { executeWithPermission } from '@/hooks/business/auth';

  defineOptions({ name: 'K8sSecurityAuthorization' });

  const authStore = useAuthStore();
  /** 按能力显隐：无授权/撤销权限的用户不渲染对应入口，也不触发其数据请求 */
  const canAssign = computed(() => authStore.hasPermission('k8s.permission.assign'));
  const canRevoke = computed(() => authStore.hasPermission('k8s.permission.revoke'));

  /** 授权主体（OneOps 平台侧）：用户 / 用户组（平台角色是权限模板，非 K8s 授权主体） */
  interface SubjectRow {
    type: 'user' | 'group';
    id: number;
    label: string;
    sub: string;
    /** 用户所属组（用户视角叠加"经组继承"的有效授权） */
    groupIds?: number[];
  }

  /** 预设档语义：roleName 为集群内真实角色，权限完全由集群内 RBAC 决定 */
  interface PresetOption {
    value: string;
    label: string;
    roleKind: 'ClusterRole' | 'Role';
    roleName: string;
    /** 是否支持"所选命名空间"范围（不选 = 全集群） */
    nsOptional: boolean;
    desc: string;
  }

  const presetOptions: PresetOption[] = [
    {
      value: 'admin',
      label: '管理员',
      roleKind: 'ClusterRole',
      roleName: 'oneops-admin',
      nsOptional: true,
      desc: '对集群所有命名空间资源、节点、存储卷、命名空间、配额的读写权限；限定命名空间时对该命名空间内全资源读写（OneOps 预设，首次授权自动创建）'
    },
    {
      value: 'readonly-admin',
      label: '只读管理员',
      roleKind: 'ClusterRole',
      roleName: 'oneops-readonly',
      nsOptional: true,
      desc: '聚合角色：内置 view 全部只读范围 + 节点/存储卷/存储类只读，随集群版本与 CRD 自动扩展；不含 secrets（OneOps 预设，首次授权自动创建）'
    },
    {
      value: 'ops',
      label: '运维人员',
      roleKind: 'ClusterRole',
      roleName: 'oneops-ops',
      nsOptional: true,
      desc: '控制台可见资源读写 + 节点/存储卷/命名空间/配额的读取与更新 + 其他资源只读；限定命名空间时不含节点/存储卷更新（OneOps 预设，首次授权自动创建）'
    },
    {
      value: 'developer',
      label: '开发人员',
      roleKind: 'ClusterRole',
      roleName: 'oneops-dev',
      nsOptional: true,
      desc: '对控制台可见资源（工作负载/网络/配置等）的读写权限，可选限定命名空间；不含 serviceaccounts 与 RBAC 资源（OneOps 预设，首次授权自动创建）'
    },
    {
      value: 'viewer',
      label: '受限用户',
      roleKind: 'ClusterRole',
      roleName: 'oneops-view',
      nsOptional: true,
      desc: '对控制台可见资源的只读权限（含日志查看），可选限定命名空间（OneOps 预设，首次授权自动创建）'
    }
  ];

  type TabKey = 'users' | 'groups' | 'records';
  const activeTab = ref<TabKey>('users');

  // ========== 主体与绑定数据（页面加载即取，k8s.permission.list 即可） ==========
  interface UserOption {
    id: number;
    username: string;
    nickname: string;
    groupIds?: number[];
  }
  const users = ref<UserOption[]>([]);
  const groups = ref<{ id: number; code: string; name: string }[]>([]);
  const bindings = ref<K8s.NativeRoleBinding[]>([]);
  const loading = ref(false);

  const userSearch = ref('');
  const groupSearch = ref('');
  const clusterFilter = ref<number>();
  const clusterOptions = ref<{ id: number; name: string }[]>([]);
  /** 集群候选懒加载标记：仅打开授权弹窗或切到授权记录 tab 时拉取 */
  const clusterOptionsLoaded = ref(false);

  /** groupId → 组名（继承授权来源标注） */
  const groupNameById = computed(() => {
    const map = new Map<number, string>();
    groups.value.forEach(g => map.set(g.id, g.name));
    return map;
  });

  const filteredUsers = computed(() => {
    const kw = userSearch.value.trim().toLowerCase();
    if (!kw) return users.value;
    return users.value.filter(
      u => u.username.toLowerCase().includes(kw) || (u.nickname || '').toLowerCase().includes(kw)
    );
  });

  const filteredGroups = computed(() => {
    const kw = groupSearch.value.trim().toLowerCase();
    if (!kw) return groups.value;
    return groups.value.filter(g => g.name.toLowerCase().includes(kw) || (g.code || '').toLowerCase().includes(kw));
  });

  /** 授权记录 tab：按集群筛选（空 = 全部） */
  const filteredRecords = computed(() => {
    if (!clusterFilter.value) return bindings.value;
    return bindings.value.filter(b => b.clusterId === clusterFilter.value);
  });

  const directUserBindings = (userId: number) =>
    bindings.value.filter(b => b.subjectType === 'user' && b.userId === userId);
  const groupBindings = (groupId: number) =>
    bindings.value.filter(b => b.subjectType === 'group' && b.groupId === groupId);
  /** 用户有效授权 = 直绑 + 所属组的绑定（继承） */
  const inheritedUserBindings = (u: UserOption) => (u.groupIds || []).flatMap(gid => groupBindings(gid));
  const userPermSummary = (u: UserOption) => {
    const direct = directUserBindings(u.id).length;
    const inherited = inheritedUserBindings(u).length;
    return { direct, inherited, total: direct + inherited };
  };
  const groupPermCount = (groupId: number) => groupBindings(groupId).length;

  const loadPageData = async () => {
    loading.value = true;
    try {
      // subject-options 控制器不使用集群 id 参数（全局主体候选），传 0 占位
      const [subjectRes, bindingRes] = await Promise.all([fetchK8sSubjectOptions(0), fetchK8sAllNativeBindings()]);
      if (!subjectRes.error && subjectRes.data) {
        users.value = subjectRes.data.users || [];
        groups.value = subjectRes.data.groups || [];
      }
      if (!bindingRes.error && bindingRes.data) {
        bindings.value = bindingRes.data || [];
      }
    } finally {
      loading.value = false;
    }
  };

  // ========== 权限管理抽屉（单主体的全部有效授权：直接 + 组继承，可逐条撤销） ==========
  const permsVisible = ref(false);
  const permsSubject = ref<SubjectRow | null>(null);

  /** 抽屉行：inherited = 经组继承（撤销入口归用户组管理，此处只读） */
  interface PermRow {
    binding: K8s.NativeRoleBinding;
    inherited: boolean;
    sourceLabel: string;
  }

  const permRows = computed<PermRow[]>(() => {
    const subject = permsSubject.value;
    if (!subject) return [];
    const rows: PermRow[] = [];
    if (subject.type === 'user') {
      directUserBindings(subject.id).forEach(b => rows.push({ binding: b, inherited: false, sourceLabel: '直接' }));
      (subject.groupIds || []).forEach(gid => {
        groupBindings(gid).forEach(b =>
          rows.push({ binding: b, inherited: true, sourceLabel: `继承：${groupNameById.value.get(gid) || gid}` })
        );
      });
    } else {
      groupBindings(subject.id).forEach(b => rows.push({ binding: b, inherited: false, sourceLabel: '直接' }));
    }
    return rows;
  });

  const openPerms = async (subject: SubjectRow) => {
    permsSubject.value = subject;
    permsVisible.value = true;
    pendingRows.value = [];
    await ensureClusterOptions();
  };

  const handleRevoke = async (row: K8s.NativeRoleBinding) => {
    try {
      await ElMessageBox.confirm(
        `确定要撤销 "${row.subjectName}" 在集群「${row.clusterName || row.clusterId}」的访问授权（${row.roleKind}/${
          row.roleName
        }）吗？集群内对应 Binding 将一并删除。`,
        '确认',
        { type: 'warning' }
      );
      await executeWithPermission('k8s.permission.revoke', async () => {
        const { error } = await revokeK8sNativeBinding(row.clusterId!, row.id);
        if (!error) {
          ElMessage.success('撤销成功');
          await loadPageData();
        }
      });
    } catch (error: unknown) {
      if (error !== 'cancel') {
        const err = error as Error;
        ElMessage.error(err.message || '撤销访问授权失败');
      }
    }
  };

  // ========== 添加权限：行式待提交清单（集群/命名空间/权限/操作，提交授权统一落库） ==========
  const submitting = ref(false);

  /** 集群候选懒加载：打开抽屉或切到授权记录 tab 时拉取 */
  const ensureClusterOptions = async () => {
    if (clusterOptionsLoaded.value) return;
    const { data, error } = await fetchK8sClusterOptions();
    if (!error && data) clusterOptions.value = data || [];
    clusterOptionsLoaded.value = true;
  };

  watch(activeTab, tab => {
    if (tab === 'records') ensureClusterOptions();
  });

  /** 待提交行：clusterId=0 表示所有集群（提交时展开为逐集群授权） */
  interface PendingPermRow {
    clusterId: number | undefined;
    namespaces: string[];
    nsOptions: string[];
    /** 预设档 value 或 'custom' */
    permission: string;
    roleKind: 'ClusterRole' | 'Role';
    roleName: string;
    roleOptions: string[];
  }

  const pendingRows = ref<PendingPermRow[]>([]);

  const addPendingRow = () =>
    pendingRows.value.push({
      clusterId: undefined,
      namespaces: [],
      nsOptions: [],
      permission: 'admin',
      roleKind: 'ClusterRole',
      roleName: '',
      roleOptions: []
    });

  const removePendingRow = (index: number) => pendingRows.value.splice(index, 1);

  /** 命名空间选择框始终渲染、仅切换禁用态，避免"文本↔选择框"切换造成的布局移动 */
  const rowNsDisabled = (row: PendingPermRow) =>
    row.clusterId === undefined ||
    row.clusterId === 0 ||
    (row.permission === 'custom' && row.roleKind === 'ClusterRole');

  const rowNsPlaceholder = (row: PendingPermRow) => {
    if (row.clusterId === undefined) return '请先选择集群';
    if (row.clusterId === 0) return '全部命名空间（集群级授权）';
    if (row.permission === 'custom' && row.roleKind === 'ClusterRole') return '集群级角色，无需命名空间';
    if (row.permission === 'custom' && row.roleKind === 'Role') return '必选（命名空间级角色）';
    return '不选 = 全部命名空间';
  };

  const loadRowRoleOptions = async (row: PendingPermRow) => {
    if (!row.clusterId || row.clusterId <= 0) {
      row.roleOptions = [];
      return;
    }
    const { data, error } = await fetchK8sRoleOptions(row.clusterId, row.roleKind);
    if (!error && data) row.roleOptions = data;
  };

  /** 行内集群变化：重置范围/角色，按需加载命名空间与角色候选 */
  const onRowClusterChange = async (row: PendingPermRow) => {
    row.namespaces = [];
    row.nsOptions = [];
    row.roleName = '';
    row.roleOptions = [];
    if (row.permission === 'custom' && row.clusterId === 0) row.permission = 'admin';
    if (!row.clusterId || row.clusterId <= 0) return;
    const nsRes = await fetchK8sClusterNamespaces(row.clusterId);
    if (!nsRes.error && nsRes.data) row.nsOptions = (nsRes.data || []).map(ns => ns.name);
    if (row.permission === 'custom') await loadRowRoleOptions(row);
  };

  const onRowPermissionChange = (row: PendingPermRow) => {
    row.namespaces = [];
    row.roleName = '';
    if (row.clusterId && row.clusterId > 0 && row.permission === 'custom') loadRowRoleOptions(row);
  };

  const onRowRoleKindChange = (row: PendingPermRow) => {
    row.roleName = '';
    row.namespaces = [];
    if (row.clusterId && row.clusterId > 0) loadRowRoleOptions(row);
  };

  /** 提交授权：校验后逐行（所有集群行再逐集群展开）创建 Binding，各条独立成败 */
  const handleSubmitPerms = async () => {
    if (!permsSubject.value) return;
    if (pendingRows.value.length === 0) {
      ElMessage.warning('请先添加权限');
      return;
    }
    for (const [i, row] of pendingRows.value.entries()) {
      const no = i + 1;
      if (row.clusterId === undefined) {
        ElMessage.warning(`第 ${no} 行：请选择集群`);
        return;
      }
      if (row.clusterId === 0 && row.permission === 'custom') {
        ElMessage.warning(`第 ${no} 行：所有集群不支持自定义权限`);
        return;
      }
      if (row.permission === 'custom' && !row.roleName) {
        ElMessage.warning(`第 ${no} 行：请选择集群内角色`);
        return;
      }
      if (row.permission === 'custom' && row.roleKind === 'Role' && row.namespaces.length === 0) {
        ElMessage.warning(`第 ${no} 行：命名空间级角色（Role）必须指定命名空间`);
        return;
      }
    }

    const subject = permsSubject.value;
    await executeWithPermission('k8s.permission.assign', async () => {
      submitting.value = true;
      try {
        let total = 0;
        let ok = 0;
        const failed: string[] = [];
        for (const row of pendingRows.value) {
          const preset = presetOptions.find(p => p.value === row.permission);
          const kind = row.permission === 'custom' ? row.roleKind : preset!.roleKind;
          const roleName = row.permission === 'custom' ? row.roleName : preset!.roleName;
          const nsList = row.clusterId === 0 ? [] : row.namespaces;
          const targets = row.clusterId === 0 ? clusterOptions.value.map(c => c.id) : [row.clusterId!];
          for (const cid of targets) {
            total++;
            const { error } = await assignK8sNativeBinding(cid, {
              subjectType: subject.type,
              userId: subject.type === 'user' ? subject.id : undefined,
              groupId: subject.type === 'group' ? subject.id : undefined,
              roleKind: kind,
              roleName,
              namespaces: nsList
            });
            if (error) failed.push(clusterOptions.value.find(c => c.id === cid)?.name || String(cid));
            else ok++;
          }
        }

        if (failed.length === 0) {
          ElMessage.success(`提交授权成功：共创建 ${ok} 条集群内 Binding，由 kube-apiserver 判定`);
        } else {
          ElMessage.warning(
            `部分授权失败（${failed.length}/${total}）：${failed.join('、')}；成功部分已生效，可重试失败项`
          );
        }
        if (ok > 0) {
          pendingRows.value = [];
          await loadPageData();
          // 全部成功即关闭抽屉；部分失败保留清单供修正重试
          if (failed.length === 0) permsVisible.value = false;
        }
      } finally {
        submitting.value = false;
      }
    });
  };

  // ========== 权限说明：档位 → 语义说明 + 对应集群内角色（单表，不再分视角 tab） ==========
  const permissionDescRows = [
    {
      name: '管理员',
      desc: '对集群所有命名空间或所选命名空间下 Kubernetes 资源的 RBAC 读写权限；集群级绑定时对集群节点、存储卷、命名空间、配额也有读写权限（命名空间级不含节点/存储卷等集群级资源）',
      roleName: 'oneops-admin'
    },
    {
      name: '只读管理员',
      desc: '对集群所有命名空间或所选命名空间下 Kubernetes 资源的 RBAC 只读权限；集群级绑定时对集群节点、存储卷、命名空间、配额也有只读权限（命名空间级不含节点/存储卷等集群级资源）',
      roleName: 'oneops-readonly（聚合）'
    },
    {
      name: '运维人员',
      desc: '对集群所有命名空间或所选命名空间下控制台可见 Kubernetes 资源的 RBAC 读写权限，对其他资源的只读权限；集群级绑定时含对集群节点、存储卷、命名空间的读取与更新权限（命名空间级不含节点/存储卷更新）',
      roleName: 'oneops-ops'
    },
    {
      name: '开发人员',
      desc: '对集群所有命名空间或所选命名空间下控制台可见 Kubernetes 资源的 RBAC 读写权限',
      roleName: 'oneops-dev'
    },
    {
      name: '受限用户',
      desc: '对集群所有命名空间或所选命名空间下控制台可见 Kubernetes 资源的 RBAC 只读权限',
      roleName: 'oneops-view'
    },
    {
      name: '自定义',
      desc: '权限由您所选择的 ClusterRole 决定，请在确定所选 ClusterRole 对各类资源的操作权限后再进行授权，以免被授权主体获得不符合预期的权限',
      roleName: '所选 ClusterRole / Role'
    }
  ];

  onMounted(() => {
    loadPageData();
  });
</script>

<template>
  <div class="flex flex-col gap-16px">
    <ElAlert
      type="info"
      show-icon
      :closable="false"
      title="以用户/用户组视角管理集群访问授权：有效权限 = 系统权限码 ∧ 集群内原生 RBAC 绑定（含经用户组继承）；此处创建集群内真实 Binding，操作经 impersonation 由 kube-apiserver 最终判定"
      class="!mb-0"
    />

    <ElTabs v-model="activeTab">
      <!-- 用户 tab：主体视角 -->
      <ElTabPane label="用户" name="users">
        <div class="mb-12px flex items-center gap-8px">
          <ElInput
            v-model="userSearch"
            :prefix-icon="Search"
            placeholder="按用户名 / 昵称搜索"
            clearable
            style="width: 240px"
          />
        </div>
        <ElTable v-loading="loading && users.length === 0" :data="filteredUsers" stripe>
          <ElTableColumn label="用户" min-width="200">
            <template #default="{ row }">
              <span>{{ row.nickname || row.username }}</span>
              <span v-if="row.nickname && row.nickname !== row.username" class="ml-4px text-12px opacity-60">
                {{ row.username }}
              </span>
            </template>
          </ElTableColumn>
          <ElTableColumn label="所属用户组" min-width="180">
            <template #default="{ row }">
              <ElTag v-for="gid in row.groupIds || []" :key="gid" size="small" effect="plain" class="mr-4px">
                {{ groupNameById.get(gid) || gid }}
              </ElTag>
              <span v-if="!(row.groupIds || []).length" class="opacity-40">-</span>
            </template>
          </ElTableColumn>
          <ElTableColumn label="有效授权" width="170">
            <template #default="{ row }">
              <ElTag :type="userPermSummary(row).total > 0 ? 'success' : 'info'" size="small" effect="plain">
                {{ userPermSummary(row).total }} 条
              </ElTag>
              <span v-if="userPermSummary(row).inherited > 0" class="ml-4px text-12px opacity-60">
                直 {{ userPermSummary(row).direct }} / 继承 {{ userPermSummary(row).inherited }}
              </span>
            </template>
          </ElTableColumn>
          <ElTableColumn label="操作" width="110" fixed="right">
            <template #default="{ row }">
              <ElButton
                link
                type="primary"
                size="small"
                @click="
                  openPerms({
                    type: 'user',
                    id: row.id,
                    label: row.nickname || row.username,
                    sub: row.username,
                    groupIds: row.groupIds
                  })
                "
              >
                权限管理
              </ElButton>
            </template>
          </ElTableColumn>
        </ElTable>
      </ElTabPane>

      <!-- 用户组 tab：主体视角 -->
      <ElTabPane label="用户组" name="groups">
        <div class="mb-12px flex items-center gap-8px">
          <ElInput
            v-model="groupSearch"
            :prefix-icon="Search"
            placeholder="按名称 / 编码搜索"
            clearable
            style="width: 240px"
          />
        </div>
        <ElTable v-loading="loading && groups.length === 0" :data="filteredGroups" stripe>
          <ElTableColumn label="用户组" min-width="200">
            <template #default="{ row }">
              <span>{{ row.name }}</span>
              <span v-if="row.code && row.code !== row.name" class="ml-4px text-12px opacity-60">{{ row.code }}</span>
            </template>
          </ElTableColumn>
          <ElTableColumn label="有效授权" width="120">
            <template #default="{ row }">
              <ElTag :type="groupPermCount(row.id) > 0 ? 'success' : 'info'" size="small" effect="plain">
                {{ groupPermCount(row.id) }} 条
              </ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn label="操作" width="110" fixed="right">
            <template #default="{ row }">
              <ElButton
                link
                type="primary"
                size="small"
                @click="openPerms({ type: 'group', id: row.id, label: row.name, sub: row.code || '' })"
              >
                权限管理
              </ElButton>
            </template>
          </ElTableColumn>
        </ElTable>
      </ElTabPane>

      <!-- 授权记录 tab：全局绑定流水（审计视角） -->
      <ElTabPane label="授权记录" name="records">
        <div class="mb-12px flex items-center gap-8px">
          <ElSelect v-model="clusterFilter" filterable clearable placeholder="全部集群" style="width: 200px">
            <ElOption v-for="c in clusterOptions" :key="c.id" :value="c.id" :label="c.name" />
          </ElSelect>
          <ElButton :icon="Refresh" @click="loadPageData">刷新</ElButton>
        </div>
        <ElTable v-loading="loading" :data="filteredRecords" stripe>
          <ElTableColumn label="主体类型" width="100">
            <template #default="{ row }">
              <ElTag :type="row.subjectType === 'user' ? 'primary' : 'warning'" size="small">
                {{ row.subjectType === 'user' ? '用户' : '用户组' }}
              </ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn label="主体" min-width="150">
            <template #default="{ row }">
              <span>{{ row.subjectName }}</span>
              <span
                v-if="row.subjectNickname && row.subjectNickname !== row.subjectName"
                class="ml-4px text-12px opacity-60"
              >
                {{ row.subjectNickname }}
              </span>
            </template>
          </ElTableColumn>
          <ElTableColumn label="集群" min-width="130">
            <template #default="{ row }">
              {{ row.clusterName || row.clusterId }}
            </template>
          </ElTableColumn>
          <ElTableColumn label="角色" min-width="220">
            <template #default="{ row }">
              <ElTag size="small" effect="plain">{{ row.roleKind }}</ElTag>
              <span class="ml-4px">{{ row.roleName }}</span>
            </template>
          </ElTableColumn>
          <ElTableColumn label="范围" width="130">
            <template #default="{ row }">
              <ElTag v-if="!row.namespace" size="small" effect="plain" type="info">全集群</ElTag>
              <span v-else>{{ row.namespace }}</span>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="createdAt" label="授权时间" width="170" />
          <ElTableColumn v-if="canRevoke" label="操作" width="100" fixed="right">
            <template #default="{ row }">
              <ElButton type="danger" link size="small" @click="handleRevoke(row)">撤销</ElButton>
            </template>
          </ElTableColumn>
        </ElTable>
      </ElTabPane>
    </ElTabs>

    <!-- 权限管理抽屉：主体头 + 已有权限 + 添加权限（待提交行清单）+ 权限说明 -->
    <ElDrawer v-model="permsVisible" title="权限管理" size="70%">
      <!-- 顶部：授权主体 -->
      <ElAlert type="info" :closable="false" class="mb-4px">
        <template #title>
          <ElTag size="small" class="mr-4px">{{ permsSubject?.type === 'user' ? '用户' : '用户组' }}</ElTag>
          {{ permsSubject?.label }}（{{ permsSubject?.sub || permsSubject?.id }}）
        </template>
      </ElAlert>

      <!-- 模块：已有权限（直接 + 组继承） -->
      <ElDivider content-position="left">已有权限（{{ permRows.length }} 条）</ElDivider>
      <ElText type="info" size="small" class="mb-8px block">
        有效权限 = 直接授权 + 经用户组继承，逐条对应集群内真实 Binding；继承条目请在对应用户组下管理
      </ElText>
      <ElTable :data="permRows" stripe size="small" class="mb-8px">
        <ElTableColumn label="来源" width="140">
          <template #default="{ row }">
            <ElTag :type="row.inherited ? 'warning' : 'success'" size="small" effect="plain">
              {{ row.sourceLabel }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="集群" min-width="110">
          <template #default="{ row }">
            {{ row.binding.clusterName || row.binding.clusterId }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="角色" min-width="190">
          <template #default="{ row }">
            <ElTag size="small" effect="plain">{{ row.binding.roleKind }}</ElTag>
            <span class="ml-4px">{{ row.binding.roleName }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn label="命名空间" width="120">
          <template #default="{ row }">
            <ElTag v-if="!row.binding.namespace" size="small" effect="plain" type="info">全部命名空间</ElTag>
            <span v-else>{{ row.binding.namespace }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="binding.createdAt" label="授权时间" width="165" />
        <ElTableColumn v-if="canRevoke" label="操作" width="90" fixed="right">
          <template #default="{ row }">
            <ElButton v-if="!row.inherited" type="danger" link size="small" @click="handleRevoke(row.binding)">
              撤销
            </ElButton>
            <span v-else class="text-12px opacity-50">组管理</span>
          </template>
        </ElTableColumn>
      </ElTable>

      <!-- 模块：添加权限（表格内交互：单元格即选择器，控件常驻仅切换禁用态） -->
      <template v-if="canAssign">
        <ElDivider content-position="left">添加权限</ElDivider>
        <ElTable :data="pendingRows" size="small" class="mb-8px">
          <ElTableColumn label="集群" min-width="170">
            <template #default="{ row }">
              <ElSelect
                v-model="row.clusterId"
                filterable
                placeholder="选择集群"
                style="width: 100%"
                @change="onRowClusterChange(row)"
              >
                <ElOption :value="0" label="所有集群" />
                <ElOption v-for="c in clusterOptions" :key="c.id" :value="c.id" :label="c.name" />
              </ElSelect>
            </template>
          </ElTableColumn>
          <ElTableColumn label="命名空间" min-width="200">
            <template #default="{ row }">
              <ElSelect
                v-model="row.namespaces"
                multiple
                filterable
                collapse-tags
                :disabled="rowNsDisabled(row)"
                :placeholder="rowNsPlaceholder(row)"
                style="width: 100%"
              >
                <ElOption v-for="ns in row.nsOptions" :key="ns" :value="ns" :label="ns" />
              </ElSelect>
              <ElText
                v-if="!rowNsDisabled(row) && row.namespaces.length > 0 && row.namespaces.length < row.nsOptions.length"
                type="info"
                size="small"
                class="block"
              >
                已选择 {{ row.namespaces.length }}/{{ row.nsOptions.length }} 项（不选 = 全部命名空间）
              </ElText>
            </template>
          </ElTableColumn>
          <ElTableColumn label="权限" min-width="130">
            <template #default="{ row }">
              <ElSelect
                v-model="row.permission"
                filterable
                placeholder="选择权限"
                style="width: 100%"
                @change="onRowPermissionChange(row)"
              >
                <ElOption v-for="p in presetOptions" :key="p.value" :value="p.value" :label="p.label" />
                <ElOption value="custom" label="自定义" :disabled="row.clusterId === 0" />
              </ElSelect>
            </template>
          </ElTableColumn>
          <ElTableColumn label="自定义角色" min-width="230">
            <template #default="{ row }">
              <template v-if="row.permission === 'custom'">
                <ElSelect v-model="row.roleKind" style="width: 100%" @change="onRowRoleKindChange(row)">
                  <ElOption value="ClusterRole" label="ClusterRole（集群级）" />
                  <ElOption value="Role" label="Role（命名空间级）" />
                </ElSelect>
                <ElSelect
                  v-model="row.roleName"
                  filterable
                  placeholder="选择集群内角色"
                  style="width: 100%; margin-top: 4px"
                >
                  <ElOption v-for="r in row.roleOptions" :key="r" :value="r" :label="r" />
                </ElSelect>
              </template>
              <ElText v-else type="info" size="small">—</ElText>
            </template>
          </ElTableColumn>
          <ElTableColumn label="操作" width="70" fixed="right">
            <template #default="{ $index }">
              <ElButton link type="danger" size="small" @click="removePendingRow($index)">删除</ElButton>
            </template>
          </ElTableColumn>
        </ElTable>
        <ElButton plain :icon="Plus" @click="addPendingRow">添加权限</ElButton>
      </template>

      <!-- 模块：权限说明（档位 → 语义 + 对应集群内角色） -->
      <ElDivider content-position="left">权限说明</ElDivider>
      <ElTable :data="permissionDescRows" size="small" border>
        <ElTableColumn prop="name" label="访问权限" width="110" />
        <ElTableColumn prop="desc" label="集群内RBAC权限" />
        <ElTableColumn prop="roleName" label="集群内角色" width="170" />
      </ElTable>

      <template #footer>
        <ElButton
          type="primary"
          :loading="submitting"
          :disabled="!canAssign || pendingRows.length === 0"
          @click="handleSubmitPerms"
        >
          提交授权
        </ElButton>
        <ElButton @click="permsVisible = false">取消</ElButton>
      </template>
    </ElDrawer>
  </div>
</template>
