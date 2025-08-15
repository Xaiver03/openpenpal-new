/**
 * Comment Report Dialog - SOTA实现
 * 评论举报对话框 - 完整的用户界面和状态管理
 */

import React, { useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Button } from "@/components/ui/button"
import { Textarea } from "@/components/ui/textarea"
import { Label } from "@/components/ui/label"
import { AlertTriangle, Flag, Loader2 } from "lucide-react"
import { useToast } from "@/hooks/use-toast"
import type { CommentReportRequestSOTA } from '@/types/comment-sota'

interface ReportCommentDialogProps {
  /** 是否显示对话框 */
  open: boolean
  /** 关闭对话框回调 */
  onOpenChange: (open: boolean) => void
  /** 评论ID */
  commentId: string
  /** 评论作者信息 */
  commentAuthor?: {
    username: string
    nickname: string
  }
  /** 举报提交回调 */
  onSubmit: (commentId: string, report: CommentReportRequestSOTA) => Promise<void>
  /** 是否正在提交 */
  isSubmitting?: boolean
}

// 举报原因选项配置
const REPORT_REASONS = [
  {
    value: 'spam' as const,
    label: '垃圾信息',
    description: '垃圾邮件、广告或重复内容'
  },
  {
    value: 'inappropriate' as const,
    label: '不当内容',
    description: '不符合社区准则的内容'
  },
  {
    value: 'offensive' as const,
    label: '冒犯性内容',
    description: '仇恨言论、骚扰或恶意攻击'
  },
  {
    value: 'false_info' as const,
    label: '虚假信息',
    description: '误导性或不准确的信息'
  },
  {
    value: 'other' as const,
    label: '其他',
    description: '其他违规行为'
  }
] as const

export function ReportCommentDialog({
  open,
  onOpenChange,
  commentId,
  commentAuthor,
  onSubmit,
  isSubmitting = false
}: ReportCommentDialogProps) {
  const { toast } = useToast()
  
  // 表单状态
  const [reason, setReason] = useState<CommentReportRequestSOTA['reason'] | ''>('')
  const [description, setDescription] = useState('')
  const [isFormValid, setIsFormValid] = useState(false)

  // 表单验证
  React.useEffect(() => {
    setIsFormValid(reason !== '' && reason !== undefined)
  }, [reason])

  // 重置表单
  const resetForm = () => {
    setReason('')
    setDescription('')
    setIsFormValid(false)
  }

  // 处理对话框关闭
  const handleOpenChange = (newOpen: boolean) => {
    if (!newOpen && !isSubmitting) {
      resetForm()
    }
    onOpenChange(newOpen)
  }

  // 处理举报提交
  const handleSubmit = async () => {
    if (!isFormValid || !reason || isSubmitting) {
      return
    }

    try {
      const reportData: CommentReportRequestSOTA = {
        reason,
        ...(description.trim() && { description: description.trim() })
      }

      await onSubmit(commentId, reportData)
      
      toast({
        title: "举报已提交",
        description: "我们会尽快审核您的举报，感谢您的反馈。",
      })
      
      handleOpenChange(false)
      resetForm()
    } catch (error) {
      toast({
        title: "举报提交失败",
        description: error instanceof Error ? error.message : "请稍后重试",
        variant: "destructive",
      })
    }
  }

  // 获取选中原因的详细信息
  const selectedReason = REPORT_REASONS.find(r => r.value === reason)

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Flag className="h-5 w-5 text-destructive" />
            举报评论
          </DialogTitle>
          <DialogDescription>
            {commentAuthor ? (
              <>举报来自 <span className="font-medium">{commentAuthor.nickname || commentAuthor.username}</span> 的评论</>
            ) : (
              '举报此评论'
            )}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          {/* 举报原因选择 */}
          <div className="space-y-2">
            <Label htmlFor="reason">举报原因 *</Label>
            <Select value={reason} onValueChange={(value) => setReason(value as CommentReportRequestSOTA['reason'])}>
              <SelectTrigger>
                <SelectValue placeholder="请选择举报原因" />
              </SelectTrigger>
              <SelectContent>
                {REPORT_REASONS.map((item) => (
                  <SelectItem key={item.value} value={item.value}>
                    <div className="flex flex-col">
                      <span className="font-medium">{item.label}</span>
                      <span className="text-xs text-muted-foreground">{item.description}</span>
                    </div>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            
            {selectedReason && (
              <div className="flex items-start gap-2 p-3 rounded-md bg-muted/50 text-sm">
                <AlertTriangle className="h-4 w-4 text-amber-500 mt-0.5 shrink-0" />
                <div>
                  <p className="font-medium">{selectedReason.label}</p>
                  <p className="text-muted-foreground">{selectedReason.description}</p>
                </div>
              </div>
            )}
          </div>

          {/* 详细描述（可选） */}
          <div className="space-y-2">
            <Label htmlFor="description">详细说明（可选）</Label>
            <Textarea
              id="description"
              placeholder="请详细描述违规行为..."
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              maxLength={500}
              rows={3}
              className="resize-none"
            />
            <div className="text-xs text-muted-foreground text-right">
              {description.length}/500
            </div>
          </div>

          {/* 提示信息 */}
          <div className="p-3 rounded-md bg-blue-50 dark:bg-blue-950/50 border border-blue-200 dark:border-blue-800">
            <div className="flex items-start gap-2">
              <AlertTriangle className="h-4 w-4 text-blue-500 mt-0.5 shrink-0" />
              <div className="text-sm text-blue-700 dark:text-blue-300">
                <p className="font-medium">举报说明</p>
                <ul className="mt-1 space-y-1 text-xs">
                  <li>• 我们会认真审核每一个举报</li>
                  <li>• 恶意举报可能导致账户受限</li>
                  <li>• 审核结果将通过消息通知您</li>
                </ul>
              </div>
            </div>
          </div>
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => handleOpenChange(false)}
            disabled={isSubmitting}
          >
            取消
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={!isFormValid || isSubmitting}
            className="min-w-[80px]"
          >
            {isSubmitting ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                提交中...
              </>
            ) : (
              '提交举报'
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

export default ReportCommentDialog