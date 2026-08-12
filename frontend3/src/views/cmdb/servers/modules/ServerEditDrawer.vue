<script setup lang="ts">
  /**
   * 编辑主机抽屉
   * 重构：
   * - defineModel('visible') 替代 :visible prop + @update:visible emit
   * - useForm() 在内部管理 form ref，不再 defineExpose
   * - emit('submitted') 替代 emit('save')，由父组件触发保存
   */

  import { watch } from 'vue';
  import type { FormRules } from 'element-plus';
  import { useForm } from '@/hooks/common/form';

  defineOptions({ name: 'ServerEditDrawer' });

  interface Props {
    serverForm: CMDB.ServerForm;
    serverFormRules: FormRules;
    submitError: string;
    serverType: 'normal' | 'cloud';
    cloudForm: CMDB.CloudServerForm;
    userCredentials: CMDB.SSHCredential[];
    systemCredentials: CMDB.SSHCredential[];
    serverTags: CMDB.ServerTag[];
    serverRooms: CMDB.ServerRoom[];
    cabinets: CMDB.Cabinet[];
    businessUnits: CMDB.BusinessUnit[];
    groupTreeForSelect: { id: number; name: string; children?: { id: number; name: string }[] }[];
    loadingAttributes: boolean;
    attributeDefinitions: Api.SystemManage.AttributeDefinition[];
    getAttributeOptions: (key: string) => { value: string; label: string }[];
    getUnifiedAttributes: () => Api.SystemManage.AttributeDefinition[];
    getAttributeValue: (id: number) => string;
    setAttributeValue: (id: number, val: string | number | boolean) => void;
    getAttributeMultiValue: (id: number) => string[];
    setAttributeMultiValue: (id: number, vals: string[]) => void;
    parseAttributeOptions: (str: string) => { value: string; label: string }[];
    handleRoomChange: (roomId: number, serverForm: CMDB.ServerForm) => void;
  }

  const props = defineProps<Props>();

  interface Emits {
    (e: 'submitted'): void;
  }

  const emit = defineEmits<Emits>();

  // P1a: defineModel('visible')
  const visible = defineModel<boolean>('visible', { default: false });

  // activeTab 双向绑定（本地管理，无需上抛）
  const activeTab = defineModel<string>('activeTab', { default: 'basic' });

  // P0a: useForm() 在内部管理 form ref
  // @ts-expect-error vue-tsc noUnusedLocals: template ref
  const { formRef, validate, restoreValidation } = useForm();

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
  <ElDrawer
    v-model="visible"
    :title="`编辑主机 - ${serverForm.hostname}`"
    direction="rtl"
    size="80%"
    destroy-on-close
  >
    <ElTabs v-model="activeTab" type="border-card">
      <!-- 基础信息 Tab -->
      <ElTabPane label="基础信息" name="basic">
        <ElForm ref="formRef" :model="serverForm" :rules="serverFormRules" label-width="100px" class="edit-form">
          <div class="edit-form-row">
            <div class="edit-form-col">
              <div class="subsection-title">主机信息</div>
              <ElFormItem label="主机名" prop="hostname" class="hostname-input">
                <ElInput v-model="serverForm.hostname" placeholder="请输入主机名" />
                <div v-if="submitError && submitError.includes('主机名')" class="form-error-text">
                  {{ submitError }}
                </div>
              </ElFormItem>
              <ElFormItem label="连接IP" prop="ip" class="ip-input">
                <ElInput v-model="serverForm.ip" placeholder="请输入连接IP" />
                <div v-if="submitError && submitError.includes('IP地址')" class="form-error-text">
                  {{ submitError }}
                </div>
              </ElFormItem>
              <ElFormItem label="内网IP">
                <ElInput v-model="serverForm.innerIp" placeholder="请输入内网IP" />
              </ElFormItem>
              <ElFormItem label="SSH端口">
                <ElInputNumber v-model="serverForm.sshPort" :min="1" :max="65535" style="width: 100%" />
              </ElFormItem>
            </div>
            <div class="edit-form-col">
              <div class="subsection-title">环境与归属</div>
              <ElFormItem label="环境">
                <ElSelect v-model="serverForm.env" placeholder="请选择环境" style="width: 100%">
                  <ElOption
                    v-for="opt in getAttributeOptions('env')"
                    :key="opt.value"
                    :label="opt.label"
                    :value="opt.value"
                  />
                </ElSelect>
              </ElFormItem>
              <ElFormItem label="业务系统">
                <ElTreeSelect
                  v-model="serverForm.businessId"
                  :data="businessUnits"
                  :props="{ label: 'name', value: 'id', children: 'children' }"
                  placeholder="请选择业务系统"
                  clearable
                  check-strictly
                  style="width: 100%"
                />
              </ElFormItem>
            </div>
          </div>

          <div class="edit-form-row">
            <div class="edit-form-col">
              <div class="subsection-title">凭证配置</div>
              <ElFormItem label="用户连接凭证" prop="credentialIds">
                <ElSelect
                  v-model="serverForm.credentialIds"
                  placeholder="请选择用户连接凭证"
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
              </ElFormItem>
              <ElFormItem label="系统运维凭证">
                <ElSelect
                  v-model="serverForm.systemCredentialId"
                  placeholder="请选择系统运维凭证"
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
              </ElFormItem>
            </div>
            <div class="edit-form-col">
              <div class="subsection-title">分组与标签</div>
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
                  placeholder="请选择分组"
                  style="width: 100%"
                />
              </ElFormItem>
              <ElFormItem label="标签">
                <ElSelect
                  v-model="serverForm.tagIds"
                  placeholder="请选择标签"
                  multiple
                  collapse-tags
                  collapse-tags-tooltip
                  style="width: 100%"
                >
                  <ElOption v-for="tag in serverTags" :key="tag.id" :label="tag.name" :value="tag.id">
                    <span>{{ tag.name }}</span>
                    <span
                      :style="{
                        marginLeft: '8px',
                        display: 'inline-block',
                        width: '12px',
                        height: '12px',
                        borderRadius: '2px',
                        backgroundColor: tag.color
                      }"
                    />
                  </ElOption>
                </ElSelect>
              </ElFormItem>
            </div>
          </div>

          <div class="edit-form-row">
            <div class="edit-form-col">
              <div class="subsection-title">位置信息</div>
              <ElFormItem label="所在机房">
                <ElSelect
                  v-model="serverForm.roomId"
                  placeholder="请选择机房"
                  clearable
                  style="width: 100%"
                  @change="handleRoomChange($event, serverForm)"
                >
                  <ElOption
                    v-for="room in serverRooms"
                    :key="room.id"
                    :label="`${room.name} (${room.code})`"
                    :value="room.id"
                  />
                </ElSelect>
              </ElFormItem>
              <ElFormItem label="所在机柜">
                <ElSelect
                  v-model="serverForm.cabinetId"
                  placeholder="请先选择机房"
                  clearable
                  :disabled="!serverForm.roomId"
                  style="width: 100%"
                >
                  <ElOption
                    v-for="cabinet in cabinets"
                    :key="cabinet.id"
                    :label="`${cabinet.name} (${cabinet.code})`"
                    :value="cabinet.id"
                  />
                </ElSelect>
              </ElFormItem>
            </div>
            <div class="edit-form-col">
              <div class="subsection-title">硬件配置</div>
              <ElFormItem label="CPU/内存/磁盘">
                <div style="display: flex; gap: 8px">
                  <ElInputNumber v-model="serverForm.cpu" :min="0" :max="1024" placeholder="CPU" style="flex: 1" />
                  <ElInputNumber
                    v-model="serverForm.memory"
                    :min="0"
                    :max="65536"
                    placeholder="内存GB"
                    style="flex: 1"
                  />
                  <ElInputNumber
                    v-model="serverForm.disk"
                    :min="0"
                    :max="999999"
                    placeholder="磁盘GB"
                    style="flex: 1"
                  />
                </div>
              </ElFormItem>
              <ElFormItem label="操作系统">
                <ElInput v-model="serverForm.os" placeholder="如 Ubuntu 22.04" style="width: 100%" />
              </ElFormItem>
              <ElFormItem label="备注">
                <ElInput v-model="serverForm.remarks" type="textarea" :rows="2" placeholder="请输入备注" />
              </ElFormItem>
            </div>
          </div>
        </ElForm>
      </ElTabPane>

      <!-- 属性配置 Tab -->
      <ElTabPane label="属性配置" name="attributes">
        <div v-if="loadingAttributes" class="py-40px text-center">
          <ElIcon class="is-loading text-32px"><icon-mdi-loading /></ElIcon>
          <div class="mt-16px">加载中...</div>
        </div>
        <div v-else-if="getUnifiedAttributes().length === 0" class="py-40px text-center text-gray-400">
          <icon-mdi-information-outline class="text-48px" />
          <div class="mt-16px">暂无可用属性</div>
          <div class="mt-8px text-12px">请先在"系统管理 → 属性管理"中配置属性</div>
        </div>
        <div v-else class="attributes-container-drawer">
          <div v-for="attr in getUnifiedAttributes()" :key="attr.id" class="attribute-form-item-compact">
            <div class="attribute-label">
              <span v-if="attr.required" class="required-mark">*</span>
              {{ attr.name }}
              <span v-if="attr.description" class="attribute-description">{{ attr.description }}</span>
            </div>
            <div class="attribute-input">
              <ElInput
                v-if="attr.type === 'text'"
                :model-value="getAttributeValue(attr.id)"
                :placeholder="attr.defaultValue || `请输入${attr.name}`"
                @change="val => setAttributeValue(attr.id, val)"
              />
              <ElInputNumber
                v-else-if="attr.type === 'number'"
                :model-value="getAttributeValue(attr.id)"
                :placeholder="attr.defaultValue || `请输入${attr.name}`"
                style="width: 100%"
                @change="val => setAttributeValue(attr.id, val)"
              />
              <ElDatePicker
                v-else-if="attr.type === 'date'"
                :model-value="getAttributeValue(attr.id)"
                type="date"
                :placeholder="attr.defaultValue || `请选择${attr.name}`"
                style="width: 100%"
                format="YYYY-MM-DD"
                value-format="YYYY-MM-DD"
                @change="val => setAttributeValue(attr.id, val)"
              />
              <ElSwitch
                v-else-if="attr.type === 'boolean'"
                :model-value="getAttributeValue(attr.id)"
                active-text="是"
                inactive-text="否"
                @change="val => setAttributeValue(attr.id, val)"
              />
              <ElSelect
                v-else-if="attr.type === 'select'"
                :model-value="getAttributeValue(attr.id)"
                :placeholder="`请选择${attr.name}`"
                style="width: 100%"
                @change="val => setAttributeValue(attr.id, val)"
              >
                <ElOption
                  v-for="opt in parseAttributeOptions(attr.options)"
                  :key="opt.value"
                  :label="opt.label"
                  :value="opt.value"
                />
              </ElSelect>
              <ElSelect
                v-else-if="attr.type === 'multiselect'"
                :model-value="getAttributeMultiValue(attr.id)"
                :placeholder="`请选择${attr.name}`"
                style="width: 100%"
                multiple
                collapse-tags
                collapse-tags-tooltip
                @change="val => setAttributeMultiValue(attr.id, val)"
              >
                <ElOption
                  v-for="opt in parseAttributeOptions(attr.options)"
                  :key="opt.value"
                  :label="opt.label"
                  :value="opt.value"
                />
              </ElSelect>
            </div>
          </div>
        </div>
      </ElTabPane>

      <!-- 云主机配置 Tab -->
      <ElTabPane label="云主机配置" name="cloud" :disabled="serverType !== 'cloud'">
        <div v-if="serverType === 'cloud'" class="edit-form">
          <ElForm label-width="100px">
            <ElFormItem label="云服务商">
              <ElSelect v-model="cloudForm.provider" style="width: 100%">
                <ElOption label="阿里云" value="aliyun" />
                <ElOption label="腾讯云" value="tencent" />
                <ElOption label="华为云" value="huawei" />
                <ElOption label="AWS" value="aws" />
                <ElOption label="其他" value="other" />
              </ElSelect>
            </ElFormItem>
            <ElFormItem label="实例名称">
              <ElInput v-model="cloudForm.instanceName" placeholder="请输入实例名称" />
            </ElFormItem>
            <ElFormItem label="实例规格">
              <ElInput v-model="cloudForm.instanceType" placeholder="如: ecs.t6-c1m2.large" />
            </ElFormItem>
            <ElFormItem label="地域">
              <ElInput v-model="cloudForm.region" placeholder="如: cn-hangzhou" />
            </ElFormItem>
            <ElFormItem label="可用区">
              <ElInput v-model="cloudForm.zone" placeholder="如: cn-hangzhou-i" />
            </ElFormItem>
            <ElFormItem label="计费类型">
              <ElSelect v-model="cloudForm.chargeType" style="width: 100%">
                <ElOption label="按量付费" value="postpay" />
                <ElOption label="包年包月" value="prepay" />
              </ElSelect>
            </ElFormItem>
          </ElForm>
        </div>
        <div v-else class="py-40px text-center text-gray-400">
          <icon-mdi-cloud-off-outline class="text-48px" />
          <div class="mt-16px">当前主机不是云主机，无需配置云服务信息</div>
        </div>
      </ElTabPane>
    </ElTabs>

    <template #footer>
      <div style="flex: 1"></div>
      <ElButton size="large" @click="visible = false">取消</ElButton>
      <ElButton type="primary" size="large" @click="handleSubmit">保存更改</ElButton>
    </template>
  </ElDrawer>
</template>

<style scoped>
  .edit-form {
    padding: 16px;
  }
  .edit-form-row {
    display: flex;
    gap: 20px;
    margin-bottom: 16px;
  }
  .edit-form-col {
    flex: 1;
    min-width: 0;
  }
  .subsection-title {
    font-size: 13px;
    font-weight: 600;
    color: #409eff;
    margin-bottom: 10px;
    padding-bottom: 4px;
    border-bottom: 1px solid #dcdfe6;
  }
  .hostname-input,
  .ip-input {
    margin-bottom: 8px;
  }
  .form-error-text {
    color: #f56c6c;
    font-size: 12px;
    line-height: 1;
    padding-top: 4px;
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
  .attributes-container-drawer {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 16px 20px;
    padding: 20px;
  }
  .attribute-form-item-compact {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 12px;
    background-color: #fafafa;
    border-radius: 6px;
    border: 1px solid #e4e7ed;
  }
  .attribute-form-item-compact .attribute-label {
    font-size: 13px;
    font-weight: 500;
    color: #303133;
    display: flex;
    align-items: center;
    gap: 4px;
  }
  .attribute-form-item-compact .attribute-input {
    width: 100%;
  }
  .required-mark {
    color: #f56c6c;
    font-size: 14px;
  }
  .attribute-description {
    font-size: 12px;
    color: #909399;
    font-weight: normal;
    margin-left: 8px;
  }
</style>
