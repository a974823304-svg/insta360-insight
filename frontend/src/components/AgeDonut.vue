<template>
  <div class="stage" ref="el" />
</template>

<script setup>
// 粉丝年龄分布环形图(中间展示"总粉丝")
//   - 关闭 outside label, 图例由父级组件右侧展示
//   - 总粉丝展示当前数据规模(默认 287.6M 与 v1.png 一致)
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import * as echarts from 'echarts'

const props = defineProps({
  data: { type: Array, default: () => [] },
  // 总粉丝(M) — 父级可传入真实数, 默认演示值 287.6M
  totalM: { type: Number, default: 287.6 },
  centerLabel: { type: String, default: '总粉丝' },
  centerValue: { type: String, default: '' }
})

const el = ref(null)
let instance = null
let resizeHandler = null

function buildOption() {
  return {
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'item',
      backgroundColor: 'rgba(19, 26, 48, 0.95)',
      borderColor: 'rgba(61, 217, 235, 0.4)',
      textStyle: { color: '#E6EBF5' },
      formatter: '{b}<br/>{c}%'
    },
    series: [
      {
        type: 'pie',
        radius: ['58%', '78%'],
        center: ['50%', '52%'],
        avoidLabelOverlap: true,
        itemStyle: {
          borderColor: '#0B1020',
          borderWidth: 3,
          borderRadius: 4
        },
        // 关闭外标签, 由右侧 legend 展示
        label: { show: false },
        labelLine: { show: false },
        data: props.data.map(d => ({
          name: d.bucket,
          value: d.share,
          itemStyle: { color: d.color }
        }))
      },
      // 中心挡板
      {
        type: 'pie',
        silent: true,
        radius: ['0%', '58%'],
        center: ['50%', '52%'],
        label: { show: false },
        itemStyle: { color: 'transparent' },
        data: [{ value: 1 }]
      }
    ],
    graphic: [
      {
        type: 'text',
        left: 'center',
        top: '46%',
        style: { text: props.centerLabel, fill: '#5A6182', fontSize: 11 }
      },
      {
        type: 'text',
        left: 'center',
        top: '53%',
        style: { text: props.centerValue !== '' ? props.centerValue : props.totalM.toFixed(1) + 'M', fill: '#E6EBF5', fontSize: 22, fontWeight: 700 }
      }
    ]
  }
}

function setOption() { if (instance) instance.setOption(buildOption(), true) }

onMounted(() => {
  instance = echarts.init(el.value)
  setOption()
  resizeHandler = () => instance.resize()
  window.addEventListener('resize', resizeHandler)
})
watch(() => [props.data, props.totalM, props.centerLabel, props.centerValue], () => setOption(), { deep: true })
onBeforeUnmount(() => {
  if (resizeHandler) window.removeEventListener('resize', resizeHandler)
  if (instance) instance.dispose()
})
</script>

<style lang="scss" scoped>
.stage { width: 100%; height: 100%; min-width: 0; min-height: 0; overflow: hidden; }
</style>
