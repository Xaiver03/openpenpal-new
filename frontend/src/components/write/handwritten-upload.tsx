'use client'

import React, { useState, useRef, useCallback } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Progress } from '@/components/ui/progress'
import { 
  Upload, 
  X, 
  FileImage, 
  Loader2, 
  CheckCircle, 
  AlertCircle,
  Image as ImageIcon,
  ZoomIn,
  RotateCw
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { apiClient } from '@/lib/api-client'
import { ocrApi } from '@/lib/api/ocr'

interface HandwrittenUploadProps {
  onImagesUploaded?: (images: UploadedImage[]) => void
  onTextExtracted?: (text: string) => void
  maxImages?: number
  maxFileSize?: number // in MB
}

interface UploadedImage {
  id: string
  url: string
  file: File
  preview: string
}

const ACCEPTED_IMAGE_TYPES = ['image/jpeg', 'image/jpg', 'image/png', 'image/webp', 'image/gif']
const DEFAULT_MAX_FILE_SIZE = 10 // 10MB
const DEFAULT_MAX_IMAGES = 5

export function HandwrittenUpload({
  onImagesUploaded,
  onTextExtracted,
  maxImages = DEFAULT_MAX_IMAGES,
  maxFileSize = DEFAULT_MAX_FILE_SIZE
}: HandwrittenUploadProps) {
  const [images, setImages] = useState<UploadedImage[]>([])
  const [uploading, setUploading] = useState(false)
  const [uploadProgress, setUploadProgress] = useState(0)
  const [error, setError] = useState<string | null>(null)
  const [isDragging, setIsDragging] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  // 处理文件选择
  const handleFileSelect = useCallback((files: FileList | null) => {
    if (!files || files.length === 0) return

    setError(null)
    const newImages: UploadedImage[] = []

    for (let i = 0; i < files.length; i++) {
      const file = files[i]

      // 检查文件类型
      if (!ACCEPTED_IMAGE_TYPES.includes(file.type)) {
        setError(`不支持的文件类型: ${file.name}`)
        continue
      }

      // 检查文件大小
      if (file.size > maxFileSize * 1024 * 1024) {
        setError(`文件过大: ${file.name} (最大 ${maxFileSize}MB)`)
        continue
      }

      // 检查图片数量
      if (images.length + newImages.length >= maxImages) {
        setError(`最多只能上传 ${maxImages} 张图片`)
        break
      }

      // 创建预览
      const reader = new FileReader()
      reader.onload = (e) => {
        const preview = e.target?.result as string
        const uploadedImage: UploadedImage = {
          id: `${Date.now()}-${i}`,
          url: '', // 上传后填充
          file,
          preview
        }
        setImages(prev => [...prev, uploadedImage])
      }
      reader.readAsDataURL(file)
    }
  }, [images.length, maxFileSize, maxImages])

  // 拖拽处理
  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(true)
  }, [])

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(false)
  }, [])

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(false)
    handleFileSelect(e.dataTransfer.files)
  }, [handleFileSelect])

  // 删除图片
  const handleRemoveImage = useCallback((id: string) => {
    setImages(prev => prev.filter(img => img.id !== id))
  }, [])

  // 上传图片到服务器
  const handleUpload = async () => {
    if (images.length === 0) return

    setUploading(true)
    setUploadProgress(0)
    setError(null)

    try {
      const uploadedImages: UploadedImage[] = []
      
      for (let i = 0; i < images.length; i++) {
        const image = images[i]
        const formData = new FormData()
        formData.append('file', image.file)
        formData.append('category', 'handwritten_letter')
        formData.append('is_public', 'false')

        // 上传单个图片
        const response = await apiClient.post<{ url: string }>('/storage/upload', formData, {
          headers: {
            'Content-Type': 'multipart/form-data'
          }
        })
        
        // 模拟进度更新
        const progress = ((i + 1) / images.length) * 100
        setUploadProgress(Math.round(progress))

        uploadedImages.push({
          ...image,
          url: (response.data as any).url
        })
      }

      // 更新图片状态
      setImages(uploadedImages)
      
      // 通知父组件
      if (onImagesUploaded) {
        onImagesUploaded(uploadedImages)
      }

      // 调用OCR服务提取文字
      if (uploadedImages.length > 0 && onTextExtracted) {
        setError(null)
        try {
          // 批量识别所有上传的图片
          const imageUrls = uploadedImages.map(img => img.url)
          const ocrResults = await ocrApi.batchRecognize(imageUrls)
          
          // 合并所有识别结果
          const combinedText = ocrResults
            .map((result, index) => {
              const pageNumber = uploadedImages.length > 1 ? `[第${index + 1}页]\n` : ''
              return pageNumber + result.text
            })
            .join('\n\n')
          
          // 通知父组件
          onTextExtracted(combinedText)
        } catch (ocrError: any) {
          console.error('OCR识别失败:', ocrError)
          setError('文字识别失败，但图片已成功上传。您可以手动输入文字内容。')
        }
      }
      
    } catch (err: any) {
      setError(err.message || '上传失败，请重试')
    } finally {
      setUploading(false)
      setUploadProgress(0)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <FileImage className="h-5 w-5" />
          上传手写信照片
        </CardTitle>
        <CardDescription>
          拍摄或上传你的手写信照片，系统将自动识别文字内容
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* 上传区域 */}
        <div
          className={cn(
            "border-2 border-dashed rounded-lg p-8 text-center cursor-pointer transition-colors",
            isDragging ? "border-primary bg-primary/5" : "border-muted-foreground/25",
            images.length >= maxImages && "opacity-50 cursor-not-allowed"
          )}
          onDragOver={handleDragOver}
          onDragLeave={handleDragLeave}
          onDrop={handleDrop}
          onClick={() => {
            if (images.length < maxImages) {
              fileInputRef.current?.click()
            }
          }}
        >
          <input
            ref={fileInputRef}
            type="file"
            accept={ACCEPTED_IMAGE_TYPES.join(',')}
            multiple
            className="hidden"
            onChange={(e) => handleFileSelect(e.target.files)}
            disabled={images.length >= maxImages}
          />
          
          <Upload className="h-10 w-10 mx-auto mb-4 text-muted-foreground" />
          <p className="text-lg font-medium mb-2">
            {isDragging ? '释放以上传' : '点击或拖拽上传图片'}
          </p>
          <p className="text-sm text-muted-foreground">
            支持 JPG、PNG、WEBP 格式，单个文件最大 {maxFileSize}MB
          </p>
          <p className="text-sm text-muted-foreground mt-1">
            最多可上传 {maxImages} 张图片（已上传 {images.length} 张）
          </p>
        </div>

        {/* 错误提示 */}
        {error && (
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {/* 图片预览 */}
        {images.length > 0 && (
          <div className="space-y-4">
            <h4 className="text-sm font-medium">已选择的图片</h4>
            <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
              {images.map((image) => (
                <div key={image.id} className="relative group">
                  <div className="aspect-[3/4] rounded-lg overflow-hidden bg-muted">
                    <img
                      src={image.preview}
                      alt="手写信预览"
                      className="w-full h-full object-cover"
                    />
                  </div>
                  <Button
                    variant="destructive"
                    size="icon"
                    className="absolute top-2 right-2 h-8 w-8 opacity-0 group-hover:opacity-100 transition-opacity"
                    onClick={() => handleRemoveImage(image.id)}
                  >
                    <X className="h-4 w-4" />
                  </Button>
                  {image.url && (
                    <div className="absolute bottom-2 left-2">
                      <CheckCircle className="h-5 w-5 text-green-500 bg-white rounded-full" />
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>
        )}

        {/* 上传进度 */}
        {uploading && (
          <div className="space-y-2">
            <div className="flex items-center justify-between text-sm">
              <span>正在上传...</span>
              <span>{uploadProgress}%</span>
            </div>
            <Progress value={uploadProgress} />
          </div>
        )}

        {/* 操作按钮 */}
        <div className="flex gap-2">
          <Button
            onClick={handleUpload}
            disabled={images.length === 0 || uploading || images.every(img => img.url)}
            className="flex-1"
          >
            {uploading ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                上传中...
              </>
            ) : images.every(img => img.url) ? (
              <>
                <CheckCircle className="mr-2 h-4 w-4" />
                已上传
              </>
            ) : (
              <>
                <Upload className="mr-2 h-4 w-4" />
                上传图片
              </>
            )}
          </Button>
        </div>

        {/* 使用提示 */}
        <div className="rounded-lg bg-muted/50 p-4 space-y-2">
          <h5 className="text-sm font-medium">拍摄建议</h5>
          <ul className="text-sm text-muted-foreground space-y-1">
            <li>• 确保光线充足，避免阴影</li>
            <li>• 将信纸放平，确保文字清晰</li>
            <li>• 尽量拍摄完整的信纸内容</li>
            <li>• 如果内容较多，可分多张拍摄</li>
          </ul>
        </div>
      </CardContent>
    </Card>
  )
}