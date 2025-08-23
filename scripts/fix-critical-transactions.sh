#!/bin/bash

# 修复关键业务逻辑的事务处理
# 重点处理积分系统、信件系统等核心模块

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SERVICES_DIR="$PROJECT_ROOT/backend/internal/services"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_info "🔄 修复关键业务逻辑的事务处理..."

# 备份函数
backup_file() {
    local file="$1"
    local backup="${file}.backup.$(date +%Y%m%d_%H%M%S)"
    cp "$file" "$backup"
    log_info "已备份: $(basename "$file") -> $(basename "$backup")"
}

# 检查文件是否存在TransactionHelper
has_transaction_helper() {
    local file="$1"
    grep -q "transactionHelper.*TransactionHelper" "$file" 2>/dev/null
}

# 添加TransactionHelper字段到服务结构
add_transaction_helper_field() {
    local file="$1"
    local service_name="$2"
    
    if has_transaction_helper "$file"; then
        log_info "$(basename "$file") 已有TransactionHelper字段"
        return 0
    fi
    
    # 查找服务结构体定义
    if grep -q "type ${service_name} struct" "$file"; then
        # 在struct中添加TransactionHelper字段
        sed -i.tmp "/type ${service_name} struct/,/^}/ {
            /db.*\*gorm\.DB/a\\
\\	transactionHelper *TransactionHelper
        }" "$file"
        rm -f "${file}.tmp"
        
        # 在构造函数中初始化TransactionHelper
        if grep -q "func New${service_name}" "$file"; then
            sed -i.tmp "/return &${service_name}{/,/}/ {
                /db.*db,/a\\
\\		transactionHelper: NewTransactionHelper(db),
            }" "$file"
            rm -f "${file}.tmp"
        fi
        
        log_success "已添加TransactionHelper到 $(basename "$file")"
    else
        log_warning "在 $(basename "$file") 中未找到 ${service_name} 结构体"
    fi
}

# 修复积分服务的事务处理
fix_credit_service() {
    local file="$SERVICES_DIR/credit_service.go"
    
    if [ ! -f "$file" ]; then
        log_warning "积分服务文件不存在: $file"
        return
    fi
    
    log_info "修复积分服务事务处理..."
    backup_file "$file"
    
    # 添加TransactionHelper
    add_transaction_helper_field "$file" "CreditService"
    
    # 替换直接的事务调用为标准化调用
    # 修复AwardCredits方法
    sed -i.tmp 's/tx := s\.db\.Begin()/err := s.transactionHelper.WithCreditTransferTransaction(ctx, "award_credits", func(tx *gorm.DB) error {/' "$file"
    
    # 修复DeductCredits方法
    sed -i.tmp 's/defer tx\.Rollback()/\/\/ Auto-rollback handled by transaction helper/' "$file"
    sed -i.tmp 's/return tx\.Commit()\.Error/return nil \/\/ Auto-commit handled by transaction helper/' "$file"
    sed -i.tmp 's/tx\.Rollback()/return err \/\/ Will trigger rollback/' "$file"
    
    # 添加context参数到方法签名（如果缺失）
    sed -i.tmp 's/func (s \*CreditService) AwardCredits(/func (s *CreditService) AwardCredits(ctx context.Context, /' "$file"
    sed -i.tmp 's/func (s \*CreditService) DeductCredits(/func (s *CreditService) DeductCredits(ctx context.Context, /' "$file"
    
    # 添加事务结束括号
    echo "// Note: Add closing bracket '}); err != nil { return err }' after transaction logic" >> "$file"
    
    rm -f "${file}.tmp"
    log_success "积分服务事务修复完成"
}

# 修复积分转账服务
fix_credit_transfer_service() {
    local file="$SERVICES_DIR/credit_transfer_service.go"
    
    if [ ! -f "$file" ]; then
        log_warning "积分转账服务文件不存在: $file"
        return
    fi
    
    log_info "修复积分转账服务事务处理..."
    backup_file "$file"
    
    # 添加TransactionHelper
    add_transaction_helper_field "$file" "CreditTransferService"
    
    # 这是关键的业务逻辑，需要SERIALIZABLE隔离级别
    sed -i.tmp 's/tx := s\.db\.Begin()/err := s.transactionHelper.WithCreditTransferTransaction(ctx, "credit_transfer", func(tx *gorm.DB) error {/' "$file"
    
    rm -f "${file}.tmp"
    log_success "积分转账服务事务修复完成"
}

# 修复信件服务
fix_letter_service() {
    local file="$SERVICES_DIR/letter_service.go"
    
    if [ ! -f "$file" ]; then
        log_warning "信件服务文件不存在: $file"
        return
    fi
    
    log_info "修复信件服务事务处理..."
    backup_file "$file"
    
    # 添加TransactionHelper
    add_transaction_helper_field "$file" "LetterService"
    
    # 信件创建需要高并发优化
    sed -i.tmp 's/tx := s\.db\.WithContext(ctx)\.Begin()/err := s.transactionHelper.WithHighConcurrencyTransaction(ctx, "create_letter", func(tx *gorm.DB) error {/' "$file"
    
    rm -f "${file}.tmp"
    log_success "信件服务事务修复完成"
}

# 修复评论服务
fix_comment_service() {
    local file="$SERVICES_DIR/comment_service.go"
    
    if [ ! -f "$file" ]; then
        log_warning "评论服务文件不存在: $file"
        return
    fi
    
    log_info "修复评论服务事务处理..."
    backup_file "$file"
    
    # 添加TransactionHelper
    add_transaction_helper_field "$file" "CommentService"
    
    # 评论创建使用高并发事务
    sed -i.tmp 's/tx := s\.db\.WithContext(ctx)\.Begin()/err := s.transactionHelper.WithHighConcurrencyTransaction(ctx, "create_comment", func(tx *gorm.DB) error {/' "$file"
    
    rm -f "${file}.tmp"
    log_success "评论服务事务修复完成"
}

# 创建事务标准化检查工具
create_transaction_check_tool() {
    local check_script="$SCRIPT_DIR/check-transaction-usage.sh"
    
    cat > "$check_script" << 'EOF'
#!/bin/bash

# 检查事务使用情况
SERVICES_DIR="../backend/internal/services"

echo "🔍 检查事务使用情况..."
echo

echo "📊 直接使用db.Begin()的文件:"
grep -l "\.Begin()" "$SERVICES_DIR"/*.go | while read file; do
    count=$(grep -c "\.Begin()" "$file")
    echo "  $(basename "$file"): $count 处"
done

echo
echo "✅ 使用TransactionHelper的文件:"
grep -l "transactionHelper\|TransactionHelper" "$SERVICES_DIR"/*.go | while read file; do
    echo "  $(basename "$file")"
done

echo
echo "⚠️  需要注意的模式:"
echo "  - 查找未处理的事务错误"
grep -n "tx\.Commit()\.Error" "$SERVICES_DIR"/*.go | head -5
echo "  - 查找手动rollback"
grep -n "tx\.Rollback()" "$SERVICES_DIR"/*.go | head -5

echo
echo "📋 建议:"
echo "  1. 将剩余的db.Begin()替换为TransactionHelper"
echo "  2. 确保所有事务操作都有适当的错误处理"
echo "  3. 考虑为不同业务场景使用不同的事务类型"
EOF

    chmod +x "$check_script"
    log_success "已创建事务检查工具: $(basename "$check_script")"
}

# 创建事务最佳实践文档
create_transaction_guide() {
    local guide_file="$PROJECT_ROOT/docs/TRANSACTION_BEST_PRACTICES.md"
    mkdir -p "$(dirname "$guide_file")"
    
    cat > "$guide_file" << 'EOF'
# 事务管理最佳实践

## 概述

本文档描述了OpenPenPal项目中数据库事务的标准化管理方法。

## TransactionHelper使用指南

### 1. 基础事务
```go
err := s.transactionHelper.WithTransaction(ctx, func(tx *gorm.DB) error {
    // 业务逻辑
    return nil // 自动提交
    // return err // 自动回滚
})
```

### 2. 积分转账（高一致性要求）
```go
err := s.transactionHelper.WithCreditTransferTransaction(ctx, "transfer_credits", func(tx *gorm.DB) error {
    // 积分扣减和增加逻辑
    return nil
})
```

### 3. 高并发场景
```go
err := s.transactionHelper.WithHighConcurrencyTransaction(ctx, "create_comment", func(tx *gorm.DB) error {
    // 高并发创建逻辑
    return nil
})
```

### 4. 只读操作
```go
err := s.transactionHelper.WithReadOnlyTransaction(ctx, "analytics_query", func(tx *gorm.DB) error {
    // 只读查询逻辑
    return nil
})
```

## 服务结构示例

```go
type YourService struct {
    db                *gorm.DB
    transactionHelper *TransactionHelper
    logger           *logger.SmartLogger
}

func NewYourService(db *gorm.DB, logger *logger.SmartLogger) *YourService {
    return &YourService{
        db:                db,
        transactionHelper: NewTransactionHelper(db),
        logger:           logger,
    }
}
```

## 事务类型选择

| 业务场景 | 推荐事务类型 | 原因 |
|---------|-------------|------|
| 积分转账 | CreditTransferTransaction | 需要最高一致性 |
| 信件创建 | HighConcurrencyTransaction | 高并发优化 |
| 订单创建 | OrderCreationTransaction | 库存检查 |
| 数据分析 | ReadOnlyTransaction | 只读优化 |
| 用户更新 | UserDataTransaction | 适中的一致性 |
| 批量操作 | BulkDataTransaction | 批量优化 |

## 错误处理

### DO ✅
```go
err := s.transactionHelper.WithTransaction(ctx, func(tx *gorm.DB) error {
    if err := tx.Create(&record).Error; err != nil {
        return fmt.Errorf("failed to create record: %w", err)
    }
    return nil
})
if err != nil {
    return err
}
```

### DON'T ❌
```go
tx := s.db.Begin()
defer tx.Rollback() // 不需要手动管理
if err := tx.Create(&record).Error; err != nil {
    return err
}
return tx.Commit().Error // 不需要手动提交
```

## 性能考虑

1. **选择合适的隔离级别**: 不是所有操作都需要SERIALIZABLE
2. **避免长时间事务**: 尽快提交以减少锁定时间
3. **批量操作**: 使用BulkDataTransaction进行批量处理
4. **只读查询**: 使用ReadOnlyTransaction减少开销

## 监控和调试

```go
// 获取事务统计
stats, err := s.transactionHelper.GetTransactionStats(ctx)
if err == nil {
    log.Printf("Transaction stats: %+v", stats)
}
```

## 迁移清单

- [ ] 替换所有db.Begin()调用
- [ ] 添加TransactionHelper到服务结构
- [ ] 更新构造函数
- [ ] 选择合适的事务类型
- [ ] 测试错误处理
- [ ] 性能验证
EOF

    log_success "已创建事务最佳实践文档: $(basename "$guide_file")"
}

# 主执行流程
main() {
    log_info "开始修复关键业务逻辑的事务处理..."
    
    # 检查services目录
    if [ ! -d "$SERVICES_DIR" ]; then
        log_error "Services目录不存在: $SERVICES_DIR"
        exit 1
    fi
    
    # 修复关键服务
    fix_credit_service
    fix_credit_transfer_service
    fix_letter_service
    fix_comment_service
    
    # 创建工具和文档
    create_transaction_check_tool
    create_transaction_guide
    
    log_success "关键业务逻辑事务修复完成!"
    echo
    log_info "后续步骤:"
    echo "1. 运行 ./scripts/check-transaction-usage.sh 检查剩余问题"
    echo "2. 查看 docs/TRANSACTION_BEST_PRACTICES.md 了解最佳实践"
    echo "3. 测试修改的服务功能"
    echo "4. 考虑将其他服务也迁移到TransactionHelper"
}

# 执行主函数
main "$@"