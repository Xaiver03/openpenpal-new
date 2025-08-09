# 🚀 OpenPenPal 测试快速开始指南

本指南帮助你快速开始为 OpenPenPal 项目编写测试。

## 📋 前置准备

1. **运行测试环境搭建脚本**
```bash
cd /path/to/openpenpal
./scripts/setup-tests.sh
```

2. **验证环境**
```bash
./run-tests.sh
```

## 🎯 第一周测试任务清单

### Day 1-2: 认证系统测试
- [ ] 完成 `auth_service_test.go`
- [ ] 完成 `auth_handler_test.go`
- [ ] 完成前端 `useAuth.test.ts`
- [ ] 完成 `LoginForm.test.tsx`

### Day 3-4: 信件核心功能测试
- [ ] 完成 `letter_service_test.go`
- [ ] 完成 `letter_handler_test.go`
- [ ] 完成 `LetterEditor.test.tsx`
- [ ] 完成 `LetterList.test.tsx`

### Day 5-7: 信使系统测试
- [ ] 完成 `courier_service_test.go`
- [ ] 完成 `courier_handler_test.go`
- [ ] 完成 `CourierDashboard.test.tsx`
- [ ] 完成 E2E 测试场景

## 📝 编写测试的步骤

### 1. 后端服务测试模板

```go
// 文件：backend/internal/services/xxx_service_test.go

package services

import (
    "testing"
    "github.com/stretchr/testify/suite"
)

type XxxServiceTestSuite struct {
    suite.Suite
    service *XxxService
    db      *gorm.DB
}

func (suite *XxxServiceTestSuite) SetupSuite() {
    // 初始化测试环境
}

func (suite *XxxServiceTestSuite) TestMethodName() {
    // 准备
    // 执行
    // 断言
}

func TestXxxServiceSuite(t *testing.T) {
    suite.Run(t, new(XxxServiceTestSuite))
}
```

### 2. 前端组件测试模板

```typescript
// 文件：frontend/src/components/xxx/Xxx.test.tsx

import { render, screen, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Xxx } from './Xxx'

describe('Xxx Component', () => {
  it('should render correctly', () => {
    render(<Xxx />)
    expect(screen.getByText('Expected Text')).toBeInTheDocument()
  })
  
  it('should handle user interaction', async () => {
    const user = userEvent.setup()
    render(<Xxx />)
    
    await user.click(screen.getByRole('button'))
    // 断言结果
  })
})
```

## 🔧 常用测试命令

### 后端测试
```bash
cd backend

# 运行特定测试文件
go test -v ./internal/services/auth_service_test.go

# 运行特定测试函数
go test -v -run TestAuthService_Login ./internal/services

# 查看覆盖率
go test -cover ./internal/services/...

# 生成覆盖率报告
make test-coverage
```

### 前端测试
```bash
cd frontend

# 运行特定测试文件
npm test -- auth.test.ts

# 监视模式
npm run test:watch

# 调试测试
npm run test:debug

# 覆盖率报告
npm run test:coverage
```

## 📊 测试覆盖率目标

| 模块 | 第1周目标 | 第2周目标 | 最终目标 |
|------|----------|----------|---------|
| 认证系统 | 80% | 90% | 95% |
| 信件管理 | 60% | 80% | 90% |
| 信使系统 | 50% | 70% | 85% |
| 其他功能 | 20% | 50% | 80% |

## ✅ 测试检查清单

每个测试文件应包含：
- [ ] 正常路径测试（Happy Path）
- [ ] 边界条件测试
- [ ] 错误处理测试
- [ ] 并发测试（如适用）
- [ ] 性能测试（如适用）

## 🚨 常见问题解决

### 1. Mock 生成失败
```bash
# 确保安装了 mockgen
go install github.com/golang/mock/mockgen@latest

# 手动生成 mock
mockgen -source=internal/services/auth_service.go -destination=internal/mocks/mock_auth_service.go
```

### 2. 前端测试找不到模块
```bash
# 清理缓存
npm cache clean --force

# 重新安装依赖
rm -rf node_modules package-lock.json
npm install
```

### 3. 数据库连接错误
```go
// 使用内存数据库进行测试
db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
```

## 📚 推荐阅读

1. [Go 测试最佳实践](https://github.com/golang/go/wiki/TestComments)
2. [React Testing Library 文档](https://testing-library.com/docs/react-testing-library/intro/)
3. [Jest 文档](https://jestjs.io/docs/getting-started)

## 🎯 本周挑战

尝试在本周内：
1. 为至少 10 个核心函数编写测试
2. 达到 30% 的总体测试覆盖率
3. 修复所有测试中发现的 bug
4. 建立 CI/CD 测试流程

加油！让我们一起提升代码质量！ 💪