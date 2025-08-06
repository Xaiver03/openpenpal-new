# BrowserMCP 安装配置指南

## 🚀 安装状态

✅ **已成功安装以下 BrowserMCP 工具**:

1. **@browsermcp/mcp** (v0.1.3)
   - 命令: `mcp-server-browsermcp`
   - 用途: 基础浏览器自动化 MCP 服务器

2. **@playwright/mcp** (v0.0.31)
   - 命令: `mcp-server-playwright`
   - 用途: Playwright 集成的浏览器自动化

3. **any-browser-mcp** (已安装)
   - 命令: `any-browser-mcp`
   - 用途: 附加到现有浏览器会话

## 🔧 环境变量配置

已添加到 `~/.zshrc` 和 `~/.bash_profile`:

```bash
# BrowserMCP Environment Variables
export BROWSERMCP_SERVER="mcp-server-browsermcp"
export PLAYWRIGHT_MCP_SERVER="mcp-server-playwright"
export ANY_BROWSER_MCP="any-browser-mcp"

# NPM Global Bin Path
export PATH="/Users/rocalight/.npm-global/bin:$PATH"
```

## 📦 安装路径

全局 npm 包安装在: `/Users/rocalight/.npm-global/`

可执行文件位于: `/Users/rocalight/.npm-global/bin/`

## 🎯 使用方法

### 1. BrowserMCP 基础服务器
```bash
mcp-server-browsermcp --help
```

### 2. Playwright MCP 服务器
```bash
mcp-server-playwright --help
```

### 3. Any Browser MCP
```bash
any-browser-mcp --help
```

## 🔄 重新加载环境变量

如需在当前终端会话中使用，请执行:

```bash
source ~/.zshrc
# 或者
source ~/.bash_profile
```

## ✅ 验证安装

运行以下命令验证安装:

```bash
echo "Available browser MCP tools:"
echo "1. $BROWSERMCP_SERVER"
echo "2. $PLAYWRIGHT_MCP_SERVER" 
echo "3. $ANY_BROWSER_MCP"
```

## 📚 下一步

现在你可以:
1. 在 Claude Code 中使用这些 MCP 工具进行浏览器自动化
2. 集成到前端测试流程中
3. 用于自动化测试和质量保证

所有工具已成功安装并配置到全局环境变量中！ 🎉