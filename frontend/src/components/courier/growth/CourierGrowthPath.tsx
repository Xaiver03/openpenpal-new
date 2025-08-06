'use client';

import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { courierGrowthAPI, type GrowthPath, type GrowthRequirement } from '@/lib/api/courier-growth';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { CheckCircle2, Circle, Lock, TrendingUp } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useRouter } from 'next/navigation';

export function CourierGrowthPath() {
  const router = useRouter();
  
  const { data: growthPath, isLoading } = useQuery({
    queryKey: ['courier-growth-path'],
    queryFn: async () => {
      const response = await courierGrowthAPI.getGrowthPath();
      return (response as any).data;
    },
  });

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto"></div>
          <p className="mt-4 text-muted-foreground">加载成长路径...</p>
        </div>
      </div>
    );
  }

  if (!growthPath) {
    return <div>无法加载成长路径</div>;
  }

  const levelNames = ['', '一级信使', '二级信使', '三级信使', '四级信使'];
  const levelColors = ['', 'bg-green-500', 'bg-blue-500', 'bg-purple-500', 'bg-orange-500'];

  return (
    <div className="space-y-6">
      {/* 当前等级卡片 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <TrendingUp className="h-5 w-5" />
            我的成长路径
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-4">
            <div className={`w-16 h-16 rounded-full ${levelColors[growthPath.current_level]} flex items-center justify-center text-white font-bold text-xl`}>
              L{growthPath.current_level}
            </div>
            <div>
              <h3 className="text-lg font-semibold">{growthPath.current_name}</h3>
              <p className="text-muted-foreground">当前等级</p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* 晋升路径 */}
      <div className="space-y-4">
        {growthPath.paths.map((path) => (
          <Card key={path.target_level} className={path.can_upgrade ? 'border-primary' : ''}>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle className="text-lg flex items-center gap-2">
                  {path.can_upgrade ? (
                    <Circle className="h-5 w-5 text-primary" />
                  ) : (
                    <Lock className="h-5 w-5 text-muted-foreground" />
                  )}
                  {path.target_name}
                </CardTitle>
                <Badge variant={path.can_upgrade ? 'default' : 'secondary'}>
                  L{path.target_level}
                </Badge>
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {/* 权限预览 */}
                <div>
                  <p className="text-sm font-medium mb-2">解锁权限</p>
                  <div className="flex flex-wrap gap-1">
                    {path.permissions.slice(0, 3).map((perm) => (
                      <Badge key={perm} variant="outline" className="text-xs">
                        {perm}
                      </Badge>
                    ))}
                    {path.permissions.length > 3 && (
                      <Badge variant="outline" className="text-xs">
                        +{path.permissions.length - 3}
                      </Badge>
                    )}
                  </div>
                </div>

                {/* 晋升条件 */}
                {path.detailed_requirements && (
                  <div>
                    <div className="flex items-center justify-between mb-2">
                      <p className="text-sm font-medium">晋升条件</p>
                      {path.completion_rate !== undefined && (
                        <span className="text-sm text-muted-foreground">
                          {Math.round(path.completion_rate)}% 完成
                        </span>
                      )}
                    </div>
                    {path.completion_rate !== undefined && (
                      <Progress value={path.completion_rate} className="mb-3" />
                    )}
                    <div className="space-y-2">
                      {path.detailed_requirements.map((req, index) => (
                        <RequirementItem key={index} requirement={req} />
                      ))}
                    </div>
                  </div>
                )}

                {/* 申请按钮 */}
                {path.can_upgrade && (
                  <Button 
                    className="w-full"
                    onClick={() => router.push(`/courier/promotion/apply?level=${path.target_level}`)}
                  >
                    申请晋升到 {path.target_name}
                  </Button>
                )}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}

function RequirementItem({ requirement }: { requirement: GrowthRequirement }) {
  const getProgressText = () => {
    if (requirement.current !== undefined && requirement.target) {
      return `${requirement.current} / ${requirement.target}`;
    }
    return requirement.completed ? '已完成' : '未完成';
  };

  return (
    <div className="flex items-start gap-2 text-sm">
      {requirement.completed ? (
        <CheckCircle2 className="h-4 w-4 text-green-500 mt-0.5" />
      ) : (
        <Circle className="h-4 w-4 text-muted-foreground mt-0.5" />
      )}
      <div className="flex-1">
        <p className={requirement.completed ? 'text-green-600' : 'text-muted-foreground'}>
          {requirement.name}
        </p>
        <p className="text-xs text-muted-foreground">{requirement.description}</p>
        <p className="text-xs font-medium mt-1">{getProgressText()}</p>
      </div>
    </div>
  );
}