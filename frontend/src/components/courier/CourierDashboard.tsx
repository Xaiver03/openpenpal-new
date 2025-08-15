'use client';

import React, { useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { courierAPI } from '@/lib/api/courier';
import { courierGrowthAPI } from '@/lib/api/courier-growth';
import { 
  TrendingUp, Package, Users, Trophy, Target, 
  CheckCircle2, Clock, AlertCircle, ArrowRight,
  Star, Award
} from 'lucide-react';
import Link from 'next/link';
import { useCreditInfo, useCreditStore } from '@/stores/credit-store';
import { formatPoints } from '@/lib/api/credit';
import { CreditLevelBadge } from '@/components/credit/credit-level-badge';

export function CourierDashboard() {
  // 获取积分信息
  const { userCredit, creditSummary } = useCreditInfo();
  const { refreshAll } = useCreditStore();

  // 初始化时刷新积分数据
  useEffect(() => {
    refreshAll();
  }, [refreshAll]);

  // 获取信使状态
  const { data: courierStatus } = useQuery({
    queryKey: ['courier-status'],
    queryFn: async () => {
      const response = await courierAPI.getStatus();
      return (response as any).data;
    },
  });

  // 获取成长进度
  const { data: growthProgress } = useQuery({
    queryKey: ['courier-growth-progress'],
    queryFn: async () => {
      const response = await courierGrowthAPI.getGrowthProgress();
      return (response as any).data;
    },
  });

  // 获取今日任务统计
  const { data: tasks } = useQuery({
    queryKey: ['courier-tasks'],
    queryFn: async () => {
      const response = await courierAPI.getTasks();
      return (response as any).data;
    },
  });

  const levelNames = ['', '一级信使', '二级信使', '三级信使', '四级信使'];
  const levelColors = ['', 'text-green-600', 'text-blue-600', 'text-purple-600', 'text-orange-600'];

  const stats = [
    {
      title: '今日完成',
      value: tasks?.filter((t: any) => t.status === 'delivered')?.length || 0,
      total: tasks?.length || 0,
      icon: Package,
      color: 'text-green-600',
    },
    {
      title: '累计积分',
      value: formatPoints(userCredit?.total || 0),
      icon: Star,
      color: 'text-yellow-600',
      subtitle: userCredit ? `等级 ${userCredit.level}` : '',
    },
    {
      title: '本周获得',
      value: formatPoints(creditSummary?.week_earned || 0),
      icon: TrendingUp,
      color: 'text-blue-600',
    },
    {
      title: '待处理任务',
      value: creditSummary?.pending_tasks || 0,
      icon: Target,
      color: 'text-purple-600',
    },
  ];

  return (
    <div className="space-y-6">
      {/* 信使信息卡片 */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>信使中心</CardTitle>
              {userCredit && (
                <div className="mt-2">
                  <CreditLevelBadge
                    level={userCredit.level}
                    totalPoints={userCredit.total}
                    showTooltip={true}
                    size="sm"
                  />
                </div>
              )}
            </div>
            <Badge className={levelColors[courierStatus?.level || 1]}>
              {levelNames[courierStatus?.level || 1]}
            </Badge>
          </div>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            {stats.map((stat, index) => (
              <div key={index} className="text-center">
                <stat.icon className={`h-8 w-8 mx-auto mb-2 ${stat.color}`} />
                <p className="text-2xl font-bold">{stat.value}</p>
                <p className="text-sm text-muted-foreground">{stat.title}</p>
                {stat.total && stat.total > 0 && (
                  <p className="text-xs text-muted-foreground">共 {stat.total}</p>
                )}
                {stat.subtitle && (
                  <p className="text-xs text-muted-foreground">{stat.subtitle}</p>
                )}
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* 晋升进度卡片 */}
      {growthProgress && growthProgress.next_level && (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle className="text-lg flex items-center gap-2">
                <TrendingUp className="h-5 w-5" />
                晋升进度
              </CardTitle>
              <Link href="/courier/growth">
                <Button variant="ghost" size="sm">
                  查看详情 <ArrowRight className="ml-1 h-4 w-4" />
                </Button>
              </Link>
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-sm">
                  距离 {levelNames[growthProgress.next_level]} 还需完成
                </span>
                <span className="text-sm font-medium">
                  {Math.round(growthProgress.completion_rate)}%
                </span>
              </div>
              <Progress value={growthProgress.completion_rate} />
              
              {growthProgress.can_upgrade && (
                <div className="bg-primary/10 rounded-lg p-4">
                  <p className="text-sm font-medium text-primary mb-2">
                    🎉 恭喜！您已满足晋升条件
                  </p>
                  <Link href={`/courier/promotion/apply?level=${growthProgress.next_level}`}>
                    <Button size="sm" className="w-full">
                      申请晋升到 {levelNames[growthProgress.next_level]}
                    </Button>
                  </Link>
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      )}

      {/* 快捷操作 */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">快捷操作</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 gap-3">
            <Link href="/courier/tasks">
              <Button variant="outline" className="w-full justify-start">
                <Package className="mr-2 h-4 w-4" />
                任务中心
              </Button>
            </Link>
            <Link href="/courier/growth">
              <Button variant="outline" className="w-full justify-start">
                <TrendingUp className="mr-2 h-4 w-4" />
                成长路径
              </Button>
            </Link>
            
            {courierStatus?.level && courierStatus.level >= 2 && (
              <Link href="/courier/subordinates">
                <Button variant="outline" className="w-full justify-start">
                  <Users className="mr-2 h-4 w-4" />
                  管理下级
                </Button>
              </Link>
            )}
            
            {courierStatus?.level && courierStatus.level >= 3 && (
              <Link href="/courier/promotion/manage">
                <Button variant="outline" className="w-full justify-start">
                  <CheckCircle2 className="mr-2 h-4 w-4" />
                  审核晋升
                </Button>
              </Link>
            )}
            
            <Link href="/courier/points">
              <Button variant="outline" className="w-full justify-start">
                <Trophy className="mr-2 h-4 w-4" />
                积分中心
              </Button>
            </Link>
            
            <Link href="/courier/performance">
              <Button variant="outline" className="w-full justify-start">
                <Target className="mr-2 h-4 w-4" />
                绩效统计
              </Button>
            </Link>
          </div>
        </CardContent>
      </Card>

      {/* 今日任务概览 */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="text-lg">今日任务</CardTitle>
            <Link href="/courier/tasks">
              <Button variant="ghost" size="sm">
                查看全部 <ArrowRight className="ml-1 h-4 w-4" />
              </Button>
            </Link>
          </div>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            {tasks?.slice(0, 3).map((task: any) => (
              <div key={task.id} className="flex items-center justify-between p-3 rounded-lg border">
                <div className="flex items-center gap-3">
                  {task.status === 'delivered' ? (
                    <CheckCircle2 className="h-5 w-5 text-green-600" />
                  ) : task.status === 'in_transit' ? (
                    <Clock className="h-5 w-5 text-blue-600" />
                  ) : (
                    <AlertCircle className="h-5 w-5 text-amber-600" />
                  )}
                  <div>
                    <p className="font-medium text-sm">{task.letterCode}</p>
                    <p className="text-xs text-muted-foreground">{task.targetLocation}</p>
                  </div>
                </div>
                <Badge variant={
                  task.status === 'delivered' ? 'success' : 
                  task.status === 'in_transit' ? 'default' : 'secondary'
                }>
                  {task.status === 'delivered' ? '已完成' :
                   task.status === 'in_transit' ? '配送中' : '待处理'}
                </Badge>
              </div>
            ))}
            
            {(!tasks || tasks.length === 0) && (
              <div className="text-center py-8 text-muted-foreground">
                暂无任务
              </div>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}