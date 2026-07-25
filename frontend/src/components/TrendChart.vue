<template>
  <div class="chart-wrap">
    <div ref="el" class="chart" />
  </div>
</template>

<script setup>
// 趋势分析(播放量)
//   - 实线 = 当期
//   - 虚线 = 上周期
//   - 异常点用大圆点 + 提示文字
//   - 支持 日/周/月 切换(粒度由父组件传入)
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import * as echarts from 'echarts'

const props = defineProps({
  data: { type: Array, default: () => [] },
  granularity: { type: String, default: 'day' } // day/week/month
})

const el = ref(null)
let instance = null
let resizeHandler = null

function buildOption() {
  const dates = props.data.map(d => d.date)
  const cur = props.data.map(d => d.views)
  const prev = props.data.map(d => d.prev_views)
  const anomalyIdx = []
  props.data.forEach((d, i) => { if (d.has_anomaly) anomalyIdx.push(i) })

  return {
    backgroundColor: 'transparent',
    grid: { left: 48, right: 18, top: 40, bottom: 28 },
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(19, 26, 48, 0.95)',
      borderColor: 'rgba(61, 217, 235, 0.4)',
      textStyle: { color: '#E6EBF5' },
      formatter: (params) => {
        const cur = params.find(p => p.seriesName === '当前周期')?.value ?? 0
        const prev = params.find(p => p.seriesName === '上周期')?.value ?? 0
        const idx = params[0]?.dataIndex
        const ratio = prev ? ((cur - prev) / prev * 100).toFixed(1) : 0
        const tip = props.data[idx]?.anomaly_tag
          ? `<div style="color:#FFB547;font-size:11px;margin-top:4px">⚡ ${props.data[idx].anomaly_tag}</div>` : ''
        return `<div style="font-weight:600">${params[0].axisValue}</div>
                <div>当前: ${(cur/1e6).toFixed(1)}M</div>
                <div style="opacity:0.6">上周期: ${(prev/1e6).toFixed(1)}M</div>
                <div style="color:${cur>=prev?'#7DD96E':'#FF6B6B'};margin-top:4px">${cur>=prev?'+':''}${ratio}%</div>
                ${tip}`
      }
    },
    legend: {
      data: ['当前周期', '上周期'],
      textStyle: { color: '#8A93B2' },
      right: 10,
      top: 0
    },
    xAxis: {
      type: 'category',
      data: dates,
      axisLine: { lineStyle: { color: 'rgba(255,255,255,0.1)' } },
      axisLabel: { color: '#8A93B2', fontSize: 11 },
      axisTick: { show: false }
    },
    yAxis: {
      type: 'value',
      axisLabel: {
        color: '#8A93B2',
        fontSize: 11,
        formatter: (v) => (v / 1e6).toFixed(0) + 'M'
      },
      splitLine: { lineStyle: { color: 'rgba(255,255,255,0.05)' } }
    },
    series: [
      {
        name: '当前周期',
        type: 'line',
        smooth: true,
        symbol: 'circle',
        symbolSize: (val, idx) => anomalyIdx.includes(idx) ? 10 : 4,
        showSymbol: true,
        lineStyle: { color: '#3DD9EB', width: 2.5 },
        itemStyle: {
          color: (params) => anomalyIdx.includes(params.dataIndex) ? '#FFB547' : '#3DD9EB',
          borderColor: '#0B1020',
          borderWidth: 2
        },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(61, 217, 235, 0.4)' },
            { offset: 1, color: 'rgba(61, 217, 235, 0)' }
          ])
        },
        data: cur,
        markPoint: {
          symbol: 'circle',
          symbolSize: 14,
          symbolOffset: [0, 0],
          itemStyle: {
            color: '#FFB547',
            borderColor: '#0B1020',
            borderWidth: 2,
            shadowBlur: 6,
            shadowColor: 'rgba(255, 181, 71, 0.5)'
          },
          label: {
            show: true,
            position: 'top',
            distance: 8,
            color: '#FFB547',
            fontSize: 10,
            fontWeight: 600,
            formatter: '⚡AI',
            backgroundColor: 'rgba(11, 16, 32, 0.7)',
            padding: [2, 4],
            borderRadius: 3
          },
          data: anomalyIdx.map(i => ({
            name: 'AI', value: 'AI', xAxis: i, yAxis: cur[i]
          }))
        }
      },
      {
        name: '上周期',
        type: 'line',
        smooth: true,
        showSymbol: false,
        lineStyle: { color: '#5A6182', width: 1.5, type: 'dashed' },
        data: prev
      }
    ]
  }
}

function setOption() {
  if (!instance) return
  instance.setOption(buildOption(), true)
}

onMounted(() => {
  instance = echarts.init(el.value, null, { renderer: 'canvas' })
  setOption()
  resizeHandler = () => instance.resize()
  window.addEventListener('resize', resizeHandler)
})

watch(() => props.data, () => setOption(), { deep: true })
watch(() => props.granularity, () => setOption())

onBeforeUnmount(() => {
  if (resizeHandler) window.removeEventListener('resize', resizeHandler)
  if (instance) instance.dispose()
})
</script>

<style lang="scss" scoped>
.chart-wrap {
  position: relative;
  width: 100%;
  height: 100%;
  min-width: 0;
  min-height: 0;
  overflow: hidden;
}
.chart {
  width: 100%;
  height: 100%;
}
</style>
