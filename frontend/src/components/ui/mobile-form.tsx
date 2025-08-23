'use client'

import * as React from 'react'
import { cn } from '@/lib/utils'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Badge } from '@/components/ui/badge'
import { 
  Eye, 
  EyeOff, 
  Check, 
  X, 
  AlertCircle,
  Search,
  Phone,
  Mail,
  User,
  MapPin,
  Calendar,
  Clock,
  Hash,
  ChevronDown,
  Plus,
  Minus
} from 'lucide-react'

// 移动端优化的输入框
interface MobileInputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string
  error?: string
  success?: string
  hint?: string
  icon?: React.ComponentType<any>
  variant?: 'default' | 'large' | 'compact'
  showCounter?: boolean
  maxLength?: number
}

export const MobileInput = React.forwardRef<HTMLInputElement, MobileInputProps>(
  ({ 
    className, 
    type = 'text', 
    label, 
    error, 
    success, 
    hint, 
    icon: Icon,
    variant = 'default',
    showCounter = false,
    maxLength,
    ...props 
  }, ref) => {
    const [showPassword, setShowPassword] = React.useState(false)
    const [value, setValue] = React.useState(props.value || '')

    const inputClasses = cn(
      'flex w-full rounded-lg border border-input bg-background text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50',
      // 移动端优化
      'touch-manipulation', // 禁用双击缩放
      'text-base', // 防止iOS缩放
      variant === 'large' && 'h-14 px-4 text-base',
      variant === 'compact' && 'h-9 px-3 text-sm',
      variant === 'default' && 'h-12 px-4',
      Icon && 'pl-12',
      type === 'password' && 'pr-12',
      error && 'border-red-500 focus-visible:ring-red-500',
      success && 'border-green-500 focus-visible:ring-green-500',
      className
    )

    const handleValueChange = (e: React.ChangeEvent<HTMLInputElement>) => {
      const newValue = e.target.value
      setValue(newValue)
      props.onChange?.(e)
    }

    return (
      <div className="space-y-2">
        {label && (
          <label className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
            {label}
            {props.required && <span className="text-red-500 ml-1">*</span>}
          </label>
        )}
        
        <div className="relative">
          {Icon && (
            <div className="absolute left-3 top-1/2 transform -translate-y-1/2">
              <Icon className="h-4 w-4 text-gray-500" />
            </div>
          )}
          
          <Input
            ref={ref}
            type={type === 'password' && showPassword ? 'text' : type}
            className={inputClasses}
            value={value}
            onChange={handleValueChange}
            maxLength={maxLength}
            // 移动端输入优化
            autoComplete={type === 'password' ? 'current-password' : 'off'}
            autoCapitalize="none"
            autoCorrect="off"
            spellCheck="false"
            {...props}
          />
          
          {type === 'password' && (
            <Button
              type="button"
              variant="ghost"
              size="sm"
              className="absolute right-2 top-1/2 transform -translate-y-1/2 h-8 w-8 p-0"
              onClick={() => setShowPassword(!showPassword)}
            >
              {showPassword ? (
                <EyeOff className="h-4 w-4" />
              ) : (
                <Eye className="h-4 w-4" />
              )}
            </Button>
          )}
          
          {showCounter && maxLength && (
            <div className="absolute right-2 top-1/2 transform -translate-y-1/2 text-xs text-gray-500">
              {String(value).length}/{maxLength}
            </div>
          )}
        </div>
        
        {/* 状态信息 */}
        {error && (
          <div className="flex items-center gap-2 text-red-600 text-sm">
            <AlertCircle className="h-4 w-4" />
            {error}
          </div>
        )}
        
        {success && (
          <div className="flex items-center gap-2 text-green-600 text-sm">
            <Check className="h-4 w-4" />
            {success}
          </div>
        )}
        
        {hint && !error && !success && (
          <p className="text-xs text-gray-500">{hint}</p>
        )}
      </div>
    )
  }
)
MobileInput.displayName = 'MobileInput'

// 移动端优化的文本域
interface MobileTextareaProps extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {
  label?: string
  error?: string
  hint?: string
  showCounter?: boolean
  maxLength?: number
  autoResize?: boolean
}

export const MobileTextarea = React.forwardRef<HTMLTextAreaElement, MobileTextareaProps>(
  ({ 
    className, 
    label, 
    error, 
    hint, 
    showCounter = false,
    maxLength,
    autoResize = false,
    ...props 
  }, ref) => {
    const [value, setValue] = React.useState(props.value || '')
    const textareaRef = React.useRef<HTMLTextAreaElement>(null)

    React.useImperativeHandle(ref, () => textareaRef.current!, [])

    const handleValueChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
      const newValue = e.target.value
      setValue(newValue)
      
      // 自动调整高度
      if (autoResize && textareaRef.current) {
        textareaRef.current.style.height = 'auto'
        textareaRef.current.style.height = `${textareaRef.current.scrollHeight}px`
      }
      
      props.onChange?.(e)
    }

    return (
      <div className="space-y-2">
        {label && (
          <label className="text-sm font-medium leading-none">
            {label}
            {props.required && <span className="text-red-500 ml-1">*</span>}
          </label>
        )}
        
        <div className="relative">
          <Textarea
            ref={textareaRef}
            className={cn(
              'touch-manipulation text-base resize-none',
              error && 'border-red-500 focus-visible:ring-red-500',
              className
            )}
            value={value}
            onChange={handleValueChange}
            maxLength={maxLength}
            autoComplete="off"
            autoCapitalize="sentences"
            spellCheck="true"
            {...props}
          />
          
          {showCounter && maxLength && (
            <div className="absolute right-2 bottom-2 text-xs text-gray-500 bg-background px-1">
              {String(value).length}/{maxLength}
            </div>
          )}
        </div>
        
        {error && (
          <div className="flex items-center gap-2 text-red-600 text-sm">
            <AlertCircle className="h-4 w-4" />
            {error}
          </div>
        )}
        
        {hint && !error && (
          <p className="text-xs text-gray-500">{hint}</p>
        )}
      </div>
    )
  }
)
MobileTextarea.displayName = 'MobileTextarea'

// 移动端选择器
interface MobileSelectProps {
  label?: string
  placeholder?: string
  options: Array<{ value: string; label: string; disabled?: boolean }>
  value?: string
  onValueChange?: (value: string) => void
  error?: string
  hint?: string
  icon?: React.ComponentType<any>
}

export function MobileSelect({ 
  label, 
  placeholder = '请选择...', 
  options, 
  value, 
  onValueChange,
  error,
  hint,
  icon: Icon 
}: MobileSelectProps) {
  const [isOpen, setIsOpen] = React.useState(false)
  const selectedOption = options.find(opt => opt.value === value)

  return (
    <div className="space-y-2">
      {label && (
        <label className="text-sm font-medium leading-none">
          {label}
        </label>
      )}
      
      <div className="relative">
        <button
          type="button"
          className={cn(
            'flex h-12 w-full items-center justify-between rounded-lg border border-input bg-background px-4 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50',
            Icon && 'pl-12',
            error && 'border-red-500 focus:ring-red-500'
          )}
          onClick={() => setIsOpen(!isOpen)}
        >
          {Icon && (
            <div className="absolute left-3 top-1/2 transform -translate-y-1/2">
              <Icon className="h-4 w-4 text-gray-500" />
            </div>
          )}
          
          <span className={selectedOption ? '' : 'text-muted-foreground'}>
            {selectedOption ? selectedOption.label : placeholder}
          </span>
          
          <ChevronDown className="h-4 w-4 opacity-50" />
        </button>
        
        {isOpen && (
          <div className="absolute top-full left-0 right-0 z-50 mt-1 bg-background border border-input rounded-lg shadow-lg max-h-60 overflow-y-auto">
            {options.map((option) => (
              <button
                key={option.value}
                type="button"
                className={cn(
                  'w-full px-4 py-3 text-left text-sm hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground first:rounded-t-lg last:rounded-b-lg',
                  option.disabled && 'opacity-50 cursor-not-allowed',
                  value === option.value && 'bg-accent text-accent-foreground'
                )}
                onClick={() => {
                  if (!option.disabled) {
                    onValueChange?.(option.value)
                    setIsOpen(false)
                  }
                }}
                disabled={option.disabled}
              >
                {option.label}
                {value === option.value && (
                  <Check className="h-4 w-4 ml-auto inline" />
                )}
              </button>
            ))}
          </div>
        )}
      </div>
      
      {/* 点击外部关闭 */}
      {isOpen && (
        <div
          className="fixed inset-0 z-40"
          onClick={() => setIsOpen(false)}
        />
      )}
      
      {error && (
        <div className="flex items-center gap-2 text-red-600 text-sm">
          <AlertCircle className="h-4 w-4" />
          {error}
        </div>
      )}
      
      {hint && !error && (
        <p className="text-xs text-gray-500">{hint}</p>
      )}
    </div>
  )
}

// 移动端数量选择器
interface MobileNumberStepperProps {
  label?: string
  value: number
  onValueChange: (value: number) => void
  min?: number
  max?: number
  step?: number
  disabled?: boolean
}

export function MobileNumberStepper({
  label,
  value,
  onValueChange,
  min = 0,
  max = 100,
  step = 1,
  disabled = false
}: MobileNumberStepperProps) {
  const canDecrease = value > min
  const canIncrease = value < max

  const handleDecrease = () => {
    if (canDecrease) {
      onValueChange(Math.max(min, value - step))
    }
  }

  const handleIncrease = () => {
    if (canIncrease) {
      onValueChange(Math.min(max, value + step))
    }
  }

  return (
    <div className="space-y-2">
      {label && (
        <label className="text-sm font-medium leading-none">
          {label}
        </label>
      )}
      
      <div className="flex items-center space-x-3">
        <Button
          type="button"
          variant="outline"
          size="icon"
          className="h-10 w-10 rounded-full"
          onClick={handleDecrease}
          disabled={disabled || !canDecrease}
        >
          <Minus className="h-4 w-4" />
        </Button>
        
        <div className="flex-1 text-center">
          <div className="text-2xl font-bold">{value}</div>
        </div>
        
        <Button
          type="button"
          variant="outline"
          size="icon"
          className="h-10 w-10 rounded-full"
          onClick={handleIncrease}
          disabled={disabled || !canIncrease}
        >
          <Plus className="h-4 w-4" />
        </Button>
      </div>
      
      <div className="text-xs text-center text-gray-500">
        {min} - {max}
      </div>
    </div>
  )
}

// 快速输入标签
interface QuickInputTagsProps {
  label?: string
  suggestions: string[]
  selectedTags: string[]
  onTagsChange: (tags: string[]) => void
  placeholder?: string
  maxTags?: number
}

export function QuickInputTags({
  label,
  suggestions,
  selectedTags,
  onTagsChange,
  placeholder = '点击选择或输入...',
  maxTags = 10
}: QuickInputTagsProps) {
  const [inputValue, setInputValue] = React.useState('')

  const addTag = (tag: string) => {
    const trimmedTag = tag.trim()
    if (trimmedTag && !selectedTags.includes(trimmedTag) && selectedTags.length < maxTags) {
      onTagsChange([...selectedTags, trimmedTag])
      setInputValue('')
    }
  }

  const removeTag = (tagToRemove: string) => {
    onTagsChange(selectedTags.filter(tag => tag !== tagToRemove))
  }

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      addTag(inputValue)
    }
  }

  const availableSuggestions = suggestions.filter(
    suggestion => !selectedTags.includes(suggestion)
  )

  return (
    <div className="space-y-3">
      {label && (
        <label className="text-sm font-medium leading-none">
          {label}
        </label>
      )}
      
      {/* 已选择的标签 */}
      {selectedTags.length > 0 && (
        <div className="flex flex-wrap gap-2">
          {selectedTags.map((tag) => (
            <Badge
              key={tag}
              variant="secondary"
              className="pl-3 pr-1 py-1 text-sm"
            >
              {tag}
              <Button
                type="button"
                variant="ghost"
                size="sm"
                className="h-4 w-4 p-0 ml-2 hover:bg-transparent"
                onClick={() => removeTag(tag)}
              >
                <X className="h-3 w-3" />
              </Button>
            </Badge>
          ))}
        </div>
      )}
      
      {/* 输入框 */}
      <div className="relative">
        <Input
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          onKeyPress={handleKeyPress}
          placeholder={placeholder}
          className="h-12 text-base"
          disabled={selectedTags.length >= maxTags}
        />
        
        {inputValue && (
          <Button
            type="button"
            variant="ghost"
            size="sm"
            className="absolute right-2 top-1/2 transform -translate-y-1/2 h-8 px-2 text-xs"
            onClick={() => addTag(inputValue)}
          >
            添加
          </Button>
        )}
      </div>
      
      {/* 建议标签 */}
      {availableSuggestions.length > 0 && (
        <div className="space-y-2">
          <p className="text-xs text-gray-500">建议标签：</p>
          <div className="flex flex-wrap gap-2">
            {availableSuggestions.slice(0, 8).map((suggestion) => (
              <Button
                key={suggestion}
                type="button"
                variant="outline"
                size="sm"
                className="h-8 text-xs"
                onClick={() => addTag(suggestion)}
                disabled={selectedTags.length >= maxTags}
              >
                {suggestion}
              </Button>
            ))}
          </div>
        </div>
      )}
      
      <p className="text-xs text-gray-500">
        {selectedTags.length}/{maxTags} 个标签
      </p>
    </div>
  )
}

// 预定义输入类型
export function PhoneInput(props: Omit<MobileInputProps, 'type' | 'icon'>) {
  return <MobileInput {...props} type="tel" icon={Phone} />
}

export function EmailInput(props: Omit<MobileInputProps, 'type' | 'icon'>) {
  return <MobileInput {...props} type="email" icon={Mail} />
}

export function UserInput(props: Omit<MobileInputProps, 'icon'>) {
  return <MobileInput {...props} icon={User} />
}

export function AddressInput(props: Omit<MobileInputProps, 'icon'>) {
  return <MobileInput {...props} icon={MapPin} />
}

export function SearchInput(props: Omit<MobileInputProps, 'icon'>) {
  return <MobileInput {...props} icon={Search} />
}