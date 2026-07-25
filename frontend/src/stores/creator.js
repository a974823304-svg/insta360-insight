// 达人分析 Store: 先 fallback 填充, 后端可达则用真实数据
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import creatorApi from '../api/creator'
import fallback from '../api/fallback-data'
import { useFilterStore } from './filter'

export const useCreatorStore = defineStore('creator', () => {
  const loading = ref(false)
  const error = ref(null)
  const usedFallback = ref(false)

  const kpi = ref([])
  const trend = ref([])
  const platforms = ref([])
  const tracks = ref([])
  const audience = ref({ age: [], gender: [] })
  const list = ref([])

  function fillWithFallback() {
    kpi.value = fallback.creatorKpi
    trend.value = fallback.creatorTrend
    platforms.value = fallback.creatorPlatforms
    tracks.value = fallback.creatorTracks
    audience.value = fallback.creatorAudience
    list.value = fallback.creatorList
    usedFallback.value = true
  }
  fillWithFallback()

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
      const [k, tr, ps, tk, au, ls] = await withTimeout(Promise.all([
        creatorApi.kpi(q),
        creatorApi.trend(q),
        creatorApi.platforms(q),
        creatorApi.tracks(q),
        creatorApi.audience(q),
        creatorApi.list(q)
      ]), 800)

      if (
        !Array.isArray(k) || !Array.isArray(tr) || !Array.isArray(ps) ||
        !Array.isArray(tk) || !au || typeof au !== 'object' || !Array.isArray(ls)
      ) {
        throw new Error('后端返回数据结构异常, 已回退到本地兜底数据')
      }

      kpi.value = k
      trend.value = tr
      platforms.value = ps
      tracks.value = tk
      audience.value = au
      list.value = ls
      usedFallback.value = false
    } catch (e) {
      usedFallback.value = true
      error.value = e?.message || '后端不可达, 已切换到本地数据'
    } finally {
      loading.value = false
    }
  }

  const hasData = computed(() => kpi.value.length > 0)

  return {
    loading, error, usedFallback, hasData,
    kpi, trend, platforms, tracks, audience, list,
    loadAll
  }
})
