<script setup lang="ts">
  import { computed, nextTick, ref, watch } from 'vue';
  import { fetchCreateUser, fetchGetRoleMenuPaths, fetchUpdateUser } from '@/service/api';
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

  // @ts-expect-error vue-tsc noUnusedLocals: template ref
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

  // admin 是内置超管入口，编辑时禁止修改其状态
  const isAdminUser = computed(() => isEdit.value && props.rowData?.username === 'admin');

  type Model = {
    username: string;
    nickname: string;
    email: string;
    phone: string;
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
      phone: '',
      roleIds: [],
      status: 'active',
      homePath: '',
      password: ''
    };
  }

  const rules = computed(() => {
    const baseRules: Record<string, App.Global.FormRule> = {
      username: defaultRequiredRule,
      status: defaultRequiredRule,
      homePath: defaultRequiredRule
    };

    // 只在新增模式下验证密码字段
    if (!isEdit.value) {
      baseRules.password = defaultRequiredRule;
    }

    return baseRules;
  });

  /** the enabled role options */
  const roleOptions = computed<CommonType.Option<number>[]>(() => {
    if (!props.allRoles || props.allRoles.length === 0) {
      return [];
    }

    return props.allRoles.map(role => ({
      label: role.name,
      value: role.id
    }));
  });

  /** the home directory options */
  const homePathOptions = ref<CommonType.Option<string>[]>([]);

  /** 按选中角色实时拉取可见叶子菜单（后端按 角色→权限→菜单 推导，与登录后菜单一致）
   *  家目录只能从有权限的菜单中选：首页（/home）仅在有首页权限时出现在候选中 */
  async function getHomePathOptions() {
    const options: CommonType.Option<string>[] = [];

    const roleIds = model.value.roleIds || [];
    if (roleIds.length > 0) {
      const { error, data } = await fetchGetRoleMenuPaths(roleIds);
      if (!error && data) {
        data.forEach(menu => {
          options.push({ label: menu.name, value: menu.path });
        });
      }
    }

    homePathOptions.value = options;

    // 当前家目录不在候选（含存量非法值如 "/"）时，重置为第一个可见菜单；
    // 无任何可见菜单时置空，提交前由必选规则拦截
    const availablePaths = options.map(opt => opt.value);
    if (!availablePaths.includes(model.value.homePath)) {
      model.value.homePath = availablePaths[0] ?? '';
    }
  }

  function handleInitModel() {
    // 🔥 关键修复：直接使用 props.operateType 判断，避免 computed 的时序问题
    const isEditMode = props.operateType === 'edit' && props.rowData;

    if (isEditMode) {
      // 编辑模式：填充用户数据

      // 🔥 关键修复：先清空，再赋值，确保响应式更新
      const newData = {
        username: props.rowData.username || '',
        nickname: props.rowData.nickname || '',
        email: props.rowData.email || '',
        phone: props.rowData.phone || '',
        roleIds: props.rowData.roleIds || [],
        status: props.rowData.status || 'active',
        homePath: props.rowData.homePath || '',
        password: ''
      };

      model.value = { ...newData };
    } else {
      // 新增模式或无数据：使用空表单
      model.value = {
        username: '',
        nickname: '',
        email: '',
        phone: '',
        roleIds: [],
        status: 'active',
        homePath: '',
        password: ''
      };
    }
  }

  async function handleSubmit() {
    await validate();

    // 准备提交数据
    const submitData = isEdit.value
      ? {
          nickname: model.value.nickname,
          email: model.value.email,
          phone: model.value.phone,
          roleIds: model.value.roleIds,
          status: model.value.status,
          homePath: model.value.homePath
        }
      : {
          username: model.value.username,
          nickname: model.value.nickname,
          email: model.value.email,
          phone: model.value.phone,
          roleIds: model.value.roleIds,
          status: model.value.status,
          homePath: model.value.homePath,
          password: model.value.password
        };

    const { error } = isEdit.value
      ? await fetchUpdateUser(userId.value, submitData)
      : await fetchCreateUser(submitData);

    if (!error) {
      window.$message?.success(isEdit.value ? $t('common.updateSuccess') : '添加成功');
      closeDrawer();
      emit('submitted');
    }
  }

  watch(
    () => visible.value,
    async newVal => {
      if (newVal) {
        // 抽屉打开时，初始化数据
        await nextTick();

        handleInitModel();
        restoreValidation();
        getHomePathOptions();
      }
    }
  );

  // 监听 rowData 变化，重新初始化表单
  watch(
    () => [props.rowData, props.operateType] as const,
    async () => {
      if (visible.value && props.rowData) {
        await nextTick();
        handleInitModel();
        restoreValidation();
        getHomePathOptions();
      }
    }
  );

  // 监听角色变化，动态更新家目录选项（含非法值重置）
  watch(
    () => model.value.roleIds,
    async () => {
      if (!visible.value) return;
      await getHomePathOptions();
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
        <ElInput
          v-model="model.email"
          :placeholder="$t('page.manage.user.form.userEmail')"
          autocomplete="off"
          name="new-email"
        />
      </ElFormItem>
      <ElFormItem :label="$t('page.manage.user.userPhone')" prop="phone">
        <ElInput
          v-model="model.phone"
          :placeholder="$t('page.manage.user.form.userPhone')"
          autocomplete="off"
          maxlength="11"
        />
      </ElFormItem>
      <ElFormItem :label="$t('page.manage.user.userStatus')" prop="status">
        <ElRadioGroup v-model="model.status" :disabled="isAdminUser">
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
        <PermissionButton
          :code="isEdit ? 'system.user.update' : 'system.user.create'"
          type="primary"
          @click="handleSubmit"
        >
          {{ $t('common.confirm') }}
        </PermissionButton>
      </ElSpace>
    </template>
  </ElDrawer>
</template>

<style scoped></style>
