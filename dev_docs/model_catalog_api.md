# Model Catalog API 接口文档

模型目录（Model Catalog）接口用于聚合定价数据与模型规格元数据，为前端模型列表/卡片页面提供统一查询能力。

数据表由 sidecar 私有维护（`sidecar/model`），主包 pricing 逻辑零侵入。

---

## 目录

- [公开接口](#公开接口)
  - [GET /api/model-catalog](#get-apimodel-catalog)
  - [GET /api/model-catalog/:model_name](#get-apimodel-catalogmodel_name)
- [管理接口](#管理接口)
  - [GET /api/model-catalog/admin/specs](#get-apimodel-catalogadminspecs)
  - [GET /api/model-catalog/admin/specs/:id](#get-apimodel-catalogadminspecsid)
  - [POST /api/model-catalog/admin/specs](#post-apimodel-catalogadminspecs)
  - [PUT /api/model-catalog/admin/specs](#put-apimodel-catalogadminspecs)
  - [DELETE /api/model-catalog/admin/specs/:id](#delete-apimodel-catalogadminspecsid)
- [CatalogModel 数据结构](#catalogmodel-数据结构)
- [ModelSpec 数据结构](#modelspec-数据结构)

---

## 公开接口

### GET /api/model-catalog

返回全部模型目录，聚合 pricing + model_spec 数据。

**认证**：无需认证（公开接口）

**请求参数**：无

**响应示例**：

```json
{
  "success": true,
  "data": [
    {
      "model_name": "gpt-4o",
      "description": "OpenAI 最新旗舰模型，多模态能力强大",
      "vendor_id": 1,
      "vendor_name": "OpenAI",
      "vendor_icon": "openai",
      "context_length": 128000,
      "max_output_tokens": 4096,
      "capabilities": ["chat", "function", "vision", "streaming"],
      "model_ratio": 5.0,
      "completion_ratio": 15.0,
      "model_price": 0,
      "cache_ratio": null,
      "create_cache_ratio": null,
      "image_ratio": null,
      "audio_ratio": null,
      "audio_completion_ratio": null,
      "enable_groups": ["default", "vip"],
      "tags": "multimodal,vision",
      "supported_endpoint_types": ["openai"],
      "billing_mode": "",
      "billing_expr": "",
      "icon": "",
      "release_date": "2024-08-15",
      "knowledge_cutoff": "2024-08",
      "parameter_count": ""
    }
  ]
}
```

---

### GET /api/model-catalog/:model_name

返回单个模型的聚合数据。

**认证**：无需认证（公开接口）

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| model_name | string | 模型名称（URL 编码） |

**响应示例**：

```json
{
  "success": true,
  "data": {
    "model_name": "gpt-4o",
    "context_length": 128000,
    "max_output_tokens": 4096,
    "capabilities": ["chat", "function", "vision"],
    ...
  }
}
```

---

## 管理接口

管理接口用于维护模型规格数据（CRUD），需管理员权限。

### GET /api/model-catalog/admin/specs

分页查询 ModelSpec 列表，支持关键词搜索。

**认证**：`AdminAuth`

**查询参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| keyword | string | 否 | 按 model_name 模糊搜索 |
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页条数，默认 20 |

**响应**：标准分页结构，`items` 为 `ModelSpec[]`。

---

### GET /api/model-catalog/admin/specs/:id

获取单条 ModelSpec 详情。

**认证**：`AdminAuth`

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 规格记录 ID |

---

### POST /api/model-catalog/admin/specs

创建模型规格记录。

**认证**：`AdminAuth`

**请求体**（JSON）：

```json
{
  "model_name": "gpt-4o",
  "context_length": 128000,
  "max_output_tokens": 4096,
  "capabilities": ["chat", "function", "vision", "streaming"],
  "description": "OpenAI 最新旗舰模型",
  "icon": "openai",
  "release_date": "2024-08-15",
  "knowledge_cutoff": "2024-08",
  "parameter_count": "",
  "status": 1
}
```

**字段说明**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| model_name | string | 是 | 模型名称，全局唯一 |
| context_length | int | 否 | 上下文窗口长度（token 数） |
| max_output_tokens | int | 否 | 单次最大输出 token 数 |
| capabilities | string[] | 否 | 能力标签，如 `["chat","vision"]` |
| description | string | 否 | 模型描述 |
| icon | string | 否 | 图标标识 |
| release_date | string | 否 | 发布日期，格式 `YYYY-MM-DD` |
| knowledge_cutoff | string | 否 | 知识截止日期，格式 `YYYY-MM` |
| parameter_count | string | 否 | 参数量，如 `"1.5B"`、`"70B"` |
| status | int | 否 | 0=禁用，1=启用，默认 1 |

---

### PUT /api/model-catalog/admin/specs

更新模型规格记录。

**认证**：`AdminAuth`

**请求体**（JSON）：与 POST 相同，必须包含 `id` 字段。

---

### DELETE /api/model-catalog/admin/specs/:id

删除模型规格记录（软删除）。

**认证**：`AdminAuth`

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 规格记录 ID |

---

## CatalogModel 数据结构

公开接口返回的聚合模型对象：

| 字段 | 类型 | 来源 | 说明 |
|------|------|------|------|
| model_name | string | pricing | 模型名称 |
| description | string | spec > pricing | 描述（spec 优先） |
| vendor_id | int | pricing | 供应商 ID |
| vendor_name | string | pricing | 供应商名称 |
| vendor_icon | string | pricing | 供应商图标 |
| context_length | int | spec | 上下文长度 |
| max_output_tokens | int | spec | 最大输出 token |
| capabilities | string[] | spec | 能力标签 |
| model_ratio | float64 | pricing | 模型倍率 |
| completion_ratio | float64 | pricing | 补全倍率 |
| model_price | float64 | pricing | 固定价格 |
| cache_ratio | float64? | pricing | 缓存读取倍率 |
| create_cache_ratio | float64? | pricing | 缓存创建倍率 |
| image_ratio | float64? | pricing | 图像倍率 |
| audio_ratio | float64? | pricing | 音频输入倍率 |
| audio_completion_ratio | float64? | pricing | 音频补全倍率 |
| enable_groups | string[] | pricing | 可用分组 |
| tags | string | pricing | 标签 |
| supported_endpoint_types | string[] | pricing | 支持的端点类型 |
| billing_mode | string | pricing | 计费模式 |
| billing_expr | string | pricing | 计费表达式 |
| icon | string | spec > pricing | 图标（spec 优先） |
| release_date | string | spec | 发布日期 |
| knowledge_cutoff | string | spec | 知识截止 |
| parameter_count | string | spec | 参数量 |
| has_variable_pricing | bool | pricing（由 doubao 价格表推导） | 是否存在按请求参数变化的多维定价，如 Seedance 2.0 按分辨率/视频输入变价 |
| price_variants | PriceVariant[] | pricing（由 doubao 价格表推导） | 多维定价档位列表，仅 has_variable_pricing 为 true 时返回 |

---

## PriceVariant 数据结构

多维定价档位：

| 字段 | 类型 | 说明 |
|------|------|------|
| resolution | string | 输出分辨率档位，如 `480p/720p`、`1080p` |
| has_video | bool | 输入是否包含视频 |
| input_price | float64 | 该档位每百万 token 输入价格（USD） |

---

## ModelSpec 数据结构

sidecar 私有数据表 `model_specs`：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int | 自增主键 |
| model_name | string(128) | 唯一索引，关联 pricing 的 model_name |
| context_length | int | 上下文窗口长度 |
| max_output_tokens | int | 单次最大输出 token |
| capabilities | text | JSON 数组字符串，如 `["chat","vision"]` |
| description | text | 模型描述 |
| icon | varchar(128) | 图标标识 |
| release_date | varchar(32) | 发布日期 |
| knowledge_cutoff | varchar(32) | 知识截止日期 |
| parameter_count | varchar(32) | 参数量 |
| status | int | 0=禁用，1=启用 |
| created_time | bigint | 创建时间戳 |
| updated_time | bigint | 更新时间戳 |
| deleted_at | datetime | 软删除标记 |

---

## 设计说明

1. **零侵入主包 pricing**：`model_spec` 表由 sidecar 独立维护，不修改 `model.Pricing` 结构体。
2. **聚合逻辑**：service 层从 `model.GetPricing()` 拉取定价数据，从 sidecar 表拉取规格数据，按 `model_name` 匹配后合并返回。
3. **数据回退**：如果某模型没有 spec 记录，`context_length`、`max_output_tokens`、`capabilities` 等字段返回零值/空值，前端可继续使用 `inferModelMetadata()` 作为兜底推断。
