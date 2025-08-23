# OpenPenPal 权限系统数据库修复方案

## 🎯 核心问题

当前权限系统与FSD规范严重不符，存在以下关键问题：

2. **权限控制机制缺失**: 无法实现基于编码段的地理权限控制
3. **RBAC系统不完整**: 缺少角色-权限映射表
4. **审批流程缺失**: 没有信使升级审批机制

## 📋 FSD权限系统要求对比

### FSD要求的角色体系
```
user → messenger1 → messenger2 → messenger3 → messenger4 → admin
```

### FSD要求的权限矩阵
| 功能/资源 | user | M1 | M2 | M3 | M4 | admin |
|-----------|------|----|----|----|----|-------|
| 查看公开信件 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| 扫码执行任务 | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ |
| 审批点位/编码 | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ |
| 信使权限管理 | ❌ | ❌ | ❌ | ✅ | ✅ | ✅ |
| 城市级开通学校 | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ |

### FSD要求的编码段权限控制
| 信使等级 | 可操作编码段 | 权限说明 |
|----------|-------------|----------|
| messenger1 | PK5F3D (完整编码) | 仅能投递自己宿舍/绑定点位 |
| messenger2 | PK5F** | 宿舍/店铺申请审批、任务分配 |
| messenger3 | PK** | 管理所有PK编码的学校片区 |
| messenger4 | 全城市(多个学校) | 开通新学校、赋予学校编码权限 |

## 🛠️ 数据库表结构修复方案

### 1. 角色系统表

```sql
-- 角色定义表
CREATE TABLE roles (
    id VARCHAR(36) PRIMARY KEY DEFAULT (uuid()),
    name VARCHAR(50) NOT NULL UNIQUE, 
    display_name VARCHAR(100) NOT NULL,
    level INTEGER NOT NULL,
    description TEXT,
    permissions TEXT, -- JSON格式存储权限列表
    is_system_role BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 插入FSD要求的系统角色
INSERT INTO roles (id, name, display_name, level, description, permissions) VALUES
('role_user', 'user', '普通用户', 1, '可以写信、回信、购买信封、编辑资料', '["write_letter", "read_letter", "buy_envelope", "edit_profile"]'),
('role_m1', 'messenger1', '一级信使', 2, '可见对应点位编码、执行扫码派送、提交派送状态', '["scan_code", "deliver_letter", "update_status"]'),
('role_m2', 'messenger2', '二级信使', 3, '管理片区投递点位、审批点位申请、查看片区内所有任务', '["manage_zone_points", "approve_applications", "view_zone_tasks"]'),
('role_m3', 'messenger3', '三级信使', 4, '管理本校信封、审批/指派一级二级信使、维护校级编码（3-4位）', '["manage_school_envelopes", "manage_messengers", "maintain_school_codes"]'),
('role_m4', 'messenger4', '四级信使', 5, '管理城市级学校开通、更新城市编码（1-2位）、设置活动信封', '["manage_city_schools", "update_city_codes", "set_event_envelopes"]'),
('role_admin', 'admin', '平台管理员', 6, '拥有所有权限，支持后台查看/审核/封禁等', '["*"]');
```

### 2. 用户角色关联表

```sql
-- 用户角色关联表
CREATE TABLE user_roles (
    id VARCHAR(36) PRIMARY KEY DEFAULT (uuid()),
    user_id VARCHAR(36) NOT NULL,
    role_id VARCHAR(36) NOT NULL,
    code_range VARCHAR(20), -- 地理权限范围，如"PK5F**"、"PK**"等
    school_code VARCHAR(10), -- 所属学校编码
    zone_code VARCHAR(10), -- 所属区域编码  
    assigned_by VARCHAR(36), -- 审批者用户ID
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP, -- 角色过期时间（可选）
    status ENUM('active', 'suspended', 'expired') DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE KEY unique_user_role (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (assigned_by) REFERENCES users(id) ON DELETE SET NULL,
    
    INDEX idx_user_roles_user_id (user_id),
    INDEX idx_user_roles_role_id (role_id),
    INDEX idx_user_roles_code_range (code_range),
    INDEX idx_user_roles_school_code (school_code)
);
```

### 3. 角色升级申请表

```sql
-- 角色升级申请表
CREATE TABLE role_upgrade_requests (
    id VARCHAR(36) PRIMARY KEY DEFAULT (uuid()),
    applicant_id VARCHAR(36) NOT NULL,
    current_role_id VARCHAR(36) NOT NULL,
    target_role_id VARCHAR(36) NOT NULL,
    requested_code_range VARCHAR(20), -- 申请的编码段权限
    motivation TEXT, -- 申请动机
    supporting_docs TEXT, -- 支持材料（JSON格式）
    
    -- 审批相关
    reviewer_id VARCHAR(36),
    review_status ENUM('pending', 'approved', 'rejected', 'withdrawn') DEFAULT 'pending',
    review_comment TEXT,
    reviewed_at TIMESTAMP,
    
    -- 系统字段
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (applicant_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (current_role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (target_role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (reviewer_id) REFERENCES users(id) ON DELETE SET NULL,
    
    INDEX idx_upgrade_requests_applicant (applicant_id),
    INDEX idx_upgrade_requests_status (review_status),
    INDEX idx_upgrade_requests_reviewer (reviewer_id)
);
```

### 4. 权限检查日志表

```sql
-- 权限检查日志表
CREATE TABLE permission_check_logs (
    id VARCHAR(36) PRIMARY KEY DEFAULT (uuid()),
    user_id VARCHAR(36) NOT NULL,
    resource VARCHAR(100) NOT NULL, -- 访问的资源
    action VARCHAR(50) NOT NULL, -- 执行的动作
    code_context VARCHAR(20), -- 编码上下文
    is_granted BOOLEAN NOT NULL,
    reason TEXT, -- 授权/拒绝原因
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    
    INDEX idx_permission_logs_user_id (user_id),
    INDEX idx_permission_logs_resource (resource),
    INDEX idx_permission_logs_created_at (created_at),
    INDEX idx_permission_logs_is_granted (is_granted)
);
```

## 🔄 现有User表修改方案

### 方案A: 平滑迁移（推荐）

```sql
-- 1. 保留现有role列作为兼容性字段
ALTER TABLE users ADD COLUMN legacy_role VARCHAR(50) AFTER role;
UPDATE users SET legacy_role = role;

-- 2. 添加主要角色ID字段
ALTER TABLE users ADD COLUMN primary_role_id VARCHAR(36) AFTER legacy_role;

-- 3. 添加FSD要求的新字段
ALTER TABLE users ADD COLUMN zone_code VARCHAR(10) AFTER school_code;
ALTER TABLE users ADD COLUMN address_code VARCHAR(10) AFTER zone_code;
ALTER TABLE users ADD COLUMN is_code_public BOOLEAN DEFAULT FALSE AFTER address_code;
ALTER TABLE users ADD COLUMN allow_ai_penpal BOOLEAN DEFAULT TRUE AFTER is_code_public;
ALTER TABLE users ADD COLUMN user_tags TEXT AFTER allow_ai_penpal; -- JSON格式存储标签数组

-- 4. 创建外键约束
ALTER TABLE users ADD FOREIGN KEY (primary_role_id) REFERENCES roles(id) ON DELETE SET NULL;

-- 5. 数据迁移脚本
UPDATE users SET primary_role_id = CASE 
    WHEN role = 'user' THEN 'role_user'
    WHEN role = 'courier_level1' THEN 'role_m1'
    WHEN role = 'courier_level2' THEN 'role_m2'
    WHEN role = 'courier_level3' THEN 'role_m3'
    WHEN role = 'courier_level4' THEN 'role_m4'
    WHEN role = 'admin' THEN 'role_admin'
    ELSE 'role_user'
END;
```

### 方案B: 直接重构（激进）

```sql
-- 直接修改role列为符合FSD的枚举值
ALTER TABLE users MODIFY COLUMN role ENUM('user', 'messenger1', 'messenger2', 'messenger3', 'messenger4', 'admin') DEFAULT 'user';

-- 数据迁移
UPDATE users SET role = CASE 
    WHEN role = 'courier_level1' THEN 'messenger1'
    WHEN role = 'courier_level2' THEN 'messenger2'
    WHEN role = 'courier_level3' THEN 'messenger3'
    WHEN role = 'courier_level4' THEN 'messenger4'
    ELSE role
END;
```

## 🚀 Go Model 更新

### 更新后的User Model

```go
// User 用户模型
type User struct {
    ID           string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
    Username     string    `json:"username" gorm:"type:varchar(50);uniqueIndex;not null"`
    Email        string    `json:"email" gorm:"type:varchar(100);uniqueIndex"`
    PasswordHash string    `json:"-" gorm:"type:varchar(255);not null"`
    Nickname     string    `json:"nickname" gorm:"type:varchar(50)"`
    Avatar       string    `json:"avatar" gorm:"type:varchar(500)"`
    
    // 角色相关
    Role           UserRole `json:"role" gorm:"type:varchar(20);not null;default:'user'"` // 兼容字段
    PrimaryRoleID  *string  `json:"primary_role_id" gorm:"type:varchar(36)"` // 主要角色ID
    LegacyRole     *string  `json:"legacy_role,omitempty" gorm:"type:varchar(50)"` // 历史兼容
    
    // 地理编码相关
    SchoolCode    string `json:"school_code" gorm:"type:varchar(20);index"`
    ZoneCode      string `json:"zone_code" gorm:"type:varchar(10);index"` // 新增
    AddressCode   string `json:"address_code" gorm:"type:varchar(10)"` // 新增
    IsCodePublic  bool   `json:"is_code_public" gorm:"default:false"` // 新增
    
    // AI和标签相关
    AllowAIPenpal bool   `json:"allow_ai_penpal" gorm:"default:true"` // 新增
    UserTags      string `json:"user_tags" gorm:"type:text"` // JSON格式标签数组
    
    // 系统字段
    IsActive     bool      `json:"is_active" gorm:"default:true"`
    LastLoginAt  *time.Time `json:"last_login_at"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
    DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

    // 关联关系
    PrimaryRole   *Role      `json:"primary_role,omitempty" gorm:"foreignKey:PrimaryRoleID"`
    UserRoles     []UserRole `json:"user_roles,omitempty" gorm:"foreignKey:UserID"`
    SentLetters   []Letter   `json:"sent_letters,omitempty" gorm:"foreignKey:UserID"`
}

// Role 角色模型
type Role struct {
    ID           string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
    Name         string    `json:"name" gorm:"type:varchar(50);uniqueIndex;not null"`
    DisplayName  string    `json:"display_name" gorm:"type:varchar(100);not null"`
    Level        int       `json:"level" gorm:"not null"`
    Description  string    `json:"description" gorm:"type:text"`
    Permissions  string    `json:"permissions" gorm:"type:text"` // JSON格式权限列表
    IsSystemRole bool      `json:"is_system_role" gorm:"default:true"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}

// UserRole 用户角色关联模型
type UserRole struct {
    ID          string     `json:"id" gorm:"primaryKey;type:varchar(36)"`
    UserID      string     `json:"user_id" gorm:"type:varchar(36);not null;index"`
    RoleID      string     `json:"role_id" gorm:"type:varchar(36);not null;index"`
    CodeRange   string     `json:"code_range" gorm:"type:varchar(20);index"` // 地理权限范围
    SchoolCode  string     `json:"school_code" gorm:"type:varchar(10);index"`
    ZoneCode    string     `json:"zone_code" gorm:"type:varchar(10)"`
    AssignedBy  *string    `json:"assigned_by" gorm:"type:varchar(36)"`
    AssignedAt  time.Time  `json:"assigned_at"`
    ExpiresAt   *time.Time `json:"expires_at"`
    Status      string     `json:"status" gorm:"type:varchar(20);default:'active'"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
    
    // 关联关系
    User       *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
    Role       *Role `json:"role,omitempty" gorm:"foreignKey:RoleID"`
    AssignedByUser *User `json:"assigned_by_user,omitempty" gorm:"foreignKey:AssignedBy"`
}

// RoleUpgradeRequest 角色升级申请模型
type RoleUpgradeRequest struct {
    ID                 string     `json:"id" gorm:"primaryKey;type:varchar(36)"`
    ApplicantID        string     `json:"applicant_id" gorm:"type:varchar(36);not null;index"`
    CurrentRoleID      string     `json:"current_role_id" gorm:"type:varchar(36);not null"`
    TargetRoleID       string     `json:"target_role_id" gorm:"type:varchar(36);not null"`
    RequestedCodeRange string     `json:"requested_code_range" gorm:"type:varchar(20)"`
    Motivation         string     `json:"motivation" gorm:"type:text"`
    SupportingDocs     string     `json:"supporting_docs" gorm:"type:text"` // JSON
    
    // 审批相关
    ReviewerID     *string    `json:"reviewer_id" gorm:"type:varchar(36)"`
    ReviewStatus   string     `json:"review_status" gorm:"type:varchar(20);default:'pending';index"`
    ReviewComment  string     `json:"review_comment" gorm:"type:text"`
    ReviewedAt     *time.Time `json:"reviewed_at"`
    
    CreatedAt      time.Time  `json:"created_at"`
    UpdatedAt      time.Time  `json:"updated_at"`
    
    // 关联关系
    Applicant    *User `json:"applicant,omitempty" gorm:"foreignKey:ApplicantID"`
    CurrentRole  *Role `json:"current_role,omitempty" gorm:"foreignKey:CurrentRoleID"`
    TargetRole   *Role `json:"target_role,omitempty" gorm:"foreignKey:TargetRoleID"`
    Reviewer     *User `json:"reviewer,omitempty" gorm:"foreignKey:ReviewerID"`
}
```

## 📝 权限检查服务

```go
// PermissionService 权限检查服务
type PermissionService struct {
    db *gorm.DB
}

func NewPermissionService(db *gorm.DB) *PermissionService {
    return &PermissionService{db: db}
}

// CheckPermission 检查用户权限
func (s *PermissionService) CheckPermission(userID, resource, action, codeContext string) (bool, error) {
    // 1. 获取用户所有有效角色
    var userRoles []UserRole
    err := s.db.Preload("Role").Where("user_id = ? AND status = 'active'", userID).
        Where("expires_at IS NULL OR expires_at > NOW()").Find(&userRoles).Error
    if err != nil {
        return false, err
    }
    
    // 2. 检查每个角色的权限
    for _, userRole := range userRoles {
        if s.hasPermission(userRole.Role, resource, action) {
            // 3. 检查编码段权限
            if s.checkCodePermission(userRole.CodeRange, codeContext) {
                // 4. 记录权限检查日志
                s.logPermissionCheck(userID, resource, action, codeContext, true, "权限检查通过")
                return true, nil
            }
        }
    }
    
    // 5. 记录拒绝日志
    s.logPermissionCheck(userID, resource, action, codeContext, false, "权限不足")
    return false, nil
}

// hasPermission 检查角色是否有指定权限
func (s *PermissionService) hasPermission(role *Role, resource, action string) bool {
    if role == nil {
        return false
    }
    
    // 管理员拥有所有权限
    if role.Name == "admin" {
        return true
    }
    
    // 解析权限JSON
    var permissions []string
    if err := json.Unmarshal([]byte(role.Permissions), &permissions); err != nil {
        return false
    }
    
    // 检查具体权限
    targetPermission := fmt.Sprintf("%s.%s", resource, action)
    for _, perm := range permissions {
        if perm == "*" || perm == targetPermission || perm == resource+".*" {
            return true
        }
    }
    
    return false
}

// checkCodePermission 检查编码段权限
func (s *PermissionService) checkCodePermission(codeRange, codeContext string) bool {
    if codeRange == "" || codeContext == "" {
        return true // 无编码限制
    }
    
    // 实现编码段匹配逻辑
    // 例如：codeRange="PK5F**", codeContext="PK5F3D" -> true
    // 例如：codeRange="PK**", codeContext="PK5F3D" -> true
    return strings.HasPrefix(codeContext, strings.Replace(codeRange, "*", "", -1))
}

// logPermissionCheck 记录权限检查日志
func (s *PermissionService) logPermissionCheck(userID, resource, action, codeContext string, isGranted bool, reason string) {
    log := PermissionCheckLog{
        ID:          uuid.New().String(),
        UserID:      userID,
        Resource:    resource,
        Action:      action,
        CodeContext: codeContext,
        IsGranted:   isGranted,
        Reason:      reason,
        CreatedAt:   time.Now(),
    }
    s.db.Create(&log)
}
```

## 🎯 迁移执行计划

### 阶段1: 表结构创建 (1天)
1. 创建roles表并插入系统角色
2. 创建user_roles表
3. 创建role_upgrade_requests表
4. 创建permission_check_logs表

### 阶段2: 数据迁移 (1天)
1. 迁移现有用户角色数据
2. 为现有信使创建编码段权限
3. 验证数据一致性

### 阶段3: 代码更新 (2天)
1. 更新Go model文件
2. 实现权限检查服务
3. 更新中间件和处理器

### 阶段4: 测试验证 (1天)
1. 权限检查功能测试
2. 角色升级流程测试
3. 编码段权限测试

---

**总时间预估**: 5个工作日  
**风险级别**: 中等（需要数据迁移）  
**影响范围**: 整个权限系统，但向后兼容