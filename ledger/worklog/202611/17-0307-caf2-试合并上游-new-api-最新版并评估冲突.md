# 20261117-0307-caf2-试合并上游-new-api-最新版并评估冲突: 试合并上游 new-api 最新版并评估冲突

- **Tags**: 上游同步, worktree, 冲突评估
- **Status**: doing · 2026-11-17 03:20:16 · @test · 0f55b754-ef6e-4a8b-a660-a84ba503773e
- **Created**: 2026-11-17 03:07:21 · @test · 0f55b754-ef6e-4a8b-a660-a84ba503773e

<!-- Header lines are maintained by the script, fixed order: Tags → Status → Created → Sealed.
     The `- **Tags**: …` line is inserted at the top by new --tags; omitted when there are no tags
     (category goes to the ledger and history only, never into this document).
     The `- **Sealed**: …` line is appended on close. -->

## Background

用户问「当前项目如果要更新到最新的 new-api 版本，会有哪些冲突点，多不多，工作量大不大」，随后要求「新建一个 worktree 分支，合并试下，计入台账」，因此做一次真实的三方合并来取证据。

- 分叉点：`948780e3`（2026-05-08，`🎨 fix(theme): align UI controls with global radius tokens`）
- 上游最新：`ce8fc49e`（2026-11-16，`fix(channel): read batch tag changes within the transaction`）
- 落后 985 个上游提交；本分支 `feat/video-fetch-mode-release` 领先 106 个提交，改动 173 个文件（+20111/-1610）

## To Do

- 解决 74 个未合并路径（后端 25 / web 根 25 / classic 23 / default 1），目前 worktree 保持冲突态未提交
- 决定 `web/classic` 去留：上游已整目录删除（提交 `31d70fca3`），保留则永久自维护 + Dockerfile 双 stage 常驻
- 豆包视频任务适配器迁到上游新的 JS 插件体系 `plugins/tasks/doubao/plugin.js`（Go 适配器已被 `eb48396d5` 删除）
- `go.mod` / `go.sum` 重新生成（`go mod tidy`），`web/bun.lock` 用 bun 重装生成
- `Dockerfile` 由「default + classic 双前端」改为上游的单 `web/` 构建，保留内部镜像仓库 `registry.sfbp.cn` 底座
- 合并后跑 `go build ./...` + 前端 `bun run build`，回归支付/微信登录/视频/模型规格四条链路

## Next

在 `.worktrees/upstream-merge` 里按后端先行的顺序解冲突：先 `go.sum`/`go.mod`（25 处冲突块全是版本号，直接重生成），再 `model/task.go:1` 的 4 个冲突块与 `model/topup.go`、`controller/topup.go` 各 3 个，然后 `go build ./...` 看编译级不兼容。

## Process

试合并命令（worktree 建在工作区内，避免污染主工作区）：

```
git fetch https://github.com/QuantumNous/new-api.git '+refs/heads/main:refs/remotes/QuantumNous/main'
git worktree add .worktrees/upstream-merge -b sync/upstream-20261117 feat/video-fetch-mode-release
cd .worktrees/upstream-merge && git merge --no-commit --no-ff QuantumNous/main   # 退出码 1，74 冲突
```

实测数据（`.tmp/merge-conflicts.txt`、`.tmp/merge-conflict-detail.csv`）：

- 未合并路径 **74**，冲突块 **70**；其中 `go.sum` 独占 25 块，扣除后真正要动脑的只有 **45** 块
- 按区域：backend 25 文件 / 61 块（含 go.sum 25）、`web/` 新根 25 文件 / 9 块、`web/classic` 23 文件 / **0 块**（全是"上游删除、我们修改"）、`web/default` 1 文件 / 0 块
- 合并一共动了 4004 个文件；26 处为 upstream 删除而保留我们版本的 modify/delete
- `git merge-tree` 预演与真实 merge 结果一致（都是 74），说明预演可信

三类结构性破坏（不是文本冲突，是架构换了）：

1. 前端双主题合并：`31d70fca3`（#6329，1605 文件、444 删除）把 `web/default` 上提到 `web/`（新 `web/src` 1510 文件，与老 `web/default/src` 同名路径 824/875），并删掉整个 `web/classic`（老 429 文件）
2. 任务适配器改 JS 沙箱插件：`eb48396d5`（2026-08-29，#7076）删除 `relay/channel/task/{ali,doubao,gemini,hailuo,jimeng,kling,sora,suno,vertex,vidu}`，改为 `relay/channel/task/jsplugin` + `plugins/tasks/doubao/plugin.js`
3. 认证与会话重构、`controller/video_proxy.go` 重写 599 行、`controller/relay.go` +651/-270、`model/user.go` +748/-266

踩坑：本机全局 gitconfig 有 `url.git@github.com:.insteadOf https://github.com/`，而沙箱里 msys ssh 无法建信号管道（Win32 error 5），导致所有 GitHub 操作被改写成 ssh 后失败；用 `GIT_CONFIG_GLOBAL` 指向空配置文件绕开。`git fetch`/`worktree` 需要写 `.git`，workspace-write 下被拒，需提权到 danger-full-access。

## Conclusion

真实试合并已落地：分支 `sync/upstream-20261117` + worktree `.worktrees/upstream-merge`，停在 74 个未合并路径的冲突态（未提交）。

工作量判断：数字看着多（占我们 173 个改动文件的 43%），但结构是"少数真冲突 + 大量删除/搬迁"——45 个有效冲突块、其中 Go 侧每个文件仅 1~4 块；真正的大头是前端双主题归一与视频任务适配器改插件两笔重写。1 人估：保留 classic 约 13~20 人日，砍掉 classic 并把 5 个定制 UI 迁到新前端约 21~31 人日。

## Execution Log

<!-- Maintained by the script: appended by new / claim / set / release / log / close. Do not edit.
     Backfill with: progress log --id 20261117-0307-caf2-试合并上游-new-api-最新版并评估冲突 --note "...". -->
- 2026-11-17 03:07:21 · new · @test · 0f55b754-ef6e-4a8b-a660-a84ba503773e · created and claimed
- 2026-11-17 03:08:13 · log · @test · 0f55b754-ef6e-4a8b-a660-a84ba503773e · 真实试合并完成：.worktrees/upstream-merge 停在 74 个未合并路径（冲突块 70，go.sum 占 25）；classic 23 文件全是上游删除类，0 冲突块
- 2026-11-17 03:20:16 · log · @test · 0f55b754-ef6e-4a8b-a660-a84ba503773e · 收益评估：上游 985 提交=472 fix/248 feat，视频任务插件化+分档定价、审计日志、多节点同步是实收益；wechat/sms/model spec 关键词 0 命中仍为自研独有；本 fork TASK_TIMEOUT_MINUTES 默认 1440 vs 上游 4320
