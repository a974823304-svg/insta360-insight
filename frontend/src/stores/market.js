// 市场洞察 Store: 先 fallback 填充, 后端可达则用真实数据
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import marketApi from '../api/market'
import fallback from '../api/fallback-data'
import { useFilterStore } from './filter'

export const useMarketStore = defineStore('market', () => {
  const loading = ref(false)
  const error = ref(null)
  const usedFallback = ref(false)

  const kpi = ref([])
  const trend = ref([])
  const competitors = ref([])
  const regions = ref([])
  const prices = ref([])
  const list = ref([])

  function fillWithFallback() {
    kpi.value = fallback.marketKpi
    trend.value = fallback.marketTrend
    competitors.value = fallback.marketCompetitors
    regions.value = fallback.marketRegions
    prices.value = fallback.marketPrices
    list.value = fallback.marketList
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
      const [k, tr, co, re, pr, ls] = await withTimeout(Promise.all([
        marketApi.kpi(q),
        marketApi.trend(q),
        marketApi.competitors(q),
        marketApi.regions(q),
        marketApi.prices(q),
        marketApi.list(q)
      ]), 800)

      if (
        !Array.isArray(k) || !Array.isArray(tr) || !Array.isArray(co) ||
        !Array.isArray(re) || !Array.isArray(pr) || !Array.isArray(ls)
      ) {
        throw new Error('后端返回数据结构异常, 已回退到本地兜底数据')
      }

      kpi.value = k
      trend.value = tr
      competitors.value = co
      regions.value = re
      prices.value = pr
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
    kpi, trend, competitors, regions, prices, list,
    loadAll
  }
})
