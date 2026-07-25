// 数据洞察 Store
// 加载顺序: 先尝试后端, 失败 / 超时则使用前端内嵌的 fallback 数据。
// 这样 Demo 在没有启动 Go 后端时也能完整展示。
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import insightApi from '../api/insight'
import fallback from '../api/fallback-data'
import { useFilterStore } from './filter'

export const useInsightStore = defineStore('insight', () => {
  const loading = ref(false)
  const error = ref(null)
  const usedFallback = ref(false)

  const kpi = ref([])
  const viewsTrend = ref([])
  const platformShare = ref([])
  const tracks = ref([])
  const radar = ref([])
  const age = ref([])
  const insights = ref([])
  const topCreators = ref([])
  const options = ref(fallback.options)

  // 立即用 fallback 填充,让 UI 不会空白
  function fillWithFallback() {
    kpi.value = fallback.kpi
    viewsTrend.value = fallback.viewsTrend
    platformShare.value = fallback.platformShare
    tracks.value = fallback.tracks
    radar.value = fallback.radar
    age.value = fallback.age
    insights.value = fallback.insights
    topCreators.value = fallback.topCreators
    options.value = fallback.options
    usedFallback.value = true
  }
  fillWithFallback()

  // 软超时: 800ms 内后端没返回, 就不再等
  function withTimeout(promise, ms) {
    return Promise.race([
      promise,
      new Promise((_, reject) => setTimeout(() => reject(new Error('timeout')), ms))
    ])
  }

  async function loadAll() {
    loading.value = true
    error.value = null
    const f = useFilterStore()
    const q = f.toQuery()
    try {
      const [k, vt, ps, tk, rd, ag, ins, tc, opt] = await withTimeout(Promise.all([
        insightApi.kpi(q),
        insightApi.viewsTrend(q),
        insightApi.platformShare(q),
        insightApi.tracks(q),
        insightApi.radar(q),
        insightApi.age(q),
        insightApi.insights(q),
        insightApi.topCreators(q),
        insightApi.options()
      ]), 800)

      // 严格校验: 任何一项不是预期结构(例如后端不可达时静态服务器返回了
      // index.html 被当成文本塞进来), 统统退回 fallback, 绝不让脏数据污染 UI。
      // 否则数组/字符串被 v-for 逐字符展开, 会出现成百上千个空卡片。
      if (
        !Array.isArray(k) || !Array.isArray(vt) || !Array.isArray(ps) ||
        !Array.isArray(tk) || !Array.isArray(rd) || !Array.isArray(ag) ||
        !Array.isArray(ins) || !Array.isArray(tc) || !opt || typeof opt !== 'object'
      ) {
        throw new Error('后端返回数据结构异常, 已回退到本地兜底数据')
      }

      // 后端有响应, 用真实数据
      kpi.value = k
      viewsTrend.value = vt
      platformShare.value = ps
      tracks.value = tk
      radar.value = rd
      age.value = ag
      insights.value = ins
      topCreators.value = tc
      options.value = opt
      usedFallback.value = false
    } catch (e) {
      // 后端没启动 / 超时 / 报错 / 数据异常: 保持初始化时填充的 fallback 数据。
      // fillWithFallback() 已在 store 创建时执行, 这里无需再赋值。
      usedFallback.value = true
      error.value = e?.message || '后端不可达, 已切换到本地数据'
    } finally {
      loading.value = false
    }
  }

  const hasData = computed(() => kpi.value.length > 0)

  return {
    loading, error, usedFallback, hasData,
    kpi, viewsTrend, platformShare, tracks, radar, age, insights, topCreators, options,
    loadAll
  }
})
