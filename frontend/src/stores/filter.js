// 全局筛选条件 Store
// 所有看板图表组件通过它读取当前筛选;修改后由 Dashboard 统一发起请求
import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useFilterStore = defineStore('filter', () => {
  // 时间范围: 数组 [start, end]
  const dateRange = ref(['2024-04-20', '2024-05-20'])

  // 多选筛选
  const regions = ref([])      // 空数组表示"全部"
  const tracks = ref([])
  const platforms = ref([])
  const ageBands = ref([])

  // 时间粒度: day / week / month (用于趋势图切换)
  const granularity = ref('day')

  // 把筛选条件序列化为后端可识别的 querystring
  function toQuery() {
    const q = {}
    if (dateRange.value?.length === 2) {
      q.date_range = dateRange.value
    }
    if (regions.value.length) q.regions = [...regions.value]
    if (tracks.value.length) q.tracks = [...tracks.value]
    if (platforms.value.length) q.platforms = [...platforms.value]
    if (ageBands.value.length) q.age_bands = [...ageBands.value]
    return q
  }

  function reset() {
    regions.value = []
    tracks.value = []
    platforms.value = []
    ageBands.value = []
  }

  return { dateRange, regions, tracks, platforms, ageBands, granularity, toQuery, reset }
})
