# mcpbrowser 简化使用指南

## 🚀 快速启动

现在你可以通过以下简单命令使用 BrowserMCP：

### 基础命令
```bash
mcpbrowser          # 启动默认 BrowserMCP 服务器
mcp                 # 相同功能 (简化别名)
```

### 完整命令参考

#### 📋 查看帮助和信息
```bash
mcpbrowser --help         # 显示完整帮助信息
mcpbrowser --version      # 显示版本信息
mcpbrowser list          # 列出所有可用的 MCP 工具
mcpbrowser test          # 测试浏览器连接和工具可用性
```

#### 🔧 启动不同的 MCP 服务器
```bash
# 启动基础 BrowserMCP 服务器 (默认)
mcpbrowser start
mcpbrowser              # 简化写法

# 启动 Playwright MCP 服务器
mcpbrowser playwright

# 启动 Any-Browser MCP (附加到现有浏览器)
mcpbrowser any
```

#### ⚙️ 高级选项
```bash
# 指定端口
mcpbrowser --port 3002
mcpbrowser playwright --port 3003

# 启用调试模式
mcpbrowser --debug
mcpbrowser playwright --debug

# 组合选项
mcpbrowser any --port 3004 --debug
```

#### 💡 简化别名
```bash
mcp              # 等同于 mcpbrowser
mcp test         # 等同于 mcpbrowser test
mcp playwright   # 等同于 mcpbrowser playwright
mcp list         # 等同于 mcpbrowser list
```

## 🎯 常用场景

### 1. 快速测试系统
```bash
mcp test
```
输出示例：
```
🧪 测试浏览器 MCP 连接...

✅ Chrome 浏览器正在运行
✅ 端口 3001 可用

🔧 MCP 工具可用性检查:
✅ mcp-server-browsermcp 已安装
✅ mcp-server-playwright 已安装
✅ any-browser-mcp 已安装
```

### 2. 启动开发环境的浏览器自动化
```bash
# 终端 1: 启动前端开发服务器
cd /Users/rocalight/同步空间/opplc/openpenpal/frontend
npm run dev

# 终端 2: 启动浏览器 MCP
mcp playwright --debug
```

### 3. 快速查看工具状态
```bash
mcp list
```

### 4. 连接到现有浏览器会话
```bash
# 先手动打开 Chrome 浏览器
# 然后运行：
mcp any
```

## 🔧 故障排除

### 常见问题和解决方案

#### 1. 命令未找到
```bash
# 重新加载环境变量
source ~/.zshrc
# 或者
source ~/.bash_profile
```

#### 2. 端口被占用
```bash
# 检查端口占用
mcp test

# 使用不同端口
mcp --port 3002
```

#### 3. Chrome 浏览器未运行
```bash
# 测试会提示
mcp test

# 手动启动 Chrome 后再运行 MCP
```

#### 4. 权限问题
```bash
# 确保脚本可执行
chmod +x /Users/rocalight/.npm-global/bin/mcpbrowser
```

## 📚 详细说明

### BrowserMCP 工具说明

1. **mcp-server-browsermcp**
   - 基础浏览器自动化服务器
   - 适合简单的浏览器操作
   - 启动: `mcp start` 或 `mcp`

2. **mcp-server-playwright**
   - Playwright 集成的浏览器自动化
   - 功能更强大，支持多浏览器
   - 启动: `mcp playwright`

3. **any-browser-mcp**
   - 附加到现有浏览器会话
   - 无需重新启动浏览器
   - 启动: `mcp any`

### 环境配置

脚本位置: `/Users/rocalight/.npm-global/bin/mcpbrowser`

环境变量已配置在:
- `~/.zshrc`
- `~/.bash_profile`

别名配置:
```bash
alias mcp="mcpbrowser"
```

## 🎉 使用示例

### OpenPenPal 前端测试流程
```bash
# 1. 检查系统状态
mcp test

# 2. 启动前端开发服务器 (另一个终端)
cd /Users/rocalight/同步空间/opplc/openpenpal/frontend && npm run dev

# 3. 启动浏览器自动化 (当前终端)
mcp playwright --debug

# 4. 在 Claude Code 中使用 MCP 工具进行自动化测试
```

### 快速调试
```bash
# 启动调试模式的基础 MCP
mcp --debug

# 在另一个终端检查状态
mcp test
```

现在你可以简单地输入 `mcpbrowser` 或 `mcp` 来使用所有 BrowserMCP 功能！ 🎯