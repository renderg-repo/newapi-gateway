# Sidecar 开发规范 — Agent 系统提示词

## 项目背景

本项目是 **new-api**，一个 AI API 网关/代理（Go 后端 + React 前端）。代码托管在 GitHub 上，存在 upstream 源仓库。我们的目标是在最小化 sync 冲突风险的前提下，通过 Sidecar 模式扩展新功能。

## 核心原则：最小侵入 + Sidecar 优先

### 1. 功能扩展优先使用 Sidecar 模式

任何新功能开发前，先评估是否可 Sidecar 化：

| 类型 | 处理方式 |
|------|---------|
| 可完整 Sidecar 化的模块（如独立服务、配置管理、独立接口） | **必须**放入 `sidecar/` 目录 |
| 必须侵入主代码的环节（如 User 模型加字段、登录入口 UI） | 只改最必要的点，改完即止 |
| 路由注册 | 在 `router/api-router.go` 末尾加一行 `sidecarRouter.RegisterXxxRoutes(apiRouter)` |

### 2. Sidecar 目录结构标准

```
sidecar/
├── controller/   # HTTP 请求处理器（与主包 controller 平行）
├── service/      # 业务逻辑（与主包 service 平行）
├── router/       # 路由注册函数
└── model/        # 可选，仅当需要 Sidecar 私有数据模型时
```

**规则：**
- Sidecar 包内遵循主包相同的分层：`Router -> Controller -> Service`
- Sidecar controller 中复用主包的 `common.ApiErrorI18n`、`common.ApiSuccess`、`i18n.MsgXxx` 等通用响应函数
- Sidecar 内部 session 写入：如果主包 `setupLogin` 未导出（小写），可在 sidecar 中**复制**其逻辑（通常 10~20 行），不要为了一行导出而增加侵入点

### 3. 侵入点白名单

以下文件允许被修改，其他文件**原则上不动**：

| 允许修改的文件 | 改动内容限制 |
|---------------|-------------|
| `model/user.go` | 仅在 User struct 末尾追加字段 + 辅助查询函数 |
| `model/option.go` | 在 `InitOptionMap()` 中追加配置项默认值；在 `updateOptionMap()` 中追加配置同步逻辑 |
| `router/api-router.go` | 在 `SetApiRouter()` 末尾追加一行 sidecar 路由注册 |
| `web/src/components/settings/SystemSetting.jsx` | 追加新配置板块（Card + Form.Section） |
| `web/src/components/auth/LoginForm.jsx` | 追加登录入口 Tab/选项 |
| `web/src/components/auth/RegisterForm.jsx` | 追加注册入口 Tab/选项 |
| `web/src/helpers/data.js` | 追加新状态到 localStorage（如有需要） |

### 4. 禁止行为

- **不要**在 `controller/`、`service/`、`model/`（根目录）、`middleware/`、`setting/` 等主包目录中直接添加新文件来扩展功能
- **不要**大幅重构主包的现有代码结构
- **不要**修改或删除项目品牌信息（QuantumNous、new-api 等）
- **不要**为了 sidecar 功能而导出主包的私有函数（小写函数），复制比导出更干净

## 后端编码规范

### 技术栈
- Go 1.22+, Gin, GORM v2
- 数据库：SQLite / MySQL >= 5.7.8 / PostgreSQL >= 9.6（必须三库兼容）

### 代码风格
- 请求 DTO 使用 struct + `binding:"required"` 标签
- 错误响应使用 `common.ApiErrorI18n(c, i18n.MsgXxx)`，成功响应使用 `common.ApiSuccess(c, data)`
- 日志使用 `common.SysLog()`
- 配置读取使用 `common.OptionMap["KeyName"]`，配置变更通过 `model.UpdateOption(key, value)`
- JSON marshal/unmarshal **必须**使用 `common/json.go` 中的包装函数，禁止直接调用 `encoding/json`

### 数据库兼容性
- 优先使用 GORM 方法（`Create`, `Find`, `Where`, `Updates`），避免 raw SQL
- 如需 raw SQL，注意三库差异（引号、布尔值等），参考 `model/main.go` 中的 `commonGroupCol`、`commonTrueVal` 等变量
- 新增字段尽量利用 GORM AutoMigrate，无需手写 migration

## 前端编码规范

### 技术栈
- React 18, Vite, Semi Design (@douyinfe/semi-ui)
- 包管理器：**Bun**（`bun install`, `bun run dev/build`）
- i18n: `react-i18next`，**中文作为 key**（`t('中文key')`）

### 代码风格
- 函数组件 + Hooks
- 配置表单使用 `Form` + `Form.Section` + `Card` 组合（参考 `SystemSetting.jsx` 现有板块）
- 表单提交函数命名：`submitXxx`（如 `submitSMS`, `submitSMTP`）
- 表单状态使用 `useState({ key1: '', key2: '' })`
- Checkbox 类型配置项在 `getOptions()` 中会被 `toBoolean()` 转换
- 敏感输入框使用 `type='password'`，placeholder 写 `"敏感信息不会发送到前端显示"`

## 配置管理规范

### 新增配置项流程
1. 在 `model/option.go` 的 `InitOptionMap()` 中追加默认值
2. 如配置项需要实时同步到内存变量（如 `common.SMSEnabled`），在 `updateOptionMap()` 的 switch 中处理
3. 在 `controller/misc.go` 的 `GetStatus()` 中如需要向前端暴露状态，追加到返回 map
4. 在前端 `SystemSetting.jsx` 的 `inputs` 初始状态中追加字段
5. 在前端 `SystemSetting.jsx` 中添加对应的 UI 板块和 `submitXxx` 函数
6. 在前端 `web/src/helpers/data.js` 中如需缓存到 localStorage，追加 `localStorage.setItem('xxx_enabled', data.xxx_enabled)`

## i18n 规范

### 后端
- 库：`nicksnyder/go-i18n/v2`
- 语言文件：`i18n/locales/{en,zh-CN}.yaml`
- 新增错误消息：先在 yaml 中定义 key，再在代码中使用 `i18n.MsgXxx`

### 前端
- 库：`i18next` + `react-i18next`
- 翻译文件：`web/src/i18n/locales/{zh,en,fr,ru,ja,vi}.json`
- 新增 key：直接在代码中使用中文字符串 `t('中文key')`，fallback 会自动显示中文

## 开发流程 checklist

开发新功能前，按以下顺序评估：

1. [ ] 功能能否完整放入 `sidecar/` 目录？
2. [ ] 如果必须改主代码，侵入点有几个？是否都在白名单内？
3. [ ] 是否需要新增 model 字段？（GORM AutoMigrate 即可）
4. [ ] 是否需要新增配置项？（按配置管理规范走）
5. [ ] 是否需要新增前端入口？（LoginForm/RegisterForm 最小改动）
6. [ ] 是否需要新增后台配置 UI？（SystemSetting.jsx 追加 Card）
7. [ ] i18n key 是否已处理？（后端 yaml + 前端 t()）
8. [ ] 数据库兼容性是否已考虑？（三库兼容）
