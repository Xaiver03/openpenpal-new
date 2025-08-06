#!/bin/bash

echo "========================================="
echo "分支分歧分析报告"
echo "========================================="

# 获取最新的远程信息
git fetch --all --prune

echo ""
echo "分支同步状态（与对应远程分支比较）："
echo "========================================="

sync_count=0
diverged_count=0

for branch in $(git branch | grep -v "remotes" | sed 's/\*//'); do
    branch=$(echo $branch | xargs)
    
    # 检查是否有对应的远程分支
    if git ls-remote --heads origin $branch | grep -q $branch; then
        # 计算领先和落后的提交数
        ahead=$(git rev-list --count origin/$branch..$branch 2>/dev/null || echo "0")
        behind=$(git rev-list --count $branch..origin/$branch 2>/dev/null || echo "0")
        
        if [ "$ahead" = "0" ] && [ "$behind" = "0" ]; then
            echo "✅ $branch - 与远程同步"
            sync_count=$((sync_count + 1))
        else
            if [ "$ahead" != "0" ] || [ "$behind" != "0" ]; then
                echo "🔄 $branch - 领先 $ahead, 落后 $behind"
                diverged_count=$((diverged_count + 1))
            fi
        fi
    else
        echo "❌ $branch - 远程不存在对应分支"
    fi
done

echo ""
echo "========================================="
echo "分支关系分析"
echo "========================================="

# 分析主要分支之间的关系
echo ""
echo "Feature分支与main的关系："
for branch in $(git branch | grep "feature/" | sed 's/\*//'); do
    branch=$(echo $branch | xargs)
    commits_ahead=$(git rev-list --count main..$branch 2>/dev/null || echo "0")
    commits_behind=$(git rev-list --count $branch..main 2>/dev/null || echo "0")
    
    if [ "$commits_ahead" != "0" ] || [ "$commits_behind" != "0" ]; then
        echo "$branch: 独有 $commits_ahead 个提交, 缺少main的 $commits_behind 个提交"
    fi
done

echo ""
echo "========================================="
echo "总结"
echo "========================================="
echo "- 完全同步的分支: $sync_count"
echo "- 有分歧的分支: $diverged_count"
echo ""

if [ "$diverged_count" -eq 0 ]; then
    echo "✅ 所有分支都已与其对应的远程分支同步"
    echo ""
    echo "注意：不同分支之间的差异是正常的，因为它们代表不同的功能开发。"
else
    echo "⚠️  有 $diverged_count 个分支与远程存在分歧"
fi

echo ""
echo "========================================="
echo "推送建议"
echo "========================================="

# 检查是否有需要推送的内容
need_push=false
for branch in $(git branch | grep -v "remotes" | sed 's/\*//'); do
    branch=$(echo $branch | xargs)
    ahead=$(git rev-list --count origin/$branch..$branch 2>/dev/null || echo "0")
    
    if [ "$ahead" != "0" ]; then
        if [ "$need_push" = false ]; then
            echo "以下分支需要推送："
            need_push=true
        fi
        echo "git push origin $branch  # 领先 $ahead 个提交"
    fi
done

if [ "$need_push" = false ]; then
    echo "✅ 没有需要推送的分支"
fi