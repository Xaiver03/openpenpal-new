#!/bin/bash

echo "========================================="
echo "检查需要推送的分支"
echo "========================================="

# 获取最新的远程信息
git fetch --all --prune

echo ""
echo "以下分支领先远程（需要推送）："
echo "========================================="

# 检查每个本地分支
for branch in $(git branch | grep -v "remotes" | sed 's/\*//'); do
    branch=$(echo $branch | xargs)
    
    # 检查是否有远程追踪分支
    remote_branch=$(git config --get branch.$branch.remote)
    if [ -z "$remote_branch" ]; then
        continue
    fi
    
    # 检查本地是否领先远程
    ahead=$(git rev-list --count origin/$branch..refs/heads/$branch 2>/dev/null || echo "0")
    
    if [ "$ahead" != "0" ]; then
        echo ""
        echo "分支: $branch"
        echo "领先远程: $ahead 个提交"
        
        # 显示领先的提交
        echo "领先的提交："
        git log origin/$branch..$branch --oneline --max-count=5
        
        # 如果超过5个提交，显示省略信息
        if [ "$ahead" -gt 5 ]; then
            remaining=$((ahead - 5))
            echo "... 还有 $remaining 个提交"
        fi
    fi
done

echo ""
echo "========================================="
echo "分支合并状态分析"
echo "========================================="

# 特别检查main分支
echo ""
echo "Main分支状态："
main_ahead=$(git rev-list --count origin/main..main 2>/dev/null || echo "0")
main_behind=$(git rev-list --count main..origin/main 2>/dev/null || echo "0")

if [ "$main_ahead" != "0" ]; then
    echo "- main分支领先远程 $main_ahead 个提交"
    echo "- 最新的5个提交："
    git log origin/main..main --oneline --max-count=5
fi

if [ "$main_behind" != "0" ]; then
    echo "- main分支落后远程 $main_behind 个提交"
fi

echo ""
echo "========================================="
echo "推送建议"
echo "========================================="

# 统计需要推送的分支
branches_to_push=0
for branch in $(git branch | grep -v "remotes" | sed 's/\*//'); do
    branch=$(echo $branch | xargs)
    ahead=$(git rev-list --count origin/$branch..refs/heads/$branch 2>/dev/null || echo "0")
    if [ "$ahead" != "0" ]; then
        branches_to_push=$((branches_to_push + 1))
    fi
done

if [ "$branches_to_push" -eq 0 ]; then
    echo "✅ 所有分支都已与远程同步，无需推送"
else
    echo "⚠️  有 $branches_to_push 个分支需要推送到远程"
    echo ""
    echo "推送命令："
    for branch in $(git branch | grep -v "remotes" | sed 's/\*//'); do
        branch=$(echo $branch | xargs)
        ahead=$(git rev-list --count origin/$branch..refs/heads/$branch 2>/dev/null || echo "0")
        if [ "$ahead" != "0" ]; then
            echo "git push origin $branch"
        fi
    done
fi