'use client';

import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { courierGrowthAPI } from '@/lib/api/courier-growth';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useToast } from '@/components/ui/use-toast';
import { Loader2, CheckCircle, XCircle, Clock, User } from 'lucide-react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { formatDistanceToNow } from 'date-fns';
import { zhCN } from 'date-fns/locale';

export function PromotionManagement() {
  const { toast } = useToast();
  const queryClient = useQueryClient();
  const [selectedRequest, setSelectedRequest] = useState<any>(null);
  const [reviewComment, setReviewComment] = useState('');
  const [reviewAction, setReviewAction] = useState<'approve' | 'reject' | null>(null);

  // 获取晋升申请列表
  const { data: requestsData, isLoading } = useQuery({
    queryKey: ['upgrade-requests'],
    queryFn: async () => {
      const response = await courierGrowthAPI.getUpgradeRequests();
      return (response as any).data;
    },
  });

  // 处理晋升申请
  const processMutation = useMutation({
    mutationFn: async (params: { requestId: string; action: 'approve' | 'reject'; comment: string }) => {
      return await courierGrowthAPI.processUpgradeRequest(params.requestId, {
        action: params.action,
        comment: params.comment,
      });
    },
    onSuccess: () => {
      toast({
        title: '处理成功',
        description: reviewAction === 'approve' ? '晋升申请已批准' : '晋升申请已驳回',
      });
      queryClient.invalidateQueries({ queryKey: ['upgrade-requests'] });
      setSelectedRequest(null);
      setReviewComment('');
      setReviewAction(null);
    },
    onError: (error: any) => {
      toast({
        title: '处理失败',
        description: error.response?.data?.message || '请稍后重试',
        variant: 'destructive',
      });
    },
  });

  const handleReview = (request: any, action: 'approve' | 'reject') => {
    setSelectedRequest(request);
    setReviewAction(action);
    setReviewComment('');
  };

  const submitReview = () => {
    if (!selectedRequest || !reviewAction) return;
    
    processMutation.mutate({
      requestId: selectedRequest.id,
      action: reviewAction,
      comment: reviewComment,
    });
  };

  const levelNames = ['', '一级信使', '二级信使', '三级信使', '四级信使'];
  const statusConfig = {
    pending: { label: '待审核', variant: 'default' as const, icon: Clock },
    approved: { label: '已通过', variant: 'success' as const, icon: CheckCircle },
    rejected: { label: '已驳回', variant: 'destructive' as const, icon: XCircle },
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    );
  }

  const requests = requestsData?.requests || [];

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>晋升申请管理</CardTitle>
        </CardHeader>
        <CardContent>
          <Tabs defaultValue="pending" className="w-full">
            <TabsList className="grid w-full grid-cols-3">
              <TabsTrigger value="pending">待审核</TabsTrigger>
              <TabsTrigger value="approved">已通过</TabsTrigger>
              <TabsTrigger value="rejected">已驳回</TabsTrigger>
            </TabsList>
            
            {['pending', 'approved', 'rejected'].map((status) => (
              <TabsContent key={status} value={status} className="space-y-4">
                {requests
                  .filter((req: any) => req.status === status)
                  .map((request: any) => (
                    <Card key={request.id}>
                      <CardContent className="pt-6">
                        <div className="flex items-start justify-between">
                          <div className="space-y-3 flex-1">
                            <div className="flex items-center gap-3">
                              <User className="h-5 w-5 text-muted-foreground" />
                              <div>
                                <p className="font-medium">信使ID: {request.courierId}</p>
                                <p className="text-sm text-muted-foreground">
                                  {levelNames[request.current_level]} → {levelNames[request.request_level]}
                                </p>
                              </div>
                            </div>
                            
                            <div>
                              <p className="text-sm font-medium mb-1">申请理由</p>
                              <p className="text-sm text-muted-foreground">{request.reason}</p>
                            </div>
                            
                            <div className="flex items-center gap-4 text-sm text-muted-foreground">
                              <span>
                                申请时间: {formatDistanceToNow(new Date(request.createdAt), {
                                  addSuffix: true,
                                  locale: zhCN,
                                })}
                              </span>
                              {request.reviewed_at && (
                                <span>
                                  审核时间: {formatDistanceToNow(new Date(request.reviewed_at), {
                                    addSuffix: true,
                                    locale: zhCN,
                                  })}
                                </span>
                              )}
                            </div>
                            
                            {request.reviewer_comment && (
                              <div className="mt-3 p-3 bg-muted rounded-md">
                                <p className="text-sm font-medium mb-1">审核意见</p>
                                <p className="text-sm">{request.reviewer_comment}</p>
                              </div>
                            )}
                          </div>
                          
                          <div className="flex items-center gap-2 ml-4">
                            <Badge variant={statusConfig[status as keyof typeof statusConfig].variant}>
                              {statusConfig[status as keyof typeof statusConfig].label}
                            </Badge>
                            
                            {status === 'pending' && (
                              <div className="flex gap-2">
                                <Button
                                  size="sm"
                                  variant="outline"
                                  onClick={() => handleReview(request, 'reject')}
                                >
                                  驳回
                                </Button>
                                <Button
                                  size="sm"
                                  onClick={() => handleReview(request, 'approve')}
                                >
                                  批准
                                </Button>
                              </div>
                            )}
                          </div>
                        </div>
                      </CardContent>
                    </Card>
                  ))}
                  
                {requests.filter((req: any) => req.status === status).length === 0 && (
                  <div className="text-center py-8 text-muted-foreground">
                    暂无{statusConfig[status as keyof typeof statusConfig].label}的申请
                  </div>
                )}
              </TabsContent>
            ))}
          </Tabs>
        </CardContent>
      </Card>

      {/* 审核对话框 */}
      <Dialog open={!!selectedRequest} onOpenChange={() => setSelectedRequest(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {reviewAction === 'approve' ? '批准晋升申请' : '驳回晋升申请'}
            </DialogTitle>
            <DialogDescription>
              请填写审核意见，这将发送给申请人
            </DialogDescription>
          </DialogHeader>
          
          <div className="space-y-4">
            <div>
              <Label>申请信息</Label>
              <p className="text-sm text-muted-foreground mt-1">
                信使 {selectedRequest?.courierId} 申请从 {levelNames[selectedRequest?.current_level]} 
                晋升到 {levelNames[selectedRequest?.request_level]}
              </p>
            </div>
            
            <div>
              <Label htmlFor="comment">审核意见</Label>
              <Textarea
                id="comment"
                placeholder={
                  reviewAction === 'approve' 
                    ? '恭喜您通过晋升审核！请继续保持优秀的工作表现...' 
                    : '您的申请暂未通过，建议您...'
                }
                value={reviewComment}
                onChange={(e) => setReviewComment(e.target.value)}
                rows={4}
              />
            </div>
          </div>
          
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setSelectedRequest(null)}
              disabled={processMutation.isPending}
            >
              取消
            </Button>
            <Button
              variant={reviewAction === 'approve' ? 'default' : 'destructive'}
              onClick={submitReview}
              disabled={!reviewComment || processMutation.isPending}
            >
              {processMutation.isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  处理中...
                </>
              ) : (
                <>
                  {reviewAction === 'approve' ? '批准' : '驳回'}
                </>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}