#!/usr/bin/env python3

"""
完整修复handlers中的response问题
"""

import os
import re

def fix_handler_file(file_path):
    print(f"处理文件: {file_path}")
    
    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()
    
    original_content = content
    
    # 1. 移除problematic import
    content = re.sub(r'.*"shared/pkg/response".*\n', '', content)
    
    # 2. 确保utils import存在
    if 'openpenpal-backend/internal/utils' not in content:
        # 在import block中添加utils
        content = re.sub(
            r'(import \([\s\S]*?)(\))',
            r'\1\t"openpenpal-backend/internal/utils"\n\2',
            content
        )
    
    # 3. 替换所有response.NewGinResponse()调用
    content = re.sub(r'resp := response\.NewGinResponse\(\)', 
                     '// Using utils response functions directly', content)
    
    # 4. 替换所有resp.方法调用
    replacements = {
        r'resp\.BadRequest\(c, ([^)]+)\)': r'utils.BadRequestResponse(c, \1, nil)',
        r'resp\.Unauthorized\(c, ([^)]+)\)': r'utils.UnauthorizedResponse(c, \1)',
        r'resp\.NotFound\(c, ([^)]+)\)': r'utils.NotFoundResponse(c, \1)',
        r'resp\.InternalServerError\(c, ([^)]+)\)': r'utils.InternalServerErrorResponse(c, "Internal server error", nil)',
        r'resp\.Success\(c, ([^)]+)\)': r'utils.SuccessResponse(c, 200, "Success", \1)',
        r'resp\.OK\(c, ([^)]+)\)': r'utils.SuccessResponse(c, 200, \1, nil)',
        r'resp\.Created\(c, ([^)]+)\)': r'utils.SuccessResponse(c, 201, "Created", \1)',
        r'resp\.CreatedWithMessage\(c, ([^,]+), ([^)]+)\)': r'utils.SuccessResponse(c, 201, \1, \2)',
        r'resp\.SuccessWithMessage\(c, ([^,]+), ([^)]+)\)': r'utils.SuccessResponse(c, 200, \1, \2)',
    }
    
    for pattern, replacement in replacements.items():
        content = re.sub(pattern, replacement, content)
    
    # 5. 特殊处理复杂情况
    content = re.sub(
        r'utils\.InternalServerErrorResponse\(c, "Internal server error", nil\)',
        r'utils.InternalServerErrorResponse(c, "Failed to process request", nil)',
        content
    )
    
    # 只有内容发生变化时才写入
    if content != original_content:
        with open(file_path, 'w', encoding='utf-8') as f:
            f.write(content)
        print(f"✅ {file_path} 修复完成")
        return True
    else:
        print(f"⚪ {file_path} 无需修改")
        return False

def main():
    backend_dir = "/Users/rocalight/同步空间/opplc/openpenpal/backend"
    
    # 需要修复的文件
    files_to_fix = [
        "internal/handlers/letter_handler.go",
        "internal/handlers/courier_handler.go", 
        "internal/handlers/user_handler.go",
        "internal/handlers/letter_handler_envelope.go",
        "internal/handlers/envelope_handler.go",
        "internal/handlers/credit_handler.go"
    ]
    
    print("🔧 开始修复handler文件中的response问题...")
    
    fixed_count = 0
    for file_rel_path in files_to_fix:
        file_path = os.path.join(backend_dir, file_rel_path)
        if os.path.exists(file_path):
            if fix_handler_file(file_path):
                fixed_count += 1
        else:
            print(f"⚠️  文件不存在: {file_path}")
    
    print(f"\n🎉 修复完成！共修复 {fixed_count} 个文件")

if __name__ == "__main__":
    main()