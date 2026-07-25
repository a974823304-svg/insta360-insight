<template>
  <div class="donut-stage" ref="el" />
</template>

<script setup>
// 平台分布环形图
//   - 中心展示总播放量
//   - 关闭 outside label, 图例由父级组件右侧展示
//   - 完全自适应父级容器尺寸, 不使用 min-height 防止溢出
//   - 使用 ResizeObserver 监听容器变化, 主动 resize 实例
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import * as echarts from 'echarts'

const props = defineProps({
  data: { type: Array, default: () => [] },
  centerLabel: { type: String, default: '总播放量' }
})

const el = ref(null)
let instance = null
let resizeObserver = null
let windowResizeHandler = null

function buildOption() {
  // 总播放量 = 所有平台 views 之和
  const total = props.data.reduce((s, d) => s + (d.views || 0), 0)
  // 格式化为 1.23B / 456M
  const totalText = total >= 1e9
    ? (total / 1e9).toFixed(2) + 'B'
    : (total / 1e6).toFixed(1) + 'M'

  return {
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'item',
      backgroundColor: 'rgba(19, 26, 48, 0.95)',
      borderColor: 'rgba(61, 217, 235, 0.4)',
      textStyle: { color: '#E6EBF5', fontSize: 12 },
      formatter: (p) => `${p.name}<br/>${(p.value/1e6).toFixed(1)}M (${p.percent}%)`
    },
    series: [
      {
        type: 'pie',
        radius: ['54%', '76%'],
        center: ['50%', '50%'],
        avoidLabelOverlap: true,
        itemStyle: {
          borderColor: '#0B1020',
          borderWidth: 2,
          borderRadius: 3
        },
        label: { show: false },
        labelLine: { show: false },
        data: props.data.map(d => ({
          name: d.platform,
          value: d.views,
          itemStyle: { color: d.color }
        }))
      },
      // 中心挡板(空心透明白), 让中心文字落在圆环正中央
      {
        type: 'pie',
        silent: true,
        radius: ['0%', '54%'],
        center: ['50%', '50%'],
        label: { show: false },
        itemStyle: { color: 'transparent' },
        data: [{ value: 1 }]
      }
    ],
    graphic: [
      {
        type: 'text',
        left: 'center',
        top: '44%',
        style: {
          text: props.centerLabel,
          fill: '#5A6182',
          fontSize: 11,
          textAlign: 'center'
        }
      },
      {
        type: 'text',
        left: 'center',
        top: '52%',
        style: {
          text: totalText,
          fill: '#E6EBF5',
          fontSize: 20,
          fontWeight: 700,
          textAlign: 'center'
        }
      }
    ]
  }
}

function setOption() { if (instance) instance.setOption(buildOption(), true) }

onMounted(() => {
  instance = echarts.init(el.value)
  setOption()
  // 监听容器尺寸变化(父级 flex 布局, 容器尺寸可能变化)
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
.donut-stage {
  width: 100%;
  height: 100%;
  min-width: 0;     // 防止 flex 父级挤压
  min-height: 0;    // 关键: 不再强制 220px, 完全跟随父级
  overflow: hidden; // 圆环绝不溢出容器
}
</style>
