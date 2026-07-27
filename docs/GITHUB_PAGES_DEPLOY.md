# GitHub Pages 部署保姆级指引（零成本）

> 目标：把前端静态站免费托管到 `https://a974823304-syg.github.io/insta360-insight/`，
> 面试官点开即看，后端不可达时自动降级为 `demo-real` 数据，不白屏。
> 真后端（Go + Python）不上公网，作为代码展示。

---

## 第 0 步：确认本地状态（已就绪）

- ✅ `frontend/vite.config.js` 已设 `base: './'`（相对路径，项目页 + 本地双击都能跑）
- ✅ 工作流文件已就位：`.github/workflows/deploy.yml`
- ✅ 源码与文档已本地 commit（未 push）
- 待你做的事：**联网 push 到 GitHub**，GitHub 自动构建并发布

---

## 第 1 步：在 GitHub 开启 Pages（GitHub Actions 模式）

1. 打开 `https://github.com/a974823304-syg/insta360-insight`
2. 进入 **Settings → Pages**（左侧边栏）
3. **Source** 选 **"GitHub Actions"**（不是 "Deploy from a branch"）
4. 保存即可（不用其它配置，工作流会接管）

> 这一步只需做一次。

---

## 第 2 步：确认 / 添加远程仓库

```bash
cd F:\workbuddy\影石
git remote -v
```

- 如果已显示 `origin  https://github.com/a974823304-syg/insta360-insight.git` → 跳过。
- 如果为空，先加远程：

```bash
git remote add origin https://github.com/a974823304-syg/insta360-insight.git
```

---

## 第 3 步：联网推送（本机直连被墙，三选一）

### 方式 A：开梯子（代理）后推送（推荐命令行党）
```bash
# 把下面 1080 换成你梯子本机端口
git config --global http.proxy http://127.0.0.1:1080
git config --global https.proxy http://127.0.0.1:1080

git push -u origin main

# 推完可恢复（可选）
git config --global --unset http.proxy
git config --global --unset https.proxy
```

### 方式 B：连手机热点
直接 `git push -u origin main`（热点一般不被墙）。

### 方式 C：GitHub Desktop（最省心，自动走系统代理 + 浏览器 OAuth）
1. 安装并登录 GitHub Desktop
2. File → Add local repository → 选 `F:\workbuddy\影石`
3. 左上选当前分支 `main`，点 **Publish repository**（已存在则点 **Push origin**）
4. 全程图形界面，不用管代理

> GitHub 自 2021 起不再接受账号密码登录 git，用 **PAT（classic，勾选 repo）** 或 **GitHub Desktop OAuth**。

---

## 第 4 步：等 Actions 跑完，拿到地址

1. 仓库页点 **Actions** 标签，看 `Deploy to GitHub Pages` 工作流是否绿色 ✅
2. 绿色后访问：`https://a974823304-syg.github.io/insta360-insight/`
3. 首次部署可能要等 1–2 分钟缓存生效，刷新即可

---

## 第 5 步：以后怎么更新

**只需再 push 一次**，GitHub 自动重新构建发布：

```bash
git add -A
git commit -m "更新：xxx"
git push
```

前端 `dist/` 由 Actions 构建，**不要手动提交 dist**（已被 `.gitignore` 排除）。

---

## 可选：自定义域名（暂缓，需花钱）

- 买域名（约 ¥50–100/年）→ DNS 解析到 GitHub Pages → Settings → Pages → Custom domain → Enforce HTTPS。
- 你当前预算敏感，建议先用默认的 `*.github.io` 免费地址，域名等拿到 offer 再考虑。

---

## 排错

| 现象 | 原因 / 解决 |
| --- | --- |
| `git push` 卡住 / `Could not resolve host` | 直连被墙，用第 3 步的方式 A/B/C |
| Actions 红 ✗：`npm ci` 失败 | 本地有 `package-lock.json`，Actions 用同版本；本地 `npm ci` 试一次 |
| 页面白屏 / 资源 404 | 确认 `vite.config.js` 的 `base: './'`（已就位）；不是 `/` |
| Pages 显示 404 | Settings → Pages → Source 必须选 **GitHub Actions**（不是 branch） |
| 打开是旧版 | Actions 还在跑，等绿；或 Ctrl+F5 强刷 |
