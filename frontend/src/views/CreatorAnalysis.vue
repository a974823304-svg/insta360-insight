<template>
  <div class="page" v-loading="store.loading">
    <header class="page-head">
      <div class="title-block">
        <h1>达人分析 <span class="info" title="数据每 5 分钟刷新一次">ⓘ</span></h1>
        <p class="sub">{{ dateRangeText }} · 全球达人营销数据洞察</p>
      </div>
    </header>

    <!-- 5 张 KPI -->
    <section class="kpi-row">
      <KpiCard v-for="k in store.kpi" :key="k.key" :kpi="k" />
    </section>

    <!-- 粉丝规模趋势(2fr) + 粉丝年龄画像(1fr) -->
    <section class="grid-row-2">
      <div class="card chart-card">
        <div class="card-head"><span class="card-title">粉丝规模趋势</span></div>
        <div class="card-body chart-body">
          <TrendChart :data="store.trend" :granularity="filter.granularity" />
        </div>
      </div>
      <div class="card">
        <div class="card-head"><span class="card-title">粉丝年龄画像</span></div>
        <div class="card-body donut-body">
          <div class="donut-chart"><AgeDonut :data="store.audience.age" centerLabel="粉丝年龄" /></div>
          <ul class="legend-col">
            <li v-for="a in store.audience.age" :key="a.bucket">
              <span class="dot" :style="{ background: a.color }" />
              <span class="lab">{{ a.bucket }}</span>
              <span class="pct num">{{ a.share }}%</span>
            </li>
          </ul>
        </div>
      </div>
    </section>

    <!-- 平台分布 / 赛道粉丝分布 / 性别分布 -->
    <section class="grid-row-3">
      <div class="card">
        <div class="card-head"><span class="card-title">平台分布</span></div>
        <div class="card-body donut-body">
          <div class="donut-chart"><PlatformDonut :data="store.platforms" centerLabel="总粉丝" /></div>
          <ul class="legend-col">
            <li v-for="p in store.platforms" :key="p.platform">
              <span class="dot" :style="{ background: p.color }" />
              <span class="lab">{{ p.platform }}</span>
              <span class="pct num">{{ p.share }}%</span>
            </li>
          </ul>
        </div>
      </div>
      <div class="card">
        <div class="card-head"><span class="card-title">赛道粉丝分布</span></div>
        <div class="card-body chart-body"><TrackBarChart :data="store.tracks" /></div>
      </div>
      <div class="card">
        <div class="card-head"><span class="card-title">粉丝性别分布</span></div>
        <div class="card-body donut-body">
          <div class="donut-chart"><AgeDonut :data="store.audience.gender" centerLabel="性别占比" centerValue="100%" /></div>
          <ul class="legend-col">
            <li v-for="g in store.audience.gender" :key="g.gender">
              <span class="dot" :style="{ background: g.color }" />
              <span class="lab">{{ g.gender }}</span>
              <span class="pct num">{{ g.share }}%</span>
            </li>
          </ul>
        </div>
      </div>
    </section>

    <!-- 达人总表 全宽 -->
    <section class="grid-row-full">
      <div class="card creators-card">
        <div class="card-head">
          <span class="card-title">达人总表</span>
          <el-input v-model="search" placeholder="搜索达人" size="small" class="search" clearable />
        </div>
        <div class="card-body">
          <TopCreatorsTable :rows="filteredList" />
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useFilterStore } from '../stores/filter'
import { useCreatorStore } from '../stores/creator'

import KpiCard from '../components/KpiCard.vue'
import TrendChart from '../components/TrendChart.vue'
import AgeDonut from '../components/AgeDonut.vue'
import PlatformDonut from '../components/PlatformDonut.vue'
import TrackBarChart from '../components/TrackBarChart.vue'
import TopCreatorsTable from '../components/TopCreatorsTable.vue'

const filter = useFilterStore()
const store = useCreatorStore()
const search = ref('')

const dateRangeText = computed(() => filter.dateRange?.length ? filter.dateRange.join(' ~ ') : '')
const filteredList = computed(() => {
  if (!search.value) return store.list
  const kw = search.value.toLowerCase()
  return store.list.filter(c => c.name.toLowerCase().includes(kw))
})

onMounted(() => store.loadAll())
</script>

<style lang="scss" scoped>
.page { display: flex; flex-direction: column; gap: 8px; }
.page-head {
  display: flex; align-items: center; justify-content: space-between;
  .title-block h1 { margin: 0; font-size: 15px; font-weight: 700; color: var(--text-primary); display: flex; gap: 6px; align-items: center;
    .info { cursor: help; color: var(--text-muted); font-size: 12px; } }
  .sub { margin: 1px 0 0; color: var(--text-muted); font-size: 10px; }
}
.kpi-row {
  display: grid; grid-template-columns: repeat(5, 1fr); gap: 6px;
  @media (max-width: 1280px) { grid-template-columns: repeat(3, 1fr); }
  @media (max-width: 768px) { grid-template-columns: repeat(2, 1fr); }
}
.card {
  background: var(--bg-elev); border: 1px solid var(--border);
  border-radius: var(--radius); padding: 8px 10px;
}
.card-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 2px; min-height: 20px;
  .card-title { font-size: 12px; font-weight: 600; color: var(--text-primary); } }
.card-body { padding-top: 0; }
.chart-body { height: 184px; display: flex; gap: 8px; align-items: center; min-width: 0; min-height: 0; overflow: hidden; }
.donut-body {
  display: grid; grid-template-columns: minmax(0, 1.1fr) minmax(0, 1fr);
  height: 170px; padding-top: 0; align-items: center; gap: 4px;
  .donut-chart { width: 100%; height: 100%; min-width: 0; min-height: 0; overflow: hidden;
    display: flex; align-items: center; justify-content: center; }
  .legend-col { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: 6px;
    li { display: flex; align-items: center; gap: 6px; font-size: 11px; color: var(--text-secondary); }
    .dot { width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0; }
    .lab { flex: 1; } .pct { color: var(--text-primary); font-weight: 600; } }
}
.grid-row-2 { display: grid; grid-template-columns: minmax(0, 2fr) minmax(0, 1fr); gap: 10px; .chart-body { height: 184px; } }
.grid-row-3 { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) minmax(0, 1fr); gap: 10px; align-items: stretch;
  .chart-body { height: 150px; flex-direction: column; align-items: stretch; }
  > .card:has(.donut-body) { display: flex; flex-direction: column; } }
.grid-row-full { display: grid; grid-template-columns: 1fr; gap: 10px; > .card { min-width: 0; }
  .creators-card .card-body { padding: 0; }
  .creators-card { ::v-deep(.el-table) { width: 100%; } ::v-deep(.el-table__body-wrapper) { overflow-x: auto; } } }
.search { width: 130px; }
@media (max-width: 1280px) { .grid-row-2, .grid-row-3 { grid-template-columns: 1fr; } }
</style>
