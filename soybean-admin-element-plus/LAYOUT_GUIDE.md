# OneOps 前端布局指南

## 🎯 概述

本指南基于参考项目 `sxdevops/frontend/` 的布局模式整理，提供标准化的页面布局方案，确保整个 OneOps 平台的界面一致性和用户体验。

## 📐 设计原则

### 核心原则
1. **一致性优先** - 同类功能页面使用相同的布局模式
2. **渐进增强** - 从基础布局开始，根据需要逐步增加复杂度
3. **响应式设计** - 所有布局必须支持移动端适配
4. **性能考虑** - 避免过度嵌套和不必要的渲染
5. **可维护性** - 使用语义化类名，便于理解和修改

### 布局层次
```
L1: 页面容器 (page-container)
 └─ L2: 功能区域 (hero-section, stats-section, content-section)
     └─ L3: 内容块 (card, panel, toolbar)
         └─ L4: UI组件 (table, form, button)
```

## 🎨 六大标准布局模式

### 1️⃣ Hero + 统计卡片 + 标签页布局

**适用场景**：监控中心、告警管理、资产管理等多功能页面

**结构示例**：
```vue
<template>
  <div class="page-container monitoring-page">
    <!-- Hero 区域 -->
    <section class="hero-section">
      <div class="hero-content">
        <div class="hero-title-row">
          <span class="hero-icon">
            <el-icon><Bell /></el-icon>
          </span>
          <h2>告警监控</h2>
          <p class="hero-desc">统一监控告警，支持聚合、抑制、屏蔽、确认与升级</p>
        </div>
      </div>
      <div class="hero-actions">
        <el-button size="small" :loading="loading" @click="refreshData">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </section>

    <!-- 统计卡片网格 -->
    <div class="stats-grid">
      <div 
        v-for="card in statCards" 
        :key="card.key" 
        class="stat-card content-card"
        :class="card.tone"
        @click="applyStatFilter(card)"
      >
        <div class="stat-value">{{ card.value }}</div>
        <div class="stat-label">{{ card.label }}</div>
        <div class="stat-desc">{{ card.desc }}</div>
      </div>
    </div>

    <!-- 标签页切换 -->
    <div class="neo-tabs theme-blue">
      <button 
        v-for="tab in tabs" 
        :key="tab.key"
        class="neo-tab-btn"
        :class="{ active: activeTab === tab.key }"
        @click="switchTab(tab.key)"
      >
        <el-icon><component :is="tab.icon" /></el-icon>
        {{ tab.label }}
      </button>
    </div>

    <!-- 内容区域 -->
    <section class="content-section">
      <!-- 过滤工具栏 -->
      <div class="toolbar">
        <el-select v-model="filters.level" placeholder="告警级别" size="small" />
        <el-select v-model="filters.status" placeholder="处理状态" size="small" />
        <el-input 
          v-model="filters.search" 
          placeholder="搜索标题 / 来源" 
          :prefix-icon="Search" 
          size="small"
        />
        <el-button type="primary" size="small" @click="handleSearch">
          <el-icon><Search /></el-icon>
          搜索
        </el-button>
        <div class="toolbar-spacer" />
        <el-button 
          v-if="canManage" 
          type="danger" 
          size="small" 
          :disabled="!selectedItems.length"
          @click="handleBatchDelete"
        >
          批量删除
        </el-button>
      </div>

      <!-- 数据表格 -->
      <el-table 
        :data="tableData" 
        class="data-table" 
        v-loading="loading"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="42" />
        <el-table-column prop="id" label="告警ID" width="70" />
        <el-table-column prop="title" label="告警标题" min-width="240" />
        <el-table-column prop="level" label="级别" width="80">
          <template #default="{ row }">
            <el-tag :type="getLevelType(row.level)" size="small">
              {{ getLevelText(row.level) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="handleView(row)">
              详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="table-pagination">
        <el-pagination 
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next"
          @current-change="handlePageChange"
          @size-change="handlePageSizeChange"
        />
      </div>
    </section>
  </div>
</template>

<style scoped lang="scss">
.monitoring-page {
  padding: var(--spacing-lg);
  background: var(--el-bg-color-page);
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.hero-section {
  background: var(--sx-gradient-hero);
  border: 1px solid var(--sx-border-soft);
  border-radius: var(--sx-card-radius);
  box-shadow: var(--sx-card-shadow);
  padding: 14px 22px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--spacing-md);
}

.hero-content {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.hero-title-row {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  flex-wrap: wrap;
}

.hero-icon {
  width: 42px;
  height: 42px;
  border-radius: var(--sx-button-radius);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  color: #fff;
  background: linear-gradient(135deg, var(--sx-primary), var(--sx-primary-light));
  box-shadow: 0 10px 20px rgba(51, 112, 255, 0.2);
}

.hero-title-row h2 {
  color: var(--sx-text-primary);
  font-size: 23px;
  font-weight: 700;
  margin: 0;
}

.hero-desc {
  color: var(--sx-text-secondary);
  font-size: 13px;
  line-height: 1.45;
  margin: 0;
  max-width: 600px;
}

.hero-actions {
  display: flex;
  gap: var(--spacing-sm);
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: var(--spacing-lg);
  margin-bottom: var(--spacing-md);
}

.stat-card {
  cursor: pointer;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
  }

  &::after {
    content: '';
    position: absolute;
    right: -24px;
    bottom: -30px;
    width: 108px;
    height: 108px;
    border-radius: 50%;
    opacity: 0.16;
  }

  &.info-card::after {
    background: radial-gradient(circle, var(--sx-primary) 0%, transparent 70%);
  }

  &.success-card::after {
    background: radial-gradient(circle, var(--sx-success) 0%, transparent 70%);
  }

  &.warning-card::after {
    background: radial-gradient(circle, var(--sx-warning) 0%, transparent 70%);
  }

  &.danger-card::after {
    background: radial-gradient(circle, var(--sx-danger) 0%, transparent 70%);
  }
}

.stat-value {
  font-size: 28px;
  line-height: 1.05;
  color: var(--sx-text-primary);
  font-weight: 700;
  margin-bottom: var(--spacing-xs);
}

.stat-label {
  color: var(--sx-text-secondary);
  font-size: 14px;
  font-weight: 600;
  margin-bottom: var(--spacing-xs);
}

.stat-desc {
  color: var(--sx-text-muted);
  font-size: 12px;
  line-height: 1.4;
}

.neo-tabs {
  display: flex;
  gap: var(--spacing-xs);
  padding: var(--spacing-xs);
  background: rgba(255, 255, 255, 0.9);
  border: 1px solid var(--sx-border-soft);
  border-radius: var(--sx-button-radius);
  box-shadow: var(--sx-card-shadow);
  margin-bottom: var(--spacing-lg);

  &.theme-blue .neo-tab-btn.active {
    color: var(--sx-primary);
    background: rgba(51, 112, 255, 0.1);
    box-shadow: inset 0 0 0 1px rgba(51, 112, 255, 0.08);
  }
}

.neo-tab-btn {
  min-height: 36px;
  padding: 0 var(--spacing-lg);
  border: none;
  background: transparent;
  border-radius: calc(var(--sx-button-radius) - 2px);
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  font-size: 14px;
  color: var(--sx-text-secondary);

  &:hover {
    color: var(--sx-text-primary);
    background: rgba(255, 255, 255, 0.8);
  }

  &.active {
    font-weight: 600;
  }
}

.content-section {
  background: var(--sx-card-bg);
  border: 1px solid var(--sx-border-soft);
  border-radius: var(--sx-card-radius);
  box-shadow: var(--sx-card-shadow);
  padding: var(--spacing-lg);
}

.toolbar {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-md);
  padding: var(--spacing-sm);
  background: var(--sx-toolbar-bg);
  border: 1px solid var(--sx-border-soft);
  border-radius: var(--sx-toolbar-radius);
  flex-wrap: wrap;
}

.toolbar-spacer {
  flex: 1 1 auto;
}

.data-table {
  width: 100%;
  border-radius: var(--sx-table-radius);
  overflow: hidden;
}

.table-pagination {
  display: flex;
  justify-content: flex-end;
  padding-top: var(--spacing-md);
}

@media (max-width: 768px) {
  .hero-section {
    flex-direction: column;
    align-items: flex-start;
  }

  .stats-grid {
    grid-template-columns: 1fr;
  }

  .toolbar {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
```

### 2️⃣ Workbench 卡片布局

**适用场景**：仪表板、查询工作台、分析页面等需要展示多个独立内容块的页面

**结构示例**：
```vue
<template>
  <div class="page-container dashboard-page">
    <!-- 主工作台卡片 -->
    <div class="workbench-card">
      <!-- 卡片头部 -->
      <div class="workbench-card-head">
        <div class="workbench-card-title">
          <strong>运行概览</strong>
          <span>聚焦所选时间范围内的调用明细与模型成本</span>
        </div>
        <div class="workbench-card-actions">
          <el-date-picker
            v-model="timeRange"
            type="datetimerange"
            format="YYYY-MM-DD HH:mm"
            placeholder="选择时间范围"
          />
          <el-button type="primary" size="small" @click="loadData">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </div>

      <!-- 内容网格 -->
      <div class="content-grid">
        <!-- 图表区域 -->
        <div class="chart-section">
          <div class="chart-grid">
            <div 
              v-for="chart in charts" 
              :key="chart.id" 
              class="chart-card content-card"
            >
              <div class="chart-card-head">
                <strong>{{ chart.title }}</strong>
                <el-tag size="small">{{ chart.total }}</el-tag>
              </div>
              <div class="chart-body">
                <div :style="{ height: chart.height }" ref="chartRefs"></div>
              </div>
              <div class="chart-legend">
                <div 
                  v-for="item in chart.legend" 
                  :key="item.key"
                  class="legend-item"
                >
                  <span class="legend-dot" :style="{ background: item.color }"></span>
                  <span class="legend-label">{{ item.label }}</span>
                  <span class="legend-value">{{ item.value }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 统计面板 -->
        <div class="stats-panel">
          <div class="panel-head">
            <span class="section-title">统计指标</span>
          </div>
          <div class="mini-stats-grid">
            <div 
              v-for="stat in miniStats" 
              :key="stat.key" 
              class="mini-stat content-card"
            >
              <span class="mini-stat-label">{{ stat.label }}</span>
              <strong class="mini-stat-value">{{ stat.value }}</strong>
            </div>
          </div>
          <div class="ranking-list">
            <div 
              v-for="item in rankings" 
              :key="item.id" 
              class="ranking-item"
            >
              <div class="ranking-main">
                <div class="ranking-title">
                  <span>{{ item.title }}</span>
                  <strong>{{ item.count }}</strong>
                </div>
                <div class="ranking-bar">
                  <span :style="{ width: `${item.percent}%` }"></span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.dashboard-page {
  padding: var(--spacing-lg);
  background: var(--el-bg-color-page);
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.workbench-card {
  background: var(--sx-gradient-card);
  border: 1px solid var(--sx-border-medium);
  border-radius: var(--sx-card-radius);
  box-shadow: var(--sx-card-shadow);
  padding: var(--spacing-lg);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.workbench-card-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--spacing-lg);
  padding-bottom: var(--spacing-md);
  border-bottom: 1px solid var(--sx-border-soft);
}

.workbench-card-title {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);

  strong {
    color: var(--sx-text-primary);
    font-size: 18px;
    font-weight: 700;
  }

  span {
    color: var(--sx-text-secondary);
    font-size: 13px;
  }
}

.workbench-card-actions {
  display: flex;
  gap: var(--spacing-sm);
  align-items: center;
}

.content-grid {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: var(--spacing-lg);
}

.chart-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: var(--spacing-md);
}

.chart-card {
  min-height: 400px;
  display: flex;
  flex-direction: column;
}

.chart-card-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-md);
}

.chart-card-head strong {
  color: var(--sx-text-primary);
  font-size: 15px;
  font-weight: 600;
}

.chart-body {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.chart-legend {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
  margin-top: var(--spacing-md);
}

.legend-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  font-size: 12px;
}

.legend-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.legend-label {
  color: var(--sx-text-secondary);
  flex: 1;
}

.legend-value {
  color: var(--sx-text-primary);
  font-weight: 600;
}

.stats-panel {
  background: var(--sx-bg-card);
  border: 1px solid var(--sx-border-soft);
  border-radius: calc(var(--sx-card-radius) - 4px);
  padding: var(--spacing-md);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.panel-head {
  margin-bottom: var(--spacing-sm);
}

.section-title {
  color: var(--sx-text-primary);
  font-size: 14px;
  font-weight: 700;
}

.mini-stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--spacing-sm);
}

.mini-stat {
  padding: var(--spacing-sm);
  text-align: center;
  background: linear-gradient(180deg, #fff 0%, #f8fafc 100%);
  border: 1px solid var(--sx-border-soft);
  border-radius: calc(var(--sx-button-radius) - 2px);
}

.mini-stat-label {
  display: block;
  color: var(--sx-text-secondary);
  font-size: 12px;
  font-weight: 600;
  margin-bottom: var(--spacing-xs);
}

.mini-stat-value {
  color: var(--sx-text-primary);
  font-size: 20px;
  font-weight: 800;
}

.ranking-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.ranking-item {
  padding: var(--spacing-sm);
  background: rgba(255, 255, 255, 0.8);
  border: 1px solid var(--sx-border-soft);
  border-radius: calc(var(--sx-button-radius) - 2px);
}

.ranking-title {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-xs);

  span {
    color: var(--sx-text-primary);
    font-size: 13px;
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  strong {
    color: var(--sx-text-primary);
    font-size: 13px;
    font-weight: 700;
  }
}

.ranking-bar {
  height: 4px;
  border-radius: 999px;
  background: #e8eef7;
  overflow: hidden;

  span {
    display: block;
    height: 100%;
    background: linear-gradient(90deg, #60a5fa, #2563eb);
    border-radius: inherit;
    transition: width 0.3s ease;
  }
}

@media (max-width: 1024px) {
  .content-grid {
    grid-template-columns: 1fr;
  }

  .chart-grid {
    grid-template-columns: 1fr;
  }

  .mini-stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
```

### 3️⃣ 标签页 + 内容卡片布局

**适用场景**：用户管理、角色管理、配置管理等需要切换不同子功能的页面

**结构示例**：
```vue
<template>
  <div class="page-container users-page">
    <!-- Hero 区域 -->
    <section class="hero-section">
      <div class="hero-content">
        <div class="hero-title-row">
          <span class="hero-icon">
            <el-icon><User /></el-icon>
          </span>
          <h2>用户管理</h2>
          <p class="hero-desc">统一维护用户、用户组、角色与权限字典</p>
        </div>
      </div>
    </section>

    <!-- 标签页 -->
    <div class="neo-tabs theme-blue">
      <button 
        v-for="tab in tabs" 
        :key="tab.key"
        class="neo-tab-btn"
        :class="{ active: activeTab === tab.key }"
        @click="switchTab(tab.key)"
      >
        <el-icon><component :is="tab.icon" /></el-icon>
        {{ tab.label }}
      </button>
    </div>

    <!-- 内容卡片 -->
    <div class="content-card users-content-card">
      <!-- 工具栏 -->
      <div class="card-toolbar">
        <div class="toolbar-head">
          <span class="toolbar-title">{{ currentSectionTitle }}</span>
          <span class="toolbar-desc">{{ currentSectionDesc }}</span>
        </div>
        <div class="toolbar-actions">
          <el-button 
            v-if="canSync" 
            @click="handleSync"
          >
            <el-icon><Refresh /></el-icon>
            同步内置权限
          </el-button>
          <el-button 
            v-if="canAdd" 
            type="primary" 
            @click="handleAdd"
          >
            <el-icon><Plus /></el-icon>
            新增
          </el-button>
        </div>
      </div>

      <!-- 搜索工具栏 -->
      <div class="search-toolbar">
        <el-input 
          v-model="searchText" 
          :placeholder="searchPlaceholder" 
          clearable
          @input="handleSearch"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
      </div>

      <!-- 数据表格 -->
      <el-table 
        :data="tableData" 
        class="data-table" 
        v-loading="loading"
      >
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="名称" min-width="150" />
        <el-table-column prop="description" label="描述" min-width="200" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'info'" size="small">
              {{ row.enabled ? '启用' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="handleEdit(row)">
              编辑
            </el-button>
            <el-button link type="danger" size="small" @click="handleDelete(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="table-pagination">
        <el-pagination 
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next"
        />
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.users-page {
  padding: var(--spacing-lg);
  background: var(--el-bg-color-page);
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.content-card {
  background: var(--sx-card-bg);
  border: 1px solid var(--sx-border-soft);
  border-radius: var(--sx-card-radius);
  box-shadow: var(--sx-card-shadow);
  padding: var(--spacing-lg);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.card-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--spacing-md);
  padding-bottom: var(--spacing-md);
  border-bottom: 1px solid var(--sx-border-soft);
}

.toolbar-head {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
}

.toolbar-title {
  color: var(--sx-text-primary);
  font-size: 16px;
  font-weight: 700;
}

.toolbar-desc {
  color: var(--sx-text-secondary);
  font-size: 12px;
}

.toolbar-actions {
  display: flex;
  gap: var(--spacing-sm);
  align-items: center;
}

.search-toolbar {
  display: flex;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm);
  background: var(--sx-toolbar-bg);
  border: 1px solid var(--sx-border-soft);
  border-radius: var(--sx-toolbar-radius);
}

.data-table {
  flex: 1;
  border-radius: var(--sx-table-radius);
}

.table-pagination {
  display: flex;
  justify-content: flex-end;
  padding-top: var(--spacing-md);
}
</style>
```

### 4️⃣ 提供者网格 + 内容网格布局

**适用场景**：日志查询、数据源选择、服务配置等需要先选择提供者再配置功能的页面

**结构示例**：
```vue
<template>
  <div class="page-container logs-page">
    <!-- Hero 区域 -->
    <section class="hero-section">
      <div class="hero-content">
        <p class="eyebrow">Log Center</p>
        <h2>统一日志工作台</h2>
        <p class="hero-desc">在 Loki、Elasticsearch 和阿里云 SLS 之间无缝切换</p>
      </div>
      <div class="hero-actions">
        <el-button @click="loadCatalog" :loading="catalogLoading">
          <el-icon><Refresh /></el-icon>
          刷新目录
        </el-button>
        <el-button type="primary" @click="runQuery" :loading="queryLoading">
          <el-icon><CaretRight /></el-icon>
          执行查询
        </el-button>
      </div>
    </section>

    <!-- 提供者选择网格 -->
    <section class="provider-grid">
      <button 
        v-for="provider in providers" 
        :key="provider.id"
        class="provider-card"
        :class="{ active: activeProvider === provider.id }"
        @click="selectProvider(provider.id)"
      >
        <div class="provider-head">
          <strong>{{ provider.name }}</strong>
          <el-tag 
            size="small" 
            :type="provider.configured ? 'success' : 'info'"
          >
            {{ provider.configured ? '已配置' : '可选' }}
          </el-tag>
        </div>
        <p class="provider-desc">{{ provider.description }}</p>
      </button>
    </section>

    <!-- 内容配置网格 -->
    <section class="content-grid">
      <!-- 连接配置面板 -->
      <div class="panel connection-panel">
        <div class="panel-head">
          <h3>连接配置</h3>
          <span>{{ currentProviderName }}</span>
        </div>

        <div class="panel-body">
          <el-form label-position="top">
            <el-form-item label="服务地址">
              <el-input 
                v-model="config.endpoint" 
                placeholder="http://localhost:9200" 
              />
            </el-form-item>
            
            <template v-if="currentProvider === 'elasticsearch'">
              <el-form-item label="认证方式">
                <el-select v-model="config.authType">
                  <el-option label="无认证" value="none" />
                  <el-option label="Basic认证" value="basic" />
                  <el-option label="API Key" value="api_key" />
                </el-select>
              </el-form-item>
              
              <el-form-item v-if="config.authType === 'basic'" label="用户名">
                <el-input v-model="config.username" />
              </el-form-item>
              
              <el-form-item v-if="config.authType === 'basic'" label="密码">
                <el-input v-model="config.password" show-password />
              </el-form-item>
            </template>
          </el-form>
        </div>
      </div>

      <!-- 查询配置面板 -->
      <div class="panel query-panel">
        <div class="panel-head">
          <h3>查询配置</h3>
          <span>{{ queryTypeLabel }}</span>
        </div>

        <div class="panel-body">
          <el-form label-position="top">
            <el-form-item label="时间范围">
              <el-date-picker
                v-model="queryTimeRange"
                type="datetimerange"
                value-format="x"
                format="YYYY-MM-DD HH:mm:ss"
                range-separator="至"
                start-placeholder="开始时间"
                end-placeholder="结束时间"
              />
            </el-form-item>
            
            <el-form-item label="结果限制">
              <el-input-number 
                v-model="queryLimit" 
                :min="20" 
                :max="2000" 
                :step="20" 
              />
            </el-form-item>
            
            <el-form-item label="查询语句">
              <el-input 
                v-model="queryStatement" 
                type="textarea" 
                :rows="4"
                placeholder="输入查询语句..."
              />
            </el-form-item>
          </el-form>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped lang="scss">
.logs-page {
  padding: var(--spacing-lg);
  background: var(--el-bg-color-page);
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.provider-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-lg);
}

.provider-card {
  background: var(--sx-card-bg);
  border: 2px solid var(--sx-border-soft);
  border-radius: var(--sx-card-radius);
  padding: var(--spacing-md);
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);

  &:hover {
    border-color: var(--sx-primary);
    box-shadow: 0 4px 12px rgba(51, 112, 255, 0.15);
    transform: translateY(-2px);
  }

  &.active {
    border-color: var(--sx-primary);
    background: linear-gradient(135deg, #f0f7ff 0%, #ffffff 100%);
    box-shadow: 0 8px 20px rgba(51, 112, 255, 0.2);
  }
}

.provider-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-xs);

  strong {
    color: var(--sx-text-primary);
    font-size: 15px;
    font-weight: 700;
  }
}

.provider-desc {
  color: var(--sx-text-secondary);
  font-size: 13px;
  line-height: 1.4;
  margin: 0;
}

.content-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--spacing-lg);
}

.panel {
  background: var(--sx-card-bg);
  border: 1px solid var(--sx-border-soft);
  border-radius: var(--sx-card-radius);
  box-shadow: var(--sx-card-shadow);
  padding: var(--spacing-md);
  display: flex;
  flex-direction: column;
}

.panel-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-md);
  padding-bottom: var(--spacing-sm);
  border-bottom: 1px solid var(--sx-border-soft);

  h3 {
    color: var(--sx-text-primary);
    font-size: 14px;
    font-weight: 700;
    margin: 0;
  }

  span {
    color: var(--sx-text-secondary);
    font-size: 12px;
  }
}

.panel-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

@media (max-width: 768px) {
  .provider-grid {
    grid-template-columns: 1fr;
  }

  .content-grid {
    grid-template-columns: 1fr;
  }
}
</style>
```

### 5️⃣ 简化 Hero + 组件布局

**适用场景**：任务工作台、独立功能页面等主要内容通过组件实现的页面

**结构示例**：
```vue
<template>
  <div class="page-container task-page">
    <!-- 简化 Hero 区域 -->
    <section class="hero-section task-hero">
      <div class="hero-content">
        <div class="hero-title-row">
          <span class="hero-icon task-icon">
            <el-icon><Operation /></el-icon>
          </span>
          <h2>任务工作台</h2>
          <p class="hero-desc">集中处理任务下发、模板复用与执行回溯</p>
        </div>
      </div>
    </section>

    <!-- 独立功能组件 -->
    <TaskCenter 
      :resources="resourceTree" 
      :loading="loading"
      @refresh="loadResources"
    />
  </div>
</template>

<style scoped lang="scss">
.task-page {
  padding: var(--spacing-lg);
  background: var(--el-bg-color-page);
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.task-hero {
  background: linear-gradient(135deg, #fbfdff 0%, #f7faff 52%, #f9fbfd 100%);
  border: 1px solid rgba(36, 91, 219, 0.09);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--spacing-md);
}

.hero-icon {
  width: 42px;
  height: 42px;
  border-radius: var(--sx-button-radius);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  color: var(--sx-primary);
  background: linear-gradient(180deg, #f3f7ff 0%, #ebf2ff 100%);
  border: 1px solid rgba(36, 91, 219, 0.12);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.8);
}

.hero-title-row {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  flex-wrap: wrap;
}

.hero-title-row h2 {
  color: var(--sx-text-primary);
  font-size: 23px;
  font-weight: 700;
  margin: 0;
}

.hero-desc {
  color: var(--sx-text-secondary);
  font-size: 13px;
  line-height: 1.45;
  margin: 0;
}
</style>
```

### 6️⃣ 纯 Workbench 卡片布局

**适用场景**：SQL查询、数据编辑、表单填写等功能单一的页面

**结构示例**：
```vue
<template>
  <div class="page-container query-page">
    <!-- 查询配置卡片 -->
    <div class="workbench-card query-card">
      <div class="workbench-card-head">
        <div class="workbench-card-title">
          <strong>只读查询台</strong>
          <span>安全的数据库查询执行环境</span>
        </div>
        <div class="workbench-card-actions">
          <el-button 
            type="primary" 
            @click="executeQuery"
            :loading="querying"
            :disabled="!canExecute"
          >
            <el-icon><CaretRight /></el-icon>
            执行查询
          </el-button>
        </div>
      </div>

      <!-- 数据源选择工具栏 -->
      <div class="workbench-toolbar">
        <div class="toolbar-left">
          <el-select 
            v-model="selectedDatasource" 
            placeholder="选择数据源" 
            style="width: 220px"
            @change="onDatasourceChange"
          >
            <el-option 
              v-for="ds in datasources" 
              :key="ds.id" 
              :label="ds.name" 
              :value="ds.id"
            >
              <span>{{ ds.name }}</span>
              <span style="color: var(--sx-text-secondary); margin-left: 8px;">
                {{ ds.type }} / {{ ds.host }}
              </span>
            </el-option>
          </el-select>
          
          <el-select 
            v-model="selectedDatabase" 
            placeholder="选择数据库" 
            style="width: 200px"
            :loading="dbLoading"
          >
            <el-option 
              v-for="db in databases" 
              :key="db" 
              :label="db" 
              :value="db" 
            />
          </el-select>
        </div>
      </div>

      <!-- SQL编辑器 -->
      <div class="editor-wrapper">
        <textarea 
          v-model="sqlContent" 
          class="sql-editor"
          placeholder="输入 SQL 查询语句..."
          rows="8"
          @keydown.ctrl.enter="executeQuery"
        ></textarea>
      </div>

      <!-- 查询提示 -->
      <div class="query-hint">
        <el-icon><InfoFilled /></el-icon>
        <span>按 Ctrl+Enter 快速执行查询 | 支持标准 SQL 语法</span>
      </div>
    </div>

    <!-- 查询结果卡片 -->
    <div v-if="queryResult || queryError" class="workbench-card result-card">
      <div class="workbench-card-head">
        <div class="workbench-card-title">
          <strong>查询结果</strong>
        </div>
        <div class="workbench-card-actions" v-if="queryResult">
          <el-tag type="info" size="small">{{ queryResult.rowCount }} 行</el-tag>
          <el-tag type="success" size="small">{{ queryResult.duration }}ms</el-tag>
        </div>
      </div>

      <!-- 错误显示 -->
      <div v-if="queryError" class="error-section">
        <el-alert 
          :title="queryError" 
          type="error" 
          show-icon 
          :closable="false" 
        />
      </div>

      <!-- 结果表格 -->
      <el-table 
        v-else 
        :data="queryResult.rows" 
        stripe 
        size="small"
        max-height="400"
      >
        <el-table-column 
          v-for="col in queryResult.columns" 
          :key="col"
          :prop="col" 
          :label="col" 
          min-width="120" 
          show-overflow-tooltip 
        />
      </el-table>
    </div>

    <!-- 查询历史卡片 -->
    <div v-if="canViewHistory" class="workbench-card history-card">
      <div class="workbench-card-head">
        <div class="workbench-card-title">
          <strong>查询历史</strong>
          <span>最近执行记录会沉淀在这里，方便回看语句与结果规模</span>
        </div>
      </div>

      <el-table 
        :data="history" 
        stripe 
        v-loading="historyLoading" 
        size="small"
      >
        <el-table-column prop="datasource_name" label="数据源" width="130" />
        <el-table-column label="类型" width="110">
          <template #default="{ row }">
            <el-tag size="small">{{ row.db_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="database" label="数据库" width="120" />
        <el-table-column prop="sql_content" label="查询内容" min-width="250" show-overflow-tooltip />
        <el-table-column prop="submitter" label="操作人" width="90" />
        <el-table-column prop="result_count" label="结果行数" width="100" />
        <el-table-column prop="duration_ms" label="耗时" width="90">
          <template #default="{ row }">
            {{ row.duration_ms }}ms
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="时间" width="170">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
      </el-table>

      <div class="table-pagination">
        <el-pagination 
          v-model:current-page="historyPage"
          v-model:page-size="historyPageSize"
          :total="historyTotal"
          layout="total, prev, pager, next"
          @current-change="loadHistory"
        />
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.query-page {
  padding: var(--spacing-lg);
  background: var(--el-bg-color-page);
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.workbench-card {
  background: var(--sx-gradient-card);
  border: 1px solid var(--sx-border-medium);
  border-radius: var(--sx-card-radius);
  box-shadow: var(--sx-card-shadow);
  padding: var(--spacing-lg);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.workbench-card-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--spacing-lg);
  padding-bottom: var(--spacing-md);
  border-bottom: 1px solid var(--sx-border-soft);
}

.workbench-card-title {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);

  strong {
    color: var(--sx-text-primary);
    font-size: 16px;
    font-weight: 700;
  }

  span {
    color: var(--sx-text-secondary);
    font-size: 12px;
  }
}

.workbench-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm);
  background: var(--sx-toolbar-bg);
  border: 1px solid var(--sx-border-soft);
  border-radius: var(--sx-toolbar-radius);
}

.toolbar-left {
  display: flex;
  gap: var(--spacing-sm);
  align-items: center;
}

.editor-wrapper {
  position: relative;
  border: 1px solid var(--sx-border-soft);
  border-radius: var(--sx-input-radius);
  overflow: hidden;
  
  &:focus-within {
    border-color: var(--sx-primary);
    box-shadow: 0 0 0 2px rgba(51, 112, 255, 0.1);
  }
}

.sql-editor {
  width: 100%;
  border: none;
  outline: none;
  padding: var(--spacing-md);
  font-family: 'JetBrains Mono', 'SFMono-Regular', Consolas, monospace;
  font-size: 13px;
  line-height: 1.6;
  resize: vertical;
  background: var(--sx-input-bg);
  color: var(--sx-text-primary);
}

.query-hint {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  color: var(--sx-text-muted);
  font-size: 12px;
  padding: var(--spacing-xs) var(--spacing-sm);
  background: rgba(248, 250, 252, 0.8);
  border-radius: calc(var(--sx-button-radius) - 4px);
}

.error-section {
  padding: var(--spacing-sm);
}

.table-pagination {
  display: flex;
  justify-content: flex-end;
  padding-top: var(--spacing-md);
}
</style>
```

## 🧩 布局组件库

### 核心组件类名规范

```scss
// 页面容器
.page-container {
  padding: var(--spacing-lg);
  background: var(--el-bg-color-page);
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

// Hero 区域
.hero-section {
  background: var(--sx-gradient-hero);
  border: 1px solid var(--sx-border-soft);
  border-radius: var(--sx-card-radius);
  box-shadow: var(--sx-card-shadow);
  padding: 14px 22px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--spacing-md);
}

// 统计卡片网格
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: var(--spacing-lg);
  margin-bottom: var(--spacing-md);
}

.stat-card {
  background: var(--sx-card-bg);
  border: 1px solid var(--sx-border-soft);
  border-radius: var(--sx-card-radius);
  padding: var(--spacing-md);
  text-align: center;
  transition: all 0.3s ease;
  
  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
  }
}

// 工作台卡片
.workbench-card {
  background: var(--sx-gradient-card);
  border: 1px solid var(--sx-border-medium);
  border-radius: var(--sx-card-radius);
  box-shadow: var(--sx-card-shadow);
  padding: var(--spacing-lg);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

// 内容卡片
.content-card {
  background: var(--sx-card-bg);
  border: 1px solid var(--sx-border-soft);
  border-radius: var(--sx-card-radius);
  box-shadow: var(--sx-card-shadow);
  padding: var(--spacing-lg);
}

// 工具栏
.toolbar {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm);
  background: var(--sx-toolbar-bg);
  border: 1px solid var(--sx-border-soft);
  border-radius: var(--sx-toolbar-radius);
  flex-wrap: wrap;
}

.toolbar-spacer {
  flex: 1 1 auto;
}

// 标签页
.neo-tabs {
  display: flex;
  gap: var(--spacing-xs);
  padding: var(--spacing-xs);
  background: rgba(255, 255, 255, 0.9);
  border: 1px solid var(--sx-border-soft);
  border-radius: var(--sx-button-radius);
  box-shadow: var(--sx-card-shadow);
  margin-bottom: var(--spacing-lg);
}

.neo-tab-btn {
  min-height: 36px;
  padding: 0 var(--spacing-lg);
  border: none;
  background: transparent;
  border-radius: calc(var(--sx-button-radius) - 2px);
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  font-size: 14px;
  color: var(--sx-text-secondary);
  
  &:hover {
    color: var(--sx-text-primary);
    background: rgba(255, 255, 255, 0.8);
  }
  
  &.active {
    font-weight: 600;
  }
}

// 面板
.panel {
  background: var(--sx-card-bg);
  border: 1px solid var(--sx-border-soft);
  border-radius: var(--sx-card-radius);
  box-shadow: var(--sx-card-shadow);
  padding: var(--spacing-md);
  display: flex;
  flex-direction: column;
}

.panel-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-md);
  padding-bottom: var(--spacing-sm);
  border-bottom: 1px solid var(--sx-border-soft);
}

// 提供者网格
.provider-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-lg);
}

.provider-card {
  background: var(--sx-card-bg);
  border: 2px solid var(--sx-border-soft);
  border-radius: var(--sx-card-radius);
  padding: var(--spacing-md);
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
  
  &:hover {
    border-color: var(--sx-primary);
    box-shadow: 0 4px 12px rgba(51, 112, 255, 0.15);
    transform: translateY(-2px);
  }
  
  &.active {
    border-color: var(--sx-primary);
    background: linear-gradient(135deg, #f0f7ff 0%, #ffffff 100%);
    box-shadow: 0 8px 20px rgba(51, 112, 255, 0.2);
  }
}

// 内容网格
.content-grid {
  display: grid;
  gap: var(--spacing-lg);
  
  &.two-column {
    grid-template-columns: 1fr 1fr;
  }
  
  &.three-column {
    grid-template-columns: repeat(3, 1fr);
  }
  
  @media (max-width: 768px) {
    grid-template-columns: 1fr;
  }
}
```

## 📋 布局选择决策树

```
开始
  │
  ├─ 页面是否需要多个独立功能区？
  │   ├─ 是 → 使用布局 1 (Hero + 统计卡片 + 标签页)
  │   └─ 否 → 继续判断
  │
  ├─ 页面主要是展示图表和统计数据？
  │   ├─ 是 → 使用布局 2 (Workbench 卡片)
  │   └─ 否 → 继续判断
  │
  ├─ 页面是否需要先选择服务/数据源？
  │   ├─ 是 → 使用布局 4 (提供者网格 + 内容网格)
  │   └─ 否 → 继续判断
  │
  ├─ 页面主要内容是通过组件实现？
  │   ├─ 是 → 使用布局 5 (简化 Hero + 组件)
  │   └─ 否 → 继续判断
  │
  ├─ 页面是单一功能（如查询、编辑）？
  │   ├─ 是 → 使用布局 6 (纯 Workbench 卡片)
  │   └─ 否 → 使用布局 3 (标签页 + 内容卡片)
```

## 🎯 使用场景对照表

| 页面类型 | 推荐布局 | 关键特征 | 示例页面 |
|---------|---------|---------|---------|
| 监控告警 | 布局1 | Hero + 统计卡片 + 多标签页 | /monitoring/alerts |
| 资产管理 | 布局1 | Hero + 统计卡片 + 标签切换 | /cmdb/hosts |
| 仪表板 | 布局2 | Workbench卡片 + 图表网格 | /dashboard |
| 查询工作台 | 布局2 | 独立卡片 + 编辑器 + 结果 | /sql/query |
| 用户管理 | 布局3 | Hero + 标签页 + 内容卡片 | /system/users |
| 角色管理 | 布局3 | Hero + 标签页 + 数据表格 | /system/roles |
| 日志查询 | 布局4 | 提供者选择 + 配置网格 | /logs/query |
| 数据源管理 | 布局4 | 提供者卡片 + 详细配置 | /datasources |
| 任务工作台 | 布局5 | 简化 Hero + 独立组件 | /tasks/workbench |
| 文件管理 | 布局5 | Hero + 文件浏览器组件 | /files/browser |
| SQL查询 | 布局6 | 纯卡片布局 + 编辑器 | /sql/editor |
| 表单编辑 | 布局6 | 表单卡片 + 提交按钮 | /forms/editor |

## 🚀 最佳实践

### 1. 组件复用策略

创建可复用的布局组件：

```vue
<!-- components/layout/PageContainer.vue -->
<template>
  <div :class="containerClass" :style="containerStyle">
    <slot />
  </div>
</template>

<script setup lang="ts">
interface Props {
  type?: 'default' | 'workbench' | 'compact';
  padding?: 'none' | 'small' | 'medium' | 'large';
}

const props = withDefaults(defineProps<Props>(), {
  type: 'default',
  padding: 'large'
});

const containerClass = computed(() => [
  'page-container',
  `page-container--${props.type}`,
  `page-container--padding-${props.padding}`
]);
</script>

<style scoped lang="scss">
.page-container {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
  
  &--padding-none { padding: 0; }
  &--padding-small { padding: var(--spacing-sm); }
  &--padding-medium { padding: var(--spacing-md); }
  &--padding-large { padding: var(--spacing-lg); }
  
  &--workbench {
    background: var(--el-bg-color-page);
    min-height: 100vh;
  }
  
  &--compact {
    gap: var(--spacing-sm);
  }
}
</style>
```

### 2. 主题变量使用

在所有布局中使用统一的 CSS 变量：

```scss
// 在组件中使用
.my-component {
  padding: var(--spacing-lg);
  background: var(--sx-card-bg);
  border: 1px solid var(--sx-border-soft);
  border-radius: var(--sx-card-radius);
  color: var(--sx-text-primary);
  
  &:hover {
    background: var(--sx-table-hover-bg);
  }
}
```

### 3. 响应式设计

所有布局必须支持移动端：

```scss
// 移动端适配
@media (max-width: 768px) {
  .page-container {
    padding: var(--spacing-md);
  }
  
  .hero-section {
    flex-direction: column;
    align-items: flex-start;
  }
  
  .stats-grid,
  .content-grid,
  .provider-grid {
    grid-template-columns: 1fr;
  }
  
  .toolbar {
    flex-direction: column;
    align-items: stretch;
  }
}
```

### 4. 性能优化

- 使用 `v-show` 替代 `v-if` 来切换标签页内容（保留状态）
- 图表组件使用 `requestAnimationFrame` 优化渲染
- 大数据表格使用虚拟滚动
- 图片和图标使用懒加载

### 5. 可访问性

- 所有交互元素都有适当的焦点状态
- 使用语义化 HTML 标签
- 提供键盘导航支持
- 确保颜色对比度符合 WCAG 2.1 标准

## 📚 相关文档

- [主题设置指南](./theme-settings-guide.md)
- [组件样式指南](./component-style-guide.md)
- [CSS 变量参考](./css-variables-reference.md)
- [响应式设计规范](./responsive-design-spec.md)

## 🔧 工具支持

### 布局验证工具

项目提供了布局验证脚本：

```bash
# 检查页面布局是否符合规范
npm run check:layout

# 自动生成布局组件
npm run gen:layout
```

### 布局调试工具

在开发环境中，可以通过以下方式调试布局：

```vue
<script setup>
// 在开发环境中显示布局调试信息
const isDev = import.meta.env.DEV;

const layoutInfo = {
  container: 'page-container',
  layout: 'hero-stats-tabs',
  breakpoints: {
    mobile: '768px',
    tablet: '1024px',
    desktop: '1280px'
  }
};
</script>

<template>
  <div v-if="isDev" class="layout-debug">
    <div>容器类型: {{ layoutInfo.container }}</div>
    <div>布局模式: {{ layoutInfo.layout }}</div>
    <div>断点信息: {{ layoutInfo.breakpoints }}</div>
  </div>
</template>
```

---

**最后更新**: 2025-01-13  
**维护者**: OneOps 前端团队  
**版本**: v1.0.0