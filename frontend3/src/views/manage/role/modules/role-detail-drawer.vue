<script setup lang="ts">
    import { ref, watch } from 'vue';
    import { ElTag, ElTree } from 'element-plus';
    import { fetchGetRolePermissions, fetchPermissionOptions } from '@/service/api';
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

    /** 将角色已有权限列表组装为 模块→页面→按钮 三层树（只读展示） */
    function buildPermTree(perms: Api.SystemManage.Permission[]): PermNode[] {
      const tree: PermNode[] = [];
      const modules = new Map<string, PermNode>();
      const pages = new Map<string, PermNode>();

      for (const p of perms) {
        if (p.level === 1) {
          const node: PermNode = { id: p.id, label: p.name, code: p.code, level: 1, children: [] };
          modules.set(p.module, node);
          tree.push(node);
        } else if (p.level === 2) {
          let mod = modules.get(p.module);
          if (!mod) {
            mod = { id: -modules.size - 1, label: p.module, code: p.module, level: 1, children: [] };
            modules.set(p.module, mod);
            tree.push(mod);
          }
          const page: PermNode = { id: p.id, label: p.name, code: p.code, level: 2, children: [] };
          pages.set(p.code, page);
          mod.children!.push(page);
        } else if (p.level === 3) {
          const parentCode = `${p.module}.${p.resource}`;
          const page = pages.get(parentCode);
          const node: PermNode = { id: p.id, label: p.name, code: p.code, level: 3 };
          if (page) {
            page.children!.push(node);
          } else {
            // 父页面权限未单独分配时挂到模块节点下
            let mod = modules.get(p.module);
            if (!mod) {
              mod = { id: -modules.size - 1, label: p.module, code: p.module, level: 1, children: [] };
              modules.set(p.module, mod);
              tree.push(mod);
            }
            mod.children!.push(node);
          }
        }
      }
      return tree;
    }

    /**
     * 加载角色权限：
     * 后端 /roles/:id/permissions 只返回权限 ID 数组，
     * 需再拉全量权限选项（含 name/code/module/level）按 ID 过滤后组装树
     */
    async function loadPermissions(roleId: number) {
      loading.value = true;
      const [permRes, idsRes] = await Promise.all([fetchPermissionOptions(), fetchGetRolePermissions(roleId)]);
      loading.value = false;

      if (permRes.error || !permRes.data || idsRes.error || !idsRes.data) {
        permissionCount.value = 0;
        permissionTree.value = [];
        return;
      }

      const idSet = new Set<number>(idsRes.data);
      const rolePerms = permRes.data.filter((p: Api.SystemManage.Permission) => idSet.has(p.id));
      permissionCount.value = rolePerms.length;
      permissionTree.value = buildPermTree(rolePerms);
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

            <ElAlert v-if="props.role?.code === 'admin'" title="超级管理员默认拥有全部权限（*.*.*），由系统内置规则生效" type="info" :closable="false" show-icon />
            <template v-else>
                <ElTree :data="permissionTree" node-key="id" :default-expand-all="true" :indent="20" :props="{ label: 'label', children: 'children' }">
                    <template #default="{ data }">
                        <div class="flex flex-1 items-center justify-between gap-8px">
                            <span>{{ data.label }}</span>
                            <span class="text-12px text-gray-400 truncate">{{ data.code }}</span>
                        </div>
                    </template>
                </ElTree>
                <ElEmpty v-if="!loading && permissionTree.length === 0" description="暂无权限" :image-size="60" />
            </template>
        </div>
    </ElDrawer>
</template>
