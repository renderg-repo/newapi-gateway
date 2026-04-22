# CLAUDE.md — new-api 项目规范

## 概述

这是一个基于 Go 语言构建的 AI API 网关/代理。它在统一 API 背后聚合了 40+ 上游 AI 提供商（OpenAI、Claude、Gemini、Azure、AWS Bedrock 等），并包含用户管理、计费、速率限制和管理后台。

## 技术栈

- **后端**：Go 1.22+、Gin Web 框架、GORM v2 ORM
- **前端**：React 18、Vite、Semi Design UI (@douyinfe/semi-ui)
- **数据库**：SQLite、MySQL、PostgreSQL（必须同时支持三者）
- **缓存**：Redis（go-redis）+ 内存缓存
- **认证**：JWT、WebAuthn/Passkeys、OAuth（GitHub、Discord、OIDC 等）
- **前端包管理器**：Bun（优先于 npm/yarn/pnpm）

## 架构

分层架构：Router -> Controller -> Service -> Model

```
router/        — HTTP 路由（API、relay、dashboard、web）
controller/    — 请求处理器
service/       — 业务逻辑
model/         — 数据模型与数据库访问（GORM）
relay/         — AI API 中继/代理，含提供商适配器
  relay/channel/ — 提供商专属适配器（openai/、claude/、gemini/、aws/ 等）
middleware/    — 认证、速率限制、CORS、日志、分发
setting/       — 配置管理（倍率、模型、运营、系统、性能）
common/        — 共享工具（JSON、加密、Redis、环境变量、速率限制等）
dto/           — 数据传输对象（请求/响应结构体）
constant/      — 常量（API 类型、渠道类型、上下文键）
types/         — 类型定义（relay 格式、文件来源、错误）
i18n/          — 后端国际化（go-i18n，en/zh）
oauth/         — OAuth 提供商实现
pkg/           — 内部包（cachex、ionet）
web/           — React 前端
  web/src/i18n/  — 前端国际化（i18next，zh/en/fr/ru/ja/vi）
```

## 国际化（i18n）

### 后端（`i18n/`）
- 库：`nicksnyder/go-i18n/v2`
- 语言：en、zh

### 前端（`web/src/i18n/`）
- 库：`i18next` + `react-i18next` + `i18next-browser-languagedetector`
- 语言：zh（回退）、en、fr、ru、ja、vi
- 翻译文件：`web/src/i18n/locales/{lang}.json` — 扁平 JSON，键为中文源字符串
- 用法：`useTranslation()` hook，在组件中调用 `t('中文key')`
- Semi UI 本地化通过 `SemiLocaleWrapper` 同步
- CLI 工具：`bun run i18n:extract`、`bun run i18n:sync`、`bun run i18n:lint`

## 规则

### 规则 1：JSON 包 — 使用 `common/json.go`

所有 JSON 序列化/反序列化操作**必须**使用 `common/json.go` 中的包装函数：

- `common.Marshal(v any) ([]byte, error)`
- `common.Unmarshal(data []byte, v any) error`
- `common.UnmarshalJsonStr(data string, v any) error`
- `common.DecodeJson(reader io.Reader, v any) error`
- `common.GetJsonType(data json.RawMessage) string`

**禁止**在业务代码中直接导入或调用 `encoding/json`。这些包装函数用于保持一致性和未来可扩展性（例如切换到更快的 JSON 库）。

注意：`json.RawMessage`、`json.Number` 以及 `encoding/json` 中的其他类型定义仍可作为类型引用，但实际的序列化/反序列化调用必须通过 `common.*` 进行。

### 规则 2：数据库兼容性 — SQLite、MySQL >= 5.7.8、PostgreSQL >= 9.6

所有数据库代码**必须**同时完全兼容三种数据库。

**使用 GORM 抽象：**
- 优先使用 GORM 方法（`Create`、`Find`、`Where`、`Updates` 等），避免原始 SQL。
- 让 GORM 处理主键生成 — 不要直接使用 `AUTO_INCREMENT` 或 `SERIAL`。

**不可避免使用原始 SQL 时：**
- 列引号差异：PostgreSQL 使用 `"column"`，MySQL/SQLite 使用 `` `column` ``。
- 对于 `group`、`key` 等保留字列，使用 `model/main.go` 中的 `commonGroupCol`、`commonKeyCol` 变量。
- 布尔值差异：PostgreSQL 使用 `true`/`false`，MySQL/SQLite 使用 `1`/`0`。使用 `commonTrueVal`/`commonFalseVal`。
- 使用 `common.UsingPostgreSQL`、`common.UsingSQLite`、`common.UsingMySQL` 标志来分支数据库特定逻辑。

**无跨数据库回退方案时禁止：**
- MySQL 专属函数（例如 `GROUP_CONCAT`，无 PostgreSQL `STRING_AGG` 等价物）
- PostgreSQL 专属运算符（例如 `@>`、`?`、`JSONB` 运算符）
- SQLite 中的 `ALTER COLUMN`（不支持 — 使用添加列的变通方案）
- 无回退方案的数据库专属列类型 — JSON 存储使用 `TEXT` 而非 `JSONB`

**迁移：**
- 确保所有迁移在三种数据库上都可用。
- 对于 SQLite，使用 `ALTER TABLE ... ADD COLUMN` 代替 `ALTER COLUMN`（参考 `model/main.go` 中的模式）。

### 规则 3：前端 — 优先使用 Bun

前端（`web/` 目录）使用 `bun` 作为首选包管理器和脚本运行器：
- `bun install` 安装依赖
- `bun run dev` 启动开发服务器
- `bun run build` 生产构建
- `bun run i18n:*` 运行 i18n 工具

### 规则 4：新渠道的 StreamOptions 支持

实现新渠道时：
- 确认该提供商是否支持 `StreamOptions`。
- 如支持，将该渠道添加到 `streamSupportedChannels`。


### 规则 6：上游 Relay 请求 DTO — 保留显式零值

对于从客户端 JSON 解析并重新序列化到上游提供商的请求结构体（尤其是 relay/convert 路径）：

- 可选标量字段**必须**使用指针类型配合 `omitempty`（例如 `*int`、`*uint`、`*float64`、`*bool`），而非非指针标量。
- 语义**必须**为：
  - 客户端 JSON 中缺少该字段 => `nil` => 序列化时省略；
  - 字段显式设置为零/假 => 非 `nil` 指针 => 必须仍发送到上游。
- 避免对可选请求参数使用非指针标量配合 `omitempty`，因为零值（`0`、`0.0`、`false`）会在序列化时被静默丢弃。
