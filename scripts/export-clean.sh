#\!/bin/bash

# 创建临时目录
TEMP_DIR="../openpenpal-clean"
CURRENT_DIR=$(pwd)

echo "Creating clean export of OpenPenPal..."

# 创建目标目录
mkdir -p $TEMP_DIR

# 复制所有文件（排除.git和大文件）
echo "Copying files..."
rsync -av --progress \
  --exclude='.git' \
  --exclude='node_modules' \
  --exclude='*.log' \
  --exclude='*.zip' \
  --exclude='*.tar.gz' \
  --exclude='dist' \
  --exclude='build' \
  --exclude='.next' \
  --exclude='uploads' \
  --exclude='tmp' \
  --exclude='*.mp4' \
  --exclude='*.mov' \
  --exclude='*.avi' \
  . $TEMP_DIR/

echo "Clean export created at: $TEMP_DIR"
echo "Size of clean export:"
du -sh $TEMP_DIR

# 在新目录中初始化git
cd $TEMP_DIR
git init
git add .
git commit -m "Initial commit - OpenPenPal project clean export"

# 添加远程仓库
git remote add origin https://github.com/Xaiver03/opp.git

echo "Ready to push. Run: cd $TEMP_DIR && git push -u origin main"
