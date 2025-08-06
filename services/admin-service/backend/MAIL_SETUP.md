# 邮件服务配置指南

## 🚀 快速开始

### 1. Gmail SMTP 配置（推荐）

**步骤：**
1. **创建或使用 Gmail 账户**
   - 访问 [Gmail](https://gmail.com) 并登录

2. **开启两步验证**
   - 访问 [Google 账户安全设置](https://myaccount.google.com/security)
   - 点击"两步验证" -> 开启

3. **生成应用密码**
   - 访问 [应用密码生成](https://myaccount.google.com/apppasswords)
   - 选择应用：**邮件**
   - 选择设备：**其他（自定义名称）**
   - 输入名称：`OpenPenPal System`
   - 点击"生成"
   - **复制 16 位应用密码**（格式：`xxxx xxxx xxxx xxxx`）

4. **配置环境变量**
   ```bash
   # 复制配置文件
   cp .env.example .env
   
   # 编辑 .env 文件
   MAIL_USERNAME=your-email@gmail.com
   MAIL_APP_PASSWORD=your-16-digit-app-password
   MAIL_FROM=noreply@openpenpal.com
   MAIL_FROM_NAME=OpenPenPal
   ```

### 2. QQ邮箱配置（备选）

**步骤：**
1. **登录 QQ 邮箱**
   - 访问 [QQ邮箱](https://mail.qq.com)

2. **开启 SMTP 服务**
   - 设置 -> 账户 -> POP3/SMTP服务 -> 开启
   - 生成**授权码**（16位字母数字组合）

3. **配置环境变量**
   ```bash
   MAIL_USERNAME=your-email@qq.com
   MAIL_PASSWORD=your-qq-auth-code
   ```

### 3. 腾讯企业邮箱配置（企业用户）

```bash
MAIL_USERNAME=your-email@your-domain.com
MAIL_PASSWORD=your-password
```

## 📧 测试邮件发送

### 启动服务测试
```bash
# 设置环境变量
export MAIL_USERNAME=your-email@gmail.com
export MAIL_APP_PASSWORD=your-app-password

# 启动后端服务
./mvnw spring-boot:run
```

### API 测试
```bash
# 发送验证码测试
curl -X POST http://localhost:8003/api/auth/send-verification-code \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"}'
```

## 🔧 配置说明

### 邮件模板位置
```
src/main/resources/templates/email/
├── verification-code.html      # 验证码邮件模板
└── registration-success.html   # 注册成功邮件模板
```

### 配置文件
- `application-mail.yml` - 邮件服务配置
- `.env` - 环境变量（需要自行创建）

### 安全配置
- 使用环境变量存储敏感信息
- 应用密码而非账户密码
- STARTTLS 加密传输

## ⚠️ 注意事项

1. **Gmail 限制**
   - 每天发送限制：500封（新账户）/ 2000封（已验证账户）
   - 建议生产环境使用企业邮箱

2. **防火墙设置**
   - 确保服务器可访问 SMTP 端口（587/465）

3. **域名配置**
   - 生产环境建议配置 SPF、DKIM 记录
   - 使用真实域名作为发件地址

## 🚀 生产环境推荐

1. **使用专业邮件服务商**
   - SendGrid
   - Amazon SES  
   - 阿里云邮件推送
   - 腾讯云邮件推送

2. **配置示例（Amazon SES）**
   ```yaml
   spring:
     mail:
       host: email-smtp.us-east-1.amazonaws.com
       port: 587
       username: ${AWS_SES_USERNAME}
       password: ${AWS_SES_PASSWORD}
   ```

## 🐛 故障排除

### 常见问题
1. **Authentication failed**
   - 检查用户名/密码是否正确
   - 确认已开启两步验证（Gmail）
   - 使用应用密码而非账户密码

2. **Connection timeout**
   - 检查网络连接
   - 确认防火墙设置
   - 尝试不同的端口（587/465）

3. **SSL/TLS errors**
   - 检查 STARTTLS 配置
   - 添加 SSL trust 配置

### 调试模式
```yaml
logging:
  level:
    org.springframework.mail: DEBUG
    javax.mail: DEBUG
```