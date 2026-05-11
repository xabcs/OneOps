<script setup lang="ts">
import { computed, ref, watch, nextTick } from 'vue';
import { fetchGetAllRoles, fetchGetMenuTree, fetchCreateUser, fetchUpdateUser } from '@/service/api';
import { useForm, useFormRules } from '@/hooks/common/form';
import { $t } from '@/locales';

defineOptions({ name: 'UserOperateDrawer' });

interface Props {
  /** the type of operation */
  operateType: UI.TableOperateType;
  /** the edit row data */
  rowData?: Api.SystemManage.User | null;
  /** all roles with menu permissions */
  allRoles?: Api.SystemManage.AllRole[];
}

const props = defineProps<Props>();

interface Emits {
  (e: 'submitted'): void;
}

const emit = defineEmits<Emits>();

const visible = defineModel<boolean>('visible', {
  default: false
});

// 动态生成用户名输入框的 autocomplete，防止浏览器自动填充
const usernameAutocomplete = computed(() => {
  return isEdit.value ? 'off' : 'new-username';
});

// 动态生成密码输入框的 name
const passwordInputName = computed(() => {
  return `new-password-${Date.now()}`;
});

function closeDrawer() {
  visible.value = false;
}

const { formRef, validate, restoreValidation } = useForm();
const { defaultRequiredRule } = useFormRules();

const title = computed(() => {
  const titles: Record<UI.TableOperateType, string> = {
    add: $t('page.manage.user.addUser'),
    edit: $t('page.manage.user.editUser')
  };
  return titles[props.operateType];
});

const userId = computed(() => props.rowData?.id || -1);

const isEdit = computed(() => props.operateType === 'edit');

type Model = {
  username: string;
  nickname: string;
  email: string;
  roleIds: number[];
  status: string;
  homePath: string;
  password?: string;
};

const model = ref(createDefaultModel());

function createDefaultModel(): Model {
  return {
    username: '',
    nickname: '',
    email: '',
    roleIds: [],
    status: 'active',
    homePath: '/',
    password: ''
  };
}

type RuleKey = 'username' | 'status' | 'password';

const rules = computed(() => {
  const baseRules: Record<string, App.Global.FormRule> = {
    username: defaultRequiredRule,
    status: defaultRequiredRule
  };

  // 只在新增模式下验证密码字段
  if (!isEdit.value) {
    baseRules.password = defaultRequiredRule;
  }

  return baseRules;
});

/** the enabled role options */
const roleOptions = ref<CommonType.Option<number>[]>([]);

/** the home directory options */
const homePathOptions = ref<CommonType.Option<string>[]>([]);

async function getRoleOptions() {
  const { error, data } = await fetchGetAllRoles();

  if (!error && data) {
    const options = data.map(item => ({
      label: item.name,
      value: item.id
    }));

    roleOptions.value = options;
  }
}

async function getHomePathOptions() {
  // 获取所有菜单
  const { error, data } = await fetchGetMenuTree();

  if (!error && data) {
    const options: CommonType.Option<string>[] = [
      { label: '首页（仪表盘）', value: '/' }
    ];

    // 获取当前用户的角色ID列表
    const userRoleIds = model.value.roleIds || [];

    if (userRoleIds.length === 0) {
      // 没有角色时，只能选择首页
      homePathOptions.value = options;
      return;
    }

    // 从所有角色中找到用户拥有的角色，收集菜单权限
    const permittedMenuIds = new Set<number>();
    const roleMap = new Map(props.allRoles?.map(r => [r.id, r]) || []);

    userRoleIds.forEach(roleId => {
      const role = roleMap.get(roleId);
      if (role && role.menuIds) {
        // menuIds 是字符串格式 "[1,2,3]"
        try {
          const menuIds = JSON.parse(role.menuIds);
          menuIds.forEach((id: number) => permittedMenuIds.add(id));
        } catch (e) {
          console.error('解析角色菜单ID失败:', e);
        }
      }
    });

    // 遍历菜单树，添加有权限的一级菜单选项
    data.forEach((menu: Api.SystemManage.Menu) => {
      // 只添加一级菜单（parentId === 0）且有权限的菜单
      if (menu.parentId === 0 && menu.status === 1 && menu.path && permittedMenuIds.has(menu.id)) {
        options.push({
          label: menu.name,
          value: menu.path
        });
      }
    });

    homePathOptions.value = options;
  }
}

function handleInitModel() {
  console.log('🔧 [handleInitModel] 开始执行', {
    isEdit: isEdit.value,
    operateType: props.operateType,
    hasRowData: !!props.rowData,
    rowData: props.rowData,
    rowDataUsername: props.rowData?.username
  });

  // 🔥 关键修复：直接使用 props.operateType 判断，避免 computed 的时序问题
  const isEditMode = props.operateType === 'edit' && props.rowData;

  if (isEditMode) {
    // 编辑模式：填充用户数据
    console.log('✏️ [handleInitModel] 使用编辑模式，rowData:', props.rowData);

    // 🔥 关键修复：先清空，再赋值，确保响应式更新
    const newData = {
      username: props.rowData.username || '',
      nickname: props.rowData.nickname || '',
      email: props.rowData.email || '',
      roleIds: props.rowData.roleIds || [],
      status: props.rowData.status || 'active',
      homePath: props.rowData.homePath || '/',
      password: ''
    };

    model.value = { ...newData };

    console.log('📝 [handleInitModel] 编辑模式赋值后的 model:', {
      username: model.value.username,
      'model.value.username': model.value.username,
      fullModel: model.value
    });
  } else {
    // 新增模式或无数据：使用空表单
    console.log('➕ [handleInitModel] 使用新增模式，清空表单');
    model.value = {
      username: '',
      nickname: '',
      email: '',
      roleIds: [],
      status: 'active',
      homePath: '/',
      password: ''
    };
  }

  console.log('📝 [handleInitModel] 最终model值:', model.value);
}

async function handleSubmit() {
  await validate();

  // 准备提交数据
  const submitData = isEdit.value
    ? {
        nickname: model.value.nickname,
        email: model.value.email,
        roleIds: model.value.roleIds,
        status: model.value.status,
        homePath: model.value.homePath
      }
    : {
        username: model.value.username,
        nickname: model.value.nickname,
        email: model.value.email,
        roleIds: model.value.roleIds,
        status: model.value.status,
        homePath: model.value.homePath,
        password: model.value.password
      };

  const { error } = isEdit.value
    ? await fetchUpdateUser(userId.value, submitData)
    : await fetchCreateUser(submitData as any);

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
    console.log('🔍 [用户操作抽屉] visible变化', {
      visible: visible.value,
      newVal,
      operateType: props.operateType,
      isEdit: isEdit.value,
      rowData: props.rowData
    });

    if (newVal) {
      // 抽屉打开时，初始化数据
      await nextTick();

      handleInitModel();
      restoreValidation();
      getRoleOptions();
      getHomePathOptions();
    }
  }
);

// 监听 rowData 变化，重新初始化表单
watch(
  () => [props.rowData, props.operateType] as const,
  async () => {
    if (visible.value && props.rowData) {
      console.log('🔍 [用户操作抽屉] rowData变化，重新初始化');
      await nextTick();
      handleInitModel();
      restoreValidation();
      getHomePathOptions();
    }
  }
);

// 监听角色变化，动态更新家目录选项
watch(
  () => model.value.roleIds,
  () => {
    if (visible.value) {
      getHomePathOptions();
      // 如果当前家目录不在新选项中，重置为首页
      const availablePaths = homePathOptions.value.map(opt => opt.value);
      if (!availablePaths.includes(model.value.homePath)) {
        model.value.homePath = '/';
      }
    }
  },
  { deep: true }
);
</script>

<template>
  <ElDrawer v-model="visible" :title="title" :size="360">
    <ElForm ref="formRef" :model="model" :rules="rules" label-position="top" autocomplete="off">
      <ElFormItem :label="$t('page.manage.user.userName')" prop="username">
        <ElInput
          v-model="model.username"
          :placeholder="$t('page.manage.user.form.userName')"
          :disabled="isEdit"
          :autocomplete="usernameAutocomplete"
        />
      </ElFormItem>
      <ElFormItem v-if="!isEdit" label="密码" prop="password">
        <ElInput
          v-model="model.password"
          type="password"
          placeholder="请输入密码"
          show-password
          :name="passwordInputName"
          autocomplete="new-password"
        />
      </ElFormItem>
      <ElFormItem :label="$t('page.manage.user.nickName')" prop="nickname">
        <ElInput v-model="model.nickname" :placeholder="$t('page.manage.user.form.nickName')" autocomplete="off" />
      </ElFormItem>
      <ElFormItem :label="$t('page.manage.user.userEmail')" prop="email">
        <ElInput v-model="model.email" :placeholder="$t('page.manage.user.form.userEmail')" autocomplete="off" name="new-email" />
      </ElFormItem>
      <ElFormItem :label="$t('page.manage.user.userStatus')" prop="status">
        <ElRadioGroup v-model="model.status">
          <ElRadio value="active">启用</ElRadio>
          <ElRadio value="inactive">禁用</ElRadio>
        </ElRadioGroup>
      </ElFormItem>
      <ElFormItem :label="$t('page.manage.user.userRole')" prop="roleIds">
        <ElSelect v-model="model.roleIds" multiple :placeholder="$t('page.manage.user.form.userRole')" class="w-full">
          <ElOption v-for="{ label, value } in roleOptions" :key="value" :label="label" :value="value" />
        </ElSelect>
      </ElFormItem>
      <ElFormItem label="家目录" prop="homePath">
        <ElSelect v-model="model.homePath" placeholder="请选择家目录" class="w-full">
          <ElOption v-for="{ label, value } in homePathOptions" :key="value" :label="label" :value="value" />
        </ElSelect>
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
