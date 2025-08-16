#!/usr/bin/env python3
import re

# 读取database.go文件
with open('backend/internal/config/database.go', 'r') as f:
    content = f.read()

# 备份原文件
with open('backend/internal/config/database.go.backup2', 'w') as f:
    f.write(content)

# 需要移除的用户ID和角色
users_to_remove = [
    ('5', 'courier'),
    ('6', 'senior_courier'),
    ('7', 'courier_coordinator'),
    ('8', 'school_admin'),
    ('11', 'courier_1'),
    ('12', 'courier_2'),
    ('13', 'courier_3')
]

# 移除这些用户的定义
for user_id, username in users_to_remove:
    # 匹配用户定义块
    pattern = rf'{{[^{{}}]*ID:\s*"{user_id}"[^{{}}]*Username:\s*"{username}"[^{{}}]*}},?\s*'
    content = re.sub(pattern, '', content, flags=re.DOTALL)

# 更新角色引用
replacements = [
    (r'models\.RoleCourier(?!Level)', 'models.RoleCourierLevel1'),
    (r'models\.RoleSeniorCourier', 'models.RoleCourierLevel2'),
    (r'models\.RoleCourierCoordinator', 'models.RoleCourierLevel3'),
    (r'models\.RoleSchoolAdmin', 'models.RoleCourierLevel3'),
]

for old, new in replacements:
    content = re.sub(old, new, content)

# 写回文件
with open('backend/internal/config/database.go', 'w') as f:
    f.write(content)

print("✅ Successfully updated database.go!")
print("\n📋 Changes made:")
print("  - Removed redundant user seeds (IDs: 5, 6, 7, 8, 11, 12, 13)")
print("  - Updated role references to use four-level courier system")
print("\n✅ Backup saved as database.go.backup2")