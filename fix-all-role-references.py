#!/usr/bin/env python3
import os
import re

# 需要更新的文件列表
files_to_update = [
    'backend/internal/services/courier_service.go',
    'backend/internal/websocket/client.go',
    'backend/internal/models/role_mapping.go'
]

# 角色替换规则
replacements = [
    # 在 case 语句中的替换
    (r'case models\.RoleCourier:', 'case models.RoleCourierLevel1:'),
    (r'case models\.RoleSeniorCourier:', 'case models.RoleCourierLevel2:'),
    (r'case models\.RoleCourierCoordinator:', 'case models.RoleCourierLevel3:'),
    (r'case models\.RoleSchoolAdmin:', 'case models.RoleCourierLevel3:'),
    
    # 在条件语句中的替换
    (r'models\.RoleCourier(?!Level)', 'models.RoleCourierLevel1'),
    (r'models\.RoleSeniorCourier', 'models.RoleCourierLevel2'),
    (r'models\.RoleCourierCoordinator', 'models.RoleCourierLevel3'),
    (r'models\.RoleSchoolAdmin', 'models.RolePlatformAdmin'),  # 学校管理员改为平台管理员
    
    # 特殊处理：多个角色在一起的情况
    (r'models\.RoleCourier, models\.RoleSeniorCourier, models\.RoleCourierCoordinator', 
     'models.RoleCourierLevel1, models.RoleCourierLevel2, models.RoleCourierLevel3, models.RoleCourierLevel4'),
    
    # HasRole 检查
    (r'HasRole\(models\.RoleCourier\)', 'HasRole(models.RoleCourierLevel1)'),
]

# 处理每个文件
for file_path in files_to_update:
    if not os.path.exists(file_path):
        print(f"⚠️  File not found: {file_path}")
        continue
    
    # 读取文件
    with open(file_path, 'r') as f:
        content = f.read()
    
    # 备份文件
    backup_path = file_path + '.role_backup'
    with open(backup_path, 'w') as f:
        f.write(content)
    
    # 应用替换
    original_content = content
    for old_pattern, new_pattern in replacements:
        content = re.sub(old_pattern, new_pattern, content)
    
    # 如果有变化，写回文件
    if content != original_content:
        with open(file_path, 'w') as f:
            f.write(content)
        print(f"✅ Updated: {file_path}")
    else:
        print(f"ℹ️  No changes needed: {file_path}")

print("\n📋 Summary:")
print("  - Updated role references to use four-level courier system")
print("  - Removed references to redundant roles")
print("  - Backups created with .role_backup extension")

# 特别处理 role_mapping.go - 如果存在的话
role_mapping_path = 'backend/internal/models/role_mapping.go'
if os.path.exists(role_mapping_path):
    print(f"\n⚠️  Note: {role_mapping_path} may need manual review for compatibility mapping")