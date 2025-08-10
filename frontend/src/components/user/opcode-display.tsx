import React, { useState } from 'react'
import { cn } from '@/lib/utils'
import { MapPin, Edit2, Eye, EyeOff } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { 
  Popover, 
  PopoverContent, 
  PopoverTrigger 
} from '@/components/ui/popover'

interface OPCodeDisplayProps {
  opCode?: string
  showPrivacy?: boolean
  editable?: boolean
  onUpdate?: (newCode: string) => void
  className?: string
}

// OP Code 解析配置
const SCHOOL_CODES = {
  'PK': '北京大学',
  'QH': '清华大学', 
  'BD': '北京交通大学',
  'RU': '中国人民大学',
  'BN': '北京师范大学'
}

const AREA_CODES = {
  '5F': '5号楼',
  '3D': '3号食堂',
  '2G': '2号门',
  '1L': '1号宿舍楼',
  '7B': '7号教学楼'
}

function parseOPCode(code: string) {
  if (!code || code.length !== 6) return null
  
  const schoolCode = code.substring(0, 2)
  const areaCode = code.substring(2, 4)  
  const pointCode = code.substring(4, 6)
  
  return {
    school: SCHOOL_CODES[schoolCode as keyof typeof SCHOOL_CODES] || schoolCode,
    area: AREA_CODES[areaCode as keyof typeof AREA_CODES] || areaCode,
    point: pointCode,
    formatted: `${schoolCode}${areaCode}${pointCode}`
  }
}

function formatOPCodeForDisplay(code: string, showPrivacy = false) {
  if (!code) return ''
  
  if (showPrivacy && code.length === 6) {
    return code.substring(0, 4) + '**'
  }
  
  return code
}

export function OPCodeDisplay({ 
  opCode, 
  showPrivacy = false, 
  editable = false,
  onUpdate,
  className 
}: OPCodeDisplayProps) {
  const [isEditing, setIsEditing] = useState(false)
  const [editValue, setEditValue] = useState(opCode || '')
  const [isPrivate, setIsPrivate] = useState(showPrivacy)
  
  const parsed = parseOPCode(opCode || '')
  const displayCode = formatOPCodeForDisplay(opCode || '', isPrivate)
  
  const handleSave = () => {
    if (editValue.length === 6 && onUpdate) {
      onUpdate(editValue.toUpperCase())
    }
    setIsEditing(false)
  }

  const handleCancel = () => {
    setEditValue(opCode || '')
    setIsEditing(false)
  }

  if (!opCode && !editable) {
    return null
  }

  return (
    <div className={cn('flex items-center gap-2', className)}>
      <MapPin className="h-4 w-4 text-amber-600" />
      
      {!isEditing ? (
        <div className="flex items-center gap-2">
          <Popover>
            <PopoverTrigger asChild>
              <button className="flex items-center gap-1 text-sm font-mono bg-amber-50 text-amber-700 px-2 py-1 rounded border border-amber-200 hover:bg-amber-100 transition-colors">
                {displayCode || '未设置'}
                {parsed && (
                  <span className="text-xs text-amber-600 ml-1">
                    ({parsed.school})
                  </span>
                )}
              </button>
            </PopoverTrigger>
            <PopoverContent className="w-64" align="start">
              <div className="space-y-2">
                <div className="font-medium text-sm">OP Code 详情</div>
                {parsed ? (
                  <div className="space-y-1 text-xs">
                    <div className="flex justify-between">
                      <span className="text-gray-500">完整编号:</span>
                      <span className="font-mono">{parsed.formatted}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-500">学校:</span>
                      <span>{parsed.school}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-500">区域:</span>
                      <span>{parsed.area}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-500">位置:</span>
                      <span>{parsed.point}</span>
                    </div>
                  </div>
                ) : (
                  <div className="text-xs text-gray-500">
                    OP Code格式：学校(2位) + 区域(2位) + 位置(2位)
                  </div>
                )}
              </div>
            </PopoverContent>
          </Popover>
          
          {opCode && (
            <button
              onClick={() => setIsPrivate(!isPrivate)}
              className="p-1 text-gray-400 hover:text-gray-600 transition-colors"
              title={isPrivate ? '显示完整编号' : '隐藏部分编号'}
            >
              {isPrivate ? <EyeOff className="h-3 w-3" /> : <Eye className="h-3 w-3" />}
            </button>
          )}
          
          {editable && (
            <button
              onClick={() => setIsEditing(true)}
              className="p-1 text-gray-400 hover:text-amber-600 transition-colors"
              title="编辑OP Code"
            >
              <Edit2 className="h-3 w-3" />
            </button>
          )}
        </div>
      ) : (
        <div className="flex items-center gap-2">
          <Input
            value={editValue}
            onChange={(e) => setEditValue(e.target.value.toUpperCase().slice(0, 6))}
            placeholder="PK5F3D"
            className="w-20 h-8 text-xs font-mono"
            maxLength={6}
          />
          <Button
            size="sm"
            onClick={handleSave}
            disabled={editValue.length !== 6}
            className="h-8 px-2 text-xs"
          >
            保存
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={handleCancel}
            className="h-8 px-2 text-xs"
          >
            取消
          </Button>
        </div>
      )}
    </div>
  )
}