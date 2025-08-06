# OpenPenPal 开发文档

## 📚 目录

- [项目概述](#项目概述)
- [技术架构](#技术架构)
- [开发环境搭建](#开发环境搭建)
- [项目结构详解](#项目结构详解)
- [核心功能开发指南](#核心功能开发指南)
- [组件开发规范](#组件开发规范)
- [API设计](#api设计)
- [数据库设计](#数据库设计)
- [部署指南](#部署指南)
- [开发工具配置](#开发工具配置)

---

## 项目概述

OpenPenPal是一个创新的校园慢社交平台，通过结合实体手写信件和数字追踪技术，为学生提供有温度的书信交流体验。

### 核心价值
- 🏮 **慢节奏社交**：回归传统书信的仪式感
- 📮 **实体触感**：强调手写信件的物理质感
- 🔗 **数字追踪**：现代技术保障投递可靠性
- 🎒 **校园网络**：基于信使系统的物流体系

---

## 技术架构

### 整体系统架构

```
┌─────────────────────────────────────────────────┐
│                Go Backend API                   │
│  (Gin + PostgreSQL + Redis + File Storage)     │
├─────────────────────────────────────────────────┤
│                REST API Layer                   │
│     /api/v1/* endpoints with JWT auth          │
├─────────────────────────────────────────────────┤
│              Next.js Frontend                   │
│        (React + TypeScript + TailwindCSS)      │
└─────────────────────────────────────────────────┘
```

### 前端架构图

```
┌─────────────────────────────────────────────────┐
│                    用户界面层                      │
├─────────────────────────────────────────────────┤
│  写信页面 │ 信箱页面 │ 信使页面 │ 写作广场 │ 设置页面    │
├─────────────────────────────────────────────────┤
│                   组件层                         │
├─────────────────────────────────────────────────┤
│ UI组件 │ 业务组件 │ 布局组件 │ 表单组件 │ 图表组件     │
├─────────────────────────────────────────────────┤
│                  工具&服务层                      │
├─────────────────────────────────────────────────┤
│  状态管理 │ HTTP客户端 │ 工具函数 │ 类型定义        │
├─────────────────────────────────────────────────┤
│                  基础设施层                       │
├─────────────────────────────────────────────────┤
│    Next.js │ React │ TypeScript │ TailwindCSS    │
└─────────────────────────────────────────────────┘
```

### 技术选型理由

| 技术 | 选择理由 |
|------|----------|
| **Go + Gin** | 高性能后端API，并发处理能力强，生态丰富 |
| **Next.js 14** | SSR/SSG支持，优秀的SEO，文件路由系统 |
| **TypeScript** | 类型安全，更好的开发体验和代码维护性 |
| **ShadCN/UI** | 现代化组件库，可定制性强，与Tailwind完美集成 |
| **TailwindCSS** | 原子化CSS，支持纸黄色主题定制 |
| **PostgreSQL** | 关系型数据库，支持复杂查询和事务 |

### 设计系统

#### 颜色主题 - 纸黄色系
- **主色调**: 琥珀色系 (amber) - 营造温暖纸质感
- **背景色**: 温暖纸黄 (#fefcf7) - 模拟信纸质感
- **强调色**: 橙色系 (orange) - 提供视觉层次
- **文字色**: 深琥珀色 (#92400e) - 确保可读性

#### 首页设计架构
1. **Hero Section**: 全屏展示核心价值主张
2. **Feature Highlights**: 四大核心功能引导
3. **Story & Vision**: 用户故事轮播展示
4. **Public Letter Wall**: 精选公开信件展示
5. **Join Us Section**: 信使招募与社区建设

---

## 开发环境搭建

### 环境要求

```bash
Node.js >= 18.17.0
pnpm >= 8.0.0
Git >= 2.30.0
```

### 安装步骤

1. **克隆项目**
```bash
git clone <repository-url>
cd openpenpal
```

2. **安装依赖**
```bash
pnpm install
```

3. **环境变量配置**
```bash
# 复制环境变量模板
cp .env.example .env.local

# 编辑环境变量
vim .env.local
```

4. **启动开发服务器**
```bash
pnpm dev
```

### 环境变量配置

```bash
# .env.local
NEXT_PUBLIC_APP_URL=http://localhost:3000
NEXT_PUBLIC_API_URL=http://localhost:3001

# 微信配置
NEXT_PUBLIC_WECHAT_APP_ID=your_wechat_app_id
WECHAT_APP_SECRET=your_wechat_app_secret

# 数据库配置
DATABASE_URL="postgresql://username:password@localhost:5432/openpenpal"

# 文件存储配置
OSS_ACCESS_KEY_ID=your_oss_access_key
OSS_ACCESS_KEY_SECRET=your_oss_secret
OSS_BUCKET=your_bucket_name
OSS_REGION=your_region

# JWT密钥
JWT_SECRET=your_jwt_secret
```

---

## 项目结构详解

```
openpenpal/
├── backend/                   # Go 后端服务
│   ├── cmd/                   # 主程序入口
│   ├── internal/              # 内部包
│   │   ├── handlers/          # HTTP 处理器
│   │   ├── models/            # 数据模型
│   │   ├── services/          # 业务逻辑
│   │   └── middleware/        # 中间件
│   ├── pkg/                   # 公共包
│   ├── configs/               # 配置文件
│   └── main.go               # 程序入口
├── frontend/                  # Next.js 前端
│   ├── src/
│   │   ├── app/              # Next.js App Router
│   │   │   ├── (main)/       # 主要页面组
│   │   │   │   ├── write/    # 写信页面
│   │   │   │   ├── mailbox/  # 信箱页面
│   │   │   │   ├── profile/  # 个人资料页面
│   │   │   │   └── settings/ # 设置页面
│   │   │   ├── about/        # 关于页面
│   │   │   ├── courier/      # 信使页面
│   │   │   ├── globals.css   # 全局样式
│   │   │   ├── layout.tsx    # 根布局
│   │   │   ├── page.tsx      # 首页
│   │   │   └── not-found.tsx # 404页面
│   ├── components/            # 组件库
│   │   ├── ui/               # 基础UI组件(ShadCN)
│   │   │   ├── button.tsx
│   │   │   ├── input.tsx
│   │   │   ├── card.tsx
│   │   │   └── ...
│   │   ├── pages/            # 页面级组件
│   │   │   ├── write-page.tsx
│   │   │   ├── deliver-page.tsx
│   │   │   └── ...
│   │   ├── layout/           # 布局组件
│   │   │   ├── header.tsx
│   │   │   ├── footer.tsx
│   │   │   └── sidebar.tsx
│   │   ├── forms/            # 表单组件
│   │   │   ├── letter-form.tsx
│   │   │   └── courier-form.tsx
│   │   └── business/         # 业务组件
│   │       ├── letter-editor.tsx
│   │       ├── qr-generator.tsx
│   │       └── status-tracker.tsx
│   ├── hooks/                # 自定义Hooks
│   │   ├── use-letter.ts
│   │   ├── use-courier.ts
│   │   └── use-auth.ts
│   ├── lib/                  # 工具函数库
│   │   ├── utils.ts          # 通用工具函数
│   │   ├── api.ts            # API客户端
│   │   ├── auth.ts           # 认证工具
│   │   ├── qr-code.ts        # 二维码工具
│   │   └── validations.ts    # 表单验证
│   ├── stores/               # 状态管理
│   │   ├── letter-store.ts
│   │   ├── user-store.ts
│   │   └── courier-store.ts
│   ├── types/                # TypeScript类型
│   │   ├── letter.ts
│   │   ├── user.ts
│   │   ├── courier.ts
│   │   └── api.ts
│   └── styles/               # 样式文件
│       ├── globals.css
│       └── components.css
├── public/                   # 静态资源
│   ├── images/
│   ├── icons/
│   └── fonts/
├── docs/                     # 项目文档
│   ├── 开发计划.md
│   ├── 开发文档.md
│   ├── API文档.md
│   └── 部署指南.md
├── tests/                    # 测试文件
│   ├── __tests__/
│   ├── __mocks__/
│   └── setup.ts
├── .env.example              # 环境变量模板
├── .gitignore
├── package.json
├── tsconfig.json
├── tailwind.config.js
├── next.config.js
└── README.md
```

---

## 核心功能开发指南

### 1. 写信功能开发

#### 页面路径
- `/write` - 新建信件
- `/write?reply_to=<code>` - 回信

#### 核心组件

```typescript
// src/components/business/letter-editor.tsx
interface LetterEditorProps {
  initialContent?: string;
  onSave: (content: string) => void;
  onGenerateCode: () => void;
  replyTo?: string;
}

export function LetterEditor({ 
  initialContent, 
  onSave, 
  onGenerateCode, 
  replyTo 
}: LetterEditorProps) {
  // WangEditor集成
  // 样式选择器
  // 自动保存功能
  // 字数统计
}
```

#### 状态管理

```typescript
// src/stores/letter-store.ts
interface LetterStore {
  currentDraft: LetterDraft | null;
  savedDrafts: LetterDraft[];
  
  // 操作方法
  saveDraft: (draft: LetterDraft) => void;
  loadDraft: (id: string) => void;
  generateCode: () => Promise<LetterCode>;
  deleteDraft: (id: string) => void;
}
```

### 2. 二维码生成功能

```typescript
// src/lib/qr-code.ts
export interface QRCodeOptions {
  size: number;
  format: 'png' | 'svg';
  errorCorrectionLevel: 'L' | 'M' | 'Q' | 'H';
}

export async function generateQRCode(
  code: string, 
  options: QRCodeOptions = defaultOptions
): Promise<string> {
  // 二维码生成逻辑
  // 贴纸样式设计
  // 打印优化
}
```

### 3. 信使扫码功能

```typescript
// src/components/business/courier-scanner.tsx
interface CourierScannerProps {
  onScanSuccess: (code: string) => void;
  onScanError: (error: Error) => void;
}

export function CourierScanner({ 
  onScanSuccess, 
  onScanError 
}: CourierScannerProps) {
  // 摄像头调用
  // 二维码识别
  // 手动输入备选
}
```

### 4. 状态追踪系统

```typescript
// src/types/letter.ts
export type LetterStatus = 
  | 'draft'        // 草稿
  | 'generated'    // 已生成编号
  | 'collected'    // 已收取
  | 'in_transit'   // 在途
  | 'delivered'    // 已送达
  | 'read';        // 已查看

export interface LetterStatusUpdate {
  codeId: string;
  status: LetterStatus;
  updatedBy: string;
  timestamp: Date;
  location?: string;
  note?: string;
}
```

---

## 组件开发规范

### 组件分类

1. **UI组件** (`src/components/ui/`)
   - 基础组件，基于ShadCN/UI
   - 无业务逻辑，纯展示
   - 支持主题定制

2. **布局组件** (`src/components/layout/`)
   - 页面结构组件
   - 响应式布局
   - 导航和侧边栏

3. **业务组件** (`src/components/business/`)
   - 包含业务逻辑
   - 与状态管理集成
   - 可复用的功能模块

4. **页面组件** (`src/components/pages/`)
   - 完整的页面组件
   - 组合多个子组件
   - 处理页面级状态

### 组件开发模板

```typescript
// src/components/business/example-component.tsx
import { useState, useCallback } from 'react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';

interface ExampleComponentProps {
  className?: string;
  variant?: 'default' | 'secondary';
  onAction?: () => void;
  children?: React.ReactNode;
}

export function ExampleComponent({
  className,
  variant = 'default',
  onAction,
  children,
  ...props
}: ExampleComponentProps) {
  const [state, setState] = useState(false);

  const handleAction = useCallback(() => {
    setState(!state);
    onAction?.();
  }, [state, onAction]);

  return (
    <div 
      className={cn(
        'relative flex items-center gap-2',
        variant === 'secondary' && 'text-muted-foreground',
        className
      )}
      {...props}
    >
      <Button onClick={handleAction}>
        {children}
      </Button>
    </div>
  );
}
```

### 组件文档规范

```typescript
/**
 * ExampleComponent - 示例组件
 * 
 * @description 用于演示组件开发规范的示例组件
 * 
 * @example
 * ```tsx
 * <ExampleComponent 
 *   variant="secondary"
 *   onAction={() => console.log('clicked')}
 * >
 *   点击我
 * </ExampleComponent>
 * ```
 */
```

---

## API设计

### RESTful API规范

#### 基础URL结构
```
/api/v1/{resource}/{id?}/{action?}
```

#### 核心API端点

```typescript
// 用户认证
POST   /api/v1/auth/login          // 用户登录
POST   /api/v1/auth/logout         // 用户登出
GET    /api/v1/auth/profile        // 获取用户信息

// 信件管理
POST   /api/v1/letters             // 创建信件草稿
GET    /api/v1/letters             // 获取信件列表
GET    /api/v1/letters/:id         // 获取信件详情
PUT    /api/v1/letters/:id         // 更新信件
DELETE /api/v1/letters/:id         // 删除信件

// 编号管理
POST   /api/v1/codes/generate      // 生成编号
GET    /api/v1/codes/:code         // 查询编号信息
PUT    /api/v1/codes/:code/status  // 更新编号状态

// 信使功能
GET    /api/v1/courier/tasks       // 获取信使任务
POST   /api/v1/courier/scan        // 扫码录入
GET    /api/v1/courier/history     // 投递历史

// 文件上传
POST   /api/v1/upload/image        // 上传图片
POST   /api/v1/upload/letter       // 上传信件照片
```

#### API响应格式

```typescript
// 成功响应
{
  "success": true,
  "data": {
    // 实际数据
  },
  "message": "操作成功",
  "timestamp": "2024-01-01T00:00:00Z"
}

// 错误响应
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "参数验证失败",
    "details": {
      "field": "email",
      "message": "邮箱格式不正确"
    }
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

#### API客户端封装

```typescript
// src/lib/api.ts
class APIClient {
  private baseURL: string;
  private token: string | null = null;

  constructor(baseURL: string) {
    this.baseURL = baseURL;
  }

  async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<APIResponse<T>> {
    // HTTP请求封装
    // 错误处理
    // 认证头添加
  }

  // 具体API方法
  letters = {
    create: (data: CreateLetterRequest) => 
      this.request<Letter>('/letters', { method: 'POST', body: JSON.stringify(data) }),
    
    getById: (id: string) => 
      this.request<Letter>(`/letters/${id}`),
    
    updateStatus: (id: string, status: LetterStatus) =>
      this.request<Letter>(`/letters/${id}/status`, { 
        method: 'PUT', 
        body: JSON.stringify({ status }) 
      })
  };
}
```

---

## 数据库设计

### 数据库ER图

```
Users ||--o{ Letters : creates
Users ||--o{ CourierTasks : performs
Letters ||--|| LetterCodes : has
Letters ||--o{ LetterPhotos : contains
LetterCodes ||--o{ StatusLogs : tracks
```

### 核心表结构

```sql
-- 用户表
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  wechat_id VARCHAR(100) UNIQUE,
  nickname VARCHAR(50) NOT NULL,
  avatar_url TEXT,
  role VARCHAR(20) DEFAULT 'user', -- user, courier, admin
  school_code VARCHAR(20),
  status VARCHAR(20) DEFAULT 'active',
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- 信件表
CREATE TABLE letters (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id),
  title VARCHAR(200),
  content TEXT,
  style VARCHAR(50) DEFAULT 'default',
  reply_to UUID REFERENCES letters(id),
  status VARCHAR(20) DEFAULT 'draft',
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- 信件编号表
CREATE TABLE letter_codes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  letter_id UUID REFERENCES letters(id),
  code VARCHAR(50) UNIQUE NOT NULL,
  qr_code_url TEXT,
  generated_at TIMESTAMP DEFAULT NOW(),
  expires_at TIMESTAMP
);

-- 状态日志表
CREATE TABLE status_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code_id UUID REFERENCES letter_codes(id),
  status VARCHAR(20) NOT NULL,
  updated_by UUID REFERENCES users(id),
  location VARCHAR(200),
  note TEXT,
  created_at TIMESTAMP DEFAULT NOW()
);

-- 信使任务表
CREATE TABLE courier_tasks (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  courier_id UUID REFERENCES users(id),
  code_id UUID REFERENCES letter_codes(id),
  task_type VARCHAR(20), -- collect, deliver
  status VARCHAR(20) DEFAULT 'pending',
  assigned_at TIMESTAMP DEFAULT NOW(),
  completed_at TIMESTAMP
);
```

### Prisma Schema

```prisma
// prisma/schema.prisma
generator client {
  provider = "prisma-client-js"
}

datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
}

model User {
  id          String   @id @default(cuid())
  wechatId    String?  @unique @map("wechat_id")
  nickname    String
  avatarUrl   String?  @map("avatar_url")
  role        Role     @default(USER)
  schoolCode  String?  @map("school_code")
  status      Status   @default(ACTIVE)
  createdAt   DateTime @default(now()) @map("created_at")
  updatedAt   DateTime @updatedAt @map("updated_at")

  letters      Letter[]
  courierTasks CourierTask[]
  statusLogs   StatusLog[]

  @@map("users")
}

enum Role {
  USER
  COURIER
  ADMIN
}

enum Status {
  ACTIVE
  INACTIVE
  BANNED
}
```

---

## 部署指南

### 开发环境部署

```bash
# 1. 安装依赖
pnpm install

# 2. 配置环境变量
cp .env.example .env.local

# 3. 启动数据库(Docker)
docker-compose up -d postgres

# 4. 运行数据库迁移
pnpm prisma migrate dev

# 5. 启动开发服务器
pnpm dev
```

### 生产环境部署

#### Docker部署

```dockerfile
# Dockerfile
FROM node:18-alpine AS base

# Install dependencies only when needed
FROM base AS deps
RUN apk add --no-cache libc6-compat
WORKDIR /app

COPY package.json pnpm-lock.yaml ./
RUN npm install -g pnpm && pnpm install --frozen-lockfile

# Rebuild the source code only when needed
FROM base AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .

RUN npm install -g pnpm && pnpm run build

# Production image
FROM base AS runner
WORKDIR /app

ENV NODE_ENV production

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

EXPOSE 3000

ENV PORT 3000

CMD ["node", "server.js"]
```

#### Vercel部署

```bash
# 1. 安装Vercel CLI
npm i -g vercel

# 2. 登录Vercel
vercel login

# 3. 部署项目
vercel --prod
```

### CI/CD配置

```yaml
# .github/workflows/deploy.yml
name: Deploy to Production

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '18'
          
      - name: Install pnpm
        uses: pnpm/action-setup@v2
        with:
          version: 8
          
      - name: Install dependencies
        run: pnpm install --frozen-lockfile
        
      - name: Run tests
        run: pnpm test
        
      - name: Build application
        run: pnpm build
        
      - name: Deploy to Vercel
        uses: amondnet/vercel-action@v20
        with:
          vercel-token: ${{ secrets.VERCEL_TOKEN }}
          vercel-org-id: ${{ secrets.VERCEL_ORG_ID }}
          vercel-project-id: ${{ secrets.VERCEL_PROJECT_ID }}
          vercel-args: '--prod'
```

---

## 开发工具配置

### VS Code配置

```json
// .vscode/settings.json
{
  "typescript.preferences.importModuleSpecifier": "relative",
  "typescript.preferences.includePackageJsonAutoImports": "auto",
  "editor.formatOnSave": true,
  "editor.defaultFormatter": "esbenp.prettier-vscode",
  "editor.codeActionsOnSave": {
    "source.fixAll.eslint": true
  },
  "emmet.includeLanguages": {
    "typescript": "html",
    "typescriptreact": "html"
  },
  "tailwindCSS.includeLanguages": {
    "typescript": "html",
    "typescriptreact": "html"
  }
}
```

### ESLint配置

```javascript
// .eslintrc.js
module.exports = {
  extends: [
    'next/core-web-vitals',
    '@typescript-eslint/recommended',
    'prettier'
  ],
  parser: '@typescript-eslint/parser',
  plugins: ['@typescript-eslint'],
  rules: {
    '@typescript-eslint/no-unused-vars': 'error',
    '@typescript-eslint/no-explicit-any': 'warn',
    'prefer-const': 'error',
    'no-console': 'warn'
  }
};
```

### Prettier配置

```json
// .prettierrc
{
  "semi": true,
  "trailingComma": "es5",
  "singleQuote": true,
  "printWidth": 80,
  "tabWidth": 2,
  "useTabs": false
}
```

---

## 测试策略

### 测试分层

1. **单元测试** - Jest + Testing Library
2. **集成测试** - API路由测试
3. **端到端测试** - Playwright
4. **组件测试** - Storybook

### 测试配置

```javascript
// jest.config.js
const nextJest = require('next/jest');

const createJestConfig = nextJest({
  dir: './',
});

const customJestConfig = {
  setupFilesAfterEnv: ['<rootDir>/tests/setup.ts'],
  moduleNameMapping: {
    '^@/(.*)$': '<rootDir>/src/$1',
  },
  testEnvironment: 'jest-environment-jsdom',
};

module.exports = createJestConfig(customJestConfig);
```

---

## 性能优化

### 代码分割

```typescript
// 页面级代码分割
const WritePage = dynamic(() => import('@/components/pages/write-page'), {
  loading: () => <PageSkeleton />,
  ssr: false
});

// 组件级代码分割
const LetterEditor = dynamic(() => import('@/components/business/letter-editor'), {
  loading: () => <EditorSkeleton />
});
```

### 图片优化

```typescript
// 使用Next.js Image组件
import Image from 'next/image';

<Image
  src="/images/letter-paper.jpg"
  alt="信纸背景"
  width={800}
  height={600}
  placeholder="blur"
  blurDataURL="data:image/jpeg;base64,..."
/>
```

### 缓存策略

```typescript
// SWR数据获取
import useSWR from 'swr';

function useLetters() {
  const { data, error } = useSWR('/api/letters', fetcher, {
    revalidateOnFocus: false,
    dedupingInterval: 10000
  });

  return {
    letters: data,
    isLoading: !error && !data,
    isError: error
  };
}
```

---

这份开发文档涵盖了OpenPenPal项目的完整开发指南，包括架构设计、开发规范、API设计、数据库设计、部署指南等关键内容。开发团队可以根据这份文档进行高效的协作开发。