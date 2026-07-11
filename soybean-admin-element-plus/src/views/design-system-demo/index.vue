<template>
  <div class="design-system-demo">
    <h1>OneOps 设计系统示例</h1>

    <!-- 卡片示例 -->
    <section class="demo-section">
      <h2>渐变卡片</h2>
      <div class="demo-grid">
        <GradientCard variant="diagonal" :enable-hover="true">
          <h3>对角线渐变卡片</h3>
          <p>使用135deg对角线渐变，配合悬停效果</p>
        </GradientCard>

        <GradientCard variant="3d" :enable-hover="true">
          <h3>立体感卡片</h3>
          <p>使用145deg渐变创造立体感，配合内阴影</p>
        </GradientCard>

        <GradientCard variant="glass" :enable-hover="true">
          <h3>玻璃拟态卡片</h3>
          <p>使用半透明渐变配合背景模糊</p>
        </GradientCard>

        <GradientCard variant="hover" :enable-hover="true">
          <h3>悬停渐变卡片</h3>
          <p>悬停时显示180deg上下渐变效果</p>
        </GradientCard>
      </div>
    </section>

    <!-- 按钮示例 -->
    <section class="demo-section">
      <h2>渐变按钮</h2>
      <div class="demo-buttons">
        <GradientButton variant="light" @click="handleButtonClick">
          <template #icon>✨</template>
          浅色按钮
        </GradientButton>

        <GradientButton variant="white" @click="handleButtonClick">
          <template #icon>🎯</template>
          白色按钮
        </GradientButton>

        <GradientButton variant="brand" @click="handleButtonClick">
          <template #icon>🚀</template>
          品牌按钮
        </GradientButton>

        <GradientButton variant="blue" @click="handleButtonClick">
          <template #icon>💡</template>
          蓝色按钮
        </GradientButton>

        <GradientButton variant="light" :disabled="true">
          禁用按钮
        </GradientButton>
      </div>
    </section>

    <!-- 面板示例 -->
    <section class="demo-section">
      <h2>渐变面板</h2>
      <div class="demo-panels">
        <GradientPanel variant="light" title="浅色面板">
          <p>这是浅色面板的内容区域，使用对角线渐变背景</p>
          <template #extra>
            <GradientButton variant="light">操作</GradientButton>
          </template>
        </GradientPanel>

        <GradientPanel variant="toolbar" title="工具栏面板">
          <p>工具栏面板使用上下渐变和内阴影效果</p>
          <template #extra>
            <span class="badge">工具</span>
          </template>
        </GradientPanel>

        <GradientPanel variant="glass" title="玻璃面板">
          <p>玻璃拟态面板，使用半透明渐变和背景模糊</p>
          <template #extra>
            <GradientButton variant="white">操作</GradientButton>
          </template>
        </GradientPanel>

        <GradientPanel variant="brand" title="品牌面板">
          <p>品牌色面板，使用品牌渐变色彩</p>
          <template #extra>
            <span class="white-text">特色</span>
          </template>
        </GradientPanel>
      </div>
    </section>

    <!-- 配色组合示例 -->
    <section class="demo-section">
      <h2>配色组合示例</h2>
      <div class="demo-grid">
        <div class="color-combo card">
          <h3>Header + Card</h3>
          <div class="combo-demo">
            <div class="demo-header">Header区域</div>
            <div class="demo-card">Card内容</div>
          </div>
        </div>

        <div class="color-combo card">
          <h3>Button + Panel</h3>
          <div class="combo-demo">
            <GradientButton variant="light">按钮</GradientButton>
            <div class="demo-panel">面板内容</div>
          </div>
        </div>

        <div class="color-combo card">
          <h3>Full Layout</h3>
          <div class="combo-demo full">
            <div class="demo-sidebar">侧边栏</div>
            <div class="demo-main">
              <div class="demo-header">头部</div>
              <div class="demo-content">内容区域</div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- 工具类示例 -->
    <section class="demo-section">
      <h2>工具类示例</h2>
      <div class="demo-grid">
        <div class="bg-gradient-diagonal border-gradient-light shadow-gradient-sm p-4 rounded-12px">
          <h4>对角线渐变背景</h4>
          <p>使用UnoCSS工具类快速创建渐变效果</p>
        </div>

        <div class="bg-gradient-header border-b border-gradient-light p-4">
          <h4>Header渐变</h4>
          <p>180deg上下渐变配合底部分隔线</p>
        </div>

        <div class="gradient-card-3d p-4 shadow-gradient-sm">
          <h4>立体卡片</h4>
          <p>145deg渐变创造立体感</p>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import GradientCard from '@/components/common/gradient-card.vue';
import GradientButton from '@/components/common/gradient-button.vue';
import GradientPanel from '@/components/common/gradient-panel.vue';

const handleButtonClick = () => {
  console.log('按钮被点击');
};
</script>

<style scoped>
.design-system-demo {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.design-system-demo > h1 {
  font-size: 32px;
  font-weight: 700;
  color: #1e293b;
  margin-bottom: 32px;
  text-align: center;
}

.demo-section {
  margin-bottom: 48px;
}

.demo-section > h2 {
  font-size: 24px;
  font-weight: 600;
  color: #1e293b;
  margin-bottom: 24px;
  padding-bottom: 8px;
  border-bottom: 2px solid #e2e8f0;
}

.demo-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 24px;
}

.demo-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  align-items: center;
}

.demo-panels {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 24px;
}

.color-combo {
  padding: 16px;
}

.combo-demo {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.demo-header {
  background: linear-gradient(180deg, rgba(248, 251, 255, 0.98), rgba(243, 247, 255, 0.94));
  border: 1px solid rgba(148, 163, 184, 0.12);
  border-radius: 8px;
  padding: 8px;
  text-align: center;
  font-size: 12px;
  color: #64748b;
}

.demo-card {
  background: linear-gradient(135deg, #ffffff, #f7fbff);
  border: 1px solid rgba(148, 163, 184, 0.12);
  box-shadow: 0 8px 18px rgba(59, 130, 246, 0.14);
  border-radius: 8px;
  padding: 8px;
  text-align: center;
  font-size: 12px;
  color: #64748b;
}

.demo-panel {
  background: linear-gradient(145deg, #eef5ff, #f6fbff);
  border: 1px solid rgba(148, 163, 184, 0.12);
  box-shadow: inset 0 1px rgba(255, 255, 255, 0.95);
  border-radius: 8px;
  padding: 8px;
  text-align: center;
  font-size: 12px;
  color: #64748b;
}

.demo-sidebar {
  background: linear-gradient(180deg, rgba(241, 246, 255, 0.98), rgba(232, 240, 255, 0.9));
  border: 1px solid rgba(148, 163, 184, 0.12);
  border-radius: 8px;
  padding: 8px;
  text-align: center;
  font-size: 10px;
  color: #5f6b7a;
}

.demo-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.demo-content {
  background: linear-gradient(135deg, #ffffff, #f7fbff);
  border: 1px solid rgba(148, 163, 184, 0.12);
  border-radius: 8px;
  padding: 8px;
  text-align: center;
  font-size: 10px;
  color: #64748b;
  flex: 1;
}

.full .combo-demo {
  flex-direction: row;
}

.badge {
  background: rgba(59, 130, 246, 0.1);
  color: #3b82f6;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.white-text {
  color: white;
}

.card {
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.card h3 {
  font-size: 16px;
  font-weight: 600;
  color: #1e293b;
  margin-bottom: 12px;
}

.card p {
  font-size: 14px;
  color: #64748b;
  line-height: 1.5;
}

@media (max-width: 768px) {
  .design-system-demo {
    padding: 16px;
  }

  .demo-grid {
    grid-template-columns: 1fr;
    gap: 16px;
  }

  .demo-panels {
    grid-template-columns: 1fr;
    gap: 16px;
  }

  .full .combo-demo {
    flex-direction: column;
  }
}
</style>