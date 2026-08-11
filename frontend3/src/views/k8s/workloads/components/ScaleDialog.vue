<script setup lang="ts">
    import { ElButton, ElDialog, ElForm, ElFormItem, ElInput, ElInputNumber, type FormInstance } from 'element-plus';

    const showScaleDialog = defineModel<boolean>('show', { required: true });

    const props = defineProps<{
      deployment: K8s.Deployment | null;
      scaleData: { replicas: number };
      scaleFormRef?: FormInstance;
    }>();

    const emit = defineEmits<{
      (e: 'submit'): void;
    }>();
</script>

<template>
    <ElDialog v-model="showScaleDialog" title="缩放 Deployment" width="500px">
        <ElForm ref="scaleFormRef" :model="scaleData" label-width="100px">
            <ElFormItem label="Deployment">
                <ElInput :value="deployment?.name" disabled />
            </ElFormItem>
            <ElFormItem label="副本数" prop="replicas">
                <ElInputNumber v-model="scaleData.replicas" :min="0" :max="100" :step="1" />
            </ElFormItem>
        </ElForm>
        <template #footer>
            <ElButton @click="showScaleDialog = false">取消</ElButton>
            <ElButton type="primary" @click="emit('submit')">确定</ElButton>
        </template>
    </ElDialog>
</template>
