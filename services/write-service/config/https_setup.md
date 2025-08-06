# HTTPS/WSS 配置指南

## 概述
为了确保数据传输安全，OpenPenPal Write Service 需要启用 HTTPS 和 WSS（WebSocket Secure）连接。

## 1. SSL证书获取

### 方式一：Let's Encrypt 免费证书（推荐）

```bash
# 安装 certbot
sudo apt-get update
sudo apt-get install certbot

# 获取证书（假设域名为 api.openpenpal.com）
sudo certbot certonly --standalone -d api.openpenpal.com

# 证书文件将保存在：
# /etc/letsencrypt/live/api.openpenpal.com/fullchain.pem
# /etc/letsencrypt/live/api.openpenpal.com/privkey.pem
```

### 方式二：自签名证书（开发环境）

```bash
# 创建证书目录
mkdir -p /path/to/certs

# 生成私钥
openssl genrsa -out /path/to/certs/server.key 2048

# 生成证书签名请求
openssl req -new -key /path/to/certs/server.key -out /path/to/certs/server.csr

# 生成自签名证书
openssl x509 -req -days 365 -in /path/to/certs/server.csr -signkey /path/to/certs/server.key -out /path/to/certs/server.crt
```

## 2. Nginx 反向代理配置

### 配置文件：/etc/nginx/sites-available/openpenpal-write-service

```nginx
server {
    listen 80;
    server_name api.openpenpal.com;
    
    # 重定向HTTP到HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name api.openpenpal.com;
    
    # SSL证书配置
    ssl_certificate /etc/letsencrypt/live/api.openpenpal.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.openpenpal.com/privkey.pem;
    
    # SSL安全配置
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-SHA256:ECDHE-RSA-AES256-SHA384;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 1d;
    ssl_session_tickets off;
    
    # HSTS头部
    add_header Strict-Transport-Security "max-age=63072000" always;
    
    # 其他安全头部
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Referrer-Policy "strict-origin-when-cross-origin";
    
    # 反向代理到FastAPI应用
    location / {
        proxy_pass http://127.0.0.1:8001;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-Host $host;
        proxy_set_header X-Forwarded-Port $server_port;
        
        # WebSocket支持
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_read_timeout 86400;
    }
    
    # WebSocket专用配置（如果需要单独处理）
    location /ws/ {
        proxy_pass http://127.0.0.1:8001;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 3600;
        proxy_send_timeout 3600;
    }
}
```

### 启用配置

```bash
# 创建软链接
sudo ln -s /etc/nginx/sites-available/openpenpal-write-service /etc/nginx/sites-enabled/

# 检查配置
sudo nginx -t

# 重载Nginx
sudo systemctl reload nginx
```

## 3. FastAPI HTTPS 配置

### 更新 uvicorn 启动参数

```python
# main.py
if __name__ == "__main__":
    import uvicorn
    
    # 生产环境HTTPS配置
    if settings.enable_https:
        uvicorn.run(
            app, 
            host="127.0.0.1",  # 只监听本地，由Nginx处理SSL
            port=8001,
            ssl_keyfile=settings.ssl_keyfile,
            ssl_certfile=settings.ssl_certfile
        )
    else:
        # 开发环境
        uvicorn.run(app, host="0.0.0.0", port=8001)
```

### 环境变量配置

```bash
# .env 文件
ENABLE_HTTPS=true
SSL_KEYFILE=/etc/letsencrypt/live/api.openpenpal.com/privkey.pem
SSL_CERTFILE=/etc/letsencrypt/live/api.openpenpal.com/fullchain.pem
FRONTEND_URL=https://app.openpenpal.com
```

## 4. WebSocket WSS 配置

### 前端连接更新

```javascript
// 从 ws:// 更新为 wss://
const websocket = new WebSocket('wss://api.openpenpal.com/ws/notifications');
```

### 连接字符串更新

```python
# config.py
class Settings:
    def __init__(self):
        # WebSocket URL配置
        if self.enable_https:
            self.websocket_url = f"wss://{self.domain}/ws"
        else:
            self.websocket_url = f"ws://{self.domain}:8001/ws"
```

## 5. 安全头部配置

### FastAPI中间件

```python
from fastapi.middleware.trustedhost import TrustedHostMiddleware
from fastapi.middleware.httpsredirect import HTTPSRedirectMiddleware

# 强制HTTPS重定向
if settings.enable_https:
    app.add_middleware(HTTPSRedirectMiddleware)

# 受信任主机
app.add_middleware(
    TrustedHostMiddleware, 
    allowed_hosts=["api.openpenpal.com", "*.openpenpal.com"]
)
```

## 6. 证书自动续期

### Let's Encrypt 自动续期

```bash
# 添加到 crontab
sudo crontab -e

# 每月1号凌晨2点检查并续期
0 2 1 * * /usr/bin/certbot renew --quiet && /usr/sbin/service nginx reload
```

## 7. 安全检查清单

- [ ] SSL证书正确安装
- [ ] HTTP自动重定向到HTTPS
- [ ] HSTS头部已配置
- [ ] 安全头部已添加
- [ ] WebSocket使用WSS协议
- [ ] 证书自动续期已配置
- [ ] 防火墙规则已更新（443端口开放）
- [ ] SSL Labs测试评分A+

## 8. 测试验证

### SSL检查命令

```bash
# 检查证书有效性
openssl s_client -connect api.openpenpal.com:443 -servername api.openpenpal.com

# 检查证书到期时间
echo | openssl s_client -connect api.openpenpal.com:443 2>/dev/null | openssl x509 -noout -dates
```

### 在线测试工具

- SSL Labs: https://www.ssllabs.com/ssltest/
- SecurityHeaders: https://securityheaders.com/

## 9. 故障排除

### 常见问题

1. **证书链不完整**：确保使用 fullchain.pem 而不是 cert.pem
2. **权限问题**：确保Nginx有读取证书文件的权限
3. **防火墙阻止**：确保443端口已开放
4. **WebSocket连接失败**：检查Nginx的WebSocket配置

### 日志查看

```bash
# Nginx错误日志
sudo tail -f /var/log/nginx/error.log

# SSL相关日志
sudo journalctl -u nginx -f | grep -i ssl
```