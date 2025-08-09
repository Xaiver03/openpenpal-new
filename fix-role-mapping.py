#!/usr/bin/env python3
import re

# 读取文件
with open('backend/internal/models/role_mapping.go', 'r') as f:
    content = f.read()

# 备份文件
with open('backend/internal/models/role_mapping.go.backup3', 'w') as f:
    f.write(content)

# 替换 RoleSchoolAdmin 引用
content = re.sub(r'RoleSchoolAdmin', 'RolePlatformAdmin', content)

# 更新 FrontendToBackend 映射中的 school_admin
content = re.sub(
    r'"school_admin":\s*RolePlatformAdmin,',
    '"school_admin":   RoleCourierLevel3, // 学校管理员映射到三级信使',
    content
)

# 更新 BackendToFrontend 映射，移除 RoleSchoolAdmin 条目
# 因为已经被替换为 RolePlatformAdmin，需要避免重复
content = re.sub(
    r'RolePlatformAdmin:\s*"school_admin",\s*\n',
    '',
    content
)

# 写回文件
with open('backend/internal/models/role_mapping.go', 'w') as f:
    f.write(content)

print("✅ Successfully updated role_mapping.go!")
print("\n📋 Changes made:")
print("  - Replaced RoleSchoolAdmin references")
print("  - Updated school_admin mapping to courier_level3")
print("  - Cleaned up duplicate entries")