<template>
  <div class="radar-stage" ref="el" />
</template>

<script setup>
// 引爆力维度 —— 雷达图
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import * as echarts from 'echarts'

const props = defineProps({
  data: { type: Array, default: () => [] }
})

const el = ref(null)
let instance = null
let resizeHandler = null

function buildOption() {
  return {
    backgroundColor: 'transparent',
    tooltip: {
      backgroundColor: 'rgba(19, 26, 48, 0.95)',
      borderColor: 'rgba(61, 217, 235, 0.4)',
      textStyle: { color: '#E6EBF5' }
    },
    radar: {
      center: ['50%', '55%'],
      radius: '62%',
      splitNumber: 4,
      // 显式声明 indicator, 避开 ECharts 5.x 自动生成 indicator 时 radarLayout.js:70
      // 偶发 "Cannot read properties of undefined (reading 'push')" 的崩溃路径
      indicator: props.data.map(d => ({ name: d.dimension, max: 100 })),
      // 维度名+数值 在外圈显示, 与 v3.png 一致
      axisName: {
        color: '#E6EBF5',
        fontSize: 11,
        formatter: (name, idx) => {
          const v = props.data[idx]?.value ?? 0
          return `{name|${name}}\n{val|${v}}`
        },
        rich: {
          name: { color: '#E6EBF5', fontSize: 11, lineHeight: 16 },
          val: { color: '#E6EBF5', fontSize: 12, fontWeight: 700, lineHeight: 16 }
        }
      },
      splitLine: { lineStyle: { color: 'rgba(255,255,255,0.08)' } },
      splitArea: { areaStyle: { color: ['rgba(255,255,255,0.02)', 'rgba(255,255,255,0.05)'] } },
      axisLine: { lineStyle: { color: 'rgba(255,255,255,0.08)' } }
    },
    series: [
      {
        type: 'radar',
        emphasis: { focus: 'self' },
        // v3.png 风格: 整体蓝色填充, 与背景对比强烈
        data: [
          {
            name: '当前数据',
            value: props.data.map(d => d.value),
            symbol: 'circle',
            symbolSize: 5,
            lineStyle: { color: '#3DD9EB', width: 2 },
            areaStyle: {
              color: {
                type: 'radial',
                x: 0.5, y: 0.5, r: 0.5,
                colorStops: [
                  { offset: 0, color: 'rgba(94, 161, 255, 0.65)' },
                  { offset: 1, color: 'rgba(61, 217, 235, 0.45)' }
                ]
              }
            },
            itemStyle: { color: '#3DD9EB' }
          },
          {
            name: '平均值',
            value: props.data.map(d => d.avg),
            symbol: 'none',
            lineStyle: { color: '#FFB547', width: 1.5, type: 'dashed' },
            areaStyle: { color: 'rgba(255, 181, 71, 0.04)' },
            itemStyle: { color: '#FFB547' }
          }
        ]
      }
    ]
  }
}

function setOption() {
  if (!instance || instance.isDisposed()) return
  try {
    instance.setOption(buildOption(), true)
  } catch (e) {
    // 兜底: ECharts 内部偶发 layout 异常, 不让它变成 unhandled promise rejection
    // eslint-disable-next-line no-console
    console.warn('[RadarChart] setOption failed:', e)
  }
}

onMounted(() => {
  instance = echarts.init(el.value)
  setOption()
  resizeHandler = () => {
    if (instance && !instance.isDisposed()) instance.resize()
  }
  window.addEventListener('resize', resizeHandler)
})
watch(() => props.data, () => setOption(), { deep: true })
onBeforeUnmount(() => {
  window.removeEventListener('resize', resizeHandler)
  if (instance) {
    instance.dispose()
    instance = null
  }
})
</script>

<style lang="scss" scoped>
.radar-stage { width: 100%; height: 100%; min-width: 0; min-height: 0; overflow: hidden; }
</style>
