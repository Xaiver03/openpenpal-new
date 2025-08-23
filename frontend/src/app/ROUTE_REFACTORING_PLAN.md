# 前端路由重构计划

## 1. 路由整合方案

### 信使系统整合 (从14个子路由整合为4个主要模块)

**当前结构：**
```
/courier/
├── tasks/
├── scan/
├── growth/
├── points/
├── promotion/apply/
├── promotion/manage/
├── building-manage/
├── zone-manage/
├── school-manage/
├── city-manage/
├── batch/
├── opcode-manage/
├── credit-manage/
└── apply/
```

**新结构：**
```
/courier/
├── dashboard/        # 新增：统一仪表板
├── tasks/           # 合并：tasks + scan
├── management/      # 新增：管理中心
│   ├── hierarchy/   # 合并：building/zone/school/city-manage
│   ├── opcode/      # 重命名：opcode-manage
│   ├── batch/       # 保持
│   └── credits/     # 重命名：credit-manage
└── profile/         # 新增：个人中心
    ├── growth/      # 保持
    ├── points/      # 保持
    └── promotion/   # 合并：promotion/apply + manage
```

### 商城系统整合

**当前结构：**
```
/(main)/shop/         # 实物商品
/(main)/credit-shop/  # 积分商品
/cart/               # 购物车（顶层）
/checkout/           # 结算（顶层）
/orders/             # 订单（顶层）
```

**新结构：**
```
/(main)/shop/
├── products/        # 商品浏览（支持切换类型）
├── cart/           # 购物车
├── checkout/       # 结算
└── orders/         # 订单管理
```

### 信件系统整合

**当前结构：**
```
/(main)/write/       # 写信
/(main)/deliver/     # 发送（实际是选择收件人）
/(main)/mailbox/     # 信箱
/(main)/read/[code]/ # 阅读
```

**新结构：**
```
/(main)/letters/
├── write/          # 写信
├── send/           # 发送（原deliver）
├── inbox/          # 收件箱（原mailbox）
├── sent/           # 已发送
└── read/[code]/    # 阅读
```

### 其他整合

1. **设置页面**：删除顶层 `/settings/`，保留 `/(main)/settings/`
2. **测试页面**：移到 `/dev/` 目录下
3. **博物馆**：删除顶层 `/museum/`，保留 `/(main)/museum/`

## 2. 后端API对应调整

### 信使API整合
```
# 当前
/api/v1/courier/tasks
/api/v1/courier/scan
/api/v1/courier/building/manage
/api/v1/courier/zone/manage
...

# 新结构
/api/v1/courier/tasks        # 包含扫描功能
/api/v1/courier/management/hierarchy
/api/v1/courier/management/opcode
/api/v1/courier/profile/growth
...
```

### 商城API整合
```
# 当前
/api/v1/shop/products
/api/v1/credit-shop/products

# 新结构
/api/v1/shop/products?type=regular
/api/v1/shop/products?type=credit
```

## 3. 重定向配置

为保持向后兼容，将在 `middleware.ts` 中添加重定向规则：

```typescript
const redirects = {
  '/mailbox': '/letters/inbox',
  '/deliver': '/letters/send',
  '/courier/scan': '/courier/tasks',
  '/courier/building-manage': '/courier/management/hierarchy',
  '/courier/opcode-manage': '/courier/management/opcode',
  '/credit-shop': '/shop?type=credit',
  // ... 更多重定向
}
```

## 4. 实施步骤

1. **Phase 1**: 创建新路由结构和组件
2. **Phase 2**: 迁移现有功能到新路由
3. **Phase 3**: 更新后端API
4. **Phase 4**: 添加重定向规则
5. **Phase 5**: 清理旧路由文件

## 5. 风险和注意事项

1. 确保所有链接更新
2. 更新测试用例
3. 更新文档
4. 通知用户路由变更
5. 监控404错误