#!/bin/bash

echo "========================================="
echo "检查本地和远程仓库同步状态"
echo "========================================="

# 获取最新的远程信息
echo "正在获取远程仓库最新信息..."
git fetch --all --prune

echo ""
echo "========================================="
echo "分支同步状态检查"
echo "========================================="

# 获取所有本地分支
local_branches=$(git branch | grep -v "remotes" | sed 's/\*//')

# 检查每个本地分支
for branch in $local_branches; do
    branch=$(echo $branch | xargs)  # 去除空格
    
    # 获取远程追踪分支
    remote_branch=$(git config --get branch.$branch.merge | sed 's|refs/heads/||')
    remote_name=$(git config --get branch.$branch.remote)
    
    if [ -z "$remote_branch" ]; then
        echo "❌ $branch - 没有设置远程追踪分支"
        continue
    fi
    
    # 检查本地和远程的差异
    ahead=$(git rev-list --count origin/$branch..HEAD 2>/dev/null || echo "0")
    behind=$(git rev-list --count HEAD..origin/$branch 2>/dev/null || echo "0")
    
    if [ "$ahead" = "0" ] && [ "$behind" = "0" ]; then
        echo "✅ $branch - 与远程同步"
    elif [ "$ahead" != "0" ] && [ "$behind" = "0" ]; then
        echo "⬆️  $branch - 领先远程 $ahead 个提交"
    elif [ "$ahead" = "0" ] && [ "$behind" != "0" ]; then
        echo "⬇️  $branch - 落后远程 $behind 个提交"
    else
        echo "🔄 $branch - 领先 $ahead 个提交，落后 $behind 个提交（需要合并）"
    fi
done

echo ""
echo "========================================="
echo "远程分支状态"
echo "========================================="

# 列出所有远程分支
echo "远程分支列表："
git branch -r | grep -v HEAD | sort

echo ""
echo "========================================="
echo "未推送的提交检查"
echo "========================================="

# 检查是否有未推送的提交
unpushed=$(git log --branches --not --remotes --oneline)
if [ -z "$unpushed" ]; then
    echo "✅ 所有提交都已推送到远程仓库"
else
    echo "❌ 发现未推送的提交："
    echo "$unpushed"
fi

echo ""
echo "========================================="
echo "工作区状态"
echo "========================================="

# 检查工作区状态
if git diff-index --quiet HEAD --; then
    echo "✅ 工作区干净，没有未提交的更改"
else
    echo "❌ 工作区有未提交的更改："
    git status --short
fi

echo ""
echo "========================================="
echo "总结"
echo "========================================="

# 统计信息
total_branches=$(git branch | wc -l | xargs)
remote_branches=$(git branch -r | grep -v HEAD | wc -l | xargs)
echo "本地分支数量: $total_branches"
echo "远程分支数量: $remote_branches"

# 检查是否完全同步
if [ -z "$unpushed" ] && git diff-index --quiet HEAD --; then
    echo ""
    echo "✅ 本地和远程仓库完全同步！"
else
    echo ""
    echo "⚠️  存在未同步的内容，请检查上面的详细信息"
fi