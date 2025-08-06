# OpenPenPal 测试脚本使用说明

## 快速开始

### 1. 环境准备
```bash
# 安装依赖
brew install jq curl

# 确保后端服务运行
make run-backend  # 或手动启动localhost:8080
```

### 2. 执行测试
```bash
# 运行完整测试
./test_scripts/appointment_test.sh

# 运行特定测试模块
./test_scripts/appointment_test.sh --role-hierarchy
./test_scripts/appointment_test.sh --login-tests
```

### 3. 测试覆盖范围

#### 功能测试
- ✅ 角色层级验证
- ✅ 任命权限逻辑
- ✅ 用户注册角色固定
- ✅ 登录认证
- ✅ 学校代码验证
- ✅ API端点可用性

#### 权限测试矩阵
```
测试场景：
1. 超级管理员 → 四级协调员任命
2. 四级协调员 → 三级高级信使任命  
3. 三级高级信使 → 二级普通信使任命
4. 二级普通信使 → 一级用户管理
```

## 测试数据

### 预设测试账号
```
# 普通用户
student001@penpal.com / student001

# 各级信使（初始均为user角色）
courier_building@penpal.com / courier001
courier_area@penpal.com / courier002  
courier_school@penpal.com / courier003
courier_city@penpal.com / courier004

# 管理员
admin@penpal.com / admin123
```

### 学校代码
```
有效代码: PKU001-PKU006
无效代码: PKU, PKU0001, 123, TEST123
```

## 测试报告

### 输出文件
- `test_report_YYYYMMDD_HHMMSS.json` - 详细测试报告
- `test_logs/` - 测试日志目录
- `screenshots/` - 测试截图（如适用）

### 报告格式
```json
{
  "test_date": "2024-07-21T14:30:00Z",
  "test_environment": {
    "api_base": "http://localhost:8080",
    "test_accounts": 6,
    "test_cases": 15
  },
  "results": {
    "passed": 14,
    "failed": 1,
    "skipped": 0
  }
}
```

## 故障排除

### 常见问题
1. **jq未安装**: `brew install jq`
2. **端口冲突**: 检查8080端口占用
3. **数据库连接**: 确保PostgreSQL运行正常
4. **测试账号不存在**: 先运行注册测试

### 调试模式
```bash
# 详细输出
DEBUG=1 ./test_scripts/appointment_test.sh

# 仅测试特定功能
./test_scripts/appointment_test.sh --test login_only
```

## 扩展测试

### 添加新测试用例
1. 编辑 `appointment_test.sh`
2. 添加测试函数
3. 在main函数中调用

### 测试数据更新
修改脚本顶部变量或创建新的测试数据文件

---
**维护**: Kimi AI Tester  
**更新**: 2024-07-21