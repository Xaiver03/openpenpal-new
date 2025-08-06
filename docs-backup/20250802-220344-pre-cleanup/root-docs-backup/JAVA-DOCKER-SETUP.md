# Java 17 和 Docker 安装配置指南

## Java 17 安装

### 选项1：继续等待Homebrew安装（推荐）
```bash
# 检查安装进度
brew list | grep openjdk

# 如果安装完成，配置Java环境
sudo ln -sfn /opt/homebrew/opt/openjdk@17/libexec/openjdk.jdk /Library/Java/JavaVirtualMachines/openjdk-17.jdk

# 添加到环境变量
echo 'export PATH="/opt/homebrew/opt/openjdk@17/bin:$PATH"' >> ~/.zshrc
echo 'export JAVA_HOME="/opt/homebrew/opt/openjdk@17"' >> ~/.zshrc
source ~/.zshrc

# 验证安装
java --version
```

### 选项2：手动下载安装（如果Homebrew太慢）
1. 访问：https://adoptium.net/temurin/releases/
2. 下载：macOS aarch64（Apple Silicon）版本的JDK 17
3. 双击DMG文件安装
4. 验证：`java --version`

### 选项3：使用Docker运行Java服务（临时方案）
```bash
# 为Admin Service创建Dockerfile
cd services/admin-service/backend
cat > Dockerfile << 'EOF'
FROM maven:3.9-eclipse-temurin-17 AS build
WORKDIR /app
COPY pom.xml .
COPY src ./src
RUN mvn clean package -DskipTests

FROM eclipse-temurin:17-jre
WORKDIR /app
COPY --from=build /app/target/*.jar app.jar
EXPOSE 8003
ENTRYPOINT ["java", "-jar", "app.jar"]
EOF

# 构建并运行
docker build -t openpenpal-admin .
docker run -d -p 8003:8003 --name admin-service openpenpal-admin
```

## Docker 配置

### Docker已成功启动！ ✅

当前Docker状态：
- 版本：28.3.2
- 状态：运行中
- 架构：darwin/arm64

### 使用Docker运行PostgreSQL和Redis（可选）
```bash
# 创建Docker网络
docker network create openpenpal-network

# 运行PostgreSQL
docker run -d \
  --name openpenpal-postgres \
  --network openpenpal-network \
  -e POSTGRES_USER=rocalight \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=openpenpal \
  -p 5432:5432 \
  postgres:15

# 运行Redis
docker run -d \
  --name openpenpal-redis \
  --network openpenpal-network \
  -p 6379:6379 \
  redis:latest

# 检查容器状态
docker ps
```

### 使用Docker Compose（推荐）
创建 `docker-compose.yml`:
```yaml
version: '3.8'
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_USER: rocalight
      POSTGRES_PASSWORD: password
      POSTGRES_DB: openpenpal
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:latest
    ports:
      - "6379:6379"

  admin-service:
    build: ./services/admin-service/backend
    ports:
      - "8003:8003"
    depends_on:
      - postgres
    environment:
      SPRING_DATASOURCE_URL: jdbc:postgresql://postgres:5432/openpenpal
      SPRING_DATASOURCE_USERNAME: rocalight
      SPRING_DATASOURCE_PASSWORD: password

volumes:
  postgres_data:
```

运行：
```bash
docker-compose up -d
```

## 快速验证

### 1. 检查Java安装
```bash
java --version
javac --version
mvn --version
```

### 2. 检查Docker
```bash
docker --version
docker ps
docker images
```

### 3. 测试Admin Service（Java安装后）
```bash
cd services/admin-service/backend
./mvnw clean spring-boot:run
```

## 下一步

1. **Java安装完成后**：
   ```bash
   # 构建Admin Service
   cd services/admin-service/backend
   mvn clean install
   
   # 启动完整生产模式
   ./startup/quick-start.sh production --auto-open
   ```

2. **使用Docker替代方案**：
   ```bash
   # 仅运行非Java服务
   ./startup/quick-start.sh demo --auto-open
   
   # 用Docker运行Admin Service
   docker run -d -p 8003:8003 openpenpal-admin
   ```

## 故障排除

### Java相关
- 如果`java --version`失败，检查PATH：`echo $PATH`
- 确保JAVA_HOME正确：`echo $JAVA_HOME`
- 重新加载配置：`source ~/.zshrc`

### Docker相关
- 如果Docker命令失败，确保Docker Desktop正在运行
- 检查Docker守护进程：`docker info`
- 重启Docker Desktop：退出应用并重新打开

### 端口冲突
```bash
# 检查端口占用
lsof -i :8003
lsof -i :5432
lsof -i :6379

# 停止占用端口的进程
kill -9 <PID>
```