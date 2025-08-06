# 使用Python 3.10作为基础镜像
FROM python:3.10-slim

# 设置工作目录
WORKDIR /app

# 安装系统依赖
RUN apt-get update && apt-get install -y \
    # Tesseract OCR
    tesseract-ocr \
    tesseract-ocr-chi-sim \
    tesseract-ocr-eng \
    # OpenCV依赖
    libgl1-mesa-glx \
    libglib2.0-0 \
    libsm6 \
    libxext6 \
    libxrender-dev \
    libgomp1 \
    # 图像处理库
    libjpeg-dev \
    libpng-dev \
    libtiff-dev \
    # 编译工具
    gcc \
    g++ \
    # 网络工具
    wget \
    curl \
    # 清理缓存
    && rm -rf /var/lib/apt/lists/*

# 设置环境变量
ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PIP_NO_CACHE_DIR=1 \
    PIP_DISABLE_PIP_VERSION_CHECK=1

# 复制requirements文件
COPY requirements.txt .

# 安装Python依赖
RUN pip install --upgrade pip && \
    pip install -r requirements.txt

# 下载PaddleOCR模型 (可选，首次运行时自动下载)
RUN python -c "from paddleocr import PaddleOCR; PaddleOCR(use_angle_cls=True, lang='ch', show_log=False)" || echo "PaddleOCR模型将在首次使用时下载"

# 下载EasyOCR模型 (可选)
RUN python -c "import easyocr; easyocr.Reader(['ch_sim', 'en'])" || echo "EasyOCR模型将在首次使用时下载"

# 创建必要的目录
RUN mkdir -p /app/uploads /app/models /app/logs

# 复制应用代码
COPY . .

# 设置权限
RUN chmod +x /app/app.py

# 创建非root用户
RUN adduser --disabled-password --gecos '' appuser && \
    chown -R appuser:appuser /app
USER appuser

# 暴露端口
EXPOSE 8004

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD curl -f http://localhost:8004/health || exit 1

# 启动命令
CMD ["gunicorn", "--bind", "0.0.0.0:8004", "--workers", "4", "--timeout", "300", "--worker-class", "sync", "app:app"]