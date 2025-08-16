#!/bin/bash
# 腾讯云服务器初始化脚本

set -euo pipefail

# 配置
DOMAIN=${1:-"openpenpal.com"}
EMAIL=${2:-"admin@openpenpal.com"}
REGION="ap-guangzhou"

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}✓${NC} $1"
}

warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# 更新系统
update_system() {
    log "更新系统软件包..."
    
    # 配置腾讯云镜像源
    cat > /etc/apt/sources.list << EOF
deb http://mirrors.tencentyun.com/ubuntu/ focal main restricted universe multiverse
deb http://mirrors.tencentyun.com/ubuntu/ focal-security main restricted universe multiverse
deb http://mirrors.tencentyun.com/ubuntu/ focal-updates main restricted universe multiverse
deb http://mirrors.tencentyun.com/ubuntu/ focal-backports main restricted universe multiverse
EOF
    
    apt-get update
    apt-get upgrade -y
    apt-get install -y \
        curl \
        wget \
        git \
        vim \
        htop \
        iotop \
        net-tools \
        unzip \
        jq \
        bc \
        python3-pip \
        software-properties-common
    
    success "系统更新完成"
}

# 配置防火墙
setup_firewall() {
    log "配置防火墙规则..."
    
    # 安装 ufw
    apt-get install -y ufw
    
    # 默认规则
    ufw default deny incoming
    ufw default allow outgoing
    
    # 允许 SSH
    ufw allow 22/tcp
    
    # 允许 HTTP 和 HTTPS
    ufw allow 80/tcp
    ufw allow 443/tcp
    
    # 允许监控端口（仅内网）
    ufw allow from 10.0.0.0/8 to any port 9090  # Prometheus
    ufw allow from 10.0.0.0/8 to any port 3002  # Grafana
    
    # 启用防火墙
    ufw --force enable
    
    success "防火墙配置完成"
}

# 安装 Docker
install_docker() {
    log "安装 Docker..."
    
    # 删除旧版本
    apt-get remove -y docker docker-engine docker.io containerd runc || true
    
    # 安装依赖
    apt-get install -y \
        apt-transport-https \
        ca-certificates \
        gnupg \
        lsb-release
    
    # 添加 Docker 官方 GPG 密钥
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
    
    # 设置 Docker 仓库
    echo \
        "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
        $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
    
    # 安装 Docker
    apt-get update
    apt-get install -y docker-ce docker-ce-cli containerd.io
    
    # 配置 Docker
    mkdir -p /etc/docker
    cat > /etc/docker/daemon.json << EOF
{
    "registry-mirrors": [
        "https://mirror.ccs.tencentyun.com"
    ],
    "log-driver": "json-file",
    "log-opts": {
        "max-size": "100m",
        "max-file": "3"
    },
    "storage-driver": "overlay2",
    "storage-opts": [
        "overlay2.override_kernel_check=true"
    ]
}
EOF
    
    # 启动 Docker
    systemctl restart docker
    systemctl enable docker
    
    # 安装 Docker Compose
    curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
    
    success "Docker 安装完成"
}

# 安装 Nginx
install_nginx() {
    log "安装 Nginx..."
    
    apt-get install -y nginx
    
    # 基础配置
    cat > /etc/nginx/nginx.conf << 'EOF'
user www-data;
worker_processes auto;
pid /run/nginx.pid;
include /etc/nginx/modules-enabled/*.conf;

events {
    worker_connections 2048;
    use epoll;
    multi_accept on;
}

http {
    # 基础设置
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;
    server_tokens off;
    client_max_body_size 100M;

    # MIME 类型
    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    # SSL 设置
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;

    # 日志设置
    access_log /var/log/nginx/access.log;
    error_log /var/log/nginx/error.log;

    # Gzip 压缩
    gzip on;
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types text/plain text/css text/xml text/javascript application/json application/javascript application/xml+rss application/rss+xml application/atom+xml image/svg+xml;

    # 安全头
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;

    # 包含站点配置
    include /etc/nginx/conf.d/*.conf;
    include /etc/nginx/sites-enabled/*;
}
EOF
    
    systemctl restart nginx
    systemctl enable nginx
    
    success "Nginx 安装完成"
}

# 安装 SSL 证书
setup_ssl() {
    log "配置 SSL 证书..."
    
    # 安装 Certbot
    apt-get install -y certbot python3-certbot-nginx
    
    # 申请证书
    certbot --nginx -d "$DOMAIN" -d "www.$DOMAIN" \
        --non-interactive \
        --agree-tos \
        --email "$EMAIL" \
        --redirect
    
    # 设置自动续期
    cat > /etc/systemd/system/certbot-renewal.service << EOF
[Unit]
Description=Certbot Renewal
After=network.target

[Service]
Type=oneshot
ExecStart=/usr/bin/certbot renew --quiet --deploy-hook "systemctl reload nginx"
EOF

    cat > /etc/systemd/system/certbot-renewal.timer << EOF
[Unit]
Description=Run Certbot Renewal twice daily

[Timer]
OnCalendar=*-*-* 00,12:00:00
RandomizedDelaySec=3600
Persistent=true

[Install]
WantedBy=timers.target
EOF
    
    systemctl daemon-reload
    systemctl enable certbot-renewal.timer
    systemctl start certbot-renewal.timer
    
    success "SSL 证书配置完成"
}

# 安装腾讯云 CLI 和 COS 工具
install_tencent_tools() {
    log "安装腾讯云工具..."
    
    # 安装腾讯云 CLI
    pip3 install tccli
    
    # 安装 COS CLI
    wget https://github.com/tencentyun/coscli/releases/latest/download/coscli-linux -O /usr/local/bin/coscli
    chmod +x /usr/local/bin/coscli
    
    # 安装云监控 Agent
    wget -O /tmp/install-monitor.sh https://cloud-monitor-1258344699.cos.ap-guangzhou.myqcloud.com/install.sh
    bash /tmp/install-monitor.sh
    
    success "腾讯云工具安装完成"
}

# 配置系统优化
optimize_system() {
    log "优化系统配置..."
    
    # 内核参数优化
    cat >> /etc/sysctl.conf << EOF

# 网络优化
net.core.somaxconn = 65535
net.ipv4.tcp_max_syn_backlog = 65535
net.ipv4.tcp_max_tw_buckets = 65535
net.ipv4.tcp_syncookies = 1
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_fin_timeout = 30
net.ipv4.tcp_keepalive_time = 1200
net.ipv4.ip_local_port_range = 10000 65000

# 文件系统优化
fs.file-max = 2097152
fs.nr_open = 2097152

# 内存优化
vm.swappiness = 10
vm.dirty_ratio = 15
vm.dirty_background_ratio = 5
EOF
    
    sysctl -p
    
    # 文件描述符限制
    cat >> /etc/security/limits.conf << EOF
* soft nofile 65535
* hard nofile 65535
* soft nproc 65535
* hard nproc 65535
EOF
    
    # 创建 swap 文件（如果内存小于 8GB）
    local total_mem=$(free -g | awk '/^Mem:/{print $2}')
    if [ "$total_mem" -lt 8 ]; then
        log "创建 Swap 文件..."
        fallocate -l 4G /swapfile
        chmod 600 /swapfile
        mkswap /swapfile
        swapon /swapfile
        echo "/swapfile none swap sw 0 0" >> /etc/fstab
    fi
    
    success "系统优化完成"
}

# 创建部署用户
create_deploy_user() {
    log "创建部署用户..."
    
    # 创建用户
    useradd -m -s /bin/bash -G docker,sudo openpenpal || true
    
    # 设置 sudo 权限
    echo "openpenpal ALL=(ALL) NOPASSWD: /usr/bin/docker, /usr/local/bin/docker-compose, /bin/systemctl" >> /etc/sudoers.d/openpenpal
    
    # 创建目录结构
    mkdir -p /home/openpenpal/{backups,logs,data,scripts}
    chown -R openpenpal:openpenpal /home/openpenpal
    
    success "部署用户创建完成"
}

# 配置日志轮转
setup_log_rotation() {
    log "配置日志轮转..."
    
    cat > /etc/logrotate.d/openpenpal << EOF
/home/openpenpal/logs/*.log {
    daily
    rotate 14
    compress
    delaycompress
    missingok
    notifempty
    create 0644 openpenpal openpenpal
    sharedscripts
    postrotate
        docker kill -s USR1 \$(docker ps -q) 2>/dev/null || true
    endscript
}
EOF
    
    success "日志轮转配置完成"
}

# 设置监控告警
setup_monitoring() {
    log "配置监控告警..."
    
    # 创建监控脚本
    cat > /home/openpenpal/scripts/monitor.sh << 'EOF'
#!/bin/bash
# 简单的监控脚本

# 检查磁盘使用率
disk_usage=$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')
if [ "$disk_usage" -gt 80 ]; then
    echo "WARNING: Disk usage is ${disk_usage}%"
fi

# 检查内存使用率
mem_usage=$(free | grep Mem | awk '{print int($3/$2 * 100.0)}')
if [ "$mem_usage" -gt 80 ]; then
    echo "WARNING: Memory usage is ${mem_usage}%"
fi

# 检查 Docker 容器状态
stopped_containers=$(docker ps -a --filter "status=exited" --format "{{.Names}}" | grep openpenpal)
if [ -n "$stopped_containers" ]; then
    echo "WARNING: Stopped containers: $stopped_containers"
fi
EOF
    
    chmod +x /home/openpenpal/scripts/monitor.sh
    
    # 添加到 crontab
    echo "*/5 * * * * /home/openpenpal/scripts/monitor.sh >> /home/openpenpal/logs/monitor.log 2>&1" | crontab -u openpenpal -
    
    success "监控配置完成"
}

# 主函数
main() {
    log "========================================="
    log "腾讯云服务器初始化"
    log "域名: $DOMAIN"
    log "========================================="
    
    # 检查是否为 root
    if [ "$EUID" -ne 0 ]; then
        echo "请使用 root 权限运行此脚本"
        exit 1
    fi
    
    # 执行初始化步骤
    update_system
    setup_firewall
    install_docker
    install_nginx
    setup_ssl
    install_tencent_tools
    optimize_system
    create_deploy_user
    setup_log_rotation
    setup_monitoring
    
    log "========================================="
    success "服务器初始化完成！"
    log ""
    log "后续步骤："
    log "1. 配置 GitHub Secrets"
    log "2. 推送代码到 main 分支触发部署"
    log "3. 访问 https://$DOMAIN 查看应用"
    log "========================================="
}

# 执行主函数
main "$@"