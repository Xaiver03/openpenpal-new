# React Hydration Error 修复指南

## 最新修复记录 (2025-08-20)

### 修复原则
1. **禁止硬编码** - 绝不使用硬编码日期如 `'2024-01-01T00:00:00Z'`
2. **使用SafeTimestamp组件** - 所有时间显示都使用SafeTimestamp组件
3. **深入分析问题** - 不简化问题，找到根本原因

### 已修复的组件

#### 1. 评论组件 (`comment-item.tsx`)
- **问题**: 使用 `formatDistanceToNow` 直接渲染相对时间
- **修复**: 替换为 `<SafeTimestamp date={comment.created_at} format="relative" />`

#### 2. 通知组件 (`notification-item.tsx`)
- **问题**: 使用自定义 `formatDistanceToNow` 函数
- **修复**: 使用 `<SafeTimestamp>` 组件

#### 3. 实时通信组件
- **websocket-status.tsx**: 修复连接时间显示
- **notification-center.tsx**: 修复通知时间显示
- **user-activity-feed.tsx**: 修复活动时间显示
- **PromotionManagement.tsx**: 修复申请时间显示

#### 4. 信使管理页面
- **ManagementPageLayout.tsx**: 
  - 问题：使用 `toLocaleDateString()` 和 `toLocaleString()`
  - 修复：使用 `<SafeTimestamp>` 组件，支持 undefined 值
- **city-manage/school-manage/zone-manage**: 
  - 移除硬编码日期，使用 undefined 作为默认值

#### 5. 博物馆页面
- **museum/page.tsx**: 已在之前修复
- **museum/entries/[id]/page.tsx**: 修复评论时间显示
- **courier/scan/page.tsx**: 修复扫描历史时间显示

### 关键学习点

1. **不要硬编码时间**：
   ```tsx
   // ❌ 错误
   joinDate: courier.createdAt || '2024-01-01T00:00:00Z'
   
   // ✅ 正确
   joinDate: courier.createdAt || undefined
   ```

2. **使用SafeTimestamp处理所有时间显示**：
   ```tsx
   // ❌ 错误
   <span>{new Date(date).toLocaleDateString()}</span>
   
   // ✅ 正确
   <SafeTimestamp date={date} format="locale" fallback="--" />
   ```

3. **处理可选值**：
   ```tsx
   {joinDate ? (
     <SafeTimestamp date={joinDate} format="locale" fallback="--" />
   ) : (
     <span>--</span>
   )}
   ```

## 错误信息
```
Error: Text content does not match server-rendered HTML.
Text content did not match. Server: "2025-08-20T14:11:22.285Z" Client: "2025-08-20T14:11:22.983Z"
```

## 问题原因
这个错误是因为服务器端渲染（SSR）和客户端渲染（CSR）的内容不一致。常见原因包括：

1. **直接使用 `new Date()` 或 `Date.now()`** - 服务器和客户端执行时间不同
2. **使用 `Math.random()`** - 每次执行结果不同
3. **条件渲染依赖于 `typeof window`** - 服务器端没有 window 对象
4. **读取 localStorage/sessionStorage** - 服务器端无法访问
5. **用户特定数据** - 如时区、语言设置等

## 解决方案

### 1. 使用 SafeTimestamp 组件

对于需要显示时间戳的地方，使用 `SafeTimestamp` 组件：

```tsx
import { SafeTimestamp } from '@/components/ui/safe-timestamp'

// 之前（会导致 hydration error）
<span>{new Date().toISOString()}</span>

// 之后（安全）
<SafeTimestamp date={new Date()} format="iso" fallback="Loading..." />
```

### 2. 使用 useStableTimestamp Hook

对于需要在组件中使用时间戳的场景：

```tsx
import { useStableTimestamp } from '@/hooks/use-stable-timestamp'

function MyComponent() {
  const timestamp = useStableTimestamp()
  
  return <div>Created at: {timestamp || 'Loading...'}</div>
}
```

### 3. 客户端专用组件

对于只需要在客户端渲染的内容：

```tsx
'use client'

import { useState, useEffect } from 'react'

function ClientOnlyTime() {
  const [mounted, setMounted] = useState(false)
  const [time, setTime] = useState('')

  useEffect(() => {
    setMounted(true)
    setTime(new Date().toLocaleString())
  }, [])

  if (!mounted) {
    return <span>--</span> // 或者返回加载占位符
  }

  return <span>{time}</span>
}
```

### 4. 使用 suppressHydrationWarning

如果确实需要不同的内容，可以抑制警告：

```tsx
<div suppressHydrationWarning>
  {new Date().toISOString()}
</div>
```

**注意**：这只是隐藏警告，不是真正的解决方案。应该谨慎使用。

## 最佳实践

1. **避免在渲染时生成动态内容**
   - 不要在组件体内直接调用 `new Date()`
   - 不要使用 `Math.random()` 生成 ID

2. **使用 useEffect 处理客户端特定逻辑**
   ```tsx
   useEffect(() => {
     // 客户端特定的代码
   }, [])
   ```

3. **为 SSR 提供默认值**
   ```tsx
   const [data, setData] = useState('Loading...') // 默认值
   
   useEffect(() => {
     setData(new Date().toISOString()) // 客户端更新
   }, [])
   ```

4. **使用稳定的 ID 生成器**
   ```tsx
   import { useId } from 'react' // React 18+
   
   function Component() {
     const id = useId() // 在 SSR 和 CSR 中保持一致
     return <div id={id}>...</div>
   }
   ```

## 调试技巧

1. **查看具体不匹配的内容**
   - 浏览器控制台会显示服务器和客户端的差异
   - 搜索时间戳格式、随机数等

2. **使用 React DevTools**
   - 检查组件树
   - 查看哪个组件导致了不匹配

3. **临时禁用 SSR**
   ```tsx
   import dynamic from 'next/dynamic'
   
   const DynamicComponent = dynamic(
     () => import('./MyComponent'),
     { ssr: false }
   )
   ```

## 常见场景修复

### 显示当前时间
```tsx
// ❌ 错误
<div>当前时间：{new Date().toLocaleString()}</div>

// ✅ 正确
import { CurrentTime } from '@/components/ui/safe-timestamp'
<div>当前时间：<CurrentTime format="locale" /></div>
```

### 显示相对时间
```tsx
// ❌ 错误
<div>{formatDistanceToNow(new Date(createdAt))}</div>

// ✅ 正确
<SafeTimestamp date={createdAt} format="relative" fallback="刚刚" />
```

### 生成唯一 ID
```tsx
// ❌ 错误
const id = `item-${Math.random()}`

// ✅ 正确
import { useId } from 'react'
const id = useId()
```

## 总结

Hydration 错误通常是因为服务器和客户端渲染了不同的内容。解决方案是：

1. 确保初始渲染内容一致
2. 将动态内容延迟到客户端渲染
3. 为 SSR 提供稳定的默认值
4. 使用专门的组件处理时间、随机数等动态内容

遵循这些原则可以避免大多数 hydration 错误。