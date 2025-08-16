#!/usr/bin/env python3

import re

file_path = "/Users/rocalight/同步空间/opplc/openpenpal/backend/internal/handlers/user_handler.go"

with open(file_path, 'r', encoding='utf-8') as f:
    content = f.read()

# Remove all resp := response.NewGinResponse() lines
content = re.sub(r'\s*resp := response\.NewGinResponse\(\)\s*\n', '', content)

# Replace resp.methods with utils.methods
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

with open(file_path, 'w', encoding='utf-8') as f:
    f.write(content)

print("✅ user_handler.go fixed")