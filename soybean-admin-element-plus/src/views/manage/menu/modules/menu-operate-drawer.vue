<script setup lang="ts">
import { computed, ref, watch, nextTick } from 'vue';
import { useForm, useFormRules } from '@/hooks/common/form';
import { $t } from '@/locales';
import { fetchCreateMenu, fetchUpdateMenu, fetchGetMenuTree } from '@/service/api';

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

// 菜单树数据
const menuTreeData = ref<Api.SystemManage.Menu[]>([]);
const menuTree = ref<Api.SystemManage.MenuTree[]>([]);
const loadingMenuTree = ref(false);

async function loadMenuTree() {
  loadingMenuTree.value = true;
  try {
    const { error, data } = await fetchGetMenuTree();
    if (!error && data && Array.isArray(data)) {
      menuTreeData.value = data as Api.SystemManage.Menu[];

      // 获取当前编辑菜单的子孙节点ID（编辑模式下）
      const excludedIds = new Set<number>();
      if (isEdit.value && props.rowData) {
        const collectChildren = (menuId: number) => {
          excludedIds.add(menuId);
          const children = menuTreeData.value.filter(m => m.parentId === menuId);
          children.forEach(child => collectChildren(child.id));
        };
        collectChildren(props.rowData.id);
      }

      // 构建树形结构，排除当前编辑节点及其子孙
      const buildTree = (menus: Api.SystemManage.Menu[], parentId = 0): Api.SystemManage.MenuTree[] => {
        return menus
          .filter(menu => menu.parentId === parentId && !excludedIds.has(menu.id))
          .map(menu => ({
            id: menu.id,
            name: menu.name,
            path: menu.path,
            icon: menu.icon,
            permission: menu.permission,
            parentId: menu.parentId,
            sort: menu.sort,
            status: menu.status,
            children: buildTree(menus, menu.id)
          }));
      };

      const tree = buildTree(menuTreeData.value);

      // 添加根菜单选项
      menuTree.value = [
        { id: 0, name: '作为一级菜单', parentId: 0, children: tree }
      ] as Api.SystemManage.MenuTree[];
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
  parentId: number;
  sort: number;
  status: number;
};

const model = ref<Model>({
  name: '',
  icon: '',
  path: '',
  permission: '',
  parentId: 0,
  sort: 1,
  status: 1
});

type RuleKey = 'name' | 'path' | 'permission' | 'sort' | 'status';

const rules: Record<RuleKey, App.Global.FormRule> = {
  name: defaultRequiredRule,
  path: defaultRequiredRule,
  permission: defaultRequiredRule,
  sort: defaultRequiredRule,
  status: defaultRequiredRule
};

function handleInitModel() {
  console.log('🔧 [菜单操作抽屉] handleInitModel', {
    operateType: props.operateType,
    isEdit: isEdit.value,
    hasRowData: !!props.rowData,
    rowData: props.rowData
  });

  if (isEdit.value && props.rowData) {
    // 编辑模式：填充数据
    model.value = {
      name: props.rowData.name || '',
      icon: props.rowData.icon || '',
      path: props.rowData.path || '',
      permission: props.rowData.permission || '',
      parentId: props.rowData.parentId ?? 0,
      sort: props.rowData.sort ?? 1,
      status: props.rowData.status ?? 1
    };
    console.log('✏️ [菜单操作抽屉] 编辑模式数据', model.value);
  } else {
    // 新增模式：使用空表单
    model.value = {
      name: '',
      icon: '',
      path: '',
      permission: '',
      parentId: props.isAddingChild ? (props.rowData?.id ?? 0) : 0,
      sort: 1,
      status: 1
    };
    console.log('➕ [菜单操作抽屉] 新增模式数据', model.value);
  }
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
  async (newVal) => {
    console.log('🔍 [菜单操作抽屉] visible变化', {
      visible: visible.value,
      newVal,
      operateType: props.operateType,
      isEdit: isEdit.value,
      rowData: props.rowData
    });

    if (newVal) {
      await nextTick();
      await loadMenuTree();
      handleInitModel();
      restoreValidation();
    }
  },
  { immediate: false }
);

// 监听 rowData 和 operateType 变化
watch(
  () => props.rowData,
  async (newRowData) => {
    console.log('🔍 [菜单操作抽屉] rowData变化', {
      visible: visible.value,
      operateType: props.operateType,
      isEdit: isEdit.value,
      hasRowData: !!newRowData,
      rowData: newRowData
    });

    if (visible.value && newRowData) {
      await nextTick();
      handleInitModel();
      restoreValidation();
    }
  },
  { immediate: true }
);
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
      <ElFormItem label="父级菜单" prop="parentId">
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
