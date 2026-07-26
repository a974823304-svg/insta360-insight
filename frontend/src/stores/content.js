// 内容分析 Store: 先 fallback 填充, 后端可达则用真实数据
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import contentApi from '../api/content'
import fallback from '../api/fallback-data'
import { useFilterStore } from './filter'

export const useContentStore = defineStore('content', () => {
  const loading = ref(false)
  const error = ref(null)
  const usedFallback = ref(false)

  const kpi = ref([])
  const trend = ref([])
  const forms = ref([])
  const topics = ref([])
  const durations = ref([])
  const list = ref([])

  function fillWithFallback() {
    kpi.value = fallback.contentKpi
    trend.value = fallback.contentTrend
    forms.value = fallback.contentForms
    topics.value = fallback.contentTopics
    durations.value = fallback.contentDurations
    list.value = fallback.contentList
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
      const [k, tr, fo, tp, du, ls] = await withTimeout(Promise.all([
        contentApi.kpi(q),
        contentApi.trend(q),
        contentApi.forms(q),
        contentApi.topics(q),
        contentApi.durations(q),
        contentApi.list(q)
      ]), 800)

      if (
        !Array.isArray(k) || !Array.isArray(tr) || !Array.isArray(fo) ||
        !Array.isArray(tp) || !Array.isArray(du) || !Array.isArray(ls)
      ) {
        throw new Error('后端返回数据结构异常, 已回退到本地兜底数据')
      }

      kpi.value = k
      trend.value = tr
      forms.value = fo
      topics.value = tp
      durations.value = du
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
    kpi, trend, forms, topics, durations, list,
    loadAll
  }
})
