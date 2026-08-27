<script setup lang="ts">
    import { onMounted, ref, watch } from 'vue';
    import { ElMessage } from 'element-plus';
    import { Key, Refresh, Search } from '@element-plus/icons-vue';
    import {
      fetchK8sClusterNamespaces,
      fetchK8sClusters,
      fetchK8sNativeClusterRoleBindings,
      fetchK8sNativeClusterRoles,
      fetchK8sNativeRoleBindings,
      fetchK8sNativeRoles,
      getK8sNativeClusterRole,
      getK8sNativeClusterRoleBinding,
      getK8sNativeRole,
      getK8sNativeRoleBinding,
      updateK8sNativeClusterRole,
      updateK8sNativeRole
    } from '@/service/api/k8s';
    import { executeWithPermission } from '@/hooks/business/auth';

    defineOptions({ name: 'K8sSecurityRoles' });

    type TabKey = 'clusterroles' | 'roles' | 'crbs' | 'rbs';

    const clusters = ref<K8s.Cluster[]>([]);
    const clusterId = ref<number>();
    const activeTab = ref<TabKey>('clusterroles');
    const loading = ref(false);
    const search = ref('');
    const namespaceFilter = ref('');
    const namespaces = ref<string[]>([]);

    const clusterRoles = ref<K8s.NativeRoleItem[]>([]);
    const roles = ref<K8s.NativeRoleItem[]>([]);
    const crbs = ref<K8s.NativeBindingItem[]>([]);
    const rbs = ref<K8s.NativeBindingItem[]>([]);

    // 详情抽屉
    const detailVisible = ref(false);
    const detailLoading = ref(false);
    const detailTitle = ref('');
    const roleDetail = ref<K8s.NativeRoleDetail | null>(null);
    const bindingDetail = ref<K8s.NativeBindingItem | null>(null);

    // 编辑弹窗（ClusterRole / Role）
    const editVisible = ref(false);
    const editSaving = ref(false);
    const editKind = ref<'ClusterRole' | 'Role'>('ClusterRole');
    const editNamespace = ref('');
    const editName = ref('');
    const editManagedBy = ref('');
    const editManifest = ref('');

    const showNamespaceFilter = () => activeTab.value === 'roles' || activeTab.value === 'rbs';

    async function loadClusters() {
      const { data, error } = await fetchK8sClusters({ page: 1, pageSize: 100 });
      if (!error && data) {
        clusters.value = data.list || [];
        if (clusters.value.length > 0 && !clusterId.value) {
          clusterId.value = clusters.value[0].id;
        }
      }
    }

    async function loadNamespaces() {
      if (!clusterId.value) return;
      const { data, error } = await fetchK8sClusterNamespaces(clusterId.value);
      if (!error && data) {
        namespaces.value = (data || []).map(ns => ns.name);
      }
    }

    async function loadList() {
      if (!clusterId.value) return;
      loading.value = true;
      try {
        if (activeTab.value === 'clusterroles') {
          const { data, error } = await fetchK8sNativeClusterRoles(clusterId.value, search.value || undefined);
          if (!error && data) clusterRoles.value = data || [];
        } else if (activeTab.value === 'roles') {
          const { data, error } = await fetchK8sNativeRoles(clusterId.value, {
            namespace: namespaceFilter.value || undefined,
            search: search.value || undefined
          });
          if (!error && data) roles.value = data || [];
        } else if (activeTab.value === 'crbs') {
          const { data, error } = await fetchK8sNativeClusterRoleBindings(clusterId.value, search.value || undefined);
          if (!error && data) crbs.value = data || [];
        } else {
          const { data, error } = await fetchK8sNativeRoleBindings(clusterId.value, {
            namespace: namespaceFilter.value || undefined,
            search: search.value || undefined
          });
          if (!error && data) rbs.value = data || [];
        }
      } finally {
        loading.value = false;
      }
    }

    watch(clusterId, () => {
      namespaceFilter.value = '';
      loadNamespaces();
      loadList();
    });

    watch(activeTab, () => loadList());

    async function openRoleDetail(kind: 'ClusterRole' | 'Role', row: K8s.NativeRoleItem) {
      if (!clusterId.value) return;
      detailVisible.value = true;
      detailLoading.value = true;
      detailTitle.value = `${kind}: ${row.name}`;
      bindingDetail.value = null;
      roleDetail.value = null;
      try {
        const { data, error } =
          kind === 'ClusterRole'
            ? await getK8sNativeClusterRole(clusterId.value, row.name)
            : await getK8sNativeRole(clusterId.value, row.namespace || '', row.name);
        if (!error && data) {
          roleDetail.value = data;
        }
      } finally {
        detailLoading.value = false;
      }
    }

    async function openBindingDetail(kind: 'CRB' | 'RB', row: K8s.NativeBindingItem) {
      if (!clusterId.value) return;
      detailVisible.value = true;
      detailLoading.value = true;
      detailTitle.value = `${kind === 'CRB' ? 'ClusterRoleBinding' : 'RoleBinding'}: ${row.name}`;
      roleDetail.value = null;
      bindingDetail.value = null;
      try {
        const { data, error } =
          kind === 'CRB'
            ? await getK8sNativeClusterRoleBinding(clusterId.value, row.name)
            : await getK8sNativeRoleBinding(clusterId.value, row.namespace || '', row.name);
        if (!error && data) {
          bindingDetail.value = data;
        }
      } finally {
        detailLoading.value = false;
      }
    }

    function openEdit() {
      const detail = roleDetail.value;
      if (!detail) return;
      editKind.value = detail.namespace ? 'Role' : 'ClusterRole';
      editNamespace.value = detail.namespace || '';
      editName.value = detail.name;
      editManagedBy.value = detail.managedBy;
      const metadata: Record<string, unknown> = { name: detail.name, labels: detail.labels || {} };
      if (detail.namespace) metadata.namespace = detail.namespace;
      editManifest.value = JSON.stringify(
        {
          apiVersion: 'rbac.authorization.k8s.io/v1',
          kind: editKind.value,
          metadata,
          rules: detail.rules
        },
        null,
        2
      );
      editVisible.value = true;
    }

    async function handleSave() {
      if (!clusterId.value) return;
      let manifest: Record<string, unknown>;
      try {
        manifest = JSON.parse(editManifest.value) as Record<string, unknown>;
      } catch {
        ElMessage.error('JSON 格式错误，请检查后重试');
        return;
      }
      await executeWithPermission('k8s.rbac.manage', async () => {
        editSaving.value = true;
        try {
          const { error } =
            editKind.value === 'ClusterRole'
              ? await updateK8sNativeClusterRole(clusterId.value!, manifest)
              : await updateK8sNativeRole(clusterId.value!, editNamespace.value, manifest);
          if (!error) {
            ElMessage.success('更新成功');
            editVisible.value = false;
            await loadList();
            if (roleDetail.value) {
              await openRoleDetail(editKind.value, {
                name: editName.value,
                namespace: editNamespace.value || undefined,
                managedBy: editManagedBy.value,
                rules: 0,
                createdAt: ''
              });
            }
          }
        } finally {
          editSaving.value = false;
        }
      });
    }

    function subjectPreview(row: K8s.NativeBindingItem) {
      return (row.subjects || []).map(s => `${s.kind}/${s.name}`).join('、');
    }

    onMounted(() => {
      loadClusters();
    });
</script>

<template>
    <div class="flex flex-col gap-16px">
        <!-- Hero 区域 -->
        <ElCard shadow="hover">
            <div class="flex items-center justify-between flex-wrap gap-12px">
                <div class="flex items-center gap-12px">
                    <ElIcon :size="24">
                        <Key />
                    </ElIcon>
                    <div class="flex flex-col gap-2px">
                        <h2 class="text-18px font-bold m-0">角色管理</h2>
                        <p class="text-13px opacity-70 m-0">
                            集群内原生 RBAC 对象代管：ClusterRole / Role / ClusterRoleBinding / RoleBinding，Helm 等外部管理的对象可查看但编辑会被还原
                        </p>
                    </div>
                </div>
                <div class="flex items-center gap-8px">
                    <ElSelect v-model="clusterId" filterable placeholder="选择集群" style="width: 240px">
                        <ElOption v-for="c in clusters" :key="c.id" :value="c.id" :label="c.name" />
                    </ElSelect>
                    <ElButton size="small" :loading="loading" @click="loadList">
                        <ElIcon>
                            <Refresh />
                        </ElIcon>
                        刷新
                    </ElButton>
                </div>
            </div>
        </ElCard>

        <ElCard shadow="hover">
            <ElTabs v-model="activeTab">
                <ElTabPane label="ClusterRole" name="clusterroles" />
                <ElTabPane label="Role（命名空间级）" name="roles" />
                <ElTabPane label="ClusterRoleBinding" name="crbs" />
                <ElTabPane label="RoleBinding（命名空间级）" name="rbs" />
            </ElTabs>

            <div class="flex items-center gap-8px mb-12px">
                <ElInput v-model="search" :prefix-icon="Search" placeholder="按名称 / 主体搜索" clearable style="width: 240px" @keyup.enter="loadList" @clear="loadList" />
                <ElSelect v-if="showNamespaceFilter()" v-model="namespaceFilter" filterable clearable placeholder="命名空间（全部）" style="width: 200px" @change="loadList">
                    <ElOption v-for="ns in namespaces" :key="ns" :value="ns" :label="ns" />
                </ElSelect>
                <ElButton type="primary" plain @click="loadList">查询</ElButton>
            </div>

            <!-- ClusterRole -->
            <ElTable v-if="activeTab === 'clusterroles'" v-loading="loading" :data="clusterRoles" stripe>
                <ElTableColumn prop="name" label="名称" min-width="260" show-overflow-tooltip />
                <ElTableColumn label="管理方" width="130">
                    <template #default="{ row }">
                        <ElTag v-if="row.managedBy" type="warning" size="small" effect="plain">{{ row.managedBy }}</ElTag>
                        <ElTag v-else type="info" size="small" effect="plain">原生</ElTag>
                    </template>
                </ElTableColumn>
                <ElTableColumn prop="rules" label="规则数" width="90" />
                <ElTableColumn prop="createdAt" label="创建时间" width="170" />
                <ElTableColumn label="操作" width="130" fixed="right">
                    <template #default="{ row }">
                        <ElButton link type="primary" size="small" @click="openRoleDetail('ClusterRole', row)">详情</ElButton>
                    </template>
                </ElTableColumn>
            </ElTable>

            <!-- Role -->
            <ElTable v-else-if="activeTab === 'roles'" v-loading="loading" :data="roles" stripe>
                <ElTableColumn prop="name" label="名称" min-width="240" show-overflow-tooltip />
                <ElTableColumn prop="namespace" label="命名空间" min-width="140" show-overflow-tooltip />
                <ElTableColumn label="管理方" width="130">
                    <template #default="{ row }">
                        <ElTag v-if="row.managedBy" type="warning" size="small" effect="plain">{{ row.managedBy }}</ElTag>
                        <ElTag v-else type="info" size="small" effect="plain">原生</ElTag>
                    </template>
                </ElTableColumn>
                <ElTableColumn prop="rules" label="规则数" width="90" />
                <ElTableColumn prop="createdAt" label="创建时间" width="170" />
                <ElTableColumn label="操作" width="130" fixed="right">
                    <template #default="{ row }">
                        <ElButton link type="primary" size="small" @click="openRoleDetail('Role', row)">详情</ElButton>
                    </template>
                </ElTableColumn>
            </ElTable>

            <!-- ClusterRoleBinding -->
            <ElTable v-else-if="activeTab === 'crbs'" v-loading="loading" :data="crbs" stripe>
                <ElTableColumn prop="name" label="名称" min-width="240" show-overflow-tooltip />
                <ElTableColumn label="主体" min-width="240" show-overflow-tooltip>
                    <template #default="{ row }">{{ subjectPreview(row) }}</template>
                </ElTableColumn>
                <ElTableColumn label="引用角色" min-width="200" show-overflow-tooltip>
                    <template #default="{ row }">{{ row.roleKind }}/{{ row.roleName }}</template>
                </ElTableColumn>
                <ElTableColumn label="管理方" width="130">
                    <template #default="{ row }">
                        <ElTag v-if="row.managedBy === 'oneops'" type="success" size="small" effect="plain">OneOps</ElTag>
                        <ElTag v-else-if="row.managedBy" type="warning" size="small" effect="plain">{{ row.managedBy }}</ElTag>
                        <ElTag v-else type="info" size="small" effect="plain">原生</ElTag>
                    </template>
                </ElTableColumn>
                <ElTableColumn prop="createdAt" label="创建时间" width="170" />
                <ElTableColumn label="操作" width="130" fixed="right">
                    <template #default="{ row }">
                        <ElButton link type="primary" size="small" @click="openBindingDetail('CRB', row)">详情</ElButton>
                    </template>
                </ElTableColumn>
            </ElTable>

            <!-- RoleBinding -->
            <ElTable v-else v-loading="loading" :data="rbs" stripe>
                <ElTableColumn prop="name" label="名称" min-width="220" show-overflow-tooltip />
                <ElTableColumn prop="namespace" label="命名空间" min-width="130" show-overflow-tooltip />
                <ElTableColumn label="主体" min-width="220" show-overflow-tooltip>
                    <template #default="{ row }">{{ subjectPreview(row) }}</template>
                </ElTableColumn>
                <ElTableColumn label="引用角色" min-width="200" show-overflow-tooltip>
                    <template #default="{ row }">{{ row.roleKind }}/{{ row.roleName }}</template>
                </ElTableColumn>
                <ElTableColumn label="管理方" width="130">
                    <template #default="{ row }">
                        <ElTag v-if="row.managedBy === 'oneops'" type="success" size="small" effect="plain">OneOps</ElTag>
                        <ElTag v-else-if="row.managedBy" type="warning" size="small" effect="plain">{{ row.managedBy }}</ElTag>
                        <ElTag v-else type="info" size="small" effect="plain">原生</ElTag>
                    </template>
                </ElTableColumn>
                <ElTableColumn prop="createdAt" label="创建时间" width="170" />
                <ElTableColumn label="操作" width="130" fixed="right">
                    <template #default="{ row }">
                        <ElButton link type="primary" size="small" @click="openBindingDetail('RB', row)">详情</ElButton>
                    </template>
                </ElTableColumn>
            </ElTable>
        </ElCard>

        <!-- 详情抽屉 -->
        <ElDrawer v-model="detailVisible" :title="detailTitle" size="60%">
            <div v-loading="detailLoading">
                <!-- 角色详情：rules 矩阵 -->
                <template v-if="roleDetail">
                    <ElAlert v-if="roleDetail.managedBy" type="warning" show-icon :closable="false" :title="`该对象由 ${roleDetail.managedBy} 管理，在此编辑可能在下次同步/升级时被还原`" class="mb-12px" />
                    <div class="flex items-center gap-8px mb-12px">
                        <ElTag v-if="roleDetail.namespace" size="small">{{ roleDetail.namespace }}</ElTag>
                        <ElTag type="info" size="small" effect="plain">{{ roleDetail.rules.length }} 条规则</ElTag>
                        <span class="text-12px opacity-60">创建于 {{ roleDetail.createdAt }}</span>
                        <PermissionButton code="k8s.rbac.manage" type="primary" size="small" class="ml-auto" @click="openEdit">编辑 rules</PermissionButton>
                    </div>
                    <ElTable :data="roleDetail.rules" stripe border>
                        <ElTableColumn label="API Groups" min-width="140">
                            <template #default="{ row }">
                                <ElTag v-for="g in row.apiGroups" :key="g" size="small" class="mr-4px mb-2px">{{ g || 'core' }}</ElTag>
                            </template>
                        </ElTableColumn>
                        <ElTableColumn label="Resources" min-width="220">
                            <template #default="{ row }">
                                <ElTag v-for="r in row.resources" :key="r" size="small" type="info" effect="plain" class="mr-4px mb-2px">{{ r }}</ElTag>
                            </template>
                        </ElTableColumn>
                        <ElTableColumn label="Verbs" min-width="220">
                            <template #default="{ row }">
                                <ElTag v-for="v in row.verbs" :key="v" size="small" type="success" effect="plain" class="mr-4px mb-2px">{{ v }}</ElTag>
                            </template>
                        </ElTableColumn>
                        <ElTableColumn label="NonResourceURLs" min-width="140">
                            <template #default="{ row }">
                                <template v-if="row.nonResourceURLs && row.nonResourceURLs.length">
                                    <ElTag v-for="u in row.nonResourceURLs" :key="u" size="small" type="warning" effect="plain" class="mr-4px mb-2px">{{ u }}</ElTag>
                                </template>
                                <span v-else class="opacity-40">-</span>
                            </template>
                        </ElTableColumn>
                    </ElTable>
                </template>

                <!-- Binding 详情：subjects + roleRef -->
                <template v-else-if="bindingDetail">
                    <div class="flex items-center gap-8px mb-12px">
                        <ElTag v-if="bindingDetail.namespace" size="small">{{ bindingDetail.namespace }}</ElTag>
                        <ElTag type="info" size="small" effect="plain">
                            {{ bindingDetail.roleKind }}/{{ bindingDetail.roleName }}
                        </ElTag>
                        <span class="text-12px opacity-60">创建于 {{ bindingDetail.createdAt }}</span>
                    </div>
                    <ElTable :data="bindingDetail.subjects" stripe border>
                        <ElTableColumn prop="kind" label="主体类型" width="140" />
                        <ElTableColumn prop="name" label="主体名称" min-width="240" />
                        <ElTableColumn prop="namespace" label="命名空间" min-width="160">
                            <template #default="{ row }">{{ row.namespace || '-' }}</template>
                        </ElTableColumn>
                    </ElTable>
                </template>
            </div>
        </ElDrawer>

        <!-- 编辑弹窗 -->
        <ElDialog v-model="editVisible" :title="`编辑 ${editKind} - ${editName}`" width="720px">
            <ElAlert v-if="editManagedBy" type="warning" show-icon :closable="false" :title="`该对象由 ${editManagedBy} 管理，保存的修改可能在下次同步/升级时被还原`" class="mb-12px" />
            <ElAlert v-else type="info" show-icon :closable="false" title="以 manifest 整体替换 rules（apiVersion/kind/metadata.name 需保留），保存前请确认 JSON 有效" class="mb-12px" />
            <ElInput v-model="editManifest" type="textarea" :rows="18" spellcheck="false" class="font-mono" />
            <template #footer>
                <ElButton @click="editVisible = false">取消</ElButton>
                <ElButton type="primary" :loading="editSaving" @click="handleSave">保存</ElButton>
            </template>
        </ElDialog>
    </div>
</template>

<style scoped></style>
