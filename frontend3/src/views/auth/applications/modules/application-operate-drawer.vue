<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import {
    createApplication,
    fetchAppTypeConfigTemplate,
    fetchSupportedAppTypes,
    updateApplication
  } from '@/service/api/application-permission';
  import { useForm, useFormRules } from '@/hooks/common/form';

  defineOptions({ name: 'AuthApplicationOperateDrawer' });

  interface Props {
    /** operation type: add | edit */
    operateType: UI.TableOperateType;
    /** editing row data */
    rowData?: Api.ApplicationPermission.Application | null;
  }

  const props = defineProps<Props>();

  interface Emits {
    (e: 'submitted'): void;
  }

  const emit = defineEmits<Emits>();

  // P1a: defineModel('visible')
  const visible = defineModel<boolean>('visible', { default: false });

  // P0a: drawer 内部 useForm()
  // @ts-expect-error vue-tsc noUnusedLocals: template ref
  const { formRef, validate, restoreValidation } = useForm();
  const { defaultRequiredRule } = useFormRules();

  const title = computed(() => (props.operateType === 'add' ? '添加应用' : '编辑应用'));

  type Model = Api.ApplicationPermission.ApplicationFormData;

  const defaultEndpoints: Api.ApplicationPermission.ApplicationEndpoints = {
    createUser: '/api/users',
    listUsers: '/api/users',
    getUser: '/api/users/{username}',
    getRoles: '/api/roles',
    assignRole: '/api/users/{username}/roles',
    revokeRole: '/api/users/{username}/roles/{roleCode}',
    getUserRoles: '/api/users/{username}/roles'
  };

  function createDefaultModel(): Model {
    return {
      id: 0,
      name: '',
      code: '',
      type: '',
      baseUrl: '',
      endpoints: { ...defaultEndpoints },
      authConfig: { type: 'token', token: '' },
      description: '',
      syncInterval: 60
    };
  }

  const model = ref<Model>(createDefaultModel());

  type RuleKey = Extract<keyof Model, 'name' | 'code' | 'type' | 'baseUrl'>;

  const rules: Record<RuleKey, App.Global.FormRule> = {
    name: defaultRequiredRule,
    code: defaultRequiredRule,
    type: defaultRequiredRule,
    baseUrl: defaultRequiredRule
  };

  const authTypeOptions = [
    { label: 'API Token', value: 'token' },
    { label: 'Basic Auth', value: 'basic' },
    { label: 'OAuth 2.0', value: 'oauth2' }
  ];

  // 支持的应用类型
  const supportedAppTypes = ref<Array<{ type: string; displayName: string }>>([]);
  const loadingAppTypes = ref(false);

  async function getSupportedAppTypes() {
    loadingAppTypes.value = true;
    try {
      const { data, error } = await fetchSupportedAppTypes();
      if (!error && data) {
        supportedAppTypes.value = data || [];
      }
    } finally {
      loadingAppTypes.value = false;
    }
  }

  // 监听应用类型变化，加载配置模板
  watch(
    () => model.value.type,
    async newType => {
      if (newType) {
        await loadConfigTemplate(newType);
      }
    }
  );

  async function loadConfigTemplate(appType: string) {
    const { data, error } = await fetchAppTypeConfigTemplate(appType);
    if (!error && data && data.configTemplate) {
      const tpl = data.configTemplate as Record<string, any>;
      if (tpl.endpoints) {
        model.value.endpoints = { ...tpl.endpoints };
      }
      if (tpl.token !== undefined) {
        model.value.authConfig = { type: 'token', token: tpl.token || '' };
      }
      if (tpl.username !== undefined) {
        model.value.authConfig = {
          type: 'basic',
          username: tpl.username || '',
          password: tpl.password || ''
        };
      }
      if (tpl.syncInterval) {
        model.value.syncInterval = tpl.syncInterval;
      }
    }
  }

  function needsEndpointsConfig() {
    const type = model.value.type;
    return type === 'generic' || type === 'jumpserver' || type === 'gitlab';
  }

  function handleInitModel() {
    model.value = createDefaultModel();
    if (props.operateType === 'edit' && props.rowData) {
      const endpoints =
        typeof props.rowData.endpoints === 'string' ? JSON.parse(props.rowData.endpoints) : props.rowData.endpoints;
      const authConfig =
        typeof props.rowData.authConfig === 'string' ? JSON.parse(props.rowData.authConfig) : props.rowData.authConfig;
      model.value = { ...props.rowData, endpoints, authConfig };
    }
  }

  function closeDrawer() {
    visible.value = false;
  }

  async function handleSubmit() {
    await validate();

    const submitData = {
      ...model.value,
      endpoints: JSON.stringify(model.value.endpoints),
      authConfig: JSON.stringify(model.value.authConfig)
    };

    if (props.operateType === 'edit') {
      const { error } = await updateApplication(model.value.id!, submitData);
      if (!error) {
        ElMessage.success('更新成功');
        closeDrawer();
        emit('submitted');
      }
    } else {
      const { error } = await createApplication(submitData);
      if (!error) {
        ElMessage.success('添加成功');
        closeDrawer();
        emit('submitted');
      }
    }
  }

  watch(visible, val => {
    if (val) {
      handleInitModel();
      restoreValidation();
      if (supportedAppTypes.value.length === 0) {
        getSupportedAppTypes();
      }
    }
  });
</script>

<template>
  <ElDrawer v-model="visible" :title="title" :width="600">
    <ElForm ref="formRef" :model="model" :rules="rules" label-width="120px">
      <ElFormItem label="应用名称" prop="name">
        <ElInput v-model="model.name" placeholder="请输入应用名称" />
      </ElFormItem>
      <ElFormItem label="应用代码" prop="code">
        <ElInput v-model="model.code" placeholder="请输入应用代码（唯一标识）" />
      </ElFormItem>
      <ElFormItem label="应用类型" prop="type">
        <ElSelect v-model="model.type" placeholder="请选择应用类型" class="w-full" :loading="loadingAppTypes">
          <ElOption v-for="item in supportedAppTypes" :key="item.type" :label="item.displayName" :value="item.type" />
        </ElSelect>
      </ElFormItem>
      <ElFormItem label="Base URL" prop="baseUrl">
        <ElInput v-model="model.baseUrl" placeholder="请输入应用访问地址，如：https://jenkins.example.com" />
      </ElFormItem>

      <ElDivider content-position="left">认证配置</ElDivider>

      <!-- Jenkins 专用: Basic Auth -->
      <template v-if="model.type === 'jenkins'">
        <ElFormItem label="用户名" required>
          <ElInput v-model="model.authConfig.username" placeholder="请输入 Jenkins 用户名" />
        </ElFormItem>
        <ElFormItem label="密码" required>
          <ElInput
            v-model="model.authConfig.password"
            type="password"
            show-password
            placeholder="请输入 Jenkins 密码或 Token"
          />
        </ElFormItem>
      </template>

      <!-- Jumpserver 专用: 支持三种认证方式 -->
      <template v-if="model.type === 'jumpserver'">
        <ElFormItem label="认证方式" required>
          <ElSelect v-model="model.authConfig.authType" placeholder="请选择认证方式" class="w-full">
            <ElOption label="AccessKey + Secret（推荐）" value="accessKey" />
            <ElOption label="Private Token" value="token" />
            <ElOption label="用户名 + 密码" value="basic" />
          </ElSelect>
        </ElFormItem>

        <!-- AccessKey + Secret 签名认证 -->
        <template v-if="model.authConfig.authType === 'accessKey'">
          <ElAlert type="info" :closable="false" class="mb-4">
            <template #title>
              <div class="text-sm">
                <strong>AccessKey 签名认证说明：</strong>
                <ul class="ml-4 mt-2 list-disc">
                  <li>需要在 JumpServer 中创建「服务账号」获取密钥</li>
                  <li>使用 HTTP Signature 签名算法，最安全可靠</li>
                  <li>长期有效，不会过期</li>
                </ul>
              </div>
            </template>
          </ElAlert>
          <ElFormItem label="Access Key ID" required>
            <ElInput v-model="model.authConfig.accessKey" placeholder="请输入 Access Key ID" />
            <span class="text-xs text-gray-500">在 JumpServer 系统设置 → 账号管理 → 服务账号中获取</span>
          </ElFormItem>
          <ElFormItem label="Access Key Secret" required>
            <ElInput
              v-model="model.authConfig.secret"
              type="password"
              show-password
              placeholder="请输入 Access Key Secret"
            />
            <span class="text-xs text-gray-500">密钥只显示一次，请妥善保管</span>
          </ElFormItem>
          <ElFormItem label="组织ID（可选）">
            <ElInput v-model="model.authConfig.orgId" placeholder="默认组织: 00000000-00000000-00000000-000000000002" />
            <span class="text-xs text-gray-500">不填写则使用默认组织</span>
          </ElFormItem>
        </template>

        <!-- Private Token 认证 -->
        <template v-if="model.authConfig.authType === 'token'">
          <ElFormItem label="Private Token" required>
            <ElInput
              v-model="model.authConfig.token"
              type="password"
              show-password
              placeholder="请输入 Private Token"
            />
            <span class="text-xs text-gray-500">在 JumpServer 个人中心 → API Token 获取（长期有效）</span>
          </ElFormItem>
        </template>

        <!-- 用户名密码认证 -->
        <template v-if="model.authConfig.authType === 'basic'">
          <ElFormItem label="用户名" required>
            <ElInput v-model="model.authConfig.username" placeholder="请输入用户名" />
          </ElFormItem>
          <ElFormItem label="密码" required>
            <ElInput v-model="model.authConfig.password" type="password" show-password placeholder="请输入密码" />
          </ElFormItem>
        </template>
      </template>

      <!-- GitLab 专用: API Token -->
      <template v-if="model.type === 'gitlab'">
        <ElFormItem label="API Token" required>
          <ElInput v-model="model.authConfig.token" type="password" show-password placeholder="请输入 API Token" />
        </ElFormItem>
      </template>

      <!-- 通用应用: 可选择认证方式 -->
      <template v-if="model.type === 'generic' || model.type === ''">
        <ElFormItem label="认证方式" required>
          <ElSelect v-model="model.authConfig.type" placeholder="请选择认证方式" class="w-full">
            <ElOption v-for="item in authTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
          </ElSelect>
        </ElFormItem>

        <ElFormItem v-if="model.authConfig.type === 'token'" label="API Token" required>
          <ElInput v-model="model.authConfig.token" type="password" show-password placeholder="请输入API Token" />
        </ElFormItem>

        <template v-if="model.authConfig.type === 'basic'">
          <ElFormItem label="用户名" required>
            <ElInput v-model="model.authConfig.username" placeholder="请输入用户名" />
          </ElFormItem>
          <ElFormItem label="密码" required>
            <ElInput v-model="model.authConfig.password" type="password" show-password placeholder="请输入密码" />
          </ElFormItem>
        </template>
      </template>

      <!-- API端点配置 - 仅对需要端点配置的应用显示 -->
      <ElDivider v-if="needsEndpointsConfig()" content-position="left">API端点配置</ElDivider>

      <!-- Jumpserver 端点配置 -->
      <template v-if="model.type === 'jumpserver'">
        <ElAlert type="info" :closable="false" class="mb-4">
          <template #title>
            <div class="text-sm">
              <strong>JumpServer 授权说明：</strong>
              <ul class="ml-4 mt-2 list-disc">
                <li>JumpServer 不需要同步角色，授权通过授权规则管理</li>
                <li>用户授权需在 JumpServer 管理界面中配置授权规则（用户/用户组 → 资产/节点）</li>
              </ul>
            </div>
          </template>
        </ElAlert>
        <ElFormItem label="获取用户列表">
          <ElInput v-model="model.endpoints.getUsers" placeholder="/api/v1/users/users/" />
        </ElFormItem>
        <ElFormItem label="获取用户组列表">
          <ElInput v-model="model.endpoints.getGroups" placeholder="/api/v1/users/groups/" />
        </ElFormItem>
        <ElFormItem label="创建用户">
          <ElInput v-model="model.endpoints.createUser" placeholder="/api/v1/users/users/" />
        </ElFormItem>
      </template>

      <!-- Generic 应用端点配置 -->
      <template v-if="model.type === 'generic'">
        <ElFormItem label="获取用户列表">
          <ElInput v-model="model.endpoints.getUsers" placeholder="/api/v1/users/users/" />
        </ElFormItem>
        <ElFormItem label="获取角色列表">
          <ElInput v-model="model.endpoints.getRoles" placeholder="/api/v1/perms/roles/" />
        </ElFormItem>
        <ElFormItem label="创建用户">
          <ElInput v-model="model.endpoints.createUser" placeholder="/api/v1/users/users/" />
        </ElFormItem>
        <ElFormItem label="分配角色">
          <ElInput v-model="model.endpoints.assignRole" placeholder="/api/v1/perms/user-grantings/" />
        </ElFormItem>
      </template>

      <!-- GitLab 特殊配置 -->
      <template v-if="model.type === 'gitlab'">
        <ElDivider content-position="left">GitLab 特殊配置</ElDivider>
        <ElFormItem label="项目ID (可选)">
          <ElInput v-model="model.endpoints.projectId" placeholder="用于项目级权限管理" />
          <span class="text-xs text-gray-500">填写项目ID可管理项目成员权限</span>
        </ElFormItem>
        <ElFormItem label="群组ID (可选)">
          <ElInput v-model="model.endpoints.groupId" placeholder="用于群组级权限管理" />
          <span class="text-xs text-gray-500">填写群组ID可管理群组成员权限</span>
        </ElFormItem>
        <ElFormItem label="获取用户列表">
          <ElInput v-model="model.endpoints.getUsers" placeholder="/api/v4/users" />
        </ElFormItem>
        <ElFormItem label="创建用户">
          <ElInput v-model="model.endpoints.createUser" placeholder="/api/v4/users" />
        </ElFormItem>
      </template>

      <ElDivider content-position="left">其他配置</ElDivider>

      <ElFormItem label="描述">
        <ElInput v-model="model.description" type="textarea" :rows="3" placeholder="请输入描述" />
      </ElFormItem>
      <ElFormItem label="同步间隔">
        <ElInputNumber v-model="model.syncInterval" :min="1" placeholder="同步间隔（分钟）" class="w-full" />
      </ElFormItem>
    </ElForm>

    <template #footer>
      <ElButton @click="closeDrawer">取消</ElButton>
      <ElButton type="primary" @click="handleSubmit">确定</ElButton>
    </template>
  </ElDrawer>
</template>
