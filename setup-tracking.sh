#!/bin/bash

echo "设置所有分支的远程追踪关系..."

# 获取所有本地分支
branches=$(git branch | grep -v "remotes" | sed 's/\*//')

for branch in $branches; do
    branch=$(echo $branch | xargs)
    
    # 检查远程是否存在对应分支
    if git ls-remote --heads origin $branch | grep -q $branch; then
        echo "设置 $branch 追踪 origin/$branch"
        git branch --set-upstream-to=origin/$branch $branch
    else
        echo "远程不存在分支 $branch，跳过"
    fi
done

echo "完成！"