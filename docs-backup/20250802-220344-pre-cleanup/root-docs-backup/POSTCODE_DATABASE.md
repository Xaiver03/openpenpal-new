# OpenPenPal Postcode数据库集成

Postcode编码系统现在支持PostgreSQL数据库持久化存储，确保测试数据稳定性和生产环境可靠性。

## 🚀 快速开始

### 1. 初始化数据库

```bash
# 使用默认配置初始化
./scripts/init-postcode-db.sh

# 或指定自定义数据库配置  
./scripts/init-postcode-db.sh --host mydb.com --user admin --database openpenpal_test
```

### 2. 启动服务

```bash
# 自动检测数据库可用性并启动相应模式
./scripts/start-with-db.sh

# 强制使用数据库模式
./scripts/start-with-db.sh --db-only

# 强制使用Mock模式（不依赖数据库）
./scripts/start-with-db.sh --mock-only

# 初始化数据库并运行测试
./scripts/start-with-db.sh --init-db --test
```

### 3. 验证集成

```bash
# 运行完整的API集成测试
python3 scripts/test-postcode-db.py
```

## 📊 数据库架构

### 表结构
- `postcode_schools` - 学校站点 (2位编码)
- `postcode_areas` - 片区 (1位编码)  
- `postcode_buildings` - 楼栋 (1位编码)
- `postcode_rooms` - 房间 (2位编码，自动生成6位完整编码)
- `postcode_courier_permissions` - 信使权限管理
- `postcode_feedbacks` - 地址反馈系统
- `postcode_stats` - 使用统计分析

### 层次关系
```
学校(PK) → 片区(A) → 楼栋(1) → 房间(01) = PKA101
学校(TH) → 片区(A) → 楼栋(1) → 房间(02) = THA102
```

## 🔧 配置选项

### 环境变量
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=openpenpal
export DB_USER=postgres
export DB_PASSWORD=password
```

### 数据库初始化选项
```bash
# 仅创建表结构
./scripts/init-postcode-db.sh --tables-only

# 仅插入测试数据
./scripts/init-postcode-db.sh --data-only

# 显示帮助
./scripts/init-postcode-db.sh --help
```

## 🧪 测试数据

### 预置学校
- **PK** - 北京大学
- **TH** - 清华大学  
- **BJ** - 北京师范大学
- **RD** - 中国人民大学

### 测试账号
- `courier1/courier123` - 一级信使 (楼栋管理权限 PKA1**)
- `courier2/courier123` - 二级信使 (片区管理权限 PKA*)
- `courier3/courier123` - 三级信使 (学校管理权限 PK*)
- `courier4/courier123` - 四级信使 (全局管理权限 **)

### 示例编码
- `PKA101` - 北京大学东区1栋101室
- `PKA102` - 北京大学东区1栋102室
- `THA101` - 清华大学紫荆区1栋101室
- `THA102` - 清华大学紫荆区1栋102室

## 📡 API端点

### 核心查询
- `GET /api/v1/postcode/{code}` - 根据6位编码查询地址
- `GET /api/v1/address/search?query={keyword}` - 模糊搜索地址

### 层次管理
- `GET /api/v1/postcode/schools` - 获取学校列表
- `GET /api/v1/postcode/schools/{school}/areas` - 获取片区列表
- `GET /api/v1/postcode/schools/{school}/areas/{area}/buildings` - 获取楼栋列表
- `GET /api/v1/postcode/schools/{school}/areas/{area}/buildings/{building}/rooms` - 获取房间列表

### 权限与统计
- `GET /api/v1/postcode/permissions/{courier_id}` - 查询信使权限
- `GET /api/v1/postcode/stats/popular` - 获取热门地址统计
- `POST /api/v1/postcode/validate` - 批量验证编码有效性

## 🔄 运行模式

### 数据库模式
- ✅ 数据持久化到PostgreSQL
- ✅ 完整的CRUD操作支持
- ✅ 复杂查询和统计分析
- ✅ 多用户并发安全
- ⚡ 需要PostgreSQL服务

### Mock模式  
- ✅ 内存中的模拟数据
- ✅ 快速启动，无依赖
- ✅ 开发和演示友好
- ⚠️ 数据不持久化
- ⚠️ 功能有限

## 🔍 故障排除

### 数据库连接问题
```bash
# 检查PostgreSQL服务状态
brew services list | grep postgresql

# 启动PostgreSQL
brew services start postgresql

# 测试连接
psql postgresql://postgres:password@localhost:5432/postgres
```

### 权限问题
```bash
# 确保数据库用户有足够权限
createdb -O postgres openpenpal
psql -d openpenpal -c "GRANT ALL PRIVILEGES ON DATABASE openpenpal TO postgres;"
```

### 端口冲突
```bash
# 检查端口占用
lsof -i :8001
lsof -i :3000

# 终止占用进程
pkill -f "uvicorn"
pkill -f "npm run dev"
```

## 📈 性能优化

### 数据库索引
主要索引已自动创建：
- `postcode_rooms.full_postcode` - 6位编码快速查询
- `postcode_rooms.school_code, area_code, building_code` - 层次查询
- `postcode_stats.popularity_score` - 热门度排序

### 查询优化
- 使用完整的6位编码查询最快
- 模糊搜索限制结果数量 (`limit`参数)
- 统计查询使用适当的时间范围

## 🛠️ 开发指南

### 添加新学校
```sql
INSERT INTO postcode_schools (id, code, name, full_name, status) VALUES
(gen_random_uuid(), 'XY', '新学校', '新学校全名', 'active');
```

### 扩展权限模式
```sql
-- 为新信使添加权限
INSERT INTO postcode_courier_permissions (id, courier_id, level, prefix_patterns, can_manage, can_create, can_review) VALUES
(gen_random_uuid(), 'new_courier', 2, ARRAY['XY*'], true, true, false);
```

### 自定义统计
```sql
-- 查询最活跃的学校
SELECT 
    s.name,
    COUNT(st.postcode) as active_addresses,
    SUM(st.delivery_count) as total_deliveries
FROM postcode_schools s
JOIN postcode_rooms r ON r.school_code = s.code
JOIN postcode_stats st ON st.postcode = r.full_postcode
GROUP BY s.code, s.name
ORDER BY total_deliveries DESC;
```

## 📝 更新日志

### v2.0.0 - 数据库集成
- ✅ 完整PostgreSQL数据库支持
- ✅ 自动化数据库初始化脚本
- ✅ 双模式启动支持（数据库/Mock）
- ✅ 完整的集成测试套件
- ✅ 生产级权限和统计系统

### v1.0.0 - Mock服务
- ✅ 内存模拟数据服务
- ✅ 基础API端点实现
- ✅ 前端界面集成

---

**需要帮助？** 查看 [故障排除](#-故障排除) 部分或运行 `./scripts/start-with-db.sh --help`