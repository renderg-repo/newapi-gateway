# 微信扫码登录 API 接口文档

微信公众号扫码登录功能，通过直接调用微信公众号 API 实现，无需外部 WeChat Server 服务。

使用 Sidecar 模式实现，对主包代码侵入最小。

---

## 目录

- [配置说明](#配置说明)
- [公开接口](#公开接口)
  - [GET /api/weixin/getQrCode](#get-apiweixinqrcode)
  - [POST /api/weixin/checkQrCode](#post-apiweixincheckqrcode)
  - [GET/POST /api/weixin/receiveMessage](#getpost-apiweixinreceivemessage)
- [数据结构](#数据结构)
  - [QRCodeSession](#qrcodesession)
- [登录流程](#登录流程)

---

## 配置说明

在管理后台 → 系统设置中配置以下参数：

| 配置项 | 说明 |
|--------|------|
| `WeChatAppID` | 微信公众号 AppID |
| `WeChatAppSecret` | 微信公众号 AppSecret |
| `WeChatReceiveToken` | 微信公众号消息回调验证 Token |

微信公众号后台配置：
- 服务器地址（URL）：`https://你的域名/api/weixin/receiveMessage`
- 令牌（Token）：与 `WeChatReceiveToken` 一致
- 消息加解密密钥：可自动生成或自定义（当前实现暂未加密）

---

## 公开接口

### GET /api/weixin/getQrCode

获取微信登录二维码，创建临时扫码会话。

**认证**：无需认证

**查询参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | int | 否 | 固定值 1（预留扩展） |

**响应示例**：

```json
{
  "success": true,
  "data": {
    "ticket": "550e8400-e29b-41d4-a716-446655440000",
    "qrcode_url": "https://mp.weixin.qq.com/cgi-bin/showqrcode?ticket=gQHq7zwAAAAAAAAAAS5odHRwOi8vd2VpeGluLnFxLmNvbS9xL01pQVJZUFUxNGlMWEV6RU5NMU52R1QAAgQY0WdlAwSAOgkA",
    "expires_in": 600
  }
}
```

**字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| ticket | string | 会话 UUID，用于后续轮询状态 |
| qrcode_url | string | 微信二维码图片 URL |
| expires_in | int | 二维码有效期（秒），固定 600 |

---

### POST /api/weixin/checkQrCode

轮询扫码状态，当用户扫码确认后返回用户信息并设置登录 Session。

**认证**：无需认证

**请求体**（JSON）：

```json
{
  "ticket": "550e8400-e29b-41d4-a716-446655440000"
}
```

**响应示例 - 等待扫码**：

```json
{
  "success": true,
  "data": {
    "status": "pending"
  }
}
```

**响应示例 - 已扫码待确认**：

```json
{
  "success": true,
  "data": {
    "status": "scanned"
  }
}
```

**响应示例 - 登录成功**：

```json
{
  "success": true,
  "data": {
    "status": "confirmed",
    "user_id": 123,
    "username": "wechat_123",
    "role": 1,
    "group": "default"
  }
}
```

**响应示例 - 已过期**：

```json
{
  "success": true,
  "data": {
    "status": "expired"
  }
}
```

**状态说明**：

| status | 说明 |
|--------|------|
| pending | 等待用户扫码 |
| scanned | 已扫码，等待确认 |
| confirmed | 登录成功 |
| expired | 二维码已过期 |

---

### GET/POST /api/weixin/receiveMessage

微信公众号消息回调接口。

**GET**：用于微信服务器验证接口有效性（返回 echostr）

**POST**：接收用户扫码事件通知

**注意**：此接口由微信服务器调用，前端不应直接调用。

**认证**：微信服务器签名验证

**查询参数**（GET）：

| 参数 | 类型 | 说明 |
|------|------|------|
| signature | string | 微信签名 |
| timestamp | string | 时间戳 |
| nonce | string | 随机数 |
| echostr | string | 随机字符串 |

**请求体**（POST，XML）：

```xml
<xml>
  <ToUserName><![CDATA[gh_123456789]]></ToUserName>
  <FromUserName><![CDATA[o1234567890abcdef]]></FromUserName>
  <CreateTime>1620000000</CreateTime>
  <MsgType><![CDATA[event]]></MsgType>
  <Event><![CDATA[SCAN]]></Event>
  <EventKey><![CDATA[12345]]></EventKey>
  <Ticket><![CDATA[gQHq7zwAAAAAAAAAAS5odHRwOi8vd2VpeGluLnFxLmNvbS9xL01pQVJZUFUxNGlMWEV6RU5NMU52R1QAAgQY0WdlAwSAOgkA]]></Ticket>
</xml>
```

---

## 数据结构

### QRCodeSession

扫码登录会话（内存存储，通过 sync.Map 实现）：

| 字段 | 类型 | 说明 |
|------|------|------|
| Ticket | string | 会话 UUID |
| SceneID | int | 微信场景 ID（用于回调匹配） |
| Status | QRCodeStatus | 会话状态 |
| OpenID | string | 用户微信 OpenID |
| UserID | int | 关联用户 ID |
| CreatedAt | time.Time | 创建时间 |
| ExpiresAt | time.Time | 过期时间 |

**QRCodeStatus 枚举**：

| 值 | 说明 |
|----|------|
| pending | 等待扫码 |
| scanned | 已扫码 |
| confirmed | 已确认 |
| expired | 已过期 |

---

## 登录流程

1. **前端调用 `GET /api/weixin/getQrCode`**
   - 后端生成 SceneID，创建 QRCodeSession
   - 调用微信 API 创建临时二维码（QR_SCENE，600s 有效期）
   - 返回 ticket、二维码图片 URL、有效期

2. **前端显示二维码并轮询**
   - 显示 `qrcode_url` 图片
   - 每 1-2 秒调用 `POST /api/weixin/checkQrCode` 传入 ticket

3. **用户扫码**
   - 用户使用微信扫描二维码
   - 微信服务器 POST 事件到 `/api/weixin/receiveMessage`

4. **后端处理回调**
   - 验证签名
   - 解析 XML 消息，提取 Event、EventKey（scene_id）、FromUserName（openid）
   - 通过 scene_id 查找 QRCodeSession
   - 通过 openid 查找或创建用户
   - 更新 session 状态为 confirmed

5. **前端轮询返回成功**
   - `checkQrCode` 接口检测到 status=confirmed
   - 返回用户信息并设置 session cookie
   - 前端跳转到首页或重定向页面

---

## 实现说明

1. **Sidecar 模式**：所有代码位于 `sidecar/` 目录下
   - `sidecar/service/wechat_qrcode.go`：业务逻辑 + 微信 API 调用
   - `sidecar/controller/wechat_qrcode.go`：HTTP 处理器
   - `sidecar/router/wechat_qrcode.go`：路由注册

2. **微信 API 调用**：
   - `GetAccessToken()`：获取访问令牌（带缓存）
   - `CreateTempQRCode(scene_id)`：创建临时二维码
   - `VerifySignature()`：验证回调签名
   - `ParseWeChatMessage()`：解析 XML 消息

3. **双 Map 存储**：
   - `ticketStore`：ticket → session（用于前端轮询）
   - `sceneStore`：scene_id → session（用于微信回调匹配）

4. **最小侵入主包**：
   - 仅在 `common/constants.go` 添加配置变量
   - 仅在 `model/option.go` 添加配置初始化和同步
   - 仅在 `router/api-router.go` 添加路由注册调用
