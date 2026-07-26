# 侧栏可收起（SideFilter Collapse）实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把左侧筛选栏做成整条可收起——收成 36px 薄竖条(可点击展开),默认展开,TopBar 右上提供 vscode 风格 Fold/Expand 按钮,两者双向同步同一状态。

**Architecture:** 在 `filter.js` 新增单一布尔 `collapsed` + `toggleCollapsed()` 作为唯一状态源;`SideFilter.vue` 用 `v-if/else` 在「展开整栏」与「收起薄竖条」两个根 `<aside>` 间切换;`TopBar.vue` 加一个按钮也调 `toggleCollapsed()`。App.vue 与 5 个数据页零改动(flex 布局自动伸缩 + grid 布局铁律已固化)。

**Tech Stack:** Vue 3 (`<script setup>`) + Pinia + Element Plus(图标 `Fold`/`Expand`) + SCSS scoped。无后端改动,无新依赖。

## Global Constraints

- 纯前端,后端零改动;`go test ./...` 不受影响,本计划不跑后端测试。
- 状态唯一来源在 `filter.js`(Pinia store),SideFilter 与 TopBar 都只读/调它,**禁止各自维护 collapsed 本地副本**。
- `collapsed` 默认 `false`(展开),**不持久化**到 localStorage。
- 收起态用 `v-if/else` 两个根 `<aside>`(隐藏筛选项 DOM 真正消失,a11y 正确),**禁止**单容器内 class 切换硬挤内容。
- 图标用 Element Plus 标准 `Fold`(收起)/`Expand`(展开),不引自定义 SVG。
- 改动仅限 3 文件:`frontend/src/stores/filter.js`、`frontend/src/components/SideFilter.vue`、`frontend/src/components/TopBar.vue`;App.vue 与 5 个 view 不动。
- 品牌色/背景/边框一律用 CSS 变量(`--bg-elev` / `--border` / `--brand` / `--text-*`),禁止写死颜色。
- 本前端工程**未配置 vitest**;验证靠 `npm run build` + 手测清单(计划中已写明),不新建测试框架。

---

### Task 1: filter.js 新增 collapsed 状态

**Files:**
- Modify: `frontend/src/stores/filter.js`(在 `granularity` 之后、`appliedRevision` 之前插入 collapsed 定义;并在 return 中导出)

**Interfaces:**
- Consumes: 无(本任务是最底层状态)
- Produces: `collapsed` (ref<boolean>, 默认 false)、`toggleCollapsed()` (() => void)。后续 Task 2/3 通过 `useFilterStore()` 的 `filter.collapsed` / `filter.toggleCollapsed()` 使用。

- [ ] **Step 1: 在 `granularity` 之后插入 collapsed 定义**

打开 `frontend/src/stores/filter.js`,在 `const granularity = ref('day')`(第 17 行)之后插入:

```js
  // 侧栏是否收起（仅控制 UI 显隐, 与筛选条件无关; 默认展开）
  const collapsed = ref(false)
  function toggleCollapsed() {
    collapsed.value = !collapsed.value
  }
```

- [ ] **Step 2: 在 return 中导出 collapsed / toggleCollapsed**

把文件末尾的 return(第 76-79 行)改为同时导出这两个:

```js
  return {
    dateRange, regions, tracks, platforms, ageBands, granularity,
    collapsed, toggleCollapsed,
    appliedRevision, appliedState, toQuery, apply, reset, isDirty
  }
```

- [ ] **Step 3: 校验语法**

Run: `cd /f/workbuddy/影石/frontend && node -e "import('pinia')" 2>/dev/null && echo ok || echo "pinia-missing(check build)"`
Expected: 输出 `ok`(本步仅确认 pinia 可用;真正编译留到 Task 4 的 `npm run build`)。

- [ ] **Step 4: 提交**

```bash
cd /f/workbuddy/影石 && git add frontend/src/stores/filter.js && git commit -m "feat(filter): 新增 collapsed 状态与 toggleCollapsed, 驱动侧栏收起"
```

---

### Task 2: SideFilter.vue 收起薄竖条分支

**Files:**
- Modify: `frontend/src/components/SideFilter.vue`(template 顶部加 `v-if/else` 包裹;script 已 import Expand 无需额外;style 末尾加 `.stub` 系列)

**Interfaces:**
- Consumes: `filter.collapsed`(ref<boolean>)、`filter.toggleCollapsed()`(来自 Task 1);`filter.isDirty`(computed<boolean>, 已存在)
- Produces: 渲染两个 `<aside>` 根节点;`.stub` / `.stub-toggle` / `.stub-dot` 样式

- [ ] **Step 1: 把现有整栏 `<aside class="side">` 加 `v-if="!filter.collapsed"`**

原第 8 行 `<aside class="side">` 改为:

```vue
  <aside v-if="!filter.collapsed" class="side">
```

(原 `<aside class="side">` 的结束 `</aside>` 在第 134 行保持不动,整个现有内容作为展开态。)

- [ ] **Step 2: 在展开态 `</aside>` 之后、`<template>` 结束前插入收起态薄竖条**

在第 134 行 `</aside>` 之后追加:

```vue
  <!-- 收起态: 36px 薄竖条, 可点击展开 -->
  <aside v-else class="stub" aria-label="展开筛选条件">
    <button class="stub-toggle" type="button"
            @click="filter.toggleCollapsed"
            :title="filter.collapsed ? '展开筛选' : '收起筛选'"
            aria-label="展开筛选条件">
      <el-icon><Expand /></el-icon>
    </button>
    <span v-if="filter.isDirty" class="stub-dot" title="有未应用更改"></span>
  </aside>
```

注意:`Expand` 图标已在组件 `<script setup>` 的图标 import 中需要可用——当前文件第 139 行仅 `import { Calendar } from '@element-plus/icons-vue'`。需在 Task 3 一并补 `Expand`/`Fold` 的 import(SideFilter 与 TopBar 都用到)。本任务先写模板/样式,import 在 Task 3 的「共享 import」步骤统一补。

- [ ] **Step 3: 在 `<style lang="scss" scoped>` 末尾追加收起态样式**

在文件 `.footer { ... }` 块(第 436-450 行)之后追加:

```scss
// === 收起态: 36px 薄竖条 ===
.stub {
  width: 36px;
  flex-shrink: 0;
  background: var(--bg-elev);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: space-between;
  padding: 12px 0;
  overflow: hidden;
}
.stub-toggle {
  width: 26px;
  height: 26px;
  display: grid;
  place-items: center;
  background: transparent;
  border: none;
  border-radius: 6px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: background 0.2s ease, color 0.2s ease;
  &:hover { background: rgba(255, 255, 255, 0.06); color: var(--brand); }
  &:focus-visible { outline: 2px solid var(--brand); outline-offset: 1px; }
  .el-icon { font-size: 16px; }
}
.stub-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--brand);
  box-shadow: 0 0 6px rgba(61, 217, 235, 0.6);
}
```

- [ ] **Step 4: 提交**

```bash
cd /f/workbuddy/影石 && git add frontend/src/components/SideFilter.vue && git commit -m "feat(sidefilter): 新增收起态 36px 薄竖条(展开入口 + 脏提示小青点)"
```

---

### Task 3: TopBar.vue 折叠按钮 + 共享图标 import

**Files:**
- Modify: `frontend/src/components/TopBar.vue`(script 补 `Fold, Expand` import;template `.right` 最左加按钮;style 加 `.panel-toggle`)

**Interfaces:**
- Consumes: `filter.collapsed`(ref<boolean>)、`filter.toggleCollapsed()`(来自 Task 1);`useFilterStore` 已在 TopBar 中可用(第 89 行 `const filter = useFilterStore()`)
- Produces: `.panel-toggle` 按钮;`Fold`/`Expand` 图标在同工程两处共用(需确保 SideFilter 与 TopBar 都 import)

- [ ] **Step 1: 补图标 import(两文件都改)**

`frontend/src/components/TopBar.vue` 第 81 行:
```js
import { Calendar, User, SwitchButton, EditPen, DataLine, Film, Position, Tickets } from '@element-plus/icons-vue'
```
改为:
```js
import { Calendar, User, SwitchButton, EditPen, DataLine, Film, Position, Tickets, Fold, Expand } from '@element-plus/icons-vue'
```

`frontend/src/components/SideFilter.vue` 第 139 行:
```js
import { Calendar } from '@element-plus/icons-vue'
```
改为:
```js
import { Calendar, Expand } from '@element-plus/icons-vue'
```
(SideFilter 只用 `Expand`;TopBar 用 `Fold` + `Expand`。)

- [ ] **Step 2: 在 `.right` 集群最左(日期选择器之前)加折叠按钮**

`frontend/src/components/TopBar.vue` 第 31 行 `<div class="right">` 之后、`<el-date-picker>` 之前插入:

```vue
      <!-- 侧栏收起/展开(vscode 风格) -->
      <button class="panel-toggle" type="button"
              @click="filter.toggleCollapsed"
              :title="filter.collapsed ? '展开筛选' : '收起筛选'"
              :aria-label="filter.collapsed ? '展开筛选' : '收起筛选'">
        <el-icon><component :is="filter.collapsed ? Expand : Fold" /></el-icon>
      </button>
```

- [ ] **Step 3: 在 `<style lang="scss" scoped>` 中加 `.panel-toggle` 样式**

在 TopBar 的 `.right { ... }` 块(第 196-200 行)之后追加:

```scss
.panel-toggle {
  width: 32px;
  height: 32px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid var(--border);
  border-radius: 6px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: background 0.2s ease, color 0.2s ease, border-color 0.2s ease;
  .el-icon { font-size: 15px; }
  &:hover { color: var(--brand); border-color: rgba(61, 217, 235, 0.4); }
  &:focus-visible { outline: 2px solid var(--brand); outline-offset: 1px; }
}
```

- [ ] **Step 4: 提交**

```bash
cd /f/workbuddy/影石 && git add frontend/src/components/TopBar.vue frontend/src/components/SideFilter.vue && git commit -m "feat(topbar): 右上 Fold/Expand 按钮控制侧栏收起, 与薄竖条双向同步"
```

---

### Task 4: 构建校验 + 手测清单

**Files:**
- 无新文件;仅验证已提交改动

**Interfaces:**
- Consumes: 已完成 Task 1/2/3 的全部改动

- [ ] **Step 1: 前端构建**

Run: `cd /f/workbuddy/影石/frontend && npm run build 2>&1 | tail -15`
Expected: 编译成功(可能有 ECharts chunk-size warning,非错误);无 TS/模板语法报错。

- [ ] **Step 2: 启动服务手测(需后端:Vite 代理 /api 到 :8080)**
- [ ] Step 2a: 启动后端 `cd /f/workbuddy/影石/backend && AUTH_DISABLE=1 ./be.exe`(后台)
- [ ] Step 2b: 启动前端 `cd /f/workbuddy/影石/frontend && npm run dev`(后台)
- [ ] Step 2c: 打开 `http://localhost:5173/`,逐项核对:
  1. 默认展开(220px 整栏可见)。
  2. 点 TopBar 右上折叠按钮 → 侧栏收成 36px 薄条,主区域明显变宽,图表重排不溢出。
  3. 点薄条上的 `Expand` 图标 → 展开回 220px。
  4. 展开态点选「平台=抖音」→ 收起 → 再展开,chip 选中态(青色)仍在。
  5. 收起态改 TopBar 日期选择器 → 展开后「时间范围」显示同步。
  6. 展开态改筛选但不点应用 → 收起 → 薄条底部青色小圆点出现。
- [ ] Step 2d: 关闭两个 dev 服务(后台 task 关闭 / taskkill),确认端口释放。

Expected: 6 项全部符合。

- [ ] **Step 3: 提交本计划文档(若尚未提交)**

Run: `cd /f/workbuddy/影石 && git add docs/superpowers/plans/2026-07-26-sidefilter-collapse.md && git commit -m "docs(plan): 侧栏可收起实现计划" || echo "已提交则跳过"`

---

## 自审要点(执行前已核对)

- 覆盖范围:spec 第 3 节的 3 个文件改动、4 节边界场景(脏提示/日期同步/chip 保留/默认展开)、5 节 YAGNI 全部映射到 Task 1-4;无遗漏。
- 占位符扫描:无 TBD/TODO;每步均含可直接复制的代码或命令。
- 类型/命名一致性:Task 1 产出的 `collapsed`/`toggleCollapsed` 在 Task 2/3 以 `filter.collapsed` / `filter.toggleCollapsed()` 一致引用;`isDirty` 为已有 computed,Task 2 直接复用;无改名。
- 图标 import:Task 3 Step 1 显式列出 SideFilter 与 TopBar 两处都需补,避免 Task 2 模板用到 `Expand` 而漏 import 编译失败。
