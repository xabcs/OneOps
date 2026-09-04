<script setup lang="ts">
  /**
   * 创建主机对话框
   * 重构：
   * - defineModel('visible') 替代 :visible prop + @update:visible emit
   * - useForm() 在内部管理 form ref，不再 defineExpose
   * - emit('submitted') 替代 emit('save')，由父组件触发保存
   */

  import { computed, watch } from 'vue';
  import type { FormRules } from 'element-plus';
  import { useForm } from '@/hooks/common/form';
  import AttributeFormItems from './AttributeFormItems.vue';

  defineOptions({ name: 'ServerFormDialog' });

  interface Props {
    serverForm: CMDB.ServerForm;
    serverFormRules: FormRules;
    submitError: string;
    userCredentials: CMDB.SSHCredential[];
    systemCredentials: CMDB.SSHCredential[];
    groupTreeForSelect: { id: number; name: string; children?: { id: number; name: string }[] }[];
    loadingAttributes: boolean;
    attributeDefinitions: Api.SystemManage.AttributeDefinition[];
    serverAttributes: Api.SystemManage.ServerAttribute[];
    getAttributeValue: (id: number) => string;
    setAttributeValue: (id: number, val: string | number | boolean) => void;
    getAttributeMultiValue: (id: number) => string[];
    setAttributeMultiValue: (id: number, vals: string[]) => void;
    getAttributeOptions: (key: string) => { value: string; label: string }[];
    parseAttributeOptions: (str: string) => { value: string; label: string }[];
    getUnifiedAttributes: () => Api.SystemManage.AttributeDefinition[];
    serverTags: CMDB.ServerTag[];
  }

  const props = defineProps<Props>();

  interface Emits {
    (e: 'submitted'): void;
  }

  const emit = defineEmits<Emits>();

  // P1a: defineModel('visible')
  const visible = defineModel<boolean>('visible', { default: false });

  // 本地 activeCollapse（与父组件同步）
  const activeCollapse = defineModel<string[]>('activeCollapse', { default: () => [] });

  // P0a: useForm() 在内部管理 form ref
  // @ts-expect-error vue-tsc noUnusedLocals: template ref
  const { formRef, validate, restoreValidation } = useForm();

  const dialogTitle = computed(() => '创建主机');

  function onClose() {
    visible.value = false;
  }

  async function handleSubmit() {
    try {
      await validate();
      emit('submitted');
    } catch {
      // 验证失败，不触发 submitted
    }
  }

  // 打开时重置验证状态
  watch(visible, val => {
    if (val) {
      restoreValidation();
    }
  });
</script>

<template>
  <ElDialog v-model="visible" :title="dialogTitle" width="600px">
    <ElForm ref="formRef" :model="serverForm" :rules="serverFormRules" label-width="100px">
      <div
        :style="{
          fontSize: '14px',
          fontWeight: 600,
          color: '#303133',
          marginBottom: '16px',
          paddingBottom: '8px',
          borderBottom: '2px solid #e4e7ed'
        }"
      >
        基础信息（必填）
      </div>

      <ElFormItem label="主机名" prop="hostname" :style="{ marginBottom: '8px' }">
        <ElInput v-model="serverForm.hostname" placeholder="请输入主机名" />
        <div
          v-if="submitError && submitError.includes('主机名')"
          :style="{ color: '#f56c6c', fontSize: '12px', lineHeight: '1', paddingTop: '4px' }"
        >
          {{ submitError }}
        </div>
      </ElFormItem>

      <ElFormItem label="连接IP" prop="ip" :style="{ marginBottom: '8px' }">
        <ElInput v-model="serverForm.ip" placeholder="请输入连接IP" />
        <div
          v-if="submitError && submitError.includes('IP地址')"
          :style="{ color: '#f56c6c', fontSize: '12px', lineHeight: '1', paddingTop: '4px' }"
        >
          {{ submitError }}
        </div>
      </ElFormItem>

      <ElFormItem label="SSH端口">
        <ElInputNumber v-model="serverForm.sshPort" :min="1" :max="65535" placeholder="默认22" style="width: 100%" />
        <div :style="{ marginTop: '4px', fontSize: '12px', color: '#909399' }">SSH 连接端口，默认 22</div>
      </ElFormItem>

      <ElFormItem label="SSH凭证" prop="credentialIds">
        <ElSelect
          v-model="serverForm.credentialIds"
          placeholder="请选择用户连接凭证（可多选）"
          style="width: 100%"
          multiple
          collapse-tags
          collapse-tags-tooltip
        >
          <ElOption
            v-for="cred in userCredentials"
            :key="cred.id"
            :label="`${cred.name}（${cred.username}）`"
            :value="cred.id"
          />
        </ElSelect>
        <div :style="{ marginTop: '4px', fontSize: '12px', color: '#909399' }">用于堡垒机 SSH 连接，受访问策略约束</div>
      </ElFormItem>

      <ElFormItem label="系统运维凭证">
        <ElSelect
          v-model="serverForm.systemCredentialId"
          placeholder="请选择系统运维凭证（可选）"
          style="width: 100%"
          clearable
        >
          <ElOption
            v-for="cred in systemCredentials"
            :key="cred.id"
            :label="`${cred.name}（${cred.username}）`"
            :value="cred.id"
          />
        </ElSelect>
        <div :style="{ marginTop: '4px', fontSize: '12px', color: '#909399' }">
          用于 Agent 部署、重启、指标采集，需 root/sudo 权限
        </div>
      </ElFormItem>

      <ElFormItem label="所属分组" prop="groupIds">
        <ElTreeSelect
          v-model="serverForm.groupIds"
          :data="groupTreeForSelect"
          :props="{ label: 'name', value: 'id', children: 'children' }"
          node-key="id"
          value-key="id"
          multiple
          show-checkbox
          check-strictly
          placeholder="请选择分组（可多选）"
          style="width: 100%"
        />
      </ElFormItem>

      <ElFormItem label="标签">
        <ElSelect
          v-model="serverForm.tagIds"
          placeholder="请选择标签（可多选）"
          style="width: 100%"
          multiple
          collapse-tags
          collapse-tags-tooltip
        >
          <ElOption v-for="tag in serverTags" :key="tag.id" :label="tag.name" :value="tag.id">
            <span>{{ tag.name }}</span>
            <span
              :style="{
                display: 'inline-block',
                width: '8px',
                height: '8px',
                borderRadius: '50%',
                backgroundColor: tag.color,
                marginLeft: '8px'
              }"
            />
          </ElOption>
        </ElSelect>
      </ElFormItem>

      <ElCollapse v-model="activeCollapse" :style="{ marginTop: '16px' }">
        <ElCollapseItem title="更多属性（可选）" name="attributes">
          <AttributeFormItems
            :attributes="getUnifiedAttributes()"
            :loading-attributes="loadingAttributes"
            :get-attribute-value="getAttributeValue"
            :set-attribute-value="setAttributeValue"
            :parse-attribute-options="parseAttributeOptions"
          />
        </ElCollapseItem>
      </ElCollapse>
    </ElForm>

    <template #footer>
      <ElButton @click="onClose">取消</ElButton>
      <ElButton type="primary" @click="handleSubmit">保存</ElButton>
    </template>
  </ElDialog>
</template>
