<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue';
import { fetchCreateMenu, fetchGetMenuTree, fetchUpdateMenu } from '@/service/api';
import { useForm, useFormRules } from '@/hooks/common/form';
import { $t } from '@/locales';

defineOptions({ name: 'MenuOperateDrawer' });

interface Props {
  operateType: UI.TableOperateType;
  rowData?: Api.SystemManage.Menu | null;
  isAddingChild?: boolean;
  parentMenuName?: string;
}

const props = defineProps<Props>();

interface Emits {
  (e: 'submitted'): void;
}

const emit = defineEmits<Emits>();

const visible = defineModel<boolean>('visible', {
  default: false
});

const { formRef, validate, restoreValidation } = useForm();
const { defaultRequiredRule } = useFormRules();

// 菜单树数据（树形结构）
const menuTreeData = ref<Api.SystemManage.MenuTree[]>([]);
const menuTree = ref<Api.SystemManage.MenuTree[]>([]);
const loadingMenuTree = ref(false);

async function loadMenuTree() {
  loadingMenuTree.value = true;
  try {
    const { error, data } = await fetchGetMenuTree();
    if (!error && data && Array.isArray(data)) {
      // 直接存储树形结构数据
      menuTreeData.value = data as Api.SystemManage.MenuTree[];

      // 获取当前编辑菜单的子孙节点ID（编辑模式下）
      const excludedIds = new Set<number>();
      if (isEdit.value && props.rowData) {
        const collectChildren = (node: Api.SystemManage.MenuTree) => {
          excludedIds.add(node.id);
          if (node.children && node.children.length > 0) {
            node.children.forEach(child => collectChildren(child));
          }
        };

        // 在树中找到当前编辑节点并递归收集子孙节点
        const findAndCollect = (nodes: Api.SystemManage.MenuTree[]): boolean => {
          for (const node of nodes) {
            if (node.id === props.rowData!.id) {
              collectChildren(node);
              return true;
            }
            if (node.children && node.children.length > 0) {
              if (findAndCollect(node.children)) return true;
            }
          }
          return false;
        };

        findAndCollect(menuTreeData.value);
      }

      // 构建用于父级菜单选择的树形结构（排除当前编辑节点及其子孙）
      const buildTreeForSelect = (nodes: Api.SystemManage.MenuTree[]): Api.SystemManage.MenuTree[] => {
        return nodes
          .filter(node => !excludedIds.has(node.id))
          .map(node => ({
            ...node,
            children: node.children ? buildTreeForSelect(node.children) : []
          }));
      };

      const tree = buildTreeForSelect(menuTreeData.value);

      // 添加根菜单选项
      menuTree.value = [{ id: 0, name: '作为一级菜单', parentId: 0, children: tree }] as Api.SystemManage.MenuTree[];
    }
  } finally {
    loadingMenuTree.value = false;
  }
}

const title = computed(() => {
  if (props.isAddingChild && props.parentMenuName) {
    return `添加子菜单 - ${props.parentMenuName}`;
  }
  const titles: Record<UI.TableOperateType, string> = {
    add: $t('page.manage.menu.addMenu'),
    edit: $t('page.manage.menu.editMenu')
  };
  return titles[props.operateType];
});

const menuId = computed(() => props.rowData?.id || -1);

const isEdit = computed(() => props.operateType === 'edit');

type Model = {
  name: string;
  icon: string;
  path: string;
  permission: string;
  menuType: string;
  parentId: number;
  sort: number;
  status: number;
};

const model = ref<Model>({
  name: '',
  icon: '',
  path: '',
  permission: '',
  menuType: 'menu',
  parentId: 0,
  sort: 1,
  status: 1
});

type RuleKey = 'name' | 'path' | 'permission' | 'menuType' | 'sort' | 'status';

const rules: Record<RuleKey, App.Global.FormRule> = {
  name: defaultRequiredRule,
  path: defaultRequiredRule,
  permission: defaultRequiredRule,
  menuType: defaultRequiredRule,
  sort: defaultRequiredRule,
  status: defaultRequiredRule
};

function handleInitModel() {

  if (isEdit.value && props.rowData) {
    // 编辑模式：填充数据
    model.value = {
      name: props.rowData.name || '',
      icon: props.rowData.icon || '',
      path: props.rowData.path || '',
      permission: props.rowData.permission || '',
      menuType: props.rowData.menuType || 'menu',
      parentId: props.rowData.parentId ?? 0,
      sort: props.rowData.sort ?? 1,
      status: props.rowData.status ?? 1
    };
  } else {
    // 新增模式：使用空表单
    // 如果是添加子菜单，从 rowData（父菜单数据）中获取父菜单ID
    // 否则从 rowData 中获取父菜单ID（正常新增时可能用户手动选择父级）
    const parent_Id = props.isAddingChild && props.rowData ? props.rowData.id : (props.rowData?.parentId ?? 0);
    const nextSort = props.isAddingChild ? getNextSort() : 1;

    model.value = {
      name: '',
      icon: '',
      path: '',
      permission: '',
      menuType: props.isAddingChild ? 'menu' : 'directory',
      parentId: parent_Id,
      sort: nextSort,
      status: 1
    };
  }
}

// 计算下一个排序值
function getNextSort(): number {
  // 当添加子菜单时，props.rowData 是父菜单的完整数据
  // 从 id 字段获取父菜单ID
  const parentId = props.rowData?.id;


  if (!parentId) return 1;

  // 如果菜单数据还没加载完成，返回默认值
  if (!menuTreeData.value || menuTreeData.value.length === 0) {
    console.warn('⚠️ [计算排序] 菜单数据未加载');
    return 1;
  }

  // 递归遍历树形结构，查找指定父菜单下的所有子菜单
  const findSiblings = (nodes: Api.SystemManage.MenuTree[], targetParentId: number): Api.SystemManage.MenuTree[] => {
    const siblings: Api.SystemManage.MenuTree[] = [];

    const traverse = (nodes: Api.SystemManage.MenuTree[]) => {
      for (const node of nodes) {
        // 找到目标父菜单的子菜单
        if (node.id == targetParentId && node.children && node.children.length > 0) {
          siblings.push(...node.children);
        }
        // 递归遍历子节点
        if (node.children && node.children.length > 0) {
          traverse(node.children);
        }
      }
    };

    traverse(nodes);
    return siblings;
  };

  const siblings = findSiblings(menuTreeData.value as Api.SystemManage.MenuTree[], parentId);


  if (siblings.length === 0) return 1;

  // 找到最大排序值并加1
  const maxSort = Math.max(...siblings.map(m => m.sort || 0));
  const nextSort = maxSort + 1;


  return nextSort;
}

function closeDrawer() {
  visible.value = false;
}

async function handleSubmit() {
  await validate();

  const { error } = isEdit.value
    ? await fetchUpdateMenu(menuId.value, model.value)
    : await fetchCreateMenu(model.value);

  if (!error) {
    window.$message?.success(isEdit.value ? $t('common.updateSuccess') : '添加成功');
    closeDrawer();
    emit('submitted');
  } else {
    window.$message?.error(isEdit.value ? '更新失败' : '添加失败');
  }
}

watch(
  () => visible.value,
  async newVal => {

    if (newVal) {
      await nextTick();
      await loadMenuTree();
      handleInitModel();
      restoreValidation();
    }
  },
  { immediate: false }
);

// 移除重复的 watch，合并到 visible watch 中
</script>

<template>
  <ElDrawer v-model="visible" :title="title" :size="450">
    <ElForm ref="formRef" :model="model" :rules="rules" label-position="top">
      <ElFormItem label="菜单名称" prop="name">
        <ElInput v-model="model.name" placeholder="请输入菜单名称" />
      </ElFormItem>
      <ElFormItem label="图标" prop="icon">
        <ElInput v-model="model.icon" placeholder="请输入图标名称，如：mdi:home" />
      </ElFormItem>
      <ElFormItem label="路由路径" prop="path">
        <ElInput v-model="model.path" placeholder="请输入路由路径，如：/system/menus" />
      </ElFormItem>
      <ElFormItem label="权限标识" prop="permission">
        <ElInput v-model="model.permission" placeholder="请输入权限标识，如：menu:system:menus" />
      </ElFormItem>
      <ElFormItem label="菜单类型" prop="menuType">
        <ElSelect v-model="model.menuType" placeholder="请选择菜单类型" class="w-full">
          <ElOption label="菜单" value="menu" />
          <ElOption label="目录" value="directory" />
        </ElSelect>
      </ElFormItem>
      <ElFormItem v-if="!isAddingChild" label="父级菜单" prop="parentId">
        <ElTreeSelect
          v-model="model.parentId"
          :data="menuTree"
          :props="{ label: 'name', value: 'id', children: 'children' }"
          :render-after-expand="false"
          check-strictly
          placeholder="选择父级菜单"
          :loading="loadingMenuTree"
          class="w-full"
          clearable
        />
      </ElFormItem>
      <ElFormItem v-else label="父级菜单">
        <ElInput :value="parentMenuName" disabled class="w-full" />
      </ElFormItem>
      <ElFormItem label="排序" prop="sort">
        <ElInputNumber v-model="model.sort" :min="1" :max="999" placeholder="请输入排序序号" class="w-full" />
      </ElFormItem>
      <ElFormItem label="状态" prop="status">
        <ElRadioGroup v-model="model.status">
          <ElRadio :value="1">启用</ElRadio>
          <ElRadio :value="0">禁用</ElRadio>
        </ElRadioGroup>
      </ElFormItem>
    </ElForm>
    <template #footer>
      <ElSpace :size="16">
        <ElButton @click="closeDrawer">{{ $t('common.cancel') }}</ElButton>
        <ElButton type="primary" @click="handleSubmit">{{ $t('common.confirm') }}</ElButton>
      </ElSpace>
    </template>
  </ElDrawer>
</template>

<style scoped></style>
