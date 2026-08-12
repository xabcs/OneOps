<script setup lang="ts">
  import { ref, watch } from 'vue';
  import { ElMessage } from 'element-plus';
  import { createNotificationChannel, updateNotificationChannel } from '@/service/api';

  defineOptions({ name: 'NotificationChannelDialog' });

  const props = defineProps<{
    mode: 'create' | 'edit';
    channel?: Monitoring.NotificationChannel | null;
  }>();

  const emit = defineEmits<{
    (e: 'submitted'): void;
  }>();

  const visible = defineModel<boolean>('visible', { default: false });

  // 通知渠道类型选项
  const channelTypeOptions = [
    { label: '邮件', value: 'email' },
    { label: '企业微信', value: 'wechat' },
    { label: '钉钉', value: 'dingtalk' },
    { label: '飞书', value: 'feishu' }
  ];

  const form = ref<Monitoring.NotificationChannelForm>({
    channelType: 'email',
    channelName: '',
    config: {} as Record<string, unknown>,
    enabled: true
  });

  const submitting = ref(false);

  // 同步 props 到表单
  watch(
    visible,
    val => {
      if (val) {
        if (props.mode === 'edit' && props.channel) {
          form.value = {
            id: props.channel.id,
            channelType: props.channel.channelType,
            channelName: props.channel.channelName,
            config: props.channel.config as Monitoring.EmailConfig,
            enabled: props.channel.enabled
          };
        } else {
          form.value = {
            channelType: 'email',
            channelName: '',
            config: {
              smtpHost: '',
              smtpPort: 587,
              username: '',
              password: '',
              from: '',
              fromName: '',
              useTLS: true,
              recipients: []
            },
            enabled: true
          };
        }
      }
    },
    { immediate: true }
  );

  async function handleSubmit() {
    if (!form.value.channelName) {
      ElMessage.warning('请输入渠道名称');
      return;
    }

    // 根据类型验证配置
    if (form.value.channelType === 'email') {
      const config = form.value.config as Monitoring.EmailConfig;
      if (!config.smtpHost || !config.from) {
        ElMessage.warning('请完善邮件配置');
        return;
      }
    } else if (form.value.channelType === 'wechat') {
      const config = form.value.config as Monitoring.WeChatConfig;
      if (!config.webhookUrl) {
        ElMessage.warning('请输入企业微信 Webhook URL');
        return;
      }
    }

    submitting.value = true;
    try {
      if (props.mode === 'create') {
        await createNotificationChannel(form.value);
        ElMessage.success('创建成功');
      } else {
        await updateNotificationChannel(form.value.id!, form.value);
        ElMessage.success('更新成功');
      }
      visible.value = false;
      emit('submitted');
    } catch {
      ElMessage.error(props.mode === 'create' ? '创建失败' : '更新失败');
    } finally {
      submitting.value = false;
    }
  }
</script>

<template>
  <ElDialog
    v-model="visible"
    :title="mode === 'create' ? '新增通知渠道' : '编辑通知渠道'"
    width="650px"
    :close-on-click-modal="false"
  >
    <ElForm :model="form" label-width="120px">
      <ElFormItem label="渠道类型" required>
        <ElSelect
          v-model="form.channelType"
          placeholder="请选择渠道类型"
          :disabled="mode === 'edit'"
        >
          <ElOption v-for="type in channelTypeOptions" :key="type.value" :label="type.label" :value="type.value" />
        </ElSelect>
      </ElFormItem>

      <ElFormItem label="渠道名称" required>
        <ElInput v-model="form.channelName" placeholder="请输入渠道名称" />
      </ElFormItem>

      <!-- 邮件配置 -->
      <template v-if="form.channelType === 'email'">
        <ElDivider content-position="left">邮件服务器配置</ElDivider>

        <ElFormItem label="SMTP 服务器" required>
          <ElInput
            v-model="(form.config as Monitoring.EmailConfig).smtpHost"
            placeholder="smtp.example.com"
          />
          <span class="ml-2 text-gray-500">服务器地址</span>
        </ElFormItem>

        <ElFormItem label="SMTP 端口" required>
          <ElInputNumber
            v-model="(form.config as Monitoring.EmailConfig).smtpPort"
            :min="1"
            :max="65535"
            controls-position="right"
          />
        </ElFormItem>

        <ElFormItem label="用户名">
          <ElInput v-model="(form.config as Monitoring.EmailConfig).username" placeholder="认证用户名" />
        </ElFormItem>

        <ElFormItem label="密码">
          <ElInput
            v-model="(form.config as Monitoring.EmailConfig).password"
            type="password"
            placeholder="认证密码"
            show-password
          />
        </ElFormItem>

        <ElFormItem label="发件人邮箱" required>
          <ElInput
            v-model="(form.config as Monitoring.EmailConfig).from"
            placeholder="noreply@example.com"
          />
        </ElFormItem>

        <ElFormItem label="发件人名称">
          <ElInput v-model="(form.config as Monitoring.EmailConfig).fromName" placeholder="OneOps 监控" />
        </ElFormItem>

        <ElFormItem label="启用 TLS">
          <ElSwitch v-model="(form.config as Monitoring.EmailConfig).useTLS" />
          <span class="ml-2 text-gray-500">使用加密连接</span>
        </ElFormItem>
      </template>

      <!-- 企业微信配置 -->
      <template v-if="form.channelType === 'wechat'">
        <ElDivider content-position="left">企业微信机器人配置</ElDivider>

        <ElFormItem label="Webhook URL" required>
          <ElInput
            v-model="(form.config as Monitoring.WeChatConfig).webhookUrl"
            type="textarea"
            :rows="3"
            placeholder="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx"
          />
        </ElFormItem>

        <ElAlert title="提示" type="info" :closable="false" show-icon class="mb-4">
          <template #default>
            <div>在企业微信群聊中添加机器人，获取 Webhook 地址。</div>
            <div>支持 @ 提醒指定人。</div>
          </template>
        </ElAlert>
      </template>

      <!-- 钉钉/飞书配置 -->
      <template v-if="form.channelType === 'dingtalk' || form.channelType === 'feishu'">
        <ElDivider content-position="left">
          {{ form.channelType === 'dingtalk' ? '钉钉' : '飞书' }} 机器人配置
        </ElDivider>

        <ElFormItem label="Webhook URL" required>
          <ElInput
            v-model="(form.config as Monitoring.WeChatConfig).webhookUrl"
            type="textarea"
            :rows="3"
            placeholder="请输入 Webhook 地址"
          />
        </ElFormItem>

        <ElAlert title="提示" type="info" :closable="false" show-icon>
          <template #default>
            <div>在对应的平台群聊中添加自定义机器人，获取 Webhook 地址。</div>
          </template>
        </ElAlert>
      </template>

      <ElFormItem label="启用">
        <ElSwitch v-model="form.enabled" />
        <span class="ml-2 text-gray-500">启用后该渠道将接收告警通知</span>
      </ElFormItem>
    </ElForm>

    <template #footer>
      <ElButton @click="visible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitting" @click="handleSubmit">确定</ElButton>
    </template>
  </ElDialog>
</template>
