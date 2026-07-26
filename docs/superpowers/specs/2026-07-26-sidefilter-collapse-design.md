# 侧栏可收起功能设计（SideFilter Collapse）

- 日期：2026-07-26
- 状态：已确认（用户拍板）
- 范围：纯前端，无后端改动

## 1. 目标

把左侧筛选栏（`SideFilter.vue`，固定 220px 宽）做成**整条可收起**。用户点一下就收成一条 36px 薄竖条，把横向空间让给主看板；再点一下展开回 220px。

确认的交互形态：

- 收起后保留一条 **36px 薄竖条（可点击）**，不是完全消失。
- **默认展开**（首次进入 = 展开态）。
- 顶部右上提供一个 **vscode 风格折叠按钮**，与薄竖条上的箭头双向同步控制同一状态。
- 不持久化（`localStorage` 跳过）。
- 窄屏自动收起（<1024px 断点）不做。

## 2. 现状（改造前）

- `App.vue`：`app-body` 是 `display:flex; flex:1`，内放 `<SideFilter>` + `<main class="app-main">`（flex:1，min-width:0）。SideFilter 宽由自身 `<aside class="side">` 的 `width:220px` 决定。
- `SideFilter.vue`：`<aside class="side">` 渲染全部筛选内容（时间范围 / 市场 / 运动赛道 / 平台 / 达人属性+粉丝画像折叠 / 操作 / 底部应用筛选+脏提示）。无收起能力。
- `TopBar.vue`：右上 `.right` 集群目前是 `<el-date-picker>` + 登录按钮（或头像下拉）。无折叠按钮。
- `filter.js`：持有 `dateRange / regions / tracks / platforms / ageBands` 及 `appliedRevision / apply / isDirty / reset`，**无 collapsed 概念**。

## 3. 设计

### 3.1 状态模型（`stores/filter.js`）

新增单一布尔 + 切换方法，与现有筛选状态平级：

```js
// 侧栏是否收起（仅控制 UI 显隐，与筛选条件无关）
const collapsed = ref(false)
function toggleCollapsed() { collapsed.value = !collapsed.value }
```

导出 `collapsed`、`toggleCollapsed`。其余（`appliedRevision / apply / isDirty / reset`）原样不动。

- **不做持久化**：默认 `false`（展开）。用户已明确跳过 localStorage 记忆。
- 不放进新 `ui.js` store（YAGNI）：这是单一布尔、与筛选强相关，合到 filter.js 足够。若将来再冒出多个 chrome 开关再拆。

### 3.2 `SideFilter.vue` — 展开 / 收起两形态

采用**条件渲染两个根 `<aside>`**（推荐 A1），而非在单容器内切换 class：

```vue
<!-- 展开态：现有全部内容 -->
<aside v-if="!filter.collapsed" class="side"> ...现有内容不变... </aside>

<!-- 收起态：薄竖条 -->
<aside v-else class="stub" aria-label="展开筛选条件">
  <button class="stub-toggle" @click="filter.toggleCollapsed" title="展开筛选条件" aria-label="展开筛选条件">
    <el-icon><Expand /></el-icon>
  </button>
  <span v-if="filter.isDirty" class="stub-dot" title="有未应用更改" />
</aside>
```

收起态 `<aside class="stub">` 规格：

- 宽度 `36px`，与展开态同 `background: var(--bg-elev)` + `border-right: 1px solid var(--border)`。
- 内部 `display:flex; flex-direction:column; align-items:center; justify-content:space-between`，上下 padding 12px。
- 顶部：`<Expand />` 图标（双线 `›`），点击 = `filter.toggleCollapsed()`；hover 时图标/背景轻微变亮。
- 底部：当 `filter.isDirty === true` 时显示一个 4px 青色圆点（`.stub-dot`），提示"有未应用更改"。
- 整个 stub 是 `<button>` 包裹图标，点击展开。
- 过渡：根 `<aside>` 切换由 `v-if` 控制，无宽度补间（简单、无闪烁、a11y 干净）。App.vue 的 flex 让主区域自动伸缩到新宽度。

**为什么是 A1（两个根元素）而非单容器 class 切换（A2）**：收起后筛选项 DOM 真正消失，键盘 tab 不会进入隐藏控件、屏幕阅读器也不会读出隐藏内容，可访问性更正确；A2 里内容 DOM 还在，会被挤成 0 宽溢出或需要额外 `overflow:hidden` 且首渲染闪烁。

### 3.3 `TopBar.vue` — 右上折叠按钮

位置：`.right` 集群**最左**（`<el-date-picker>` 之前）。

```vue
<button class="panel-toggle"
        :title="filter.collapsed ? '展开筛选' : '收起筛选'"
        @click="filter.toggleCollapsed">
  <el-icon><component :is="filter.collapsed ? Expand : Fold" /></el-icon>
</button>
```

- 图标：`collapsed===false` 显示 `<Fold />`（收起），`collapsed===true` 显示 `<Expand />`（展开）。与 stub 上的 `<Expand />` 保持语义一致（收起态下两个位置都提示"展开"）。
- 样式：32×32 方块、`inline-flex`、`background: rgba(255,255,255,0.05)`、圆角 6px、hover 时边框/图标变 `--brand` 弱高亮（仍用自定义按钮风格，不引 ElButton 以免默认 padding 过大）。
- import 补充：`Fold, Expand` 从 `@element-plus/icons-vue`。
- 与 stub 双向同步：两者都只调 `filter.toggleCollapsed()`，状态唯一来源在 store，天然同步。

### 3.4 `App.vue` — 不改

`app-body` 已是 flex，`<main>` `flex:1 min-width:0`，SideFilter 不论 220px 还是 36px，主区都自动占满。**零修改**。

### 3.5 数据页图表布局 — 不改

insight / creator / content / market / brand 五页均使用 grid + `minmax(0,1fr)`（项目布局铁律已固化），侧栏 220→36 让主区域多 184px，五页自动吸取新空间，无需改任何 view。

## 4. 边界场景

| 场景 | 行为 |
|---|---|
| 收起时筛选 chip 状态 | 全保留在 store；展开后 chip 态与收起前一致 |
| 收起时 TopBar 日期选择器 | 仍可用（`v-model="filter.dateRange"`），改完存 store，展开后 SideFilter「时间范围」立即同步 |
| 收起后「应用筛选 / 重置」 | footer 不可见；产生的脏变更在 stub 底部小青点点亮 |
| 首次进入 | `collapsed=false`（展开） |
| 键盘可达性 | 收起时 stub 有 `aria-label`；stub 按钮 focus 有 outline；tab 不进入隐藏筛选项 |
| 屏幕阅读器 | 收起时隐藏筛选项 DOM 不存在，不会被读出 |

## 5. 不做（YAGNI）

- ❌ 收起时主区域浮「点此展开」提示条（足够明显，不画蛇添足）
- ❌ 折叠时锁定筛选改动（TopBar 日期选择器可改本身就是特性）
- ❌ 快捷键折叠（Ctrl+B 等）
- ❌ 窄屏 <1024px 自动折叠
- ❌ localStorage 记忆展开态

## 6. 影响面 / 改动量

纯前端，触及 3 个文件、约 50 行：

- `frontend/src/stores/filter.js`：+`collapsed` ref + `toggleCollapsed()`（+9 行）
- `frontend/src/components/SideFilter.vue`：新增 `.stub` 分支与 `.stub-toggle / .stub-dot` 样式（+30 行）
- `frontend/src/components/TopBar.vue`：+`panel-toggle` 按钮与样式、补 `Fold/Expand` import（+12 行）

后端零改动；无新依赖；无新 store。

## 7. 验证

- 后端无改动，`go test ./...` 不受影响（本任务不跑）。
- 前端手测（启动 Vite）：
  1. 默认展开；点 TopBar 折叠按钮 → 侧栏收成 36px 薄条、主区域变宽、图表重排。
  2. 点薄条 → 展开回 220px。
  3. 展开态选「抖音」→ 收起 → 展开，chip 选中态仍在。
  4. 收起态改 TopBar 日期 → 展开后「时间范围」显示同步。
  5. 展开态改筛选但不点应用 → 收起 → 薄条底部青点出现。
- `npm run build` 通过（编译校验）。

## 8. 后续 / 关联

- 本任务不涉及真实数据接入（adapter/MockAdapter 已就绪，见 2026-07-25 设计）。
- 与本项目的「全局筛选器联动契约」（docs/superpowers/分析页模板说明.md）正交：本功能只动 UI 显隐，不动筛选数据流。
