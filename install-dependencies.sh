#!/bin/bash

# OpenPenPal 依赖安装脚本
# 自动安装项目所需的所有依赖

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查命令是否存在
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# 检查并安装 Homebrew (macOS)
install_homebrew() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        if ! command_exists brew; then
            log_info "安装 Homebrew..."
            /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
            
            # 添加 Homebrew 到 PATH
            if [[ -f /opt/homebrew/bin/brew ]]; then
                eval "$(/opt/homebrew/bin/brew shellenv)"
            elif [[ -f /usr/local/bin/brew ]]; then
                eval "$(/usr/local/bin/brew shellenv)"
            fi
        fi
    fi
}

# 安装系统级依赖
install_system_dependencies() {
    log_info "检查并安装系统级依赖..."
    
    # 检查操作系统
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        install_homebrew
        
        # 安装 Java 17
        if ! command_exists java || ! java -version 2>&1 | grep -q "version \"17"; then
            log_info "安装 Java 17..."
            brew install openjdk@17
            
            # 创建符号链接
            sudo ln -sfn /opt/homebrew/opt/openjdk@17/libexec/openjdk.jdk /Library/Java/JavaVirtualMachines/openjdk-17.jdk
            
            # 添加到 PATH
            echo 'export PATH="/opt/homebrew/opt/openjdk@17/bin:$PATH"' >> ~/.zshrc
            export PATH="/opt/homebrew/opt/openjdk@17/bin:$PATH"
            
            log_success "Java 17 安装完成"
        else
            log_success "Java 17 已安装"
        fi
        
        # 检查 Maven
        if ! command_exists mvn; then
            log_info "安装 Maven..."
            brew install maven
            log_success "Maven 安装完成"
        else
            log_success "Maven 已安装"
        fi
        
        # 修复 PostgreSQL 服务
        if brew services list | grep -q "postgresql@14.*error"; then
            log_warning "修复 PostgreSQL@14 错误..."
            brew services stop postgresql@14 || true
            brew services start postgresql@15 || true
            log_success "PostgreSQL 服务已修复"
        fi
        
        # 安装 Docker (可选)
        if ! command_exists docker && ! [[ -x "/Applications/Docker.app/Contents/Resources/bin/docker" ]]; then
            read -p "是否安装 Docker Desktop？(y/n) " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                log_info "安装 Docker Desktop..."
                brew install --cask docker
                log_success "Docker Desktop 安装完成，请手动启动 Docker.app"
            fi
        else
            log_success "Docker 已安装"
        fi
        
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux
        if command_exists apt-get; then
            # Debian/Ubuntu
            log_info "更新包列表..."
            sudo apt-get update
            
            # 安装 Java 17
            if ! command_exists java || ! java -version 2>&1 | grep -q "version \"17"; then
                log_info "安装 Java 17..."
                sudo apt-get install -y openjdk-17-jdk
                log_success "Java 17 安装完成"
            fi
            
            # 安装 Maven
            if ! command_exists mvn; then
                log_info "安装 Maven..."
                sudo apt-get install -y maven
                log_success "Maven 安装完成"
            fi
            
        elif command_exists yum; then
            # RedHat/CentOS
            log_warning "请手动安装 Java 17 和 Maven"
        fi
    fi
}

# 安装 Go 依赖
install_go_dependencies() {
    log_info "安装 Go 依赖..."
    
    if command_exists go; then
        # 主后端服务
        if [[ -f "backend/go.mod" ]]; then
            log_info "安装主后端服务 Go 依赖..."
            cd backend
            go mod download
            go mod tidy
            cd ..
            log_success "主后端服务 Go 依赖安装完成"
        fi
        
        # Courier Service
        if [[ -f "services/courier-service/go.mod" ]]; then
            log_info "安装 Courier Service Go 依赖..."
            cd services/courier-service
            go mod download
            go mod tidy
            cd ../..
            log_success "Courier Service Go 依赖安装完成"
        fi
        
        # Gateway Service
        if [[ -f "services/gateway/go.mod" ]]; then
            log_info "安装 Gateway Service Go 依赖..."
            cd services/gateway
            go mod download
            go mod tidy
            cd ../..
            log_success "Gateway Service Go 依赖安装完成"
        fi
    else
        log_warning "Go 未安装，跳过 Go 依赖安装"
    fi
}

# 安装 Python 依赖
install_python_dependencies() {
    log_info "安装 Python 依赖..."
    
    if command_exists python3; then
        # Write Service
        if [[ -f "services/write-service/requirements.txt" ]]; then
            log_info "安装 Write Service Python 依赖..."
            cd services/write-service
            
            # 创建虚拟环境（如果不存在）
            if [[ ! -d "venv" ]]; then
                python3 -m venv venv
            fi
            
            # 激活虚拟环境并安装依赖
            source venv/bin/activate
            pip install --upgrade pip
            pip install -r requirements.txt
            deactivate
            
            cd ../..
            log_success "Write Service Python 依赖安装完成"
        fi
        
        # OCR Service
        if [[ -f "services/ocr-service/requirements.txt" ]]; then
            log_info "安装 OCR Service Python 依赖..."
            cd services/ocr-service
            
            # 创建虚拟环境（如果不存在）
            if [[ ! -d "venv" ]]; then
                python3 -m venv venv
            fi
            
            # 激活虚拟环境并安装依赖
            source venv/bin/activate
            pip install --upgrade pip
            pip install -r requirements.txt
            deactivate
            
            cd ../..
            log_success "OCR Service Python 依赖安装完成"
        fi
    else
        log_warning "Python3 未安装，跳过 Python 依赖安装"
    fi
}

# 安装 Java 依赖
install_java_dependencies() {
    log_info "安装 Java 依赖..."
    
    if command_exists java && command_exists mvn; then
        # Admin Service
        if [[ -f "services/admin-service/backend/pom.xml" ]]; then
            log_info "构建 Admin Service..."
            cd services/admin-service/backend
            mvn clean install -DskipTests
            cd ../../..
            log_success "Admin Service 构建完成"
        fi
    else
        log_warning "Java 或 Maven 未安装，跳过 Java 依赖安装"
    fi
}

# 安装前端依赖
install_frontend_dependencies() {
    log_info "检查前端依赖..."
    
    if command_exists npm; then
        # 主前端
        if [[ -f "frontend/package.json" ]]; then
            if [[ ! -d "frontend/node_modules" ]] || [[ ! -f "frontend/package-lock.json" ]]; then
                log_info "安装前端依赖..."
                cd frontend
                npm install
                cd ..
                log_success "前端依赖安装完成"
            else
                log_success "前端依赖已安装"
            fi
        fi
        
        # Admin 前端
        if [[ -f "frontend-admin/package.json" ]]; then
            if [[ ! -d "frontend-admin/node_modules" ]]; then
                log_info "安装 Admin 前端依赖..."
                cd frontend-admin
                npm install
                cd ..
                log_success "Admin 前端依赖安装完成"
            fi
        fi
    else
        log_warning "npm 未安装，跳过前端依赖安装"
    fi
}

# 检查数据库服务
check_database_services() {
    log_info "检查数据库服务状态..."
    
    # PostgreSQL
    if command_exists psql; then
        if pgrep -x "postgres" > /dev/null; then
            log_success "PostgreSQL 正在运行"
        else
            log_warning "PostgreSQL 未运行，尝试启动..."
            if [[ "$OSTYPE" == "darwin"* ]]; then
                brew services start postgresql@15 || brew services start postgresql
            else
                sudo systemctl start postgresql || sudo service postgresql start
            fi
        fi
    else
        log_warning "PostgreSQL 未安装"
    fi
    
    # Redis
    if command_exists redis-cli; then
        if redis-cli ping > /dev/null 2>&1; then
            log_success "Redis 正在运行"
        else
            log_warning "Redis 未运行，尝试启动..."
            if [[ "$OSTYPE" == "darwin"* ]]; then
                brew services start redis
            else
                sudo systemctl start redis || sudo service redis start
            fi
        fi
    else
        log_warning "Redis 未安装"
    fi
}

# 生成依赖状态报告
generate_status_report() {
    log_info "生成依赖状态报告..."
    
    echo "# 依赖安装状态报告" > DEPENDENCY-STATUS.md
    echo "" >> DEPENDENCY-STATUS.md
    echo "生成时间: $(date '+%Y-%m-%d %H:%M:%S')" >> DEPENDENCY-STATUS.md
    echo "" >> DEPENDENCY-STATUS.md
    
    echo "## 系统级依赖" >> DEPENDENCY-STATUS.md
    echo "" >> DEPENDENCY-STATUS.md
    
    # 检查各种依赖
    for cmd in go node npm python3 java mvn psql redis-cli docker; do
        if command_exists $cmd; then
            version=$($cmd --version 2>&1 | head -n1 || echo "版本未知")
            echo "- ✅ $cmd: $version" >> DEPENDENCY-STATUS.md
        else
            echo "- ❌ $cmd: 未安装" >> DEPENDENCY-STATUS.md
        fi
    done
    
    echo "" >> DEPENDENCY-STATUS.md
    echo "## 服务依赖状态" >> DEPENDENCY-STATUS.md
    echo "" >> DEPENDENCY-STATUS.md
    
    # 检查各服务依赖
    [[ -d "frontend/node_modules" ]] && echo "- ✅ 前端依赖已安装" >> DEPENDENCY-STATUS.md || echo "- ❌ 前端依赖未安装" >> DEPENDENCY-STATUS.md
    [[ -f "backend/go.sum" ]] && echo "- ✅ 主后端 Go 依赖已配置" >> DEPENDENCY-STATUS.md || echo "- ❌ 主后端 Go 依赖未配置" >> DEPENDENCY-STATUS.md
    [[ -d "services/write-service/venv" ]] && echo "- ✅ Write Service Python 环境已创建" >> DEPENDENCY-STATUS.md || echo "- ❌ Write Service Python 环境未创建" >> DEPENDENCY-STATUS.md
    [[ -d "services/admin-service/backend/target" ]] && echo "- ✅ Admin Service 已构建" >> DEPENDENCY-STATUS.md || echo "- ❌ Admin Service 未构建" >> DEPENDENCY-STATUS.md
    
    log_success "依赖状态报告已生成: DEPENDENCY-STATUS.md"
}

# 主函数
main() {
    echo "=========================================="
    echo "   OpenPenPal 依赖安装脚本"
    echo "=========================================="
    echo ""
    
    # 检查是否在项目根目录
    if [[ ! -f "package.json" ]] || [[ ! -d "backend" ]]; then
        log_error "请在 OpenPenPal 项目根目录运行此脚本"
        exit 1
    fi
    
    # 询问安装模式
    echo "请选择安装模式:"
    echo "1) 完整安装 (推荐) - 安装所有依赖"
    echo "2) 快速安装 - 仅安装必要依赖"
    echo "3) 修复安装 - 修复已知问题"
    echo "4) 仅检查 - 不安装，仅生成状态报告"
    read -p "请输入选项 (1-4): " -n 1 -r
    echo ""
    
    case $REPLY in
        1)
            log_info "开始完整安装..."
            install_system_dependencies
            install_go_dependencies
            install_python_dependencies
            install_java_dependencies
            install_frontend_dependencies
            check_database_services
            ;;
        2)
            log_info "开始快速安装..."
            install_go_dependencies
            install_frontend_dependencies
            check_database_services
            ;;
        3)
            log_info "开始修复安装..."
            install_system_dependencies
            check_database_services
            ;;
        4)
            log_info "仅检查模式..."
            ;;
        *)
            log_error "无效选项"
            exit 1
            ;;
    esac
    
    # 生成状态报告
    generate_status_report
    
    echo ""
    echo "=========================================="
    echo "   依赖安装完成！"
    echo "=========================================="
    echo ""
    echo "下一步："
    echo "1. 查看 DEPENDENCY-STATUS.md 了解详细状态"
    echo "2. 运行 ./startup/quick-start.sh production 启动项目"
    echo ""
}

# 运行主函数
main "$@"