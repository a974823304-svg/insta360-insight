<template>
  <div class="page" v-loading="store.loading">
    <header class="page-head">
      <div class="title-block">
        <h1>内容分析 <span class="info" title="数据每 5 分钟刷新一次">ⓘ</span></h1>
        <p class="sub">{{ dateRangeText }} · 全平台内容营销数据洞察</p>
      </div>
    </header>

    <!-- 5 张 KPI -->
    <section class="kpi-row">
      <KpiCard v-for="k in store.kpi" :key="k.key" :kpi="k" />
    </section>

    <!-- 内容播放量趋势(全宽) -->
    <section class="grid-row-wide">
      <div class="card trend-card">
        <div class="card-head"><span class="card-title">内容播放量趋势</span></div>
        <div class="card-body chart-body">
          <TrendChart :data="store.trend" :granularity="filter.granularity" />
        </div>
      </div>
    </section>

    <!-- 内容形式分布 / 话题热度分布 / 内容时长分布 -->
    <section class="grid-row-3">
      <div class="card donut-card">
        <div class="card-head"><span class="card-title">内容形式分布</span></div>
        <div class="card-body donut-body">
          <div class="donut-chart"><PlatformDonut :data="store.forms" centerLabel="内容量" /></div>
          <ul class="legend-col">
            <li v-for="f in store.forms" :key="f.platform">
              <span class="dot" :style="{ background: f.color }" />
              <span class="lab">{{ f.platform }}</span>
              <span class="pct num">{{ f.share }}%</span>
            </li>
          </ul>
        </div>
      </div>
      <div class="card bar-card">
        <div class="card-head"><span class="card-title">话题热度分布</span></div>
        <div class="card-body chart-body"><TrackBarChart :data="store.topics" /></div>
      </div>
      <div class="card donut-card">
        <div class="card-head"><span class="card-title">内容时长分布</span></div>
        <div class="card-body donut-body">
          <div class="donut-chart"><AgeDonut :data="store.durations" centerLabel="时长占比" center-value="100%" /></div>
          <ul class="legend-col">
            <li v-for="d in store.durations" :key="d.bucket">
              <span class="dot" :style="{ background: d.color }" />
              <span class="lab">{{ d.bucket }}</span>
              <span class="pct num">{{ d.share }}%</span>
            </li>
          </ul>
        </div>
      </div>
    </section>

    <!-- 内容总表 全宽 -->
    <section class="grid-row-full">
      <div class="card list-card">
        <div class="card-head"><span class="card-title">内容总表</span></div>
        <div class="card-body"><DataTable :columns="contentColumns" :rows="store.list" searchable row-key="id" /></div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, watch } from 'vue'
import { useFilterStore } from '../stores/filter'
import { useContentStore } from '../stores/content'

import KpiCard from '../components/KpiCard.vue'
import TrendChart from '../components/TrendChart.vue'
import PlatformDonut from '../components/PlatformDonut.vue'
import AgeDonut from '../components/AgeDonut.vue'
import TrackBarChart from '../components/TrackBarChart.vue'
import DataTable from '../components/DataTable.vue'

const filter = useFilterStore()
const store = useContentStore()

onMounted(() => store.loadAll())
// 筛选应用后自动重拉本页数据
watch(() => filter.appliedRevision, () => store.loadAll())

const dateRangeText = computed(() => filter.dateRange?.length ? filter.dateRange.join(' ~ ') : '')

function fmtInt(v) {
  if (v == null) return '-'
  const n = Number(v)
  if (n >= 1e9) return (n / 1e9).toFixed(2) + 'B'
  if (n >= 1e6) return (n / 1e6).toFixed(1) + 'M'
  if (n >= 1e3) return (n / 1e3).toFixed(1) + 'K'
  return String(n)
}

const contentColumns = [
  { key: 'title', label: '标题', sortable: false },
  { key: 'form', label: '形式', sortable: false },
  { key: 'topic', label: '话题', sortable: false },
  { key: 'views', label: '播放量', sortable: true, formatter: (r, c, v) => fmtInt(v) },
  { key: 'engagement', label: '互动率', sortable: true, formatter: (r, c, v) => (v ?? 0) + '%' },
  { key: 'isHit', label: '是否爆款', sortable: false, formatter: (r, c, v) => (v ? '🔥 爆款' : '—') }
]

store.loadAll()
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
  border-radius: var(--radius); padding: 8px 10px; min-width: 0;
}
.card-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 2px; min-height: 20px;
  .card-title { font-size: 12px; font-weight: 600; color: var(--text-primary); } }
.card-body { padding-top: 0; }
.chart-body { height: 200px; display: flex; gap: 8px; align-items: center; min-width: 0; min-height: 0; overflow: hidden; }
.donut-body {
  display: grid; grid-template-columns: minmax(0, 1.1fr) minmax(0, 1fr);
  height: 180px; padding-top: 0; align-items: center; gap: 4px;
  .donut-chart { width: 100%; height: 100%; min-width: 0; min-height: 0; overflow: hidden;
    display: flex; align-items: center; justify-content: center; }
  .legend-col { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: 6px;
    li { display: flex; align-items: center; gap: 6px; font-size: 11px; color: var(--text-secondary); }
    .dot { width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0; }
    .lab { flex: 1; } .pct { color: var(--text-primary); font-weight: 600; } }
}
.grid-row-wide { display: grid; grid-template-columns: 1fr; gap: 10px;
  > .card { min-width: 0; } }
.grid-row-3 { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) minmax(0, 1fr); gap: 10px; align-items: stretch;
  .chart-body { height: 180px; flex-direction: column; align-items: stretch; } }
.grid-row-full { display: grid; grid-template-columns: 1fr; gap: 10px;
  > .card { min-width: 0; } }
.list-card { .card-body { height: 420px; padding: 0; } }
@media (max-width: 1280px) { .grid-row-wide, .grid-row-3 { grid-template-columns: 1fr; } }
</style>
