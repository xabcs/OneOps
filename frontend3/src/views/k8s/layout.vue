<script setup lang="ts">
  import { computed, watch } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import {
    Bell,
    Box,
    Connection,
    CopyDocument,
    DataAnalysis,
    DataLine,
    Document,
    DocumentCopy,
    Folder,
    Monitor,
    Operation,
    Service,
    Setting,
    Tools,
    TrendCharts
  } from '@element-plus/icons-vue';

  const route = useRoute();
  const router = useRouter();

  // 当前激活的菜单
  const activeMenu = computed(() => {
    return route.path;
  });

  // 菜单选择处理
  const handleMenuSelect = (index: string) => {
    router.push(index);
  };

  // 监听路由变化，更新菜单激活状态
  watch(
    () => route.path,
    newPath => {}
  );
</script>

<template>
  <div class="k8s-layout">
    <!-- 左侧菜单 -->
    <ElMenu :default-active="activeMenu" class="k8s-menu" @select="handleMenuSelect">
      <ElMenuItem index="/k8s/clusters">
        <ElIcon><Monitor /></ElIcon>
        <span>集群管理</span>
      </ElMenuItem>

      <ElMenuItem index="/k8s/nodes">
        <ElIcon><Connection /></ElIcon>
        <span>节点管理</span>
      </ElMenuItem>

      <ElMenuItem index="/k8s/namespaces">
        <ElIcon><Folder /></ElIcon>
        <span>命名空间</span>
      </ElMenuItem>

      <ElMenuItem index="/k8s/pods">
        <ElIcon><Box /></ElIcon>
        <span>Pod管理</span>
      </ElMenuItem>

      <ElMenuItem index="/k8s/services">
        <ElIcon><Service /></ElIcon>
        <span>服务管理</span>
      </ElMenuItem>

      <ElMenuItem index="/k8s/deployments">
        <ElIcon><CopyDocument /></ElIcon>
        <span>部署管理</span>
      </ElMenuItem>

      <ElSubMenu index="monitoring">
        <template #title>
          <ElIcon><DataAnalysis /></ElIcon>
          <span>监控</span>
        </template>
        <ElMenuItem index="/k8s/monitoring/pods">
          <ElIcon><TrendCharts /></ElIcon>
          <span>Pod监控</span>
        </ElMenuItem>
        <ElMenuItem index="/k8s/monitoring/nodes">
          <ElIcon><DataLine /></ElIcon>
          <span>节点监控</span>
        </ElMenuItem>
      </ElSubMenu>

      <!-- 新增：诊断菜单 -->
      <ElMenuItem index="/k8s/diagnostic" class="diagnostic-menu-item">
        <ElIcon><Operation /></ElIcon>
        <span>诊断中心</span>
        <ElTag class="new-tag" size="small" type="danger">NEW</ElTag>
      </ElMenuItem>

      <ElSubMenu index="tools">
        <template #title>
          <ElIcon><Tools /></ElIcon>
          <span>工具</span>
        </template>
        <ElMenuItem index="/k8s/logs">
          <ElIcon><Document /></ElIcon>
          <span>日志查询</span>
        </ElMenuItem>
        <ElMenuItem index="/k8s/events">
          <ElIcon><Bell /></ElIcon>
          <span>事件查询</span>
        </ElMenuItem>
        <ElMenuItem index="/k8s/yaml">
          <ElIcon><DocumentCopy /></ElIcon>
          <span>YAML编辑器</span>
        </ElMenuItem>
      </ElSubMenu>

      <ElMenuItem index="/k8s/settings">
        <ElIcon><Setting /></ElIcon>
        <span>设置</span>
      </ElMenuItem>
    </ElMenu>

    <!-- 右侧内容区域 -->
    <div class="k8s-content">
      <RouterView v-slot="{ Component }">
        <Transition name="fade" mode="out-in">
          <component :is="Component" />
        </Transition>
      </RouterView>
    </div>
  </div>
</template>

<style scoped lang="scss">
  .k8s-layout {
    display: flex;
    height: 100vh;
    background-color: #f5f7fa;

    .k8s-menu {
      width: 240px;
      height: 100%;
      background-color: #fff;
      border-right: 1px solid #e6e6e6;
      overflow-y: auto;

      .el-menu-item {
        height: 50px;
        line-height: 50px;
        padding: 0 20px;

        &:hover {
          background-color: #ecf5ff;
        }

        &.is-active {
          background-color: #e6f7ff;
          border-right: 3px solid #409eff;
          color: #409eff;
        }

        .el-icon {
          margin-right: 8px;
        }
      }

      // 诊断菜单项特殊样式
      .diagnostic-menu-item {
        position: relative;
        background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        color: #fff;
        border-radius: 8px;
        margin: 8px 12px;
        padding: 12px 16px;

        &:hover {
          background: linear-gradient(135deg, #764ba2 0%, #667eea 100%);
          transform: translateY(-1px);
          box-shadow: 0 4px 8px rgb(102 126 234 / 30%);
        }

        &.is-active {
          background: linear-gradient(135deg, #5a67d8 0%, #6b46c1 100%);
          box-shadow: 0 6px 12px rgb(102 126 234 / 40%);
        }

        .el-icon {
          color: #fff;
        }

        .new-tag {
          position: absolute;
          right: 12px;
          top: 50%;
          transform: translateY(-50%);
          animation: pulse 2s infinite;
        }
      }

      .el-sub-menu {
        .el-sub-menu__title {
          height: 50px;
          line-height: 50px;
          padding: 0 20px;

          .el-icon {
            margin-right: 8px;
          }
        }
      }
    }

    .k8s-content {
      flex: 1;
      overflow: hidden;
      padding: 20px;
    }
  }

  // 新标签动画
  @keyframes pulse {
    0%,
    100% {
      opacity: 1;
      transform: translateY(-50%) scale(1);
    }

    50% {
      opacity: 0.8;
      transform: translateY(-50%) scale(1.1);
    }
  }

  // 路由切换动画
  .fade-enter-active,
  .fade-leave-active {
    transition: opacity 0.3s ease;
  }

  .fade-enter-from,
  .fade-leave-to {
    opacity: 0;
  }
</style>
