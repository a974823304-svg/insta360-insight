<template>
  <!--
    左侧筛选栏 (v3.png 紧凑版)
    ----------------------
    宽度从 260 → 220, 内边距 18 → 14, 块间距 18 → 12
    标签字号 12 → 11, 控件高度统一减小, 适配一屏装下
  -->
  <aside class="side">
    <div class="head">
      <span class="title">筛选条件</span>
      <el-link type="primary" :underline="false" class="reset" @click="onReset">重置</el-link>
    </div>

    <!-- 时间范围: 自定义紧凑显示, 两个日期完整可见 -->
    <section class="block">
      <div class="label">时间范围</div>
      <el-popover
        placement="bottom-start"
        :width="320"
        trigger="click"
        popper-class="date-popper"
      >
        <template #reference>
          <div class="date-display">
            <el-icon class="cal-icon"><Calendar /></el-icon>
            <span class="d1">{{ formatShort(filter.dateRange?.[0]) }}</span>
            <span class="arr">→</span>
            <span class="d2">{{ formatShort(filter.dateRange?.[1]) }}</span>
          </div>
        </template>
        <el-date-picker
          v-model="filter.dateRange"
          type="daterange"
          range-separator="→"
          start-placeholder="开始"
          end-placeholder="结束"
          format="YYYY-MM-DD"
          value-format="YYYY-MM-DD"
          size="default"
          style="width: 100%;"
        />
      </el-popover>
    </section>

    <!-- 市场(地区别名, 与 v3.png 保持一致) -->
    <section class="block">
      <div class="label">市场</div>
      <div class="chips">
        <button
          v-for="r in store.options.regions"
          :key="r.value"
          class="chip"
          :class="{ active: filter.regions.includes(r.value) || (!filter.regions.length && r.value === '全球') }"
          @click="toggle('regions', r.value)"
        >{{ r.label }}</button>
      </div>
    </section>

    <!-- 运动赛道 -->
    <section class="block">
      <div class="label">运动赛道</div>
      <div class="grid">
        <button
          v-for="t in store.options.tracks"
          :key="t.value"
          class="track"
          :class="{ active: filter.tracks.includes(t.value) || (!filter.tracks.length && t.value === '骑行') }"
          @click="toggle('tracks', t.value)"
        >
          <span class="emoji">{{ emojiOf(t.value) }}</span>
          <span class="text">{{ t.label }}</span>
        </button>
      </div>
    </section>

    <!-- 平台: 3 个国内平台(抖音/B站/小红书), 横排 3 列 -->
    <section class="block">
      <div class="label">平台</div>
      <div class="platforms">
        <button
          v-for="p in store.options.platforms"
          :key="p.value"
          class="platform"
          :class="{ active: filter.platforms.includes(p.value) }"
          :title="p.label"
          @click="toggle('platforms', p.value)"
        >
          <span class="logo" :style="{ background: platformBg(p.value) }">
            {{ platformInitial(p.value) }}
          </span>
          <span class="name">{{ p.label }}</span>
        </button>
      </div>
    </section>

    <!-- 折叠面板:达人属性 / 粉丝画像(预留扩展) -->
    <el-collapse v-model="open" class="collapses">
      <el-collapse-item title="达人属性" name="prop">
        <div class="muted">粉丝量级 / 合作次数 / 配合度 — 后续可下钻</div>
      </el-collapse-item>
      <el-collapse-item title="粉丝画像" name="audience">
        <div class="muted">年龄 / 性别 / 地域 / 兴趣标签</div>
      </el-collapse-item>
    </el-collapse>

    <!-- 操作: @提及的 TOP 达人 + 标签 chips(双列网格, 绝不重叠) -->
    <section class="block ops-block">
      <div class="ops-head">
        <span class="label-inline">操作</span>
        <span class="ops-count">{{ opsRows.length }} 个达人 · {{ opsTagTotal }} 个标签</span>
      </div>
      <ul class="ops-list">
        <li v-for="row in opsRows" :key="row.id" class="op-row" :title="row.name + ' · ' + row.tags.join(' ')">
          <span class="op-name">{{ row.name }}</span>
          <div class="op-tags">
            <span v-for="t in row.tags" :key="t" class="op-tag">#{{ t }}</span>
          </div>
        </li>
      </ul>
    </section>

    <div class="footer">
      <el-button type="primary" size="small" class="apply" @click="onApply">应用筛选</el-button>
    </div>
  </aside>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { Calendar } from '@element-plus/icons-vue'
import { useFilterStore } from '../stores/filter'
import { useInsightStore } from '../stores/insight'

const filter = useFilterStore()
const store = useInsightStore()
const open = ref([])

// "操作" 区域: @提及的 TOP 5 达人 + 各自 tags
// 用 .slice(0, 5) 限制行数, 避免侧边栏过长
// data tags 形如 ['#骑行','#极限','#摄影'] -> 用 .slice(1) 去 #, template 渲染时再加
const opsRows = computed(() => {
  const list = store.topCreators || []
  return list.slice(0, 5).map((c, idx) => {
    const tags = (c.tags || []).map(t => String(t).replace(/^#/, ''))
    return {
      id: c.id ?? idx,
      name: c.name,
      tags
    }
  })
})
const opsTagTotal = computed(() =>
  opsRows.value.reduce((sum, r) => sum + r.tags.length, 0)
)

function formatShort(s) {
  if (!s) return '—'
  // 2024-04-20 -> 04-20
  return s.slice(5)
}

onMounted(() => {
  // 第一次进入页面,先加载可选项
  if (!store.options.regions.length) {
    store.loadAll()
  }
})

function toggle(key, value) {
  const arr = filter[key]
  const idx = arr.indexOf(value)
  if (idx >= 0) arr.splice(idx, 1)
  else arr.push(value)
}

function onReset() {
  filter.reset()
}

async function onApply() {
  await store.loadAll()
}

function emojiOf(track) {
  return {
    '滑雪': '⛷️',
    '冲浪': '🏄',
    '骑行': '🚴',
    '潜水': '🤿',
    '攀岩': '🧗'
  }[track] || '🏅'
}

function platformInitial(p) {
  return { '抖音': '抖', 'B站': 'B', '小红书': '小' }[p] || '?'
}
function platformBg(p) {
  return {
    '抖音':   'linear-gradient(135deg,#000000 0%,#25F4EE 100%)',
    'B站':    'linear-gradient(135deg,#00A1D6 0%,#FB7299 100%)',
    '小红书': 'linear-gradient(135deg,#FF2442 0%,#FFB6C1 100%)'
  }[p] || 'linear-gradient(135deg,#5EA1FF,#7B61FF)'
}
</script>

<style lang="scss" scoped>
.side {
  width: 220px;                       // 紧凑: 260 → 220
  flex-shrink: 0;
  background: var(--bg-elev);
  border-right: 1px solid var(--border);
  padding: 14px 12px;                  // 紧凑: 18/16 → 14/12
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 12px;                           // 紧凑: 18 → 12
}
.head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  .title {
    font-size: 13px;
    font-weight: 600;
    color: var(--text-primary);
  }
  .reset {
    font-size: 11px;
  }
}
.block {
  .label {
    font-size: 11px;                    // 紧凑: 12 → 11
    color: var(--text-secondary);
    margin-bottom: 6px;                  // 紧凑: 8 → 6
  }
}
.picker {
  width: 100%;
  --el-text-color-regular: var(--text-primary);
  --el-text-color-placeholder: var(--text-muted);
  --el-border-color: var(--border);
  --el-fill-color-blank: rgba(19, 26, 48, 0.6);
}
// 自定义紧凑日期显示(展开前): 完整显示两个日期 MM-DD
.date-display {
  display: flex;
  align-items: center;
  gap: 4px;
  width: 100%;
  height: 28px;
  padding: 0 8px;
  background: rgba(19, 26, 48, 0.6);
  border: 1px solid var(--border);
  border-radius: 6px;
  cursor: pointer;
  font-size: 11px;
  color: var(--text-primary);
  transition: border-color 0.2s;
  &:hover { border-color: var(--brand); }
  .cal-icon { font-size: 12px; color: var(--text-muted); flex-shrink: 0; }
  .d1, .d2 { flex: 1; text-align: center; font-variant-numeric: tabular-nums; }
  .arr { color: var(--text-muted); flex-shrink: 0; }
}
.chips {
  display: flex;
  flex-wrap: wrap;
  gap: 5px;
}
.chip {
  padding: 4px 10px;
  font-size: 11px;
  border-radius: 999px;
  background: var(--bg-elev-2);
  border: 1px solid var(--border);
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s ease;
  &:hover { color: var(--text-primary); }
  &.active {
    color: #0B1020;
    background: var(--brand);
    border-color: var(--brand);
    font-weight: 600;
  }
}
.grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 6px;
}
.track {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 6px 8px;
  background: var(--bg-elev-2);
  border: 1px solid var(--border);
  border-radius: 8px;
  color: var(--text-secondary);
  cursor: pointer;
  font-size: 11px;
  transition: all 0.2s ease;
  .emoji { font-size: 13px; }
  .text { font-weight: 500; }
  &:hover { color: var(--text-primary); }
  &.active {
    background: rgba(61, 217, 235, 0.12);
    border-color: rgba(61, 217, 235, 0.4);
    color: var(--brand);
  }
}
.platforms {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;        // 3 列横排, 适配 3 个国内平台
  gap: 5px;
}
.platform {
  display: flex;
  flex-direction: column;                     // logo 上, 名字下, 紧凑
  align-items: center;
  gap: 4px;
  padding: 6px 2px;
  background: var(--bg-elev-2);
  border: 1px solid var(--border);
  border-radius: 8px;
  color: var(--text-secondary);
  cursor: pointer;
  font-size: 11px;
  transition: all 0.2s ease;
  .logo {
    width: 22px;
    height: 22px;
    border-radius: 5px;
    color: #fff;
    display: grid;
    place-items: center;
    font-weight: 700;
    font-size: 11px;
    text-shadow: 0 1px 2px rgba(0,0,0,0.4);   // 抖音黑底白字可读
  }
  .name { font-weight: 500; }
  &:hover { color: var(--text-primary); border-color: rgba(255,255,255,0.18); }
  &.active {
    background: var(--bg-elev-3);
    border-color: var(--brand);
    color: var(--text-primary);
  }
}
.collapses {
  --el-collapse-bg-color: transparent;
  --el-collapse-border-color: var(--border);
  --el-collapse-header-bg-color: transparent;
  --el-collapse-content-bg-color: transparent;
  --el-text-color-primary: var(--text-primary);
}
.muted { color: var(--text-muted); font-size: 11px; padding: 4px 0; }

// === 操作区: 双列网格, 达人名 + tag chips 强制不重叠 ===
.ops-block {
  .ops-head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    margin-bottom: 6px;
    .label-inline {
      font-size: 11px;
      color: var(--text-secondary);
    }
    .ops-count {
      font-size: 10px;
      color: var(--text-muted);
    }
  }
  .ops-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 4px;
    max-height: 320px;       // 超出可滚动, 不撑爆侧边栏
    overflow-y: auto;
    padding-right: 2px;
  }
  .op-row {
    display: grid;
    grid-template-columns: 72px minmax(0, 1fr);   // 关键: 左 72 固定, 右弹性 1fr + minmax 0 防溢出
    gap: 6px;
    align-items: center;
    padding: 5px 7px;
    background: var(--bg-elev-2);
    border: 0.5px solid rgba(255, 255, 255, 0.06);
    border-radius: 4px;
    transition: background 0.18s ease, border-color 0.18s ease;
    cursor: pointer;
    &:hover {
      background: rgba(61, 217, 235, 0.08);
      border-color: rgba(61, 217, 235, 0.32);
    }
  }
  .op-name {
    font-size: 10.5px;
    font-weight: 500;
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .op-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 3px;
    min-width: 0;            // grid cell 内的 flex 必须有 min-width 0 才能正确收缩
  }
  .op-tag {
    font-size: 9.5px;
    line-height: 1.4;
    padding: 1px 5px;
    background: rgba(61, 217, 235, 0.16);
    color: var(--brand);
    border-radius: 3px;
    white-space: nowrap;
    font-weight: 500;
  }
}
.footer {
  margin-top: auto;
  padding-top: 10px;
  border-top: 1px solid var(--border);
  .apply {
    width: 100%;
  }
}
</style>