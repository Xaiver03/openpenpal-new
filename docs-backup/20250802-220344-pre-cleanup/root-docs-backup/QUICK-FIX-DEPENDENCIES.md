# OpenPenPal 快速修复依赖指南

根据依赖检查报告，以下是快速修复缺失依赖的命令：

## 1. 安装 Java 17 (必需)
```bash
# macOS
brew install openjdk@17
sudo ln -sfn /opt/homebrew/opt/openjdk@17/libexec/openjdk.jdk /Library/Java/JavaVirtualMachines/openjdk-17.jdk
echo 'export PATH="/opt/homebrew/opt/openjdk@17/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc

# 验证安装
java -version  # 应显示 version "17.x.x"
```

## 2. 下载 Go 依赖 (必需)
```bash
# 主后端服务
cd backend && go mod download && cd ..

# Courier Service  
cd services/courier-service && go mod download && cd ../..

# Gateway Service
cd services/gateway && go mod download && cd ../..
```

## 3. 安装 Python 依赖 (必需)
```bash
# Write Service
cd services/write-service
source venv/bin/activate
pip install -r requirements.txt
deactivate
cd ../..

# OCR Service
cd services/ocr-service  
source venv/bin/activate
pip install -r requirements.txt
deactivate
cd ../..
```

## 4. 构建 Java Admin Service (必需)
```bash
# 安装 Maven（如果未安装）
brew install maven

# 构建 Admin Service
cd services/admin-service/backend
mvn clean install -DskipTests
cd ../../..
```

## 5. 修复 PostgreSQL 服务 (推荐)
```bash
# 停止有问题的 PostgreSQL 14
brew services stop postgresql@14

# 使用 PostgreSQL 15
brew services start postgresql@15

# 验证连接
psql -U $(whoami) -d openpenpal -c "SELECT version();"
```

## 6. 一键执行所有修复
```bash
# 使用自动化脚本
./install-dependencies.sh
# 选择选项 1 (完整安装)
```

## 验证所有依赖
```bash
# 检查系统依赖
which java && java -version
which mvn && mvn -version  
which go && go version
which python3 && python3 --version
which psql && psql --version
which redis-cli && redis-cli --version

# 检查服务状态
ps aux | grep postgres
redis-cli ping

# 检查依赖文件
ls -la backend/go.sum
ls -la services/write-service/venv/
ls -la services/admin-service/backend/target/
```

## 快速启动项目
```bash
# 所有依赖安装完成后
./startup/quick-start.sh production --auto-open
```

## 常见问题

### Java 版本不对
```bash
# 查看所有 Java 版本
/usr/libexec/java_home -V

# 设置 JAVA_HOME
export JAVA_HOME=$(/usr/libexec/java_home -v 17)
```

### Maven 构建失败
```bash
# 清理并重新构建
cd services/admin-service/backend
mvn clean
rm -rf ~/.m2/repository  # 清理本地仓库缓存
mvn install -DskipTests
```

### Python 虚拟环境问题
```bash
# 重建虚拟环境
cd services/write-service
rm -rf venv
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```