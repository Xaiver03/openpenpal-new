#!/bin/bash

# 推送所有本地分支到远程仓库的脚本

echo "========================================="
echo "推送所有本地分支到远程仓库"
echo "========================================="

# 设置大缓冲区以处理大文件
git config http.postBuffer 3221225472

# 首先推送当前分支的最新提交
echo "推送当前分支的最新提交..."
git push origin feature/ai-system-optimization-sota

# 获取所有本地分支列表
branches=$(git branch | grep -v "remotes" | sed 's/\*//')

# 推送每个分支
for branch in $branches; do
    branch=$(echo $branch | xargs)  # 去除空格
    echo ""
    echo "========================================="
    echo "推送分支: $branch"
    echo "========================================="
    
    # 推送分支到远程
    if git push origin "$branch" --force-with-lease; then
        echo "✓ 成功推送分支: $branch"
    else
        echo "✗ 推送分支失败: $branch"
        echo "尝试使用 --force 选项..."
        if git push origin "$branch" --force; then
            echo "✓ 使用 --force 成功推送分支: $branch"
        else
            echo "✗ 无法推送分支: $branch"
        fi
    fi
done

# 推送所有标签
echo ""
echo "========================================="
echo "推送所有标签..."
echo "========================================="
if git push origin --tags; then
    echo "✓ 成功推送所有标签"
else
    echo "✗ 推送标签失败"
fi

echo ""
echo "========================================="
echo "推送完成！"
echo "========================================="

# 显示远程分支状态
echo ""
echo "远程分支状态："
git branch -r

echo ""
echo "本地分支状态："
git branch -a