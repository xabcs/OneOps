<script setup lang="ts">
  import { ref, watch } from 'vue';
  import { ElTag, ElTree } from 'element-plus';
  import { fetchGetPermissionTree, fetchGetRolePermissions } from '@/service/api';
  import { getRoleTagType } from '../../user/modules/user-columns';

  defineOptions({ name: 'RoleDetailDrawer' });

  interface Props {
    /** 角色数据（列表行） */
    role?: Api.SystemManage.Role | null;
  }

  const props = defineProps<Props>();

  const visible = defineModel<boolean>('visible', {
    default: false
  });

  interface PermNode {
    id: number;
    label: string;
    code: string;
    level: number;
    children?: PermNode[];
  }

  const permissionTree = ref<PermNode[]>([]);
  const permissionCount = ref(0);
  const loading = ref(false);

  /**
   * 从全量权限树中裁出角色拥有的子树（只读展示）：
   * 叶子（Level 3）按已绑定 ID 勾选保留，分组节点保留含勾选叶子的路径
   */
  function filterPermTree(nodes: Api.SystemManage.Permission[], idSet: Set<number>): PermNode[] {
    const result: PermNode[] = [];
    for (const node of nodes) {
      if (!node.children || node.children.length === 0) {
        if (idSet.has(node.id)) {
          result.push({ id: node.id, label: node.name, code: node.code, level: node.level });
        }
      } else {
        const children = filterPermTree(node.children, idSet);
        if (children.length > 0) {
          result.push({ id: node.id, label: node.name, code: node.code, level: node.level, children });
        }
      }
    }
    return result;
  }

  /**
   * 加载角色权限：
   * 后端 /roles/:id/permissions 返回权限 ID 数组，
   * 与全量权限树（模块→资源→按钮）求交，裁出已拥有权限的子树
   */
  async function loadPermissions(roleId: number) {
    loading.value = true;
    const [treeRes, idsRes] = await Promise.all([fetchGetPermissionTree(), fetchGetRolePermissions(roleId)]);
    loading.value = false;

    if (treeRes.error || !treeRes.data || idsRes.error || !idsRes.data) {
      permissionCount.value = 0;
      permissionTree.value = [];
      return;
    }

    // 角色只绑定叶子（Level 3）权限，分组节点不计入数量
    const leafIds = new Set<number>();
    const collectLeaves = (nodes: Api.SystemManage.Permission[]) => {
      for (const node of nodes) {
        if (!node.children || node.children.length === 0) {
          leafIds.add(node.id);
        } else {
          collectLeaves(node.children);
        }
      }
    };
    collectLeaves(treeRes.data);

    const idSet = new Set<number>(idsRes.data.filter(id => leafIds.has(id)));
    permissionCount.value = idSet.size;
    permissionTree.value = filterPermTree(treeRes.data, idSet);
  }

  watch(
    () => visible.value,
    val => {
      if (val && props.role) {
        if (props.role.code === 'admin') {
          // admin 权限在鉴权层通配生效，无关联数据
          permissionTree.value = [];
          permissionCount.value = 0;
        } else {
          loadPermissions(props.role.id);
        }
      }
    }
  );
</script>

<template>
  <ElDrawer v-model="visible" :title="`角色详情：${props.role?.name ?? ''}`" size="480px">
    <div v-loading="loading" class="flex flex-col gap-20px">
      <!-- 基本信息 -->
      <ElDescriptions :column="2" border size="small">
        <ElDescriptionsItem label="角色编码">{{ props.role?.code }}</ElDescriptionsItem>
        <ElDescriptionsItem label="状态">
          <ElTag :type="props.role?.status === 1 ? 'success' : 'danger'" size="small">
            {{ props.role?.status === 1 ? '启用' : '禁用' }}
          </ElTag>
        </ElDescriptionsItem>
        <ElDescriptionsItem label="角色描述" :span="2">{{ props.role?.description || '-' }}</ElDescriptionsItem>
        <ElDescriptionsItem label="绑定用户" :span="2">
          <div v-if="props.role?.users?.length" class="flex flex-wrap items-center gap-4px">
            <ElTag v-for="u in props.role.users" :key="u.id" size="small" :type="getRoleTagType(u.username)">
              {{ u.username }}
            </ElTag>
          </div>
          <span v-else class="text-gray">-</span>
        </ElDescriptionsItem>
      </ElDescriptions>

      <!-- 权限列表 -->
      <div class="flex items-center justify-between">
        <span class="text-15px font-bold">权限列表</span>
        <ElTag v-if="props.role?.code !== 'admin'" type="info" size="small">{{ permissionCount }} 项</ElTag>
      </div>

      <ElAlert
        v-if="props.role?.code === 'admin'"
        title="超级管理员默认拥有全部权限（*.*.*），由系统内置规则生效"
        type="info"
        :closable="false"
        show-icon
      />
      <template v-else>
        <ElTree
          :data="permissionTree"
          node-key="id"
          :default-expand-all="true"
          :indent="20"
          :props="{ label: 'label', children: 'children' }"
        >
          <template #default="{ data }">
            <div class="flex flex-1 items-center justify-between gap-8px">
              <span>{{ data.label }}</span>
              <span class="truncate text-12px text-gray-400">{{ data.code }}</span>
            </div>
          </template>
        </ElTree>
        <ElEmpty v-if="!loading && permissionTree.length === 0" description="暂无权限" :image-size="60" />
      </template>
    </div>
  </ElDrawer>
</template>
