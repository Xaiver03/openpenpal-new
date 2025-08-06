# OpenPenPal Mock数据说明

本目录包含了OpenPenPal系统的所有Mock测试数据，用于开发和测试环境。

## 数据文件说明

### 1. Postcode系统数据 (`postcode_test_data.sql`)
- **学校数据**: 4所大学的基础信息（北京大学、清华大学等）
- **片区数据**: 各学校的区域划分（东区、西区、南区、北区）
- **楼栋数据**: 各片区的建筑信息（1栋、2栋、教学楼等）
- **房间数据**: 具体的房间地址编码（PKA101、THB201等）
- **权限数据**: 信使权限配置和管理范围
- **反馈数据**: 地址反馈和错误报告
- **统计数据**: 使用频率和热门度数据

### 2. 信使管理数据 (`courier_management_data.sql`)
包含四个层级的信使管理数据：

#### 城市级信使 (Level 4 - 管理城市)
- 北京市总管理员：`beijing_city_manager`
- 上海市总管理员：`shanghai_city_manager`
- 广州市总管理员：`guangzhou_city_manager`
- 深圳市总管理员：`shenzhen_city_manager`

#### 学校级信使 (Level 3 - 管理学校)
- 北京大学校级信使：`university_peking_manager`
- 清华大学校级信使：`university_tsinghua_manager`
- 中国人民大学校级信使：`university_renda_manager`
- 北京师范大学校级信使：`university_beishi_manager`

#### 片区级信使 (Level 2 - 管理片区)
- 东区片区信使：`zone_a_manager`
- 西区片区信使：`zone_b_manager`
- 南区片区信使：`zone_c_manager`
- 北区片区信使：`zone_d_manager`

#### 楼栋级信使 (Level 1 - 管理楼栋)
- A栋楼栋信使：`building_a_courier`
- B栋楼栋信使：`building_b_courier`
- C栋楼栋信使：`building_c_courier`
- D栋楼栋信使：`building_d_courier`

#### 统计数据
- 各级别的统计信息：总数量、活跃数量、配送数、待处理任务等
- 性能指标：平均评分、成功率、覆盖率等

### 3. 博物馆信件数据 (`museum_letters_data.sql`)
#### 展览数据
- 冬日温暖信件展：`mock_exhibition_winter`
- 友谊永恒主题展：`mock_exhibition_friendship`
- 致未来的自己：`mock_exhibition_future`
- 校园时光记忆展：`mock_exhibition_campus`（筹备中）

#### 信件数据
每个展览包含3-4封精选信件：
- **冬日温暖主题**: 雪夜温暖、热茶友情、图书馆约定
- **友谊主题**: 生日惊喜、深夜谈心、食堂默契
- **未来主题**: 二十年后的自己、毕业十年后、给未来的妈妈
- **校园主题**: 初入校园、期末疯狂、社团收获

### 4. 广场公开信件数据 (`plaza_public_letters_data.sql`)
按风格分类的公开信件：

#### 未来风格 (future)
- 写给三年后的自己
- 关于梦想这件小事
- 青春就是现在

#### 温暖风格 (warm)
- 致正在迷茫的你
- 食堂里的小确幸
- 给帮助过我的陌生人

#### 故事风格 (story)
- 一个关于友谊的故事
- 那些让我成长的错误

#### 漂流风格 (drift)
- 漂流到远方的思念
- 关于孤独的思考

## 用户认证数据

### 测试账号
- **学生用户**: `alice/secret`, `bob/password123`
- **管理员**: `admin/admin123`
- **信使用户**: `courier1/courier123` (Level 1), `courier2/courier123` (Level 2), `courier3/courier123` (Level 3), `courier4/courier123` (Level 4)

## 数据导入方法

### 方法1: 使用导入脚本（推荐）
```bash
# 导入所有数据
./scripts/import-all-mock-data.sh

# 自定义数据库配置
./scripts/import-all-mock-data.sh --host localhost --user postgres --password mypass

# 跳过某些数据类型
./scripts/import-all-mock-data.sh --skip-postcode --skip-museum

# 查看将要执行的操作（不实际执行）
./scripts/import-all-mock-data.sh --dry-run
```

### 方法2: 手动导入
```bash
# 使用现有的Postcode初始化脚本
./scripts/init-postcode-db.sh

# 手动导入其他数据
psql -U postgres -d openpenpal -f services/database/seed_data/courier_management_data.sql
psql -U postgres -d openpenpal -f services/database/seed_data/museum_letters_data.sql
psql -U postgres -d openpenpal -f services/database/seed_data/plaza_public_letters_data.sql
```

### 方法3: 使用Mock服务
Mock服务已经包含了所有这些数据的API接口版本：
```bash
# 启动Mock服务
node scripts/simple-mock-services.js

# API endpoints:
# - http://localhost:8000/api/v1/auth/login
# - http://localhost:8000/api/v1/postcode/schools
# - http://localhost:8000/api/v1/letters/public
# - 等等...
```

## 数据特点

### 完整性
- 涵盖系统所有核心功能模块
- 包含不同状态的测试数据（active, pending, frozen）
- 提供完整的关联关系数据

### 真实性
- 使用真实的学校名称和地址结构
- 包含合理的统计数据和时间戳
- 模拟真实的用户行为和内容

### 多样性
- 不同层级的信使数据
- 多种风格的信件内容
- 各种状态和类型的测试场景

### 可扩展性
- 模块化的数据文件结构
- 可以选择性导入特定模块
- 方便添加新的测试数据

## 注意事项

1. **数据前缀**: 所有Mock数据都使用`mock_`前缀，便于识别和清理
2. **密码安全**: 测试账号使用简单密码，仅用于开发环境
3. **数据一致性**: 各模块数据之间保持关联关系的一致性
4. **索引优化**: 已添加必要的数据库索引以提高查询性能

## 清理Mock数据

如需清理所有Mock数据：
```sql
-- 清理所有Mock数据
DELETE FROM couriers WHERE id LIKE 'mock_%';
DELETE FROM users WHERE id LIKE 'mock_%';
DELETE FROM letters WHERE id LIKE 'mock_%';
DELETE FROM postcode_schools WHERE id LIKE 'mock_%' OR id LIKE '550e8400%';
DELETE FROM museum_letters WHERE id LIKE 'mock_%';
DELETE FROM public_letters WHERE id LIKE 'mock_plaza_%';
-- 等等...
```

## 更新日志

- **2024-01-25**: 创建完整的Mock数据体系
- **2024-01-25**: 添加信使管理四级数据
- **2024-01-25**: 增加博物馆和广场数据
- **2024-01-25**: 创建统一的导入脚本