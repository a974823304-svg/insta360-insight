// 全局筛选条件 Store
// 所有看板图表组件通过它读取当前筛选;修改后由 Dashboard 统一发起请求
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

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

  // 侧栏是否收起（仅控制 UI 显隐, 与筛选条件无关; 默认展开）
  const collapsed = ref(false)
  function toggleCollapsed() {
    collapsed.value = !collapsed.value
  }

  // 联动信号: 每次 apply 自增, 各数据页 watch 此值触发重拉
  const appliedRevision = ref(0)
  // 已应用快照(用于脏检查)
  const appliedState = ref({
    dateRange: [...dateRange.value],
    regions: [],
    tracks: [],
    platforms: [],
    ageBands: []
  })

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

  // 应用筛选: 存快照 + 自增 revision(触发页面重拉)
  function apply() {
    appliedState.value = {
      dateRange: [...dateRange.value],
      regions: [...regions.value],
      tracks: [...tracks.value],
      platforms: [...platforms.value],
      ageBands: [...ageBands.value]
    }
    appliedRevision.value++
  }

  function reset() {
    regions.value = []
    tracks.value = []
    platforms.value = []
    ageBands.value = []
    apply() // 重置后视为已应用(全量), 页面重拉
  }

  // 脏检查: 当前筛选与已应用快照不同 -> 有未应用更改
  const isDirty = computed(() => {
    const a = appliedState.value
    const eq = (x, y) => x.length === y.length && x.every((v, i) => v === y[i])
    return (
      !eq(dateRange.value, a.dateRange) ||
      !eq(regions.value, a.regions) ||
      !eq(tracks.value, a.tracks) ||
      !eq(platforms.value, a.platforms) ||
      !eq(ageBands.value, a.ageBands)
    )
  })

  return {
    dateRange, regions, tracks, platforms, ageBands, granularity,
    collapsed, toggleCollapsed,
    appliedRevision, appliedState, toQuery, apply, reset, isDirty
  }
})
