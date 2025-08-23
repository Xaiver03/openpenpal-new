# OpenPenPal 架构图集

本文档包含 OpenPenPal 项目的各类架构图，使用 Mermaid 格式便于直接渲染。

## 1. 产品功能架构图

```mermaid
graph TB
    subgraph "核心业务层"
        A[用户账户体系] --> B[信件管理系统]
        A --> C[信使配送系统]
        B --> D[博物馆展示系统]
        C --> E[OP码定位系统]
        
        B --> F[写作服务]
        B --> G[OCR识别服务]
        
        C --> H[4级信使层级]
        H --> H1[L1楼宇信使]
        H --> H2[L2片区信使]
        H --> H3[L3校区信使]
        H --> H4[L4城市信使]
        
        E --> I[地理编码服务]
        E --> J[隐私保护机制]
    end
    
    subgraph "支撑系统层"
        K[认证授权系统] --> A
        L[信用积分系统] --> A
        L --> B
        L --> C
        
        M[实时通信系统] --> C
        M --> N[WebSocket服务]
        
        O[AI服务集成] --> B
        O --> P[内容审核]
        O --> Q[智能推荐]
    end
    
    subgraph "管理系统层"
        R[管理后台] --> S[用户管理]
        R --> T[信使管理]
        R --> U[内容审核]
        R --> V[数据统计]
        
        S --> A
        T --> C
        U --> B
    end
    
    subgraph "基础设施层"
        W[数据存储层] --> W1[PostgreSQL]
        W --> W2[Redis缓存]
        W --> W3[文件存储]
        
        X[监控告警] --> Y[性能监控]
        X --> Z[日志分析]
        X --> AA[健康检查]
    end
    
    A -.-> W
    B -.-> W
    C -.-> W
    K -.-> W2
    M -.-> W2
```

## 2. 系统架构图

```mermaid
graph TB
    subgraph "前端应用"
        FE[Next.js 14<br/>3000端口]
        ADMIN[管理后台<br/>Vue.js]
    end
    
    subgraph "API网关"
        GW[Gateway<br/>Go - 8000端口]
    end
    
    subgraph "微服务层"
        MS1[主业务服务<br/>Go Gin - 8080]
        MS2[信使服务<br/>Go - 8002]
        MS3[写作服务<br/>Python - 8001]
        MS4[管理服务<br/>Java - 8003]
        MS5[OCR服务<br/>Python - 8004]
    end
    
    subgraph "数据层"
        DB[(PostgreSQL 15<br/>5432端口)]
        CACHE[(Redis<br/>6379端口)]
        FS[文件存储系统]
    end
    
    subgraph "基础设施"
        DOCKER[Docker容器]
        MONITOR[监控系统]
        LOG[日志系统]
    end
    
    FE --> GW
    ADMIN --> MS4
    
    GW --> MS1
    GW --> MS2
    GW --> MS3
    GW --> MS4
    GW --> MS5
    
    MS1 --> DB
    MS1 --> CACHE
    MS2 --> DB
    MS3 --> DB
    MS4 --> DB
    MS5 --> CACHE
    
    MS1 --> FS
    MS3 --> FS
    
    DOCKER -.-> MS1
    DOCKER -.-> MS2
    DOCKER -.-> MS3
    DOCKER -.-> MS4
    DOCKER -.-> MS5
    
    MONITOR -.-> DOCKER
    LOG -.-> DOCKER
```

## 3. 信件生命周期流程图

```mermaid
stateDiagram-v2
    [*] --> 创建草稿: 用户开始写信
    创建草稿 --> 编辑内容: 填写内容
    编辑内容 --> 选择收件人: 完成编辑
    选择收件人 --> 生成信件码: 确认收件人
    生成信件码 --> 分配OP码: 系统处理
    分配OP码 --> 待取件: 创建配送任务
    
    待取件 --> 已接单: 信使接受任务
    已接单 --> 已取件: 扫码取件
    已取件 --> 配送中: 开始配送
    配送中 --> 已送达: 扫码投递
    已送达 --> [*]: 完成
    
    待取件 --> 已取消: 用户取消
    已接单 --> 已取消: 异常取消
    已取消 --> [*]
```

## 4. 4级信使权限层级图

```mermaid
graph TD
    L4[L4 城市总监<br/>全城管理权限] --> L3_1[L3 北大校区信使<br/>PK** 管理权限]
    L4 --> L3_2[L3 清华校区信使<br/>QH** 管理权限]
    L4 --> L3_3[L3 其他校区信使]
    
    L3_1 --> L2_1[L2 北大东区信使<br/>PK1* 管理权限]
    L3_1 --> L2_2[L2 北大西区信使<br/>PK2* 管理权限]
    
    L2_1 --> L1_1[L1 北大1号楼信使<br/>PK11 配送权限]
    L2_1 --> L1_2[L1 北大2号楼信使<br/>PK12 配送权限]
    
    L2_2 --> L1_3[L1 北大5号楼信使<br/>PK25 配送权限]
    L2_2 --> L1_4[L1 北大6号楼信使<br/>PK26 配送权限]
    
    style L4 fill:#f9f,stroke:#333,stroke-width:4px
    style L3_1 fill:#bbf,stroke:#333,stroke-width:2px
    style L3_2 fill:#bbf,stroke:#333,stroke-width:2px
    style L2_1 fill:#dfd,stroke:#333,stroke-width:2px
    style L2_2 fill:#dfd,stroke:#333,stroke-width:2px
```

## 5. OP码编码系统结构图

```mermaid
graph LR
    subgraph "OP码结构 AABBCC"
        A[AA<br/>学校代码] --> A1[PK - 北京大学]
        A --> A2[QH - 清华大学]
        A --> A3[BD - 北京交通大学]
        
        B[BB<br/>区域代码] --> B1[5F - 5号楼]
        B --> B2[3D - 3号食堂]
        B --> B3[2G - 2号门]
        
        C[CC<br/>位置代码] --> C1[3D - 303室]
        C --> C2[1A - 1区A座]
        C --> C3[12 - 12号位]
    end
    
    D[示例: PK5F3D] --> E[北京大学 5号楼 303室]
    
    F[隐私级别] --> F1[完整显示: PK5F3D]
    F --> F2[部分隐私: PK5F**]
    F --> F3[高度隐私: PK****]
```

## 6. 数据流架构图

```mermaid
graph TB
    subgraph "数据采集层"
        DC1[用户行为数据]
        DC2[信件数据]
        DC3[信使配送数据]
        DC4[系统日志数据]
    end
    
    subgraph "数据处理层"
        DP1[实时流处理]
        DP2[批处理ETL]
        DP3[数据清洗]
        DP4[数据聚合]
    end
    
    subgraph "数据存储层"
        DS1[(PostgreSQL<br/>业务数据)]
        DS2[(Redis<br/>缓存数据)]
        DS3[(文件系统<br/>附件数据)]
        DS4[(日志存储<br/>系统日志)]
    end
    
    subgraph "数据服务层"
        DA1[查询服务]
        DA2[分析服务]
        DA3[推荐服务]
        DA4[报表服务]
    end
    
    subgraph "数据应用层"
        APP1[业务大屏]
        APP2[运营报表]
        APP3[用户画像]
        APP4[智能推荐]
    end
    
    DC1 --> DP1
    DC2 --> DP1
    DC3 --> DP2
    DC4 --> DP3
    
    DP1 --> DS2
    DP2 --> DS1
    DP3 --> DS4
    DP4 --> DS1
    
    DS1 --> DA1
    DS2 --> DA1
    DS1 --> DA2
    DS3 --> DA1
    
    DA1 --> APP1
    DA2 --> APP2
    DA3 --> APP4
    DA4 --> APP2
```

## 7. 安全架构图

```mermaid
graph TB
    subgraph "访问层"
        U1[普通用户]
        U2[信使用户]
        U3[管理员]
    end
    
    subgraph "网关层"
        GW1[API网关<br/>速率限制]
        GW2[WAF<br/>攻击防护]
        GW3[负载均衡]
    end
    
    subgraph "认证授权层"
        AUTH1[JWT认证]
        AUTH2[RBAC权限]
        AUTH3[OAuth2.0]
    end
    
    subgraph "应用安全层"
        SEC1[输入验证]
        SEC2[SQL注入防护]
        SEC3[XSS防护]
        SEC4[CSRF防护]
    end
    
    subgraph "数据安全层"
        DATA1[数据加密]
        DATA2[敏感数据脱敏]
        DATA3[备份恢复]
        DATA4[访问审计]
    end
    
    subgraph "基础安全"
        INFRA1[HTTPS/TLS]
        INFRA2[VPN接入]
        INFRA3[防火墙]
        INFRA4[入侵检测]
    end
    
    U1 --> GW1
    U2 --> GW1
    U3 --> GW2
    
    GW1 --> AUTH1
    GW2 --> AUTH1
    GW3 --> AUTH1
    
    AUTH1 --> SEC1
    AUTH2 --> SEC1
    
    SEC1 --> DATA1
    SEC2 --> DATA1
    SEC3 --> DATA1
    SEC4 --> DATA1
    
    DATA1 -.-> INFRA1
    DATA2 -.-> INFRA1
    DATA3 -.-> INFRA3
    DATA4 -.-> INFRA4
```

## 8. 部署架构图

```mermaid
graph TB
    subgraph "开发环境"
        DEV1[本地开发]
        DEV2[Docker Compose]
        DEV3[热重载]
    end
    
    subgraph "CI/CD流水线"
        CI1[代码提交]
        CI2[自动化测试]
        CI3[代码扫描]
        CI4[构建镜像]
        CI5[推送仓库]
    end
    
    subgraph "测试环境"
        TEST1[功能测试]
        TEST2[集成测试]
        TEST3[性能测试]
        TEST4[安全测试]
    end
    
    subgraph "生产环境"
        subgraph "Region A"
            PROD1[应用服务器组]
            PROD2[数据库主节点]
            PROD3[缓存集群]
        end
        
        subgraph "Region B"
            PROD4[应用服务器组]
            PROD5[数据库从节点]
            PROD6[缓存集群]
        end
        
        LB[负载均衡器]
        CDN[CDN分发]
    end
    
    subgraph "监控运维"
        MON1[Prometheus]
        MON2[Grafana]
        MON3[ELK Stack]
        MON4[告警系统]
    end
    
    DEV1 --> CI1
    CI1 --> CI2
    CI2 --> CI3
    CI3 --> CI4
    CI4 --> CI5
    
    CI5 --> TEST1
    TEST1 --> TEST2
    TEST2 --> TEST3
    TEST3 --> TEST4
    
    TEST4 --> LB
    LB --> PROD1
    LB --> PROD4
    
    CDN --> LB
    
    PROD1 -.-> MON1
    PROD4 -.-> MON1
    MON1 --> MON2
    MON1 --> MON4
```

## 9. 信用系统架构图

```mermaid
graph TB
    subgraph "信用获取途径"
        EARN1[写信获得]
        EARN2[收信获得]
        EARN3[任务完成]
        EARN4[活动奖励]
        EARN5[充值获得]
    end
    
    subgraph "信用账户体系"
        ACC1[总积分]
        ACC2[可用积分]
        ACC3[冻结积分]
        ACC4[过期积分]
    end
    
    subgraph "信用使用场景"
        USE1[商城兑换]
        USE2[优先配送]
        USE3[特权功能]
        USE4[转账赠送]
    end
    
    subgraph "信用管理规则"
        RULE1[获取规则]
        RULE2[过期规则]
        RULE3[冻结规则]
        RULE4[转账规则]
    end
    
    subgraph "信用数据表"
        DB1[(credit_accounts)]
        DB2[(credit_transactions)]
        DB3[(credit_activities)]
        DB4[(credit_expiration)]
    end
    
    EARN1 --> ACC1
    EARN2 --> ACC1
    EARN3 --> ACC1
    EARN4 --> ACC1
    EARN5 --> ACC1
    
    ACC1 --> ACC2
    ACC1 --> ACC3
    ACC1 --> ACC4
    
    ACC2 --> USE1
    ACC2 --> USE2
    ACC2 --> USE3
    ACC2 --> USE4
    
    RULE1 --> DB3
    RULE2 --> DB4
    RULE3 --> DB2
    RULE4 --> DB2
    
    DB1 --> ACC1
    DB2 --> ACC2
    DB3 --> EARN4
    DB4 --> ACC4
```

## 10. 技术栈全景图

```mermaid
mindmap
  root((OpenPenPal技术栈))
    前端技术
      Next.js 14
      TypeScript 5.3
      React 18
      Tailwind CSS
      Zustand状态管理
      TanStack Query
    后端技术
      Go语言
        Gin框架
        GORM ORM
        WebSocket
      Python
        FastAPI
        SQLAlchemy
      Java
        Spring Boot
        Spring Security
    数据存储
      PostgreSQL 15
      Redis 7
      本地文件系统
    基础设施
      Docker容器化
      Nginx反向代理
      监控系统
        Prometheus
        Grafana
        Jaeger
      日志系统
        智能日志
        ELK Stack
    安全体系
      JWT认证
      RBAC权限
      HTTPS/TLS
      API网关
    开发工具
      Git版本控制
      GitHub Actions
      ESLint/Prettier
      单元测试框架
```

---

## 架构图使用说明

### 1. 查看方式
- **Markdown编辑器**: 支持Mermaid的编辑器可直接预览
- **在线工具**: 可使用 [Mermaid Live Editor](https://mermaid.live/)
- **VS Code**: 安装Mermaid插件即可预览
- **GitHub**: 直接支持Mermaid渲染

### 2. 导出格式
- **PNG/SVG**: 使用Mermaid工具导出
- **PDF**: 通过浏览器打印功能
- **PPT**: 导出图片后插入演示文档

### 3. 自定义修改
- 修改节点文字: 直接编辑方括号内的内容
- 修改连接关系: 调整箭头方向和类型
- 修改样式: 使用style语句自定义颜色和样式

### 4. 架构图类型说明
- **功能架构图**: 展示产品功能模块关系
- **系统架构图**: 展示技术组件关系
- **流程图**: 展示业务流程和状态转换
- **部署架构图**: 展示系统部署结构
- **数据流图**: 展示数据流转路径

---

*最后更新: 2025-08-21*  
*维护团队: OpenPenPal Architecture Team*