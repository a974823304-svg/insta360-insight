<template>
  <div class="page" v-loading="store.loading">
    <!-- 顶部条:数据总览标题 + 操作 -->
    <header class="page-head">
      <div class="title-block">
        <h1>数据总览 <span class="info" title="数据每 5 分钟刷新一次">ⓘ</span></h1>
        <p class="sub">{{ dateRangeText }} · 全球达人营销数据洞察</p>
      </div>
      <div class="actions">
        <el-link type="primary" :underline="false" class="compare">⇆ 与上期对比</el-link>
        <el-link type="primary" :underline="false" class="export">⇪ 导出</el-link>
      </div>
    </header>

    <!-- 5 张 KPI 卡片 -->
    <section class="kpi-row">
      <KpiCard v-for="k in store.kpi" :key="k.key" :kpi="k" />
    </section>

    <!-- 趋势分析 + AI 洞察 -->
    <section class="grid-row-2">
      <div class="card chart-card">
        <div class="card-head">
          <div class="card-title-row">
            <span class="card-title">趋势分析</span>
            <el-select v-model="trendMetric" size="small" class="metric-select">
              <el-option label="播放量" value="views" />
              <el-option label="互动量" value="engagement" />
              <el-option label="粉丝增长" value="followers" />
            </el-select>
          </div>
          <div class="head-actions">
            <!-- 图例(实线+虚线) 由 echarts 在画布右上角渲染, 头部不再重复 -->
            <el-button-group class="granularity">
              <el-button size="small" :type="filter.granularity === 'day' ? 'primary' : 'default'" @click="filter.granularity = 'day'">日</el-button>
              <el-button size="small" :type="filter.granularity === 'week' ? 'primary' : 'default'" @click="filter.granularity = 'week'">周</el-button>
              <el-button size="small" :type="filter.granularity === 'month' ? 'primary' : 'default'" @click="filter.granularity = 'month'">月</el-button>
            </el-button-group>
          </div>
        </div>
        <div class="card-body chart-body">
          <TrendChart :data="store.viewsTrend" :granularity="filter.granularity" />
        </div>
      </div>
      <div class="card insights-card">
        <AIInsights :items="store.insights" />
      </div>
    </section>

    <!-- 三联图:平台分布 / 赛道表现 / 引爆力 -->
    <section class="grid-row-3">
      <div class="card">
        <div class="card-head">
          <div class="card-title">平台分布</div>
          <el-link type="primary" :underline="false" class="more">详情 →</el-link>
        </div>
        <div class="card-body donut-body">
          <div class="donut-chart">
            <PlatformDonut :data="store.platformShare" />
          </div>
          <ul class="legend-col">
            <li v-for="p in store.platformShare" :key="p.platform">
              <span class="dot" :style="{ background: p.color }" />
              <span class="lab">{{ p.platform }}</span>
              <span class="pct num">{{ p.share }}%</span>
            </li>
          </ul>
        </div>
      </div>
      <div class="card">
        <div class="card-head">
          <div>
            <div class="card-title-row">
              <span class="card-title">运动赛道表现</span>
              <el-select v-model="trackMetric" size="small" class="metric-select">
                <el-option label="播放量" value="views" />
                <el-option label="互动量" value="engagement" />
              </el-select>
            </div>
          </div>
        </div>
        <div class="card-body chart-body">
          <TrackBarChart :data="store.tracks" />
        </div>
      </div>
      <div class="card">
        <div class="card-head">
          <div class="card-title">引爆力维度表现</div>
          <div class="card-sub">
            <span class="dash" /> 平均值
          </div>
        </div>
        <div class="card-body chart-body">
          <RadarChart :data="store.radar" />
        </div>
      </div>
    </section>

    <!-- 热门达人 TOP 10 全宽 -->
    <section class="grid-row-full">
      <div class="card creators-card">
        <div class="card-head">
          <div class="card-title">热门达人 TOP 10</div>
          <el-input v-model="creatorSearch" placeholder="搜索达人" size="small" class="search" clearable />
        </div>
        <div class="card-body">
          <TopCreatorsTable />
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import { useFilterStore } from '../stores/filter'
import { useInsightStore } from '../stores/insight'

import KpiCard from '../components/KpiCard.vue'
import TrendChart from '../components/TrendChart.vue'
import AIInsights from '../components/AIInsights.vue'
import PlatformDonut from '../components/PlatformDonut.vue'
import TrackBarChart from '../components/TrackBarChart.vue'
import RadarChart from '../components/RadarChart.vue'
import TopCreatorsTable from '../components/TopCreatorsTable.vue'

const filter = useFilterStore()
const store = useInsightStore()

const creatorSearch = ref('')
const trendMetric = ref('views')
const trackMetric = ref('views')

const dateRangeText = computed(() => {
  if (!filter.dateRange?.length) return ''
  return filter.dateRange.join(' ~ ')
})
onMounted(() => {
  // 第一次进入:并行拉所有数据
  store.loadAll()
})
// 筛选应用后(应用筛选按钮)自动重拉本页数据
watch(() => filter.appliedRevision, () => store.loadAll())
</script>

<style lang="scss" scoped>
/* 全页紧凑版: 适配 1080p 显示器一屏装下完整看板(无滚动) */
.page {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.page-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  .title-block h1 {
    margin: 0;
    font-size: 15px;
    font-weight: 700;
    color: var(--text-primary);
    display: flex;
    align-items: center;
    gap: 6px;
    .info {
      cursor: help;
      color: var(--text-muted);
      font-size: 12px;
    }
  }
  .sub {
    margin: 1px 0 0;
    color: var(--text-muted);
    font-size: 10px;
  }
  .actions {
    display: flex;
    align-items: center;
    gap: 14px;
    .compare, .export { font-size: 10px; }
  }
}

// 5 个 KPI 卡片横向排列(超出可滚动)
.kpi-row {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 6px;
  @media (max-width: 1280px) { grid-template-columns: repeat(3, 1fr); }
  @media (max-width: 768px)  { grid-template-columns: repeat(2, 1fr); }
}

// 通用卡片
.card {
  background: var(--bg-elev);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 8px 10px;
}
.card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 2px;
  min-height: 20px;
  .card-title-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .card-title {
    font-size: 12px;
    font-weight: 600;
    color: var(--text-primary);
  }
  .card-sub {
    margin-top: 1px;
    font-size: 10px;
    color: var(--text-muted);
    display: flex;
    align-items: center;
    gap: 4px;
    .dash {
      display: inline-block;
      width: 12px;
      height: 0;
      border-top: 1.5px dashed var(--warning);
    }
  }
  .more { font-size: 10px; }
  .search { width: 130px; }
  .head-actions {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .legend-inline {
    display: flex;
    align-items: center;
    gap: 10px;
    .lg-item {
      display: flex;
      align-items: center;
      gap: 4px;
      font-size: 10px;
      color: var(--text-secondary);
    }
    .lg-line {
      display: inline-block;
      width: 12px;
      height: 2px;
      border-radius: 2px;
    }
    .lg-dash {
      display: inline-block;
      width: 12px;
      height: 0;
      border-top: 1.5px dashed var(--text-muted);
    }
    .granularity {
      margin-left: 6px;
    }
  }
}
.metric-select {
  width: 88px;
}
.card-body { padding-top: 0; }
.chart-body {
  height: 170px;
  display: flex;
  gap: 8px;
  align-items: center;
}

// 环形图卡片专用布局: 左 圆环(居中、不溢出) + 右 图例(单列、紧凑)
// 用 grid 替代 flex, 让两列宽度严格比例且不被内部内容撑爆
.donut-body {
  display: grid;
  grid-template-columns: minmax(0, 1.1fr) minmax(0, 1fr);  // 圆环略宽于图例
  height: 170px;
  padding-top: 0;
  align-items: center;
  gap: 4px;
  .donut-chart {
    width: 100%;
    height: 100%;
    min-width: 0;
    min-height: 0;
    overflow: hidden;
    display: flex;        // 强制 ECharts 容器居中
    align-items: center;
    justify-content: center;
  }
  .legend-col {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
    li {
      display: flex;
      align-items: center;
      gap: 6px;
      font-size: 11px;
      color: var(--text-secondary);
      .dot {
        width: 7px;
        height: 7px;
        border-radius: 50%;
        flex-shrink: 0;
      }
      .lab { flex: 1; }
      .pct {
        color: var(--text-primary);
        font-weight: 600;
      }
    }
  }
}

// 第二行:趋势(2 份) + AI 洞察(1 份)
.grid-row-2 {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 10px;
  .chart-body { height: 184px; }
  .insights-card { padding: 0; border: none; background: transparent; }
}

// 第三行:3 个等宽图
.grid-row-3 {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  gap: 10px;
  align-items: stretch;  // 三张卡同高
  .chart-body { height: 150px; flex-direction: column; align-items: stretch; }
  // 平台分布这张卡内部用 donut-body, 高度 170
  > .card:has(.donut-body) { display: flex; flex-direction: column; }
}

// 第四行:热门达人 TOP 10 全宽(粉丝画像已移除)
.grid-row-full {
  display: grid;
  grid-template-columns: 1fr;
  gap: 10px;
  > .card { min-width: 0; }             // 防止内部内容撑爆 grid cell
  .creators-card .card-body { padding: 0; }
  .creators-card {
    // 全宽后, 表格给足空间展示所有 TOP 10 列, 不挤压不重叠
    ::v-deep(.el-table) { width: 100%; }
    ::v-deep(.el-table__body-wrapper) { overflow-x: auto; }
  }
}

@media (max-width: 1280px) {
  .grid-row-2, .grid-row-3 {
    grid-template-columns: 1fr;
  }
}
</style>
