<script setup lang="ts">
  /**
   * 动态属性表单项
   * 从 index.vue 拆分：根据属性定义渲染不同类型的输入控件
   */

  defineProps<{
    attributes: Api.SystemManage.AttributeDefinition[];
    loadingAttributes: boolean;
    getAttributeValue: (id: number) => string;
    setAttributeValue: (id: number, val: string | number | boolean) => void;
    getAttributeMultiValue: (id: number) => string[];
    setAttributeMultiValue: (id: number, vals: string[]) => void;
    parseAttributeOptions: (optionsStr: string) => { value: string; label: string }[];
  }>();
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
