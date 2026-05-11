<script setup lang="ts">
import { computed, shallowRef, watch, nextTick } from 'vue';
import { fetchGetMenuTree, fetchUpdateRole } from '@/service/api';
import { $t } from '@/locales';

defineOptions({ name: 'MenuAuthModal' });

interface Props {
  /** the roleId */
  roleId: number;
  /** the role data */
  roleData?: Api.SystemManage.Role | null;
}

const props = defineProps<Props>();

interface Emits {
  (e: 'submitted'): void;
}

const emit = defineEmits<Emits>();

const visible = defineModel<boolean>('visible', {
  default: false
});

function closeModal() {
  visible.value = false;
}

const title = computed(() => $t('common.edit') + $t('page.manage.role.menuAuth'));

const tree = shallowRef<any[]>([]);
const treeRef = shallowRef<any>(null);

async function getTree() {
  const { error, data } = await fetchGetMenuTree();

  if (!error && data && Array.isArray(data)) {
    tree.value = data;
  } else {
    console.error('获取菜单树失败:', error);
    tree.value = [];
  }
}

const checks = shallowRef<number[]>([]);

async function getChecks() {
  if (!props.roleData) {
    console.error('角色数据为空');
    checks.value = [];
    return;
  }

  try {
    // 后端返回的menuIds是JSON字符串，需要解析
    let menuIds: number[] = [];

    if (props.roleData.menuIds && props.roleData.menuIds !== 'null') {
      const parsed = JSON.parse(props.roleData.menuIds);
      menuIds = Array.isArray(parsed) ? parsed : [];
    }

    // 等待Tree组件完全渲染
    await nextTick();

    // 从完整的菜单ID列表中提取叶子节点ID
    // Element Plus Tree 在 check-strictly="false" 模式下
    // setCheckedKeys 需要传入叶子节点的ID数组
    const leafKeys = getLeafKeysFromTree(tree.value, menuIds);
    checks.value = leafKeys;

    // 再次等待DOM更新，确保Tree组件已经渲染完成
    await nextTick();
    treeRef.value?.setCheckedKeys(leafKeys);
  } catch (e) {
    console.error('解析menuIds失败:', e);
    checks.value = [];
    await nextTick();
    treeRef.value?.setCheckedKeys([]);
  }
}

// 从完整的菜单ID列表中提取叶子节点ID
function getLeafKeysFromTree(treeData: any[], selectedIds: number[]) {
  const leafKeys: number[] = [];

  const traverse = (nodes: any[]) => {
    nodes.forEach(node => {
      const isParent = node.children && node.children.length > 0;
      const isSelected = selectedIds.includes(node.id);

      if (isParent) {
        // 如果是父节点，递归处理子节点
        traverse(node.children);
      } else if (isSelected) {
        // 如果是叶子节点且被选中，添加到列表
        leafKeys.push(node.id);
      }
    });
  };

  traverse(treeData);
  return leafKeys;
}

async function handleSubmit() {
  if (!treeRef.value) return;

  // 获取完全选中的节点（叶子节点）
  const checkedKeys = treeRef.value.getCheckedKeys() || [];
  // 获取半选中的节点（父节点）
  const halfCheckedKeys = treeRef.value.getHalfCheckedKeys() || [];
  // 合并所有选中的节点
  const allKeys = [...checkedKeys, ...halfCheckedKeys];

  const { error } = await fetchUpdateRole(props.roleId, {
    menuIds: allKeys
  });

  if (!error) {
    window.$message?.success($t('common.modifySuccess'));
    closeModal();
    emit('submitted');
  } else {
    window.$message?.error('保存权限失败');
  }
}

async function init() {
  // 先加载菜单树数据
  await getTree();
  // 菜单树加载完成后再设置选中状态
  await getChecks();
  // 额外延迟，确保Tree组件完全渲染
  await new Promise(resolve => setTimeout(resolve, 100));
}

watch(visible, async val => {
  if (val) {
    await init();
  }
});
</script>

<template>
  <ElDialog v-model="visible" :title="title" preset="card" class="w-480px">
    <ElTree
      ref="treeRef"
      :data="tree"
      node-key="id"
      show-checkbox
      default-expand-all
      :props="{ label: 'name', children: 'children' }"
      :check-strictly="false"
      class="h-280px overflow-y-auto"
    />
    <template #footer>
      <ElSpace class="w-full justify-end">
        <ElButton size="small" class="mt-16px" @click="closeModal">
          {{ $t('common.cancel') }}
        </ElButton>
        <ElButton type="primary" size="small" class="mt-16px" @click="handleSubmit">
          {{ $t('common.confirm') }}
        </ElButton>
      </ElSpace>
    </template>
  </ElDialog>
</template>

<style scoped></style>
