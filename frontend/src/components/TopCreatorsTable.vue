<template>
  <div class="creators">
    <el-table
      :data="rows"
      :default-sort="{ prop: 'rank', order: 'ascending' }"
      size="small"
      :header-cell-style="{ background: 'transparent', color: '#8A93B2', fontSize: '11px', fontWeight: 500 }"
      :cell-style="{ background: 'transparent', borderColor: 'rgba(255,255,255,0.04)', padding: '4px 0' }"
      empty-text="加载中..."
    >
      <el-table-column label="排名" prop="rank" width="80" sortable>
        <template #default="{ row }">
          <span class="rank" :class="`r${row.rank}`">
            <span v-if="row.rank <= 3">👑</span>
            <span v-else class="num">{{ row.rank }}</span>
          </span>
        </template>
      </el-table-column>
      <el-table-column label="达人" min-width="200">
        <template #default="{ row }">
          <div class="creator">
            <div class="avatar" :class="{ block: row.blacklist }">
              {{ row.avatar }}
            </div>
            <div class="info">
              <div class="name">
                {{ row.name }}
                <span v-if="row.blacklist" class="tag-bl">黑名单</span>
              </div>
              <div class="meta">
                <span class="plat" :class="`plat-${platSlug(row.platform)}`">{{ row.platform }}</span>
              </div>
            </div>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="粉丝数" prop="followers" width="120" align="right" sortable>
        <template #default="{ row }">
          <span class="num">{{ formatNum(row.followers) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="播放量" prop="total_views" width="120" align="right" sortable>
        <template #default="{ row }">
          <span class="num">{{ formatNum(row.total_views) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="互动率" prop="engagement" width="100" align="right" sortable>
        <template #default="{ row }">
          <span class="num">{{ row.engagement.toFixed(2) }}%</span>
        </template>
      </el-table-column>
      <el-table-column label="近30天长率" prop="growth_30d" width="130" align="right" sortable>
        <template #default="{ row }">
          <span class="delta" :class="row.growth_30d >= 0 ? 'up' : 'down'">
            {{ row.growth_30d >= 0 ? '↑' : '↓' }} {{ Math.abs(row.growth_30d).toFixed(1) }}%
          </span>
        </template>
      </el-table-column>
      <el-table-column label="引爆力评分" prop="explosive" width="180">
        <template #default="{ row }">
          <div class="score">
            <span class="num s">{{ row.explosive.toFixed(1) }}</span>
            <div class="bar">
              <div class="fill" :style="{ width: row.explosive + '%' }" />
            </div>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="主要赛道" min-width="240">
        <template #default="{ row }">
          <el-tag v-for="t in row.tags" :key="t" size="small" effect="plain" class="mr-1">
            {{ t }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="120" fixed="right">
        <template #default>
          <el-button link type="primary" size="small">查看详情</el-button>
        </template>
      </el-table-column>
    </el-table>
    <div class="more">
      <el-link type="primary" :underline="false">查看更多 ⌄</el-link>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useInsightStore } from '../stores/insight'

const props = defineProps({
  rows: { type: Array, default: null }
})
const store = useInsightStore()

// 默认回退 insight store(向后兼容); 传入 rows 时显示自定义列表(如达人分析页)
const rows = computed(() => props.rows ?? store.topCreators)

function formatNum(n) {
  if (n >= 1e6) return (n / 1e6).toFixed(1) + 'M'
  if (n >= 1e3) return (n / 1e3).toFixed(0) + 'K'
  return n
}

// 平台名 -> CSS class slug (CSS class 不支持中文)
// 抖音 -> douyin, B站 -> bilibili, 小红书 -> xiaohongshu
function platSlug(p) {
  return { '抖音': 'douyin', 'B站': 'bilibili', '小红书': 'xiaohongshu' }[p] || 'unknown'
}
</script>

<style lang="scss" scoped>
.creators {
  background: var(--bg-elev);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 8px 12px 4px;
}
.rank {
  display: inline-block;
  width: 22px;
  text-align: center;
  font-weight: 700;
  &.r1, &.r2, &.r3 { font-size: 16px; }
}
.creator {
  display: flex;
  align-items: center;
  gap: 8px;
  .avatar {
    width: 26px;
    height: 26px;
    border-radius: 50%;
    background: var(--bg-elev-2);
    display: grid;
    place-items: center;
    font-size: 15px;
    &.block { filter: grayscale(0.7); opacity: 0.7; }
  }
  .info { display: flex; flex-direction: column; }
  .name {
    font-size: 12px;
    font-weight: 600;
    color: var(--text-primary);
    display: flex;
    align-items: center;
    gap: 5px;
  }
  .tag-bl {
    font-size: 10px;
    padding: 0 5px;
    background: rgba(255, 107, 107, 0.18);
    color: var(--danger);
    border-radius: 4px;
  }
  .plat {
    font-size: 10px;
    padding: 0 5px;
    border-radius: 4px;
    background: var(--bg-elev-2);
    color: var(--text-secondary);
  }
  .plat-youtube { color: #FF3D5A; }
  .plat-tiktok { color: #3DD9EB; }
  .plat-instagram { color: #E91E63; }
  // 3 个国内平台: 抖音(品牌色) / B站(蓝粉) / 小红书(红粉)
  .plat-douyin { color: #FE2C55; }
  .plat-bilibili { color: #00A1D6; }
  .plat-xiaohongshu { color: #FF2442; }
}
.num { font-family: var(--font-num); font-feature-settings: 'tnum'; }
.delta {
  font-weight: 600;
  font-size: 11px;
  &.up { color: var(--success); }
  &.down { color: var(--danger); }
}
.score {
  display: flex;
  align-items: center;
  gap: 6px;
  .s { font-weight: 700; color: var(--text-primary); font-size: 11px; min-width: 24px; }
  .bar {
    flex: 1;
    height: 3px;
    background: var(--bg-elev-2);
    border-radius: 999px;
    overflow: hidden;
  }
  .fill {
    height: 100%;
    background: linear-gradient(90deg, #5EA1FF, #3DD9EB);
    border-radius: 999px;
  }
}
.mr-1 { margin-right: 3px; }
.more {
  text-align: center;
  padding: 4px 0 2px;
  font-size: 11px;
}
</style>
