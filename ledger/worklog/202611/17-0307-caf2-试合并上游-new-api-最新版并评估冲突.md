# 20261117-0307-caf2-试合并上游-new-api-最新版并评估冲突: 试合并上游 new-api 最新版并评估冲突

- **Tags**: 上游同步, worktree, 冲突评估
- **Status**: doing · 2026-11-17 05:40:16 · @test · 0f55b754-ef6e-4a8b-a660-a84ba503773e
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

- 剩余 5 个前端冲突未解（`web/bun.lock`、`web/src/components/ui/skeleton.tsx`、`web/src/features/models/index.tsx`、`web/src/features/system-settings/billing/section-registry.tsx`、`web/src/features/system-settings/models/ratio-settings-card.tsx`）
- `go build ./...` 仍有两处 API 变更未适配：`sidecar/service/model_catalog.go:159,205` 的 `perfmetrics.QuerySummaryAll(24)` 需改为 `(24, nil)`；`service/hook.go:58,102` 的 `types.NewAPIError` 已迁到 `relaykit/types`
- 若日后重启：豆包/移动云加密视频逻辑需迁入 `plugins/tasks/doubao/plugin.js`（Go 适配器已随上游删除，我方改动存档在 `.tmp/doubao-our-changes.patch`）
- 若日后重启：`web/classic` 23 个定制 UI（微信扫码登录、微信/支付宝支付设置与充值、模型规格页、视频定价、登录注册表单）需迁到新 `web/src/features/*`

## Next

已暂停，等待是否重启的决定；重启时先补完上面两处 API 适配跑通 `go build ./...`，再动前端。

## Process

（接上）冲突解决进展与踩到的上游 API 变更：

- 已解 **69/74**：后端 25 个全解（`go.sum` 用 `go mod tidy` 重生成；`go.mod` 取上游版本并保留我方 `smartwalle/alipay`；`.gitignore`/`Dockerfile` 取并集并改单 `web/` 构建）
- 保留我方特性的具体处理：`middleware/distributor.go` 只在 `shouldSelectChannel` 内校验 token 模型限制（去掉上游重复的外层校验）；`model/task.go` 合并 `VideoFetchMode` 与 `PluginState/PollFailures`；`model/topup.go` 保留 `OnTopUpSuccess` 钩子并用上游 `WalletQuotaFromDecimalStrict` 包装我方 `Money` 计算；`relay/relay_task.go` 保留 `VideoFetchMode` 赋值与 `publicTaskResultURL`
- 按上游架构丢弃的：`model/option.go` 的 `theme` 分支（上游已删 `system_setting.UpdateAndSyncTheme`）、`controller/user.go` 的 `checkUpdatePassword`（上游新流程接管）、`relay/channel/task/doubao/*`（上游删除，改 JS 插件）
- i18n：取上游新 `keys.go` 后补回 21 个我方仍在用但上游已删的 key（`common.invalid_params`、`user.phone_already_bound`、`payment.*` 等），并在 en/zh-CN/zh-TW 三个 yaml 补译文
- 编译阶段暴露的上游 API 变更：`common.UsingPostgreSQL` → `common.UsingMainDatabase(common.DatabaseTypePostgreSQL)`；`model.RecordLog` 第三参由 string 变 `*common.Message`；`types.NewAPIError` 迁至 `relaykit/types`；`perfmetrics.QuerySummaryAll` 增加 `groups []string` 参数

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

完成了一次完整试合并（74 冲突）与其中 69 个的落地解决（后端 25 个全解、编译已推进到只剩 2 处上游 API 适配），最终由用户决定**不升级**：收益不足以抵消前端两套界面归一 + 视频链路改插件架构的成本。分支 `sync/upstream-20261117` 与 worktree `.worktrees/upstream-merge` 保留为暂停态、未提交；未来若重启，从 To Do 里的两处编译适配直接接着走即可。

工作量判断（供未来决策参考）：数字看着多（占我们 173 个改动文件的 43%），但结构是"少数真冲突 + 大量删除/搬迁"——45 个有效冲突块、Go 侧每文件仅 1~4 块；真正的大头是前端双主题归一与任务适配器改插件两笔重写。1 人估：保留 classic 约 13~20 人日，砍掉 classic 并把 5 个定制 UI 迁到新前端约 21~31 人日。

## Execution Log

<!-- Maintained by the script: appended by new / claim / set / release / log / close. Do not edit.
     Backfill with: progress log --id 20261117-0307-caf2-试合并上游-new-api-最新版并评估冲突 --note "...". -->
- 2026-11-17 03:07:21 · new · @test · 0f55b754-ef6e-4a8b-a660-a84ba503773e · created and claimed
- 2026-11-17 03:08:13 · log · @test · 0f55b754-ef6e-4a8b-a660-a84ba503773e · 真实试合并完成：.worktrees/upstream-merge 停在 74 个未合并路径（冲突块 70，go.sum 占 25）；classic 23 文件全是上游删除类，0 冲突块
- 2026-11-17 03:20:16 · log · @test · 0f55b754-ef6e-4a8b-a660-a84ba503773e · 收益评估：上游 985 提交=472 fix/248 feat，视频任务插件化+分档定价、审计日志、多节点同步是实收益；wechat/sms/model spec 关键词 0 命中仍为自研独有；本 fork TASK_TIMEOUT_MINUTES 默认 1440 vs 上游 4320
- 2026-11-17 05:40:16 · log · @test · 0f55b754-ef6e-4a8b-a660-a84ba503773e · 用户决定不升级（收益不足）；已解 69/74 冲突，后端 25 个全解、go build 只剩 sidecar/model_catalog.go:159 与 service/hook.go:58 两处 API 适配，worktree 保留为暂停态未提交
