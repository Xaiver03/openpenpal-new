'use client'

import { useState, useRef } from 'react'
import { Upload, X, Camera } from 'lucide-react'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import { uploadAvatar, removeAvatar } from '@/lib/api/user'
import { useAuth } from '@/stores/user-store'
import { toast } from 'sonner'
import { cn } from '@/lib/utils'

interface AvatarUploadProps {
  currentAvatar?: string
  username?: string
  nickname?: string
  onAvatarChange?: (avatarUrl: string | null) => void
  className?: string
}

export function AvatarUpload({ 
  currentAvatar, 
  username, 
  nickname,
  onAvatarChange,
  className 
}: AvatarUploadProps) {
  const [isUploading, setIsUploading] = useState(false)
  const [previewUrl, setPreviewUrl] = useState<string | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const { refreshUser } = useAuth()

  const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    // Validate file type
    const allowedTypes = ['image/jpeg', 'image/jpg', 'image/png', 'image/gif', 'image/webp']
    if (!allowedTypes.includes(file.type)) {
      toast.error('只支持 JPG、PNG、GIF 和 WEBP 格式的图片')
      return
    }

    // Validate file size (5MB)
    if (file.size > 5 * 1024 * 1024) {
      toast.error('图片大小不能超过 5MB')
      return
    }

    // Create preview
    const reader = new FileReader()
    reader.onload = (e) => {
      setPreviewUrl(e.target?.result as string)
    }
    reader.readAsDataURL(file)

    // Upload file
    setIsUploading(true)
    try {
      const response = await uploadAvatar(file)
      if (response.avatar_url) {
        toast.success('头像上传成功')
        onAvatarChange?.(response.avatar_url)
        await refreshUser()
      }
    } catch (error) {
      console.error('Avatar upload failed:', error)
      toast.error('头像上传失败，请重试')
      setPreviewUrl(null)
    } finally {
      setIsUploading(false)
    }
  }

  const handleRemoveAvatar = async () => {
    if (!currentAvatar && !previewUrl) return

    setIsUploading(true)
    try {
      await removeAvatar()
      toast.success('头像已移除')
      setPreviewUrl(null)
      onAvatarChange?.(null)
      await refreshUser()
    } catch (error) {
      console.error('Avatar removal failed:', error)
      toast.error('移除头像失败，请重试')
    } finally {
      setIsUploading(false)
    }
  }

  const displayAvatar = previewUrl || currentAvatar
  const displayName = nickname || username || 'U'

  return (
    <div className={cn("flex flex-col items-center space-y-4", className)}>
      <div className="relative group">
        <Avatar className="h-24 w-24 border-2 border-border">
          <AvatarImage src={displayAvatar} alt={displayName} />
          <AvatarFallback className="text-xl">
            {displayName.slice(0, 2).toUpperCase()}
          </AvatarFallback>
        </Avatar>
        
        <div className="absolute inset-0 flex items-center justify-center bg-black/60 rounded-full opacity-0 group-hover:opacity-100 transition-opacity">
          <Button
            size="icon"
            variant="ghost"
            className="text-white hover:text-white hover:bg-white/20"
            onClick={() => fileInputRef.current?.click()}
            disabled={isUploading}
          >
            <Camera className="h-5 w-5" />
          </Button>
        </div>

        {displayAvatar && (
          <Button
            size="icon"
            variant="secondary"
            className="absolute -top-2 -right-2 h-6 w-6 rounded-full"
            onClick={handleRemoveAvatar}
            disabled={isUploading}
          >
            <X className="h-3 w-3" />
          </Button>
        )}
      </div>

      <input
        ref={fileInputRef}
        type="file"
        accept="image/jpeg,image/jpg,image/png,image/gif,image/webp"
        onChange={handleFileSelect}
        className="hidden"
      />

      <div className="flex items-center space-x-2">
        <Button
          size="sm"
          variant="outline"
          onClick={() => fileInputRef.current?.click()}
          disabled={isUploading}
        >
          <Upload className="h-4 w-4 mr-2" />
          上传头像
        </Button>
        
        {displayAvatar && (
          <Button
            size="sm"
            variant="outline"
            onClick={handleRemoveAvatar}
            disabled={isUploading}
          >
            移除头像
          </Button>
        )}
      </div>

      <p className="text-xs text-muted-foreground text-center">
        支持 JPG、PNG、GIF、WEBP 格式，最大 5MB
      </p>
    </div>
  )
}