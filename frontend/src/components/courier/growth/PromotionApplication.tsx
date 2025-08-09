'use client';

import React, { useState } from 'react';
import { useMutation, useQuery } from '@tanstack/react-query';
import { courierGrowthAPI } from '@/lib/api/courier-growth';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { useToast } from '@/components/ui/use-toast';
import { Loader2, Send, CheckCircle2, AlertCircle } from 'lucide-react';
import { useRouter, useSearchParams } from 'next/navigation';

export function PromotionApplication() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const targetLevel = parseInt(searchParams.get('level') || '2');
  const { toast } = useToast();
  
  const [reason, setReason] = useState('');
  const [evidence, setEvidence] = useState<Record<string, any>>({});

  // 获取当前等级和晋升要求
  const { data: levelInfo } = useQuery({
    queryKey: ['courier-level-check'],
    queryFn: async () => {
      const response = await courierGrowthAPI.checkLevel();
      return (response as any).data;
    },
  });

  const { data: growthProgress } = useQuery({
    queryKey: ['courier-growth-progress'],
    queryFn: async () => {
      const response = await courierGrowthAPI.getGrowthProgress();
      return (response as any).data;
    },
  });

  // 提交晋升申请
  const submitMutation = useMutation({
    mutationFn: async () => {
      return await courierGrowthAPI.submitUpgradeRequest({
        request_level: targetLevel,
        reason,
        evidence,
      });
    },
    onSuccess: () => {
      toast({
        title: '申请提交成功',
        description: '您的晋升申请已提交，请等待审核。',
      });
      router.push('/courier/growth');
    },
    onError: (error: any) => {
      toast({
        title: '申请提交失败',
        description: error.response?.data?.message || '请稍后重试',
        variant: 'destructive',
      });
    },
  });

  const levelNames = ['', '一级信使', '二级信使', '三级信使', '四级信使'];

  if (!levelInfo || !growthProgress) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    );
  }

  const canApply = growthProgress.can_upgrade && growthProgress.next_level === targetLevel;

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>申请晋升到 {levelNames[targetLevel]}</CardTitle>
          <CardDescription>
            从 {levelNames[levelInfo.level]} 晋升到 {levelNames[targetLevel]}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* 晋升条件检查 */}
          <div>
            <h3 className="font-medium mb-3">晋升条件检查</h3>
            <div className="space-y-2">
              {growthProgress.requirements.map((req: any, index: number) => (
                <div key={index} className="flex items-center gap-2 text-sm">
                  {req.completed ? (
                    <CheckCircle2 className="h-4 w-4 text-green-500" />
                  ) : (
                    <AlertCircle className="h-4 w-4 text-amber-500" />
                  )}
                  <span className={req.completed ? 'text-green-600' : 'text-amber-600'}>
                    {req.name}: {req.current || 0} / {req.target}
                  </span>
                </div>
              ))}
            </div>
            
            {!canApply && (
              <Alert className="mt-4">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>
                  您还未满足所有晋升条件，请继续努力完成要求后再申请。
                </AlertDescription>
              </Alert>
            )}
          </div>

          {/* 申请理由 */}
          <div className="space-y-2">
            <Label htmlFor="reason">申请理由</Label>
            <Textarea
              id="reason"
              placeholder="请详细说明您申请晋升的理由，包括您的工作表现、对团队的贡献等..."
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              rows={6}
              disabled={!canApply}
            />
            <p className="text-xs text-muted-foreground">
              请至少输入50个字符，详细说明您的晋升理由
            </p>
          </div>

          {/* 补充材料 */}
          <div className="space-y-2">
            <Label>补充材料（可选）</Label>
            <div className="text-sm text-muted-foreground">
              您可以提供额外的证明材料，如优秀表现截图、用户好评等
            </div>
            {/* 这里可以添加文件上传组件 */}
          </div>

          {/* 提交按钮 */}
          <div className="flex gap-2">
            <Button
              variant="outline"
              onClick={() => router.back()}
              disabled={submitMutation.isPending}
            >
              取消
            </Button>
            <Button
              onClick={() => submitMutation.mutate()}
              disabled={!canApply || reason.length < 50 || submitMutation.isPending}
            >
              {submitMutation.isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  提交中...
                </>
              ) : (
                <>
                  <Send className="mr-2 h-4 w-4" />
                  提交申请
                </>
              )}
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* 晋升说明 */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">晋升说明</CardTitle>
        </CardHeader>
        <CardContent className="text-sm text-muted-foreground space-y-2">
          <p>1. 晋升申请提交后，将由上级信使进行审核</p>
          <p>2. 审核周期一般为3-5个工作日</p>
          <p>3. 审核结果将通过系统通知发送给您</p>
          <p>4. 如果申请被驳回，您可以在满足条件后重新申请</p>
          <p>5. 晋升成功后，您将获得新的权限和责任</p>
        </CardContent>
      </Card>
    </div>
  );
}