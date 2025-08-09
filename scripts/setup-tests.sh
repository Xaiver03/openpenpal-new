#!/bin/bash

# OpenPenPal 测试环境搭建脚本
# 该脚本会自动安装所需的测试依赖并创建测试结构

set -e

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🚀 OpenPenPal 测试环境搭建工具${NC}"
echo -e "${BLUE}================================${NC}"

# 检查当前目录
if [ ! -f "go.mod" ] && [ ! -f "frontend/package.json" ]; then
    echo -e "${RED}❌ 请在项目根目录运行此脚本${NC}"
    exit 1
fi

# 1. 后端测试环境搭建
echo -e "\n${YELLOW}📦 设置后端测试环境...${NC}"

if [ -f "go.mod" ]; then
    cd backend 2>/dev/null || cd .
    
    echo -e "${GREEN}✓ 安装 Go 测试依赖${NC}"
    go get -u github.com/stretchr/testify
    go get -u github.com/golang/mock/mockgen
    go get -u github.com/DATA-DOG/go-sqlmock
    
    # 创建测试目录结构
    echo -e "${GREEN}✓ 创建测试目录结构${NC}"
    mkdir -p internal/{mocks,testutils,testdata}
    mkdir -p test/{integration,fixtures}
    
    # 创建 Makefile
    if [ ! -f "Makefile" ]; then
        echo -e "${GREEN}✓ 创建 Makefile${NC}"
        cat > Makefile << 'EOF'
.PHONY: test test-coverage test-unit test-integration mock clean

# 运行所有测试
test:
	@echo "🧪 运行所有测试..."
	@go test -v -race ./...

# 运行测试并生成覆盖率报告
test-coverage:
	@echo "📊 生成测试覆盖率报告..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ 覆盖率报告已生成: coverage.html"

# 只运行单元测试
test-unit:
	@echo "🧪 运行单元测试..."
	@go test -v -short ./...

# 运行集成测试
test-integration:
	@echo "🔗 运行集成测试..."
	@go test -v -run Integration ./...

# 生成 mocks
mock:
	@echo "🤖 生成 Mock 文件..."
	@mockgen -source=internal/services/auth_service.go -destination=internal/mocks/mock_auth_service.go -package=mocks
	@mockgen -source=internal/services/letter_service.go -destination=internal/mocks/mock_letter_service.go -package=mocks
	@mockgen -source=internal/services/courier_service.go -destination=internal/mocks/mock_courier_service.go -package=mocks
	@echo "✅ Mock 文件生成完成"

# 清理测试文件
clean:
	@echo "🧹 清理测试文件..."
	@rm -f coverage.out coverage.html
	@rm -rf internal/mocks/*
	@echo "✅ 清理完成"

# 运行特定的测试
test-service:
	@echo "🧪 运行服务层测试..."
	@go test -v ./internal/services/...

test-handler:
	@echo "🧪 运行处理器测试..."
	@go test -v ./internal/handlers/...

# 基准测试
bench:
	@echo "⚡ 运行基准测试..."
	@go test -bench=. -benchmem ./...

# 检查测试覆盖率是否达标
check-coverage:
	@echo "📊 检查测试覆盖率..."
	@go test -coverprofile=coverage.out ./... > /dev/null 2>&1
	@go tool cover -func=coverage.out | grep total | awk '{print "当前覆盖率: " $$3}'
	@rm coverage.out
EOF
    fi
    
    cd - > /dev/null
fi

# 2. 前端测试环境搭建
echo -e "\n${YELLOW}📦 设置前端测试环境...${NC}"

if [ -d "frontend" ]; then
    cd frontend
    
    echo -e "${GREEN}✓ 检查测试依赖${NC}"
    npm list @testing-library/react @testing-library/jest-dom jest 2>/dev/null || {
        echo -e "${YELLOW}📦 安装缺失的测试依赖...${NC}"
        npm install --save-dev @testing-library/react@latest
        npm install --save-dev @testing-library/user-event@latest
        npm install --save-dev @testing-library/jest-dom@latest
        npm install --save-dev @types/jest@latest
    }
    
    # 创建测试目录结构
    echo -e "${GREEN}✓ 创建测试目录结构${NC}"
    mkdir -p src/{components,hooks,stores,lib}/__tests__
    mkdir -p tests/{unit,integration,fixtures,utils}
    
    # 创建测试配置文件
    if [ ! -f "jest.setup.js" ]; then
        echo -e "${GREEN}✓ 创建 jest.setup.js${NC}"
        cat > jest.setup.js << 'EOF'
import '@testing-library/jest-dom'

// Mock Next.js router
jest.mock('next/navigation', () => ({
  useRouter() {
    return {
      push: jest.fn(),
      replace: jest.fn(),
      prefetch: jest.fn(),
      back: jest.fn(),
    }
  },
  usePathname() {
    return '/'
  },
  useSearchParams() {
    return new URLSearchParams()
  },
}))

// Mock window.matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: jest.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: jest.fn(),
    removeListener: jest.fn(),
    addEventListener: jest.fn(),
    removeEventListener: jest.fn(),
    dispatchEvent: jest.fn(),
  })),
})

// 全局测试工具
global.fetch = jest.fn()

// 清理
afterEach(() => {
  jest.clearAllMocks()
})
EOF
    fi
    
    # 添加测试脚本到 package.json
    echo -e "${GREEN}✓ 更新 package.json 测试脚本${NC}"
    node -e "
    const fs = require('fs');
    const pkg = JSON.parse(fs.readFileSync('package.json', 'utf8'));
    
    pkg.scripts = {
      ...pkg.scripts,
      'test': 'jest',
      'test:watch': 'jest --watch',
      'test:coverage': 'jest --coverage',
      'test:ci': 'jest --ci --coverage --maxWorkers=2',
      'test:debug': 'node --inspect-brk ./node_modules/.bin/jest --runInBand',
      'test:unit': 'jest --testPathPattern=unit',
      'test:integration': 'jest --testPathPattern=integration',
    };
    
    fs.writeFileSync('package.json', JSON.stringify(pkg, null, 2));
    "
    
    cd ..
fi

# 3. 创建通用测试脚本
echo -e "\n${YELLOW}📝 创建测试运行脚本...${NC}"

cat > run-tests.sh << 'EOF'
#!/bin/bash

# OpenPenPal 测试运行器

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

MODE=${1:-all}

case $MODE in
  backend)
    echo -e "${YELLOW}🧪 运行后端测试...${NC}"
    cd backend && make test-coverage
    ;;
  
  frontend)
    echo -e "${YELLOW}🧪 运行前端测试...${NC}"
    cd frontend && npm run test:coverage
    ;;
  
  e2e)
    echo -e "${YELLOW}🧪 运行 E2E 测试...${NC}"
    cd frontend && npm run test:e2e
    ;;
  
  all)
    echo -e "${YELLOW}🧪 运行所有测试...${NC}"
    
    echo -e "\n${GREEN}后端测试：${NC}"
    cd backend && make test
    
    echo -e "\n${GREEN}前端测试：${NC}"
    cd ../frontend && npm test
    
    echo -e "\n${GREEN}✅ 所有测试完成！${NC}"
    ;;
  
  watch)
    echo -e "${YELLOW}👀 监视模式...${NC}"
    echo "1) 后端测试监视"
    echo "2) 前端测试监视"
    read -p "选择 (1/2): " choice
    
    case $choice in
      1) cd backend && watch -n 2 go test ./... ;;
      2) cd frontend && npm run test:watch ;;
    esac
    ;;
  
  *)
    echo "用法: $0 [backend|frontend|e2e|all|watch]"
    exit 1
    ;;
esac
EOF

chmod +x run-tests.sh

# 4. 创建测试数据生成器
echo -e "\n${YELLOW}🏭 创建测试数据生成器...${NC}"

mkdir -p scripts/test-data

cat > scripts/test-data/generate-test-data.js << 'EOF'
#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

// 生成测试用户数据
function generateTestUsers(count = 10) {
  const roles = ['user', 'courier_level1', 'courier_level2', 'courier_level3', 'courier_level4', 'admin'];
  const users = [];
  
  for (let i = 0; i < count; i++) {
    users.push({
      id: `test-user-${i}`,
      username: `testuser${i}`,
      email: `test${i}@example.com`,
      password: 'Test123!',
      nickname: `测试用户${i}`,
      role: roles[i % roles.length],
      schoolCode: 'BJDX',
      isActive: true,
    });
  }
  
  return users;
}

// 生成测试信件数据
function generateTestLetters(count = 20) {
  const letters = [];
  const statuses = ['draft', 'generated', 'collected', 'in_transit', 'delivered'];
  
  for (let i = 0; i < count; i++) {
    letters.push({
      id: `test-letter-${i}`,
      title: `测试信件 ${i}`,
      content: `这是第 ${i} 封测试信件的内容...`,
      status: statuses[i % statuses.length],
      senderOPCode: `PK5F${String(i % 10).padStart(2, '0')}`,
      recipientOPCode: `PK3D${String(i % 10).padStart(2, '0')}`,
      createdAt: new Date(Date.now() - i * 86400000).toISOString(),
    });
  }
  
  return letters;
}

// 生成测试任务数据
function generateTestTasks(count = 15) {
  const tasks = [];
  const priorities = ['normal', 'urgent'];
  const statuses = ['pending', 'accepted', 'collected', 'in_transit', 'delivered'];
  
  for (let i = 0; i < count; i++) {
    tasks.push({
      id: `test-task-${i}`,
      letterCode: `LC${String(100000 + i).padStart(6, '0')}`,
      title: `投递任务 ${i}`,
      priority: priorities[i % priorities.length],
      status: statuses[i % statuses.length],
      pickupOPCode: `PK5F${String(i % 10).padStart(2, '0')}`,
      deliveryOPCode: `PK3D${String(i % 10).padStart(2, '0')}`,
      reward: 10 + (i % 5) * 5,
    });
  }
  
  return tasks;
}

// 保存测试数据
const testData = {
  users: generateTestUsers(),
  letters: generateTestLetters(),
  tasks: generateTestTasks(),
  timestamp: new Date().toISOString(),
};

// 输出到文件
const outputDir = path.join(__dirname, '../../test-data');
if (!fs.existsSync(outputDir)) {
  fs.mkdirSync(outputDir, { recursive: true });
}

fs.writeFileSync(
  path.join(outputDir, 'test-data.json'),
  JSON.stringify(testData, null, 2)
);

console.log('✅ 测试数据已生成：test-data/test-data.json');
console.log(`📊 生成了 ${testData.users.length} 个用户，${testData.letters.length} 封信件，${testData.tasks.length} 个任务`);
EOF

chmod +x scripts/test-data/generate-test-data.js

# 5. 创建 CI 配置
echo -e "\n${YELLOW}🔧 创建 CI 配置...${NC}"

mkdir -p .github/workflows

if [ ! -f ".github/workflows/test.yml" ]; then
    cat > .github/workflows/test.yml << 'EOF'
name: Tests

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    
    strategy:
      matrix:
        node-version: [18.x]
        go-version: [1.21.x]
    
    services:
      postgres:
        image: postgres:14
        env:
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: openpenpal_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Setup Go ${{ matrix.go-version }}
      uses: actions/setup-go@v4
      with:
        go-version: ${{ matrix.go-version }}
    
    - name: Setup Node.js ${{ matrix.node-version }}
      uses: actions/setup-node@v3
      with:
        node-version: ${{ matrix.node-version }}
    
    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-
    
    - name: Cache npm dependencies
      uses: actions/cache@v3
      with:
        path: ~/.npm
        key: ${{ runner.os }}-node-${{ hashFiles('**/package-lock.json') }}
        restore-keys: |
          ${{ runner.os }}-node-
    
    - name: Install backend dependencies
      run: |
        cd backend
        go mod download
    
    - name: Install frontend dependencies
      run: |
        cd frontend
        npm ci
    
    - name: Run backend tests
      env:
        DATABASE_URL: postgres://postgres:postgres@localhost:5432/openpenpal_test?sslmode=disable
      run: |
        cd backend
        make test-coverage
    
    - name: Run frontend tests
      run: |
        cd frontend
        npm run test:ci
    
    - name: Upload backend coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./backend/coverage.out
        flags: backend
        name: backend-coverage
    
    - name: Upload frontend coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./frontend/coverage/lcov.info
        flags: frontend
        name: frontend-coverage
    
    - name: Run E2E tests
      run: |
        cd frontend
        npx playwright install --with-deps chromium
        npm run test:e2e
EOF
fi

# 6. 创建测试报告生成器
echo -e "\n${YELLOW}📊 创建测试报告生成器...${NC}"

cat > generate-test-report.sh << 'EOF'
#!/bin/bash

# 测试报告生成器

set -e

echo "📊 生成测试报告..."

# 创建报告目录
mkdir -p test-reports

# 生成时间戳
TIMESTAMP=$(date +"%Y-%m-%d_%H-%M-%S")
REPORT_FILE="test-reports/test-report-${TIMESTAMP}.html"

# 开始生成 HTML 报告
cat > $REPORT_FILE << 'HTML_START'
<!DOCTYPE html>
<html>
<head>
    <title>OpenPenPal 测试报告</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1, h2 { color: #333; }
        .summary { background: #f0f0f0; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .pass { color: green; }
        .fail { color: red; }
        .coverage { margin: 20px 0; }
        .coverage-bar { width: 300px; height: 20px; background: #ddd; border-radius: 10px; overflow: hidden; }
        .coverage-fill { height: 100%; background: #4CAF50; }
        table { border-collapse: collapse; width: 100%; margin: 20px 0; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background: #f0f0f0; }
    </style>
</head>
<body>
    <h1>OpenPenPal 测试报告</h1>
    <p>生成时间: <script>document.write(new Date().toLocaleString())</script></p>
HTML_START

# 收集后端测试结果
if [ -d "backend" ]; then
    echo "<h2>后端测试结果</h2>" >> $REPORT_FILE
    echo "<div class='summary'>" >> $REPORT_FILE
    cd backend
    go test ./... -json | go-test-report >> ../$REPORT_FILE 2>/dev/null || echo "<p>后端测试数据暂无</p>" >> ../$REPORT_FILE
    cd ..
    echo "</div>" >> $REPORT_FILE
fi

# 收集前端测试结果
if [ -d "frontend" ]; then
    echo "<h2>前端测试结果</h2>" >> $REPORT_FILE
    echo "<div class='summary'>" >> $REPORT_FILE
    if [ -f "frontend/coverage/coverage-summary.json" ]; then
        node -e "
        const coverage = require('./frontend/coverage/coverage-summary.json');
        const total = coverage.total;
        console.log('<table>');
        console.log('<tr><th>类型</th><th>覆盖率</th><th>覆盖/总数</th></tr>');
        ['lines', 'statements', 'functions', 'branches'].forEach(type => {
            const data = total[type];
            const pct = data.pct;
            const color = pct >= 80 ? 'pass' : pct >= 60 ? 'warning' : 'fail';
            console.log(\`<tr><td>\${type}</td><td class='\${color}'>\${pct}%</td><td>\${data.covered}/\${data.total}</td></tr>\`);
        });
        console.log('</table>');
        " >> $REPORT_FILE
    else
        echo "<p>前端测试数据暂无</p>" >> $REPORT_FILE
    fi
    echo "</div>" >> $REPORT_FILE
fi

# 结束 HTML
cat >> $REPORT_FILE << 'HTML_END'
    <h2>测试建议</h2>
    <ul>
        <li>确保所有关键路径都有测试覆盖</li>
        <li>为新功能编写测试用例</li>
        <li>定期运行测试确保代码质量</li>
        <li>目标：80% 以上的测试覆盖率</li>
    </ul>
</body>
</html>
HTML_END

echo "✅ 测试报告已生成: $REPORT_FILE"
open $REPORT_FILE 2>/dev/null || echo "请手动打开报告文件查看"
EOF

chmod +x generate-test-report.sh

# 完成
echo -e "\n${GREEN}✅ 测试环境搭建完成！${NC}"
echo -e "\n${BLUE}可用命令：${NC}"
echo -e "  ${GREEN}./run-tests.sh${NC}         - 运行所有测试"
echo -e "  ${GREEN}./run-tests.sh backend${NC} - 只运行后端测试"
echo -e "  ${GREEN}./run-tests.sh frontend${NC}- 只运行前端测试"
echo -e "  ${GREEN}./run-tests.sh e2e${NC}     - 运行 E2E 测试"
echo -e "  ${GREEN}./run-tests.sh watch${NC}   - 监视模式"
echo -e "  ${GREEN}./generate-test-report.sh${NC} - 生成测试报告"
echo -e "\n${BLUE}后端测试命令（在 backend 目录下）：${NC}"
echo -e "  ${GREEN}make test${NC}          - 运行所有测试"
echo -e "  ${GREEN}make test-coverage${NC} - 生成覆盖率报告"
echo -e "  ${GREEN}make mock${NC}          - 生成 Mock 文件"
echo -e "\n${BLUE}前端测试命令（在 frontend 目录下）：${NC}"
echo -e "  ${GREEN}npm test${NC}           - 运行测试"
echo -e "  ${GREEN}npm run test:watch${NC} - 监视模式"
echo -e "  ${GREEN}npm run test:coverage${NC} - 生成覆盖率报告"

echo -e "\n${YELLOW}📚 下一步：${NC}"
echo -e "1. 运行 ${GREEN}./run-tests.sh${NC} 验证测试环境"
echo -e "2. 查看 ${GREEN}docs/TESTING_GUIDE.md${NC} 了解详细测试指南"
echo -e "3. 开始编写测试用例提升覆盖率"