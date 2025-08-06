#!/bin/bash

# Docker Desktop for Mac 路径设置脚本

echo "🐳 设置Docker Desktop路径..."

# Docker命令的常见位置
DOCKER_PATHS=(
    "/Applications/Docker.app/Contents/Resources/bin"
    "$HOME/.docker/bin"
    "/usr/local/bin"
)

# 查找Docker安装
DOCKER_FOUND=false
for path in "${DOCKER_PATHS[@]}"; do
    if [ -f "$path/docker" ]; then
        echo "✅ 找到Docker: $path"
        DOCKER_FOUND=true
        
        # 检查是否已在PATH中
        if echo $PATH | grep -q "$path"; then
            echo "   Docker路径已在PATH中"
        else
            echo "   添加到PATH..."
            export PATH="$path:$PATH"
            
            # 提示用户永久添加
            echo ""
            echo "💡 要永久添加到PATH，请将以下行添加到 ~/.zshrc 或 ~/.bash_profile:"
            echo ""
            echo "   export PATH=\"$path:\$PATH\""
            echo ""
        fi
        break
    fi
done

if [ "$DOCKER_FOUND" = false ]; then
    echo "❌ 未找到Docker Desktop"
    echo "   请从 https://www.docker.com/products/docker-desktop 下载并安装"
    exit 1
fi

# 检查Docker是否运行
if docker info &> /dev/null; then
    echo "✅ Docker正在运行"
    docker version --format "   版本: {{.Server.Version}}"
else
    echo "❌ Docker未运行"
    echo "   请启动Docker Desktop应用"
    echo ""
    echo "   尝试启动Docker Desktop..."
    open -a Docker
    echo "   请等待Docker完全启动后再运行脚本"
fi

echo ""
echo "🎯 Docker设置完成！"