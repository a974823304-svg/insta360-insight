<template>
  <div class="bar-stage" ref="el" />
</template>

<script setup>
// 运动赛道表现 —— 横向条形图
//   - 渐变填充
//   - 末端展示数值
//   - 完全自适应父级, 不使用 min-height 防止溢出
//   - 使用 ResizeObserver 监听容器变化, 主动 resize
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import * as echarts from 'echarts'

const props = defineProps({
  data: { type: Array, default: () => [] }
})

const el = ref(null)
let instance = null
let resizeObserver = null
let windowResizeHandler = null

function buildOption() {
  // 倒序排列, 让最长的在顶部
  const list = [...props.data].sort((a, b) => a.views - b.views)
  return {
    backgroundColor: 'transparent',
    // 紧凑 grid, 给 5 条 bar 留足够空间
    grid: { left: 52, right: 60, top: 4, bottom: 4, containLabel: false },
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      backgroundColor: 'rgba(19, 26, 48, 0.95)',
      borderColor: 'rgba(61, 217, 235, 0.4)',
      textStyle: { color: '#E6EBF5', fontSize: 12 },
      formatter: (p) => `${p[0].name}<br/>${(p[0].value/1e6).toFixed(0)}M`
    },
    xAxis: {
      type: 'value',
      show: false
    },
    yAxis: {
      type: 'category',
      data: list.map(d => d.track),
      axisLine: { show: false },
      axisTick: { show: false },
      axisLabel: { color: '#E6EBF5', fontSize: 11 }
    },
    series: [
      {
        type: 'bar',
        barWidth: 10,
        barCategoryGap: '30%',
        showBackground: true,
        backgroundStyle: { color: 'rgba(255,255,255,0.04)', borderRadius: 5 },
        itemStyle: {
          borderRadius: 5,
          color: (params) => {
            const d = list[params.dataIndex]
            return new echarts.graphic.LinearGradient(0, 0, 1, 0, [
              { offset: 0, color: d.color + '55' },
              { offset: 1, color: d.color }
            ])
          }
        },
        label: {
          show: true,
          position: 'right',
          distance: 6,
          color: '#E6EBF5',
          fontSize: 11,
          fontWeight: 600,
          formatter: (p) => (p.value / 1e6).toFixed(0) + 'M'
        },
        data: list.map(d => d.views)
      }
    ]
  }
}

function setOption() { if (instance) instance.setOption(buildOption(), true) }

onMounted(() => {
  instance = echarts.init(el.value)
  setOption()
  if (typeof ResizeObserver !== 'undefined') {
    resizeObserver = new ResizeObserver(() => {
      if (instance) instance.resize()
    })
    resizeObserver.observe(el.value)
  }
  windowResizeHandler = () => instance && instance.resize()
  window.addEventListener('resize', windowResizeHandler)
})
watch(() => props.data, () => setOption(), { deep: true })
onBeforeUnmount(() => {
  if (resizeObserver) resizeObserver.disconnect()
  if (windowResizeHandler) window.removeEventListener('resize', windowResizeHandler)
  if (instance) instance.dispose()
})
</script>

<style lang="scss" scoped>
.bar-stage {
  width: 100%;
  height: 100%;
  min-width: 0;
  min-height: 0;   // 关键: 不强制高度, 完全跟随父级
  overflow: hidden;
}
</style>
