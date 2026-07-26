// 品牌分析 Store: 先 fallback 填充, 后端可达则用真实数据
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import brandApi from '../api/brand'
import fallback from '../api/fallback-data'
import { useFilterStore } from './filter'

export const useBrandStore = defineStore('brand', () => {
  const loading = ref(false)
  const error = ref(null)
  const usedFallback = ref(false)

  const kpi = ref([])
  const trend = ref([])
  const platforms = ref([])
  const sentiment = ref([])
  const keywords = ref([])
  const list = ref([])

  function fillWithFallback() {
    kpi.value = fallback.brandKpi
    trend.value = fallback.brandTrend
    platforms.value = fallback.brandPlatforms
    sentiment.value = fallback.brandSentiment
    keywords.value = fallback.brandKeywords
    list.value = fallback.brandList
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
      const [k, tr, pl, se, kw, ls] = await withTimeout(Promise.all([
        brandApi.kpi(q),
        brandApi.trend(q),
        brandApi.platforms(q),
        brandApi.sentiment(q),
        brandApi.keywords(q),
        brandApi.list(q)
      ]), 800)

      if (
        !Array.isArray(k) || !Array.isArray(tr) || !Array.isArray(pl) ||
        !Array.isArray(se) || !Array.isArray(kw) || !Array.isArray(ls)
      ) {
        throw new Error('后端返回数据结构异常, 已回退到本地兜底数据')
      }

      kpi.value = k
      trend.value = tr
      platforms.value = pl
      sentiment.value = se
      keywords.value = kw
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
    kpi, trend, platforms, sentiment, keywords, list,
    loadAll
  }
})
