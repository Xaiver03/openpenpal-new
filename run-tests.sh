#!/bin/bash

# OpenPenPal 测试运行器

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

MODE=${1:-all}

case $MODE in
  backend)
    echo -e "${YELLOW}🧪 运行后端测试...${NC}"
    cd backend && make test-coverage
    ;;
  
  frontend)
    echo -e "${YELLOW}🧪 运行前端测试...${NC}"
    cd frontend && npm run test:coverage
    ;;
  
  e2e)
    echo -e "${YELLOW}🧪 运行 E2E 测试...${NC}"
    cd frontend && npm run test:e2e
    ;;
  
  all)
    echo -e "${YELLOW}🧪 运行所有测试...${NC}"
    
    echo -e "\n${GREEN}后端测试：${NC}"
    cd backend && make test
    
    echo -e "\n${GREEN}前端测试：${NC}"
    cd ../frontend && npm test
    
    echo -e "\n${GREEN}✅ 所有测试完成！${NC}"
    ;;
  
  watch)
    echo -e "${YELLOW}👀 监视模式...${NC}"
    echo "1) 后端测试监视"
    echo "2) 前端测试监视"
    read -p "选择 (1/2): " choice
    
    case $choice in
      1) cd backend && watch -n 2 go test ./... ;;
      2) cd frontend && npm run test:watch ;;
    esac
    ;;
  
  *)
    echo "用法: $0 [backend|frontend|e2e|all|watch]"
    exit 1
    ;;
esac
