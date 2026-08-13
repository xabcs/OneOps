<script setup lang="ts">
  import { computed, nextTick, ref, watch } from 'vue';
  import { ElMessage, ElNotification } from 'element-plus';
  import { fetchAssignRolePermissions, fetchGetRolePermissions, fetchPermissionOptions } from '@/service/api';

  defineOptions({ name: 'PermissionAssignModal' });

  interface Props {
    /** the roleId */
    roleId: number;
    /** the role data */
    roleData?: Api.SystemManage.Role | null;
    /** 是否为只读查看模式 */
    viewOnly?: boolean;
  }

  const props = defineProps<Props>();

  const visible = defineModel<boolean>('visible', {
    default: false
  });

  function closeModal() {
    visible.value = false;
  }

  const title = computed(() =>
    props.viewOnly
      ? `查看角色 "${props.roleData?.name || ''}" 的权限`
      : `为角色 "${props.roleData?.name || ''}" 分配权限`
  );

  interface PermissionNode {
    id: number;
    label: string;
    code: string;
    module: string;
    resource: string;
    action: string;
    level: number;
    children?: PermissionNode[];
  }

  const permissionTree = ref<PermissionNode[]>([]);
  const loading = ref(false);
  const treeRef = ref();

  // 强制刷新的计数器（用于触发统计数字更新）
  const refreshCounter = ref(0);

  // 获取选中权限数量
  function getSelectedCount(): number {
    // 触发响应式更新
    void refreshCounter.value;

    if (!treeRef.value) return 0;
    const checkedKeys = treeRef.value.getCheckedKeys() || [];
    const halfCheckedKeys = treeRef.value.getHalfCheckedKeys() || [];
    const allKeys = [...checkedKeys, ...halfCheckedKeys];
    // 过滤虚拟节点
    return allKeys.filter((id: number) => id > 0).length;
  }

  // 获取所有权限并构建树
  async function getAllPermissions() {
    loading.value = true;
    const { error, data } = await fetchPermissionOptions();

    if (!error && data) {
      const permissions = data;

      if (permissions.length === 0) {
        console.warn('⚠️ 权限列表为空，请先在权限管理中创建权限');
        ElMessage.warning('暂无权限数据，请先在权限管理中创建权限');
        permissionTree.value = [];
        loading.value = false;
        return;
      }

      // 构建权限树：3层结构 Module → Page → Button
      const tree: PermissionNode[] = [];
      const moduleMap = new Map<string, PermissionNode>();
      const pageMap = new Map<string, PermissionNode>();

      permissions.forEach((perm: Api.SystemManage.Permission) => {
        // 按 level 字段分级处理
        if (perm.level === 1) {
          // === Level 1: 模块级权限 ===
          if (!moduleMap.has(perm.module)) {
            const moduleNode: PermissionNode = {
              id: perm.id, // ✅ 使用真实权限ID
              label: perm.name,
              code: perm.code,
              module: perm.module,
              resource: perm.resource || perm.module,
              action: perm.action || 'view',
              level: 1,
              children: []
            };
            moduleMap.set(perm.module, moduleNode);
            tree.push(moduleNode);
          }
        } else if (perm.level === 2) {
          // === Level 2: 页面级权限 ===
          // 确保父模块存在
          if (!moduleMap.has(perm.module)) {
            console.warn('⚠️ 父模块不存在，创建默认模块:', perm.module);
            const moduleNode: PermissionNode = {
              id: -moduleMap.size - 1, // ⚠️ 默认模块使用虚拟ID
              label: getModuleLabel(perm.module),
              code: perm.module,
              module: perm.module,
              resource: perm.module,
              action: 'view',
              level: 1,
              children: []
            };
            moduleMap.set(perm.module, moduleNode);
            tree.push(moduleNode);
          }

          // 创建页面节点（使用真实权限ID）
          const pageNode: PermissionNode = {
            id: perm.id, // ✅ 使用真实权限ID
            label: perm.name,
            code: perm.code,
            module: perm.module,
            resource: perm.resource,
            action: perm.action || 'view',
            level: 2,
            children: []
          };
          pageMap.set(perm.code, pageNode);

          // 将页面节点添加到父模块下
          const moduleNode = moduleMap.get(perm.module)!;
          moduleNode.children!.push(pageNode);
        } else if (perm.level === 3) {
          // === Level 3: 按钮级权限 ===
          // 找到父页面节点，添加按钮
          const parentPageCode = `${perm.module}.${perm.resource}`;
          const pageNode = pageMap.get(parentPageCode);

          if (pageNode) {
            // 父页面存在，添加按钮
            const buttonNode: PermissionNode = {
              id: perm.id, // ✅ 使用真实权限ID
              label: perm.name,
              code: perm.code,
              module: perm.module,
              resource: perm.resource,
              action: perm.action,
              level: 3
            };
            pageNode.children!.push(buttonNode);
          } else {
            // 父页面不存在，尝试创建（兜底逻辑）
            console.warn('⚠️ 未找到父页面，自动创建:', parentPageCode);

            if (!moduleMap.has(perm.module)) {
              const moduleNode: PermissionNode = {
                id: -moduleMap.size - 1, // ⚠️ 虚拟ID
                label: getModuleLabel(perm.module),
                code: perm.module,
                module: perm.module,
                resource: perm.module,
                action: 'view',
                level: 1,
                children: []
              };
              moduleMap.set(perm.module, moduleNode);
              tree.push(moduleNode);
            }

            // 创建页面节点
            const autoPageNode: PermissionNode = {
              id: -moduleMap.size - pageMap.size - 1, // ⚠️ 虚拟ID
              label: getResourceLabel(perm.resource, permissions),
              code: parentPageCode,
              module: perm.module,
              resource: perm.resource,
              action: 'view',
              level: 2,
              children: []
            };
            pageMap.set(parentPageCode, autoPageNode);

            const moduleNode = moduleMap.get(perm.module)!;
            moduleNode.children!.push(autoPageNode);

            // 添加按钮节点（使用真实权限ID）
            const buttonNode: PermissionNode = {
              id: perm.id, // ✅ 使用真实权限ID
              label: perm.name,
              code: perm.code,
              module: perm.module,
              resource: perm.resource,
              action: perm.action,
              level: 3
            };
            autoPageNode.children!.push(buttonNode);
          }
        }
      });

      permissionTree.value = tree;

      // 打印所有真实节点的ID，方便调试
      const allRealIds: number[] = [];
      function collectIds(nodes: PermissionNode[]) {
        nodes.forEach(node => {
          if (node.id > 0) {
            allRealIds.push(node.id);
          }
          if (node.children && node.children.length > 0) {
            collectIds(node.children);
          }
        });
      }
      collectIds(tree);
    } else {
      console.error('❌ 获取权限失败:', error);
    }

    loading.value = false;
  }

  // 获取角色已拥有的权限
  async function getRolePermissions() {
    loading.value = true;

    const { error, data } = await fetchGetRolePermissions(props.roleId);

    if (!error && data) {
      // 等待权限树完全渲染
      await nextTick();
      await new Promise(resolve => setTimeout(resolve, 200));

      if (treeRef.value) {
        try {
          // 后端返回格式：{ success: true, data: [1, 3, 4, 8] }
          // 需要提取 data 字段中的数组
          let rolePermissionIds: number[] = [];

          if (Array.isArray(data)) {
            // 直接就是数组
            rolePermissionIds = data;
          } else if (typeof data === 'object' && data !== null) {
            // 可能是包装对象，尝试提取 data 字段
            if (Array.isArray(data.data)) {
              rolePermissionIds = data.data;
            } else if ('data' in data && typeof data.data === 'object') {
              console.warn('⚠️ data.data 不是数组:', data.data);
            }
          }

          // 🔥 关键修复：只设置 level=3 的按钮权限，过滤掉模块和页面权限
          // 因为选中父节点会自动选中所有子节点，导致所有权限被选中
          const buttonLevelPermissions = rolePermissionIds.filter(id => {
            // 在权限树中查找这个节点，只保留 level=3 的按钮节点
            const nodeLevel = findNodeLevel(permissionTree.value, id);
            const isButtonLevel = nodeLevel === 3;
            if (!isButtonLevel && id > 0) {
            }
            return isButtonLevel;
          });

          // 过滤出有效的权限ID（正整数）
          const validPermissionIds = buttonLevelPermissions.filter((id: number) => {
            const numId = typeof id === 'string' ? Number.parseInt(id, 10) : id;
            const isValid = Number.isInteger(numId) && numId > 0;
            if (!isValid) {
              console.warn('⚠️ 无效的权限ID:', id, typeof id);
            }
            return isValid;
          });

          // 先清空当前选中状态
          treeRef.value.setCheckedKeys([]);

          // 设置选中状态 - 只设置叶子节点的选中状态
          // ElTree 会自动处理父节点的半选状态
          if (validPermissionIds.length > 0) {
            treeRef.value.setCheckedKeys(validPermissionIds);

            // 验证设置结果
            const currentCheckedKeys = treeRef.value.getCheckedKeys();

            // 如果选中数量不匹配，给出警告
            if (currentCheckedKeys.length !== validPermissionIds.length) {
              console.warn('⚠️ 选中数量不匹配:', {
                预期: validPermissionIds.length,
                实际: currentCheckedKeys.length,
                未选中的IDs: validPermissionIds.filter(id => !currentCheckedKeys.includes(id))
              });
            }
          } else {
          }
        } catch (err) {
          console.error('❌ 设置权限选中状态时出错:', err);
        }
      } else {
        console.error('❌ treeRef.value 不存在');
      }
    } else {
      console.error('❌ 获取角色权限失败:', error);
    }

    loading.value = false;
  }

  // 提交权限分配
  async function handleSubmit() {
    // 使用官方推荐方式：通过 ref 获取选中的键
    if (!treeRef.value) {
      ElMessage.error('树组件未初始化');
      return;
    }

    // 获取选中的节点ID（包含半选中的父节点）
    const checkedKeys = treeRef.value.getCheckedKeys();
    const halfCheckedKeys = treeRef.value.getHalfCheckedKeys();

    // 合并完全选中和半选中的节点
    const allKeys = [...checkedKeys, ...halfCheckedKeys];

    // 过滤掉虚拟节点（负数ID），只保留真实权限ID
    const realPermissionIds = allKeys.filter((id: number) => id > 0);

    if (realPermissionIds.length === 0) {
      ElMessage.warning('请选择至少一个权限');
      return;
    }

    loading.value = true;

    try {
      const { error } = await fetchAssignRolePermissions(props.roleId, {
        permissionIds: realPermissionIds
      });

      loading.value = false;

      if (!error) {
        ElNotification({
          title: '分配成功',
          message: `成功为角色分配 ${realPermissionIds.length} 个权限`,
          type: 'success',
          duration: 3000,
          position: 'top-right'
        });
        closeModal();
      }
    } catch (err) {
      loading.value = false;
      console.error('❌ 权限分配失败:', err);
    }
  }

  // 监听模态框显示
  watch(
    () => visible.value,
    async val => {
      if (val && props.roleId) {
        await getAllPermissions();
        await getRolePermissions();
      }
    }
  );

  // 权限级别标签类型
  function getLevelTagType(level: number): '' | 'success' | 'warning' | 'info' | 'danger' {
    const types: Record<number, '' | 'success' | 'warning' | 'info' | 'danger'> = {
      1: '',
      2: 'success',
      3: 'warning',
      4: 'info'
    };
    return types[level] || 'info';
  }

  // 权限级别文本
  function getLevelText(level: number): string {
    const texts: Record<number, string> = {
      1: '模块',
      2: '页面',
      3: '按钮',
      4: 'API'
    };
    return texts[level] || '未知';
  }

  // 模块名称映射
  function getModuleLabel(module: string): string {
    const moduleMap: Record<string, string> = {
      system: '系统管理',
      auth: '授权中心',
      cmdb: '资产管理',
      monitoring: '监控中心',
      k8s: 'K8s管理',
      audit: '审计中心'
    };
    return moduleMap[module] || module;
  }

  // 资源（页面）名称映射 - 从权限数据中动态提取
  function getResourceLabel(resource: string, permissions: Api.SystemManage.Permission[]): string {
    // 查找该资源的第一个权限，从中提取资源中文名称
    const firstPerm = permissions.find(p => p.resource === resource);
    if (!firstPerm) {
      console.warn('⚠️ 未找到资源权限:', resource);
      return resource;
    }

    // 从权限名称中提取资源名称
    // "查看用户" → "用户管理"
    // "创建角色" → "角色管理"
    let resourceName = firstPerm.name;

    // 移除常见的操作前缀
    const allPrefixes = [
      '查看',
      '列表',
      '查询',
      '创建',
      '新增',
      '添加',
      '编辑',
      '修改',
      '更新',
      '删除',
      '移除',
      '重置',
      '分配'
    ];

    for (const prefix of allPrefixes) {
      if (resourceName.startsWith(prefix)) {
        resourceName = resourceName.substring(prefix.length);
        break;
      }
    }

    // 如果提取到了有意义的名称，返回"XXX管理"格式
    if (resourceName && resourceName !== firstPerm.name && resourceName.trim().length > 0) {
      const result = `${resourceName.trim()}管理`;
      return result;
    }

    console.warn('⚠️ 无法提取，使用原始标识符:', resource);
    return resource;
  }

  // ElTree 节点选中状态变化事件处理
  function handleCheckChange() {
    // 触发统计数字更新
    refreshCounter.value++;
  }

  // 查找节点级别
  function findNodeLevel(nodes: PermissionNode[], targetId: number): number | null {
    let result: number | null = null;

    function traverse(nodeList: PermissionNode[]) {
      if (result !== null) return; // 已找到，停止搜索

      nodeList.forEach(node => {
        if (node.id === targetId) {
          result = node.level;
          return;
        }

        if (node.children && node.children.length > 0) {
          traverse(node.children);
        }
      });
    }

    traverse(nodes);
    return result;
  }
</script>

<template>
  <ElDialog v-model="visible" :title="title" width="900px" :close-on-click-modal="false">
    <div v-loading="loading" class="permission-assign-container">
      <div class="tip-text">
        <span v-if="props.viewOnly">查看角色拥有的权限。权限按模块 → 页面 → 按钮层级组织。</span>
        <span v-else>勾选权限节点，为角色分配相应的操作权限。权限按模块 → 页面 → 按钮层级组织。</span>
      </div>

      <div class="permission-stats">
        <ElText type="info">
          已选择
          <ElText type="primary">{{ getSelectedCount() }}</ElText>
          个权限
        </ElText>
      </div>

      <ElTree
        ref="treeRef"
        :data="permissionTree"
        node-key="id"
        :show-checkbox="!props.viewOnly"
        :props="{
          label: 'label',
          children: 'children'
        }"
        class="permission-tree"
        :default-expand-all="true"
        :indent="24"
        @check-change="!props.viewOnly ? handleCheckChange : undefined"
      >
        <template #default="{ data }">
          <span class="custom-tree-node">
            <span class="node-label">{{ data.label }}</span>
            <ElTag v-if="data.level" :type="getLevelTagType(data.level)" size="small" class="level-tag">
              {{ getLevelText(data.level) }}
            </ElTag>
            <span class="node-code">{{ data.code }}</span>
          </span>
        </template>
      </ElTree>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <ElButton @click="closeModal">{{ props.viewOnly ? '关闭' : '取消' }}</ElButton>
        <ElButton v-if="!props.viewOnly" type="primary" :loading="loading" @click="handleSubmit">确定</ElButton>
      </div>
    </template>
  </ElDialog>
</template>

<style scoped lang="scss">
  .permission-assign-container {
    padding: 16px 0;
  }

  .tip-text {
    padding: 12px 16px;
    background: var(--el-fill-color-light);
    border-radius: 4px;
    margin-bottom: 16px;
    color: var(--el-text-color-secondary);
    font-size: 14px;
  }

  .permission-stats {
    padding: 12px 16px;
    background: var(--el-fill-color);
    border-radius: 4px;
    margin-bottom: 16px;
  }

  .permission-tree {
    max-height: 500px;
    overflow-y: auto;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 4px;
    padding: 12px;

    :deep(.el-tree-node__content) {
      height: 32px;
    }

    :deep(.custom-tree-node) {
      display: flex;
      align-items: center;
      gap: 8px;
      flex: 1;
      font-size: 14px;
    }

    :deep(.node-label) {
      font-weight: 500;
    }

    :deep(.level-tag) {
      margin-left: 8px;
    }

    :deep(.node-code) {
      color: var(--el-text-color-secondary);
      font-size: 12px;
      font-family: 'Courier New', monospace;
      margin-left: auto;
    }
  }

  .dialog-footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
  }
</style>
