# 微信/支付宝官方支付接入方案

## 一、方案概述

接入微信支付（Native 扫码支付）和支付宝（电脑网站支付）官方 API，作为独立于易支付的直连接入选项。

**核心原则**：完整 Sidecar 化，复用现有 TopUp 订单模型和支付基础设施。

---

## 二、Sidecar 目录结构

```
sidecar/
├── controller/
│   ├── wechat_pay.go      # 微信下单 + 回调 + 订单查询
│   └── alipay.go          # 支付宝下单 + 回调 + 订单查询
├── service/
│   ├── wechat_pay.go      # 微信 SDK 封装（签名、验签、统一下单）
│   └── alipay.go          # 支付宝 SDK 封装（签名、验签、创建订单）
└── router/
    └── payment_native.go  # 路由注册函数
```

---

## 三、侵入点清单（白名单内）

### 3.1 model/option.go

**位置**：`InitOptionMap()` 末尾追加默认值；`updateOptionMap()` 追加配置同步逻辑。

**新增配置项**：

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `WechatPayEnabled` | bool | false | 微信支付总开关 |
| `WechatPayAppId` | string | "" | 微信公众号/小程序 AppID |
| `WechatPayMchId` | string | "" | 微信支付商户号 |
| `WechatPayApiKey` | string | "" | API v3 密钥 |
| `WechatPayApiV3Key` | string | "" | API v3 平台证书密钥 |
| `WechatPayUnitPrice` | float64 | 7.3 | 每单位价格（人民币） |
| `WechatPayMinTopUp` | int | 1 | 最小充值金额 |
| `AlipayEnabled` | bool | false | 支付宝总开关 |
| `AlipayAppId` | string | "" | 支付宝应用 ID |
| `AlipayPrivateKey` | string | "" | 应用私钥（PKCS1/PKCS8） |
| `AlipayPublicKey` | string | "" | 支付宝公钥 |
| `AlipayGatewayUrl` | string | "" | 网关地址（沙箱/正式） |
| `AlipayUnitPrice` | float64 | 7.3 | 每单位价格（人民币） |
| `AlipayMinTopUp` | int | 1 | 最小充值金额 |

> **注意**：微信支付证书（`apiclient_cert.pem` / `apiclient_key.pem`）通过 API v3 的微信平台证书自动获取，无需手动上传文件。

### 3.2 setting/ 包新增

新建 `setting/payment_native.go`：

```go
package setting

var (
    WechatPayEnabled   bool
    WechatPayAppId     string
    WechatPayMchId     string
    WechatPayApiKey    string
    WechatPayApiV3Key  string
    WechatPayUnitPrice float64 = 7.3
    WechatPayMinTopUp  int     = 1

    AlipayEnabled      bool
    AlipayAppId        string
    AlipayPrivateKey   string
    AlipayPublicKey    string
    AlipayGatewayUrl   string
    AlipayUnitPrice    float64 = 7.3
    AlipayMinTopUp     int     = 1
)
```

### 3.3 router/api-router.go

在 `SetApiRouter()` 末尾追加一行：

```go
sidecarRouter.RegisterNativePaymentRoutes(apiRouter)
```

### 3.4 controller/topup.go:GetTopUpInfo()

在现有支付方式追加逻辑之后，追加微信/支付宝官方的启用状态判断，将官方渠道加入 `payMethods` 返回列表：

```go
// 微信支付
if setting.WechatPayEnabled && setting.WechatPayAppId != "" && setting.WechatPayMchId != "" {
    payMethods = append(payMethods, map[string]string{
        "name":      "微信支付",
        "type":      "wechatpay_native",
        "color":     "rgba(var(--semi-green-5), 1)",
        "min_topup": strconv.Itoa(setting.WechatPayMinTopUp),
    })
}

// 支付宝
if setting.AlipayEnabled && setting.AlipayAppId != "" && setting.AlipayPrivateKey != "" {
    payMethods = append(payMethods, map[string]string{
        "name":      "支付宝",
        "type":      "alipay_page",
        "color":     "rgba(var(--semi-blue-5), 1)",
        "min_topup": strconv.Itoa(setting.AlipayMinTopUp),
    })
}
```

同时追加返回字段：

```go
"enable_wechatpay_topup": setting.WechatPayEnabled && ...,
"enable_alipay_topup":    setting.AlipayEnabled && ...,
"wechatpay_min_topup":    setting.WechatPayMinTopUp,
"alipay_min_topup":       setting.AlipayMinTopUp,
```

### 3.5 controller/misc.go:GetStatus()

可选：如前端需要在全局状态中获取支付开关，追加：

```go
"wechatpay_enabled": setting.WechatPayEnabled,
"alipay_enabled":    setting.AlipayEnabled,
```

---

## 四、Sidecar 新增文件详情

### 4.1 sidecar/service/wechat_pay.go

**依赖**：`github.com/silenceper/wechat/v2/pay/notify` + `github.com/wechatpay-apiv3/wechatpay-go`（官方 SDK v3）

**功能**：
- `GetWechatPayClient()` — 初始化微信支付客户端（从 `common.OptionMap` 读取配置）
- `CreateNativeOrder(tradeNo string, amount float64, description string) (codeUrl string, err error)` — 统一下单，返回扫码链接
- `VerifyNotify(body []byte, signature string) bool` — 回调验签
- `GetUnitPrice() float64` — 读取配置单价
- `GetMinTopUp() int64` — 读取最小充值金额

### 4.2 sidecar/service/alipay.go

**依赖**：`github.com/smartwalle/alipay/v3`

**功能**：
- `GetAlipayClient()` — 初始化支付宝客户端
- `CreatePageOrder(tradeNo string, amount float64, subject string, returnUrl string) (payUrl string, err error)` — 创建电脑网站支付订单，返回跳转链接
- `VerifyNotify(params url.Values) bool` — 回调参数验签
- `GetUnitPrice() float64`
- `GetMinTopUp() int64`

### 4.3 sidecar/controller/wechat_pay.go

**接口**：

| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| POST | `/api/user/wechatpay/pay` | UserAuth + CriticalRateLimit | 创建订单，返回 code_url（扫码链接） |
| POST | `/api/wechatpay/notify` | 公开 | 微信支付异步回调 |

**Request/Response**：

```go
type WechatPayRequest struct {
    Amount int64 `json:"amount" binding:"required"`
}
```

下单流程：
1. 校验最小充值金额
2. 调用 `getPayMoney()`（复用主包逻辑，单价用 `WechatPayUnitPrice`，币种 CNY）
3. 生成订单号 `WXP-{userId}-{timestamp}-{random}`
4. 创建 `model.TopUp` 记录，`PaymentMethod = "wechatpay_native"`
5. 调用微信统一下单，获取 `code_url`
6. 返回 `{ code_url, trade_no }`

回调流程：
1. 读取请求 body + 签名头
2. `service.VerifyNotify()` 验签
3. `LockOrder(tradeNo)` / `UnlockOrder(tradeNo)`（复用主包）
4. 查询本地订单，校验状态
5. 调用 `model.RechargeWechatPay(tradeNo)`（新增）完成充值
6. 返回微信要求的响应格式

### 4.4 sidecar/controller/alipay.go

**接口**：

| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| POST | `/api/user/alipay/pay` | UserAuth + CriticalRateLimit | 创建订单，返回跳转 URL |
| POST | `/api/alipay/notify` | 公开 | 支付宝异步回调 |
| GET  | `/api/alipay/return` | 公开 | 支付宝同步返回（仅做页面跳转） |

下单流程与微信类似，返回的是支付宝收银台跳转 URL。

回调流程：
1. 解析 URL query / form body 参数
2. `service.VerifyNotify()` 验签
3. `LockOrder` / `UnlockOrder`
4. `model.RechargeAlipay(tradeNo)`（新增）完成充值
5. 返回 `"success"`

### 4.5 sidecar/router/payment_native.go

```go
package router

import (
    "github.com/QuantumNous/new-api/middleware"
    "github.com/QuantumNous/new-api/sidecar/controller"
    "github.com/gin-gonic/gin"
)

func RegisterNativePaymentRoutes(apiRouter *gin.RouterGroup) {
    // Webhooks (no auth)
    apiRouter.POST("/wechatpay/notify", controller.WechatPayNotify)
    apiRouter.POST("/alipay/notify", controller.AlipayNotify)
    apiRouter.GET("/alipay/return", controller.AlipayReturn)

    // User payment endpoints
    userRoute := apiRouter.Group("/user")
    {
        userRoute.POST("/wechatpay/pay", middleware.UserAuth(), middleware.CriticalRateLimit(), controller.RequestWechatPay)
        userRoute.POST("/alipay/pay", middleware.UserAuth(), middleware.CriticalRateLimit(), controller.RequestAlipay)
    }
}
```

---

## 五、模型层新增

### 5.1 model/topup.go 追加两个充值完成函数

参考现有 `RechargeStripe()` / `RechargeWaffo()` / `RechargeCreem()` 的写法：

```go
func RechargeWechatPay(tradeNo string) error
func RechargeAlipay(tradeNo string) error
```

各约 30 行，逻辑完全一致：
1. 开启事务 + 行级锁 `FOR UPDATE`
2. 校验 `PaymentMethod` 匹配
3. 校验状态为 pending
4. 计算 `quotaToAdd = Amount * QuotaPerUnit`
5. 更新订单状态为 success
6. `UPDATE user SET quota = quota + ?`
7. 事务外调用 `OnTopUpSuccess` 钩子
8. 记录日志

---

## 六、前端改动

### 6.1 web/src/components/settings/SystemSetting.jsx

追加一个新的 Card（参考 SMS 配置 Card），包含：

**微信支付配置 Section**：
- Switch: 启用微信支付
- Input: AppID
- Input: 商户号
- Input(type=password): API v3 密钥
- Input(type=password): API v3 平台证书密钥
- InputNumber: 单价（默认 7.3）
- InputNumber: 最小充值金额（默认 1）

**支付宝配置 Section**：
- Switch: 启用支付宝
- Input: AppID
- Input(type=password): 应用私钥
- Input(type=password): 支付宝公钥
- Input: 网关地址（默认 `https://openapi.alipay.com/gateway.do`）
- InputNumber: 单价（默认 7.3）
- InputNumber: 最小充值金额（默认 1）

表单提交函数：`submitWechatPay()` / `submitAlipay()`

### 6.2 web/src/components/topup/index.jsx

**改动极小**：

1. 状态初始化追加：
```js
const [enableWechatPay, setEnableWechatPay] = useState(false);
const [enableAlipay, setEnableAlipay] = useState(false);
const [wechatPayMinTopUp, setWechatPayMinTopUp] = useState(1);
const [alipayMinTopUp, setAlipayMinTopUp] = useState(1);
```

2. `getTopupInfo()` 中读取新字段并 `setState`

3. 新增两个充值函数：
```js
const wechatPayTopUp = async () => { /* 调用 /api/user/wechatpay/pay，获取 code_url，展示二维码 */ }
const alipayTopUp = async () => { /* 调用 /api/user/alipay/pay，获取 pay_url，window.open */ }
```

### 6.3 web/src/components/topup/RechargeCard.jsx

**不需要改**。现有 `payMethods` 动态渲染逻辑已经支持任意支付方式，`alipay` / `wxpay` 的图标和颜色已有。只需确保 `wechatpay_native` / `alipay_page` 在 `payMethods` 列表中，按钮会自动渲染。

> 唯一可能需要改的是：如果希望前端对 `wechatpay_native` 和 `wxpay` 使用同一个微信图标，在 RechargeCard.jsx 的 icon 判断中追加 `=== 'wechatpay_native'` 分支即可（同理 `alipay_page`）。

### 6.4 web/src/components/topup/modals/TopupHistoryModal.jsx

追加支付方式名称映射：

```js
const paymentMethodNames = {
    // ... existing
    wechatpay_native: '微信支付',
    alipay_page: '支付宝',
};
```

---

## 七、i18n 处理

### 后端

在 `i18n/locales/en.yaml` 和 `zh-CN.yaml` 中追加（如有需要的新错误消息）：

```yaml
# zh-CN.yaml
MsgWechatPayConfigError: "微信支付配置错误"
MsgAlipayConfigError: "支付宝配置错误"
```

### 前端

使用中文 key：`t('微信支付')`、`t('支付宝')` 等，自动 fallback 到中文。

---

## 八、数据库兼容性

- **无新增表/字段**：完全复用 `model.TopUp`
- **无 raw SQL**：所有操作通过 GORM 完成
- **三库兼容**：无需特殊处理

---

## 九、接口汇总

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/user/wechatpay/pay` | 微信下单 |
| POST | `/api/wechatpay/notify` | 微信回调 |
| POST | `/api/user/alipay/pay` | 支付宝下单 |
| POST | `/api/alipay/notify` | 支付宝回调 |
| GET  | `/api/alipay/return` | 支付宝同步返回 |

---

## 十、开发 Checklist

- [ ] `setting/payment_native.go` 新增配置变量
- [ ] `model/option.go` 追加配置默认值和同步逻辑
- [ ] `model/topup.go` 追加 `RechargeWechatPay()` / `RechargeAlipay()`
- [ ] `controller/topup.go:GetTopUpInfo()` 追加支付方式
- [ ] `router/api-router.go` 注册 Sidecar 路由
- [ ] `sidecar/service/wechat_pay.go` 实现
- [ ] `sidecar/service/alipay.go` 实现
- [ ] `sidecar/controller/wechat_pay.go` 实现
- [ ] `sidecar/controller/alipay.go` 实现
- [ ] `sidecar/router/payment_native.go` 注册路由
- [ ] `web/src/components/settings/SystemSetting.jsx` 追加配置 UI
- [ ] `web/src/components/topup/index.jsx` 追加充值逻辑
- [ ] `web/src/components/topup/modals/TopupHistoryModal.jsx` 追加名称映射
- [ ] i18n yaml 追加错误消息
- [ ] `go.mod` 引入微信/支付宝 SDK
