<script setup lang="ts">
    /**
     * 创建主机对话框
     * 从 index.vue 拆分：新增主机的表单对话框
     */

    import { ref } from 'vue';
    import AttributeFormItems from './AttributeFormItems.vue';

    const props = defineProps<{
      visible: boolean;
      serverForm: CMDB.ServerForm;
      serverFormRules: Record<string, unknown>;
      submitError: string;
      activeCollapse: string[];
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
    }>();

    const emit = defineEmits<{
      (e: 'update:visible', val: boolean): void;
      (e: 'save'): void;
    }>();

    const serverFormRef = ref<InstanceType<typeof import('element-plus')['ElForm']>>(null);

    function onClose() {
      emit('update:visible', false);
    }

    defineExpose({ serverFormRef });
</script>

<template>
    <ElDialog :model-value="visible" title="创建主机" width="600px" @update:model-value="onClose">
        <ElForm ref="serverFormRef" :model="serverForm" :rules="serverFormRules" label-width="100px">
            <div class="form-section-title">基础信息（必填）</div>

            <ElFormItem label="主机名" prop="hostname" class="hostname-input">
                <ElInput v-model="serverForm.hostname" placeholder="请输入主机名" />
                <div v-if="submitError && submitError.includes('主机名')" class="form-error-text" style="color: #f56c6c; font-size: 12px; line-height: 1; padding-top: 4px">
                    {{ submitError }}
                </div>
            </ElFormItem>

            <ElFormItem label="连接IP" prop="ip" class="ip-input">
                <ElInput v-model="serverForm.ip" placeholder="请输入连接IP" />
                <div v-if="submitError && submitError.includes('IP地址')" class="form-error-text" style="color: #f56c6c; font-size: 12px; line-height: 1; padding-top: 4px">
                    {{ submitError }}
                </div>
            </ElFormItem>

            <ElFormItem label="SSH端口">
                <ElInputNumber v-model="serverForm.sshPort" :min="1" :max="65535" placeholder="默认22" style="width: 100%" />
                <div class="mt-4px text-12px text-gray-400">SSH 连接端口，默认 22</div>
            </ElFormItem>

            <ElFormItem label="SSH凭证" prop="credentialIds">
                <ElSelect v-model="serverForm.credentialIds" placeholder="请选择用户连接凭证（可多选）" style="width: 100%" multiple collapse-tags collapse-tags-tooltip>
                    <ElOption v-for="cred in userCredentials" :key="cred.id" :label="`${cred.name}（${cred.username}）`" :value="cred.id" />
                </ElSelect>
                <div class="mt-4px text-12px text-gray-400">用于堡垒机 SSH 连接，受访问策略约束</div>
            </ElFormItem>

            <ElFormItem label="系统运维凭证">
                <ElSelect v-model="serverForm.systemCredentialId" placeholder="请选择系统运维凭证（可选）" style="width: 100%" clearable>
                    <ElOption v-for="cred in systemCredentials" :key="cred.id" :label="`${cred.name}（${cred.username}）`" :value="cred.id" />
                </ElSelect>
                <div class="mt-4px text-12px text-gray-400">用于 Agent 部署、重启、指标采集，需 root/sudo 权限</div>
            </ElFormItem>

            <ElFormItem label="环境">
                <ElSelect v-model="serverForm.env" placeholder="请选择环境" style="width: 100%">
                    <ElOption v-for="opt in getAttributeOptions('env')" :key="opt.value" :label="opt.label" :value="opt.value" />
                </ElSelect>
            </ElFormItem>

            <ElFormItem label="所属分组" prop="groupIds">
                <ElTreeSelect v-model="serverForm.groupIds" :data="groupTreeForSelect" :props="{ label: 'name', value: 'id', children: 'children' }" node-key="id" value-key="id" multiple show-checkbox check-strictly placeholder="请选择分组（可多选）" style="width: 100%" />
            </ElFormItem>

            <ElCollapse v-model="activeCollapse" class="mt-16px">
                <ElCollapseItem title="更多属性（可选）" name="attributes">
                    <AttributeFormItems :attributes="getUnifiedAttributes()" :loading-attributes="loadingAttributes" :get-attribute-value="getAttributeValue" :set-attribute-value="setAttributeValue" :get-attribute-multi-value="getAttributeMultiValue" :set-attribute-multi-value="setAttributeMultiValue" :parse-attribute-options="parseAttributeOptions" />
                </ElCollapseItem>
            </ElCollapse>
        </ElForm>

        <template #footer>
            <ElButton @click="onClose">取消</ElButton>
            <ElButton type="primary" @click="emit('save')">保存</ElButton>
        </template>
    </ElDialog>
</template>

<style scoped>
    .form-section-title {
      font-size: 14px;
      font-weight: 600;
      color: #303133;
      margin-bottom: 16px;
      padding-bottom: 8px;
      border-bottom: 2px solid #e4e7ed;
    }
    .hostname-input,
    .ip-input {
      margin-bottom: 8px;
    }
    .form-error-text {
      animation: shake 0.5s;
    }
    @keyframes shake {
      0%,
      100% {
        transform: translateX(0);
      }
      25% {
        transform: translateX(-4px);
      }
      75% {
        transform: translateX(4px);
      }
    }
</style>
