<template>
  <div class="data-table">
    <div v-if="searchable" class="dt-toolbar">
      <el-input v-model="kw" placeholder="搜索" size="small" clearable class="dt-search" />
    </div>
    <el-table :data="pagedRows" style="width: 100%" size="small" :row-key="rowKey"
      :default-sort="{ prop: firstSortable, order: 'descending' }" height="100%">
      <el-table-column v-for="col in columns" :key="col.key"
        :prop="col.key" :label="col.label"
        :width="col.width" :align="col.align || 'left'"
        :sortable="col.sortable || false"
        :formatter="col.formatter">
        <template #default="scope" v-if="col.slot">
          <component :is="col.slot" :row="scope.row" />
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  columns: { type: Array, required: true }, // [{ key, label, sortable?, width?, align?, formatter? }]
  rows: { type: Array, default: () => [] },
  searchable: { type: Boolean, default: false },
  rowKey: { type: String, default: 'id' }
})

const kw = ref('')
const firstSortable = computed(() => props.columns.find(c => c.sortable)?.key || '')
const pagedRows = computed(() => {
  if (!props.searchable || !kw.value) return props.rows
  const k = kw.value.toLowerCase()
  return props.rows.filter(r =>
    props.columns.some(c => String(r[c.key] ?? '').toLowerCase().includes(k))
  )
})
</script>

<style lang="scss" scoped>
.data-table { width: 100%; height: 100%; display: flex; flex-direction: column; gap: 6px;
  .dt-toolbar { display: flex; justify-content: flex-end; }
  .dt-search { width: 160px; }
  :deep(.el-table) { background: transparent; --el-table-border-color: var(--border);
    --el-table-header-bg-color: transparent; --el-table-bg-color: transparent;
    --el-table-tr-bg-color: transparent; --el-table-row-hover-bg-color: rgba(61,217,235,0.08);
    color: var(--text-secondary); font-size: 12px; }
  :deep(.el-table th.el-table__cell) { background: transparent; color: var(--text-muted); font-weight: 600; }
  :deep(.el-table .cell) { padding: 4px 8px; }
  :deep(.el-table .num) { font-variant-numeric: tabular-nums; }
}
</style>
