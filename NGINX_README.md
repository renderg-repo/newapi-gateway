# Nginx 反向代理配置

本目录包含多个 nginx 配置文件，用于同时服务 APIToken 前端页面和 newapi-gateway API。

## 📁 配置文件说明

| 文件 | 用途 |
|------|------|
| `nginx.conf` | 完整版生产配置 |
| `nginx-simple.conf` | 简化版配置，适合快速测试 |
| `nginx-docker.conf` | Docker 环境专用配置 |

## 🚀 快速开始

### 前提条件

1. **APIToken (Next.js 前端)** 运行在 `http://127.0.0.1:3000`
2. **newapi-gateway (Go 后端)** 运行在 `http://127.0.0.1:3001`

### 方式一：直接使用 nginx（推荐本地开发）

```bash
# 1. 启动 APIToken
cd ../APIToken
npm install
npm run dev  # 默认端口 3000

# 2. 启动 newapi-gateway (在另一个终端)
cd ../newapi-gateway
go mod download
go run main.go  # 确保监听 3001 端口

# 3. 启动 nginx
nginx -c $(pwd)/nginx-simple.conf

# 4. 访问
# - 前端页面: http://localhost
# - API 测试: curl http://localhost/api/status
```

### 方式二：使用 Docker Compose

```bash
# 1. 先启动 APIToken (宿主机)
cd ../APIToken
npm install
npm run dev  # 监听 3000

# 2. 修改 newapi-gateway 端口为 3001
# 编辑 main.go 或配置文件，确保监听 3001

# 3. 启动 docker-compose
cd ../newapi-gateway
docker-compose -f docker-compose-nginx.yml up -d

# 4. 访问
# - 前端页面: http://localhost
# - API: http://localhost/v1/...
```

## 🔌 暴露的 API 接口

### OpenAI 兼容 API (`/v1/*`)

| 路径 | 功能 |
|------|------|
| `/v1/chat/completions` | 聊天 |
| `/v1/completions` | 补全 |
| `/v1/embeddings` | 嵌入 |
| `/v1/images/generations` | 图像生成 |
| `/v1/images/edits` | 图像编辑 |
| `/v1/audio/transcriptions` | 音频转录 |
| `/v1/audio/translations` | 音频翻译 |
| `/v1/audio/speech` | 语音生成 |
| `/v1/moderations` | 内容审查 |
| `/v1/models` | 模型列表 |
| `/v1/realtime` | 实时语音 (WebSocket) |
| `/v1/rerank` | 重排 |

### 其他 API

| 路径 | 功能 |
|------|------|
| `/api/*` | 管理后台 API |
| `/v1beta/*` | Gemini 兼容 API |
| `/mj/*` | Midjourney API |
| `/suno/*` | Suno API |
| `/pg/*` | Playground API |

## ⚙️ 端口修改

如果你的服务运行在不同端口，修改配置中的 upstream：

```nginx
upstream apitoken {
    server 127.0.0.1:YOUR_FRONTEND_PORT;
}

upstream newapi {
    server 127.0.0.1:YOUR_BACKEND_PORT;
}
```

## 🔧 nginx 常用命令

```bash
# 启动
nginx -c /path/to/config.conf

# 重新加载配置（不中断服务）
nginx -s reload

# 停止
nginx -s stop

# 测试配置文件语法
nginx -t -c /path/to/config.conf
```

## 📝 注意事项

1. **大文件上传**: 配置中设置了 `client_max_body_size 100M`，可根据需要调整
2. **超时时间**: API 的超时时间设置较长（300s-600s），以支持长对话
3. **WebSocket**: 已配置支持 WebSocket，用于实时语音等功能
4. **生产环境**: 建议添加 SSL 证书，使用 HTTPS

## 🔒 生产环境建议

1. 添加 SSL 证书（Let's Encrypt）
2. 配置基本的访问控制
3. 启用更详细的访问日志
4. 添加速率限制
5. 配置防火墙规则

## 🆘 故障排查

```bash
# 查看 nginx 错误日志
tail -f /var/log/nginx/error.log

# 查看访问日志
tail -f /var/log/nginx/access.log

# 测试端口连通性
curl -v http://127.0.0.1:3000  # 测试前端
curl -v http://127.0.0.1:3001/api/status  # 测试后端
```
