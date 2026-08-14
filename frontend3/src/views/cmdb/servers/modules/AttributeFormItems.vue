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
  <div v-if="loadingAttributes" :style="{ padding: '12px 0', textAlign: 'center' }">
    <ElIcon class="is-loading"><icon-mdi-loading /></ElIcon>
    <span :style="{ marginLeft: '8px' }">加载中...</span>
  </div>
  <ElEmpty v-else-if="attributes.length === 0" description="暂无可用属性，请先在系统管理-属性管理中配置" />
  <div v-else :style="{ display: 'flex', flexDirection: 'column', gap: '16px' }">
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
      <div v-for="attr in attributes" :key="attr.id" :style="{ display: 'flex', flexDirection: 'column', gap: '8px' }">
        <div :style="{ fontSize: '14px', fontWeight: 500, color: '#303133', display: 'flex', alignItems: 'center', gap: '4px' }">
          <span v-if="attr.required" :style="{ color: '#f56c6c', fontSize: '14px' }">*</span>
          {{ attr.name }}
          <span v-if="attr.description" :style="{ fontSize: '12px', color: '#909399', fontWeight: 'normal', marginLeft: '8px' }">{{ attr.description }}</span>
        </div>
        <div :style="{ width: '100%' }">
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
