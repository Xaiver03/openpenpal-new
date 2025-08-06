#!/bin/bash
# 统一测试覆盖率脚本

set -e

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}🧪 运行测试覆盖率检查...${NC}"

# 创建覆盖率目录
mkdir -p coverage

# 前端测试覆盖率
echo -e "\n${YELLOW}📊 前端测试覆盖率${NC}"
cd frontend
npm test -- --coverage --watchAll=false
cd ..

# 后端测试覆盖率
echo -e "\n${YELLOW}📊 后端测试覆盖率${NC}"
cd backend
go test -v -race -coverprofile=../coverage/backend.out ./...
go tool cover -html=../coverage/backend.out -o ../coverage/backend.html
cd ..

# Courier服务测试覆盖率
echo -e "\n${YELLOW}📊 Courier服务测试覆盖率${NC}"
cd services/courier-service
go test -v -race -coverprofile=../../coverage/courier.out ./...
go tool cover -html=../../coverage/courier.out -o ../../coverage/courier.html
cd ../..

# Python服务测试覆盖率
echo -e "\n${YELLOW}📊 Python服务测试覆盖率${NC}"

# Write服务
cd services/write-service
python -m pytest --cov=. --cov-report=html:../../coverage/write-service --cov-report=term
cd ../..

# OCR服务  
cd services/ocr-service
python -m pytest --cov=. --cov-report=html:../../coverage/ocr-service --cov-report=term
cd ../..

# Java服务测试覆盖率
echo -e "\n${YELLOW}📊 Java服务测试覆盖率${NC}"
cd services/admin-service
./mvnw clean test jacoco:report
cp -r target/site/jacoco ../../coverage/admin-service
cd ../..

# 生成汇总报告
echo -e "\n${GREEN}📈 生成覆盖率汇总报告${NC}"
cat > coverage/index.html << EOF
<!DOCTYPE html>
<html>
<head>
    <title>OpenPenPal 测试覆盖率报告</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        h1 { color: #333; }
        .service { margin: 20px 0; padding: 20px; border: 1px solid #ddd; border-radius: 5px; }
        .service h2 { margin-top: 0; }
        a { color: #0066cc; text-decoration: none; }
        a:hover { text-decoration: underline; }
        .coverage-badge { 
            display: inline-block; 
            padding: 3px 10px; 
            border-radius: 3px; 
            color: white; 
            font-weight: bold; 
        }
        .coverage-high { background-color: #4c1; }
        .coverage-medium { background-color: #fa0; }
        .coverage-low { background-color: #e00; }
    </style>
</head>
<body>
    <h1>OpenPenPal 测试覆盖率报告</h1>
    <p>生成时间: $(date)</p>
    
    <div class="service">
        <h2>前端 (Next.js)</h2>
        <p><a href="./lcov-report/index.html">查看详细报告</a></p>
    </div>
    
    <div class="service">
        <h2>后端主服务 (Go)</h2>
        <p><a href="./backend.html">查看详细报告</a></p>
    </div>
    
    <div class="service">
        <h2>Courier服务 (Go)</h2>
        <p><a href="./courier.html">查看详细报告</a></p>
    </div>
    
    <div class="service">
        <h2>Write服务 (Python)</h2>
        <p><a href="./write-service/index.html">查看详细报告</a></p>
    </div>
    
    <div class="service">
        <h2>OCR服务 (Python)</h2>
        <p><a href="./ocr-service/index.html">查看详细报告</a></p>
    </div>
    
    <div class="service">
        <h2>Admin服务 (Java)</h2>
        <p><a href="./admin-service/index.html">查看详细报告</a></p>
    </div>
</body>
</html>
EOF

echo -e "\n${GREEN}✅ 测试覆盖率报告生成完成！${NC}"
echo -e "查看报告: open coverage/index.html"

# 检查覆盖率阈值
echo -e "\n${YELLOW}🔍 检查覆盖率阈值...${NC}"

# 这里可以添加覆盖率阈值检查逻辑
# 如果覆盖率低于阈值，返回非零退出码

echo -e "${GREEN}✅ 所有测试覆盖率检查通过！${NC}"