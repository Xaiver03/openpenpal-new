# OpenPenPal 启动脚本索引

本文档列出了所有可用的启动脚本和工具。

## 主要启动脚本

### 新统一启动系统 (推荐)

| 脚本 | 描述 | 用法 |
|------|------|------|
| `启动 OpenPenPal 集成.command` | macOS主启动器，提供多种选择 | 双击运行或 `./启动\ OpenPenPal\ 集成.command` |
| `startup/openpenpal-launcher.command` | 图形化启动菜单 | 双击运行或 `./startup/openpenpal-launcher.command` |
| `startup/quick-start.sh` | 一键启动脚本 | `./startup/quick-start.sh [模式]` |

### 启动模式

- **development**: 开发模式，完整微服务环境
- **production**: 生产模式，包含管理后台
- **simple**: 简化模式，最小服务集
- **demo**: 演示模式，自动打开浏览器

### 管理工具

| 脚本 | 描述 | 用法 |
|------|------|------|
| `startup/stop-all.sh` | 停止所有服务 | `./startup/stop-all.sh [--force]` |
| `startup/check-status.sh` | 检查服务状态 | `./startup/check-status.sh [--detailed]` |
| `startup/install-deps.sh` | 安装项目依赖 | `./startup/install-deps.sh [--force]` |

### 专用启动脚本

| 脚本 | 描述 | 用法 |
|------|------|------|
| `startup/start-simple-mock.sh` | 简化Mock服务 | `./startup/start-simple-mock.sh` |
| `startup/start-integration.sh` | 传统集成模式 | `./startup/start-integration.sh` |

## 快速开始

### 首次运行
```bash
# 1. 安装依赖
./startup/install-deps.sh

# 2. 启动演示模式
./startup/quick-start.sh demo --auto-open
```

### 开发环境
```bash
# 启动开发模式
./startup/quick-start.sh development --auto-open
```

### 检查状态
```bash
# 查看服务状态
./startup/check-status.sh

# 持续监控
./startup/check-status.sh --continuous
```

### 停止服务
```bash
# 正常停止
./startup/stop-all.sh

# 强制停止
./startup/stop-all.sh --force
```

## 配置文件

- `startup/startup-config.json`: 服务配置
- `startup/environment-vars.sh`: 环境变量
- `startup/utils.sh`: 工具函数

## 日志文件

所有日志文件位于 `logs/` 目录：
- `logs/frontend.log`: 前端日志
- `logs/simple-mock.log`: 简化Mock服务日志
- `logs/*.pid`: 进程ID文件

## 兼容性说明

原有的启动脚本已被新系统替代：
- `simple-start.js` → `startup/start-simple-mock.sh`
- `start-integration.sh` → `startup/start-integration.sh`
- `stop-integration.sh` → `startup/stop-all.sh`

如需使用原脚本，仍可通过兼容性包装器访问。

## 故障排查

1. **端口被占用**: 使用 `./startup/stop-all.sh --force` 清理
2. **依赖问题**: 运行 `./startup/install-deps.sh --force --cleanup`
3. **服务启动失败**: 查看 `logs/*.log` 文件
4. **权限问题**: 运行 `chmod +x startup/*.sh`

## 技术支持

- 查看日志: `tail -f logs/*.log`
- 检查进程: `./startup/check-status.sh --detailed`
- 完整重启: `./startup/stop-all.sh && ./startup/quick-start.sh`
