<script setup lang="ts">
  /**
   * 动态属性表单项
   * 根据属性定义渲染不同类型的输入控件
   * 支持 readonly 模式用于详情页只读展示
   */

  interface Props {
    attributes: Api.SystemManage.AttributeDefinition[];
    loadingAttributes: boolean;
    getAttributeValue: (id: number) => string;
    setAttributeValue?: (id: number, val: string | number | boolean) => void;
    parseAttributeOptions: (optionsStr: string) => { value: string; label: string }[];
    readonly?: boolean;
  }

  const props = withDefaults(defineProps<Props>(), {
    readonly: false
  });

  /** 获取 select 选项的 label（用于 readonly 模式展示） */
  function getSelectLabel(attr: Api.SystemManage.AttributeDefinition): string {
    const value = props.getAttributeValue(attr.id);
    if (!value) return '-';
    const options = props.parseAttributeOptions(attr.options);
    const opt = options.find(o => o.value === value);
    return opt?.label || value;
  }

  /** 获取 boolean 值的展示文案 */
  function getBooleanText(attr: Api.SystemManage.AttributeDefinition): string {
    const value = props.getAttributeValue(attr.id);
    if (!value) return '-';
    return value === 'true' ? '是' : '否';
  }
</script>

<template>
  <div v-if="loadingAttributes" class="py-12px text-center">
    <ElIcon class="is-loading"><icon-mdi-loading /></ElIcon>
    <span class="ml-8px">加载中...</span>
  </div>
  <div v-else-if="attributes.length === 0" class="py-12px text-center text-gray-400">
    暂无可用属性，请先在"系统管理 → 属性管理"中配置
  </div>
  <div v-else class="attributes-container">
    <!-- readonly 模式 -->
    <template v-if="readonly">
      <ElDescriptions v-for="attr in attributes" :key="attr.id" :column="1" border size="small">
        <ElDescriptionsItem :label="attr.name">
          <span v-if="attr.type === 'select'">{{ getSelectLabel(attr) }}</span>
          <span v-else-if="attr.type === 'boolean'">{{ getBooleanText(attr) }}</span>
          <span v-else>{{ getAttributeValue(attr.id) || '-' }}</span>
        </ElDescriptionsItem>
      </ElDescriptions>
    </template>

    <!-- 编辑模式 -->
    <template v-else>
      <div v-for="attr in attributes" :key="attr.id" class="attribute-item">
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
            style="width: 100%"
            @change="val => setAttributeValue?.(attr.id, val)"
          />
          <ElInputNumber
            v-else-if="attr.type === 'number'"
            :model-value="Number(getAttributeValue(attr.id)) || undefined"
            :placeholder="attr.defaultValue || `请输入${attr.name}`"
            style="width: 100%"
            @change="val => setAttributeValue?.(attr.id, val)"
          />
          <ElSwitch
            v-else-if="attr.type === 'boolean'"
            :model-value="getAttributeValue(attr.id) === 'true'"
            active-text="是"
            inactive-text="否"
            @change="val => setAttributeValue?.(attr.id, val ? 'true' : 'false')"
          />
          <ElSelect
            v-else-if="attr.type === 'select'"
            :model-value="getAttributeValue(attr.id)"
            :placeholder="`请选择${attr.name}`"
            style="width: 100%"
            @change="val => setAttributeValue?.(attr.id, val)"
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
    </template>
  </div>
</template>

<style scoped>
  .attributes-container {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
  .attribute-item {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .attribute-label {
    font-size: 14px;
    font-weight: 500;
    color: #303133;
    display: flex;
    align-items: center;
    gap: 4px;
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
  .attribute-input {
    width: 100%;
  }
</style>
