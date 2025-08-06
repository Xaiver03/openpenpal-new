import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/zh-cn'

// 配置dayjs
dayjs.extend(relativeTime)
dayjs.locale('zh-cn')

/**
 * 格式化日期时间
 * @param date 日期字符串或Date对象
 * @param format 格式字符串，默认 'YYYY-MM-DD HH:mm:ss'
 * @returns 格式化后的日期字符串
 */
export function formatDate(date: string | Date, format = 'YYYY-MM-DD HH:mm:ss'): string {
  if (!date) return ''
  return dayjs(date).format(format)
}

/**
 * 格式化为相对时间
 * @param date 日期字符串或Date对象
 * @returns 相对时间字符串，如 "3小时前"
 */
export function formatRelativeTime(date: string | Date): string {
  if (!date) return ''
  return dayjs(date).fromNow()
}

/**
 * 格式化日期（不包含时间）
 * @param date 日期字符串或Date对象
 * @returns 格式化后的日期字符串，如 "2024-01-21"
 */
export function formatDateOnly(date: string | Date): string {
  if (!date) return ''
  return dayjs(date).format('YYYY-MM-DD')
}

/**
 * 格式化时间（不包含日期）
 * @param date 日期字符串或Date对象
 * @returns 格式化后的时间字符串，如 "15:30:45"
 */
export function formatTimeOnly(date: string | Date): string {
  if (!date) return ''
  return dayjs(date).format('HH:mm:ss')
}

/**
 * 获取今天的日期范围
 * @returns 今天的开始和结束时间
 */
export function getTodayRange(): [string, string] {
  const today = dayjs()
  return [
    today.startOf('day').toISOString(),
    today.endOf('day').toISOString()
  ]
}

/**
 * 获取本周的日期范围
 * @returns 本周的开始和结束时间
 */
export function getThisWeekRange(): [string, string] {
  const today = dayjs()
  return [
    today.startOf('week').toISOString(),
    today.endOf('week').toISOString()
  ]
}

/**
 * 获取本月的日期范围
 * @returns 本月的开始和结束时间
 */
export function getThisMonthRange(): [string, string] {
  const today = dayjs()
  return [
    today.startOf('month').toISOString(),
    today.endOf('month').toISOString()
  ]
}

/**
 * 获取最近N天的日期范围
 * @param days 天数
 * @returns 最近N天的开始和结束时间
 */
export function getRecentDaysRange(days: number): [string, string] {
  const today = dayjs()
  return [
    today.subtract(days - 1, 'day').startOf('day').toISOString(),
    today.endOf('day').toISOString()
  ]
}

/**
 * 判断日期是否是今天
 * @param date 日期字符串或Date对象
 * @returns 是否是今天
 */
export function isToday(date: string | Date): boolean {
  if (!date) return false
  return dayjs(date).isSame(dayjs(), 'day')
}

/**
 * 判断日期是否是昨天
 * @param date 日期字符串或Date对象
 * @returns 是否是昨天
 */
export function isYesterday(date: string | Date): boolean {
  if (!date) return false
  return dayjs(date).isSame(dayjs().subtract(1, 'day'), 'day')
}

/**
 * 判断日期是否是本周
 * @param date 日期字符串或Date对象
 * @returns 是否是本周
 */
export function isThisWeek(date: string | Date): boolean {
  if (!date) return false
  return dayjs(date).isSame(dayjs(), 'week')
}

/**
 * 判断日期是否是本月
 * @param date 日期字符串或Date对象
 * @returns 是否是本月
 */
export function isThisMonth(date: string | Date): boolean {
  if (!date) return false
  return dayjs(date).isSame(dayjs(), 'month')
}

/**
 * 计算两个日期之间的天数差
 * @param startDate 开始日期
 * @param endDate 结束日期
 * @returns 天数差
 */
export function getDaysDiff(startDate: string | Date, endDate: string | Date): number {
  if (!startDate || !endDate) return 0
  return dayjs(endDate).diff(dayjs(startDate), 'day')
}

/**
 * 计算两个日期之间的小时数差
 * @param startDate 开始日期
 * @param endDate 结束日期
 * @returns 小时数差
 */
export function getHoursDiff(startDate: string | Date, endDate: string | Date): number {
  if (!startDate || !endDate) return 0
  return dayjs(endDate).diff(dayjs(startDate), 'hour')
}

/**
 * 获取日期的显示文本
 * @param date 日期字符串或Date对象
 * @returns 显示文本，如 "今天 15:30", "昨天 15:30", "2024-01-20 15:30"
 */
export function getDisplayText(date: string | Date): string {
  if (!date) return ''
  
  const targetDate = dayjs(date)
  const now = dayjs()
  
  if (targetDate.isSame(now, 'day')) {
    return `今天 ${targetDate.format('HH:mm')}`
  } else if (targetDate.isSame(now.subtract(1, 'day'), 'day')) {
    return `昨天 ${targetDate.format('HH:mm')}`
  } else if (targetDate.isSame(now, 'year')) {
    return targetDate.format('MM-DD HH:mm')
  } else {
    return targetDate.format('YYYY-MM-DD HH:mm')
  }
}

/**
 * 验证日期字符串是否有效
 * @param dateString 日期字符串
 * @returns 是否有效
 */
export function isValidDate(dateString: string): boolean {
  return dayjs(dateString).isValid()
}

/**
 * 获取时间戳
 * @param date 日期字符串或Date对象，默认为当前时间
 * @returns 时间戳（毫秒）
 */
export function getTimestamp(date?: string | Date): number {
  return dayjs(date).valueOf()
}

/**
 * 从时间戳创建日期
 * @param timestamp 时间戳（毫秒）
 * @returns dayjs对象
 */
export function fromTimestamp(timestamp: number) {
  return dayjs(timestamp)
}

/**
 * 格式化持续时间
 * @param seconds 秒数
 * @returns 格式化后的持续时间，如 "2小时30分钟"
 */
export function formatDuration(seconds: number): string {
  if (seconds < 60) {
    return `${seconds}秒`
  } else if (seconds < 3600) {
    const minutes = Math.floor(seconds / 60)
    const remainingSeconds = seconds % 60
    return remainingSeconds > 0 ? `${minutes}分${remainingSeconds}秒` : `${minutes}分钟`
  } else {
    const hours = Math.floor(seconds / 3600)
    const remainingMinutes = Math.floor((seconds % 3600) / 60)
    return remainingMinutes > 0 ? `${hours}小时${remainingMinutes}分钟` : `${hours}小时`
  }
}

/**
 * 获取日期选择器的快捷选项
 * @returns 快捷选项配置
 */
export function getDatePickerShortcuts() {
  return [
    {
      text: '今天',
      value: () => {
        const today = dayjs()
        return [today.startOf('day').toDate(), today.endOf('day').toDate()]
      }
    },
    {
      text: '昨天',
      value: () => {
        const yesterday = dayjs().subtract(1, 'day')
        return [yesterday.startOf('day').toDate(), yesterday.endOf('day').toDate()]
      }
    },
    {
      text: '最近7天',
      value: () => {
        const today = dayjs()
        return [today.subtract(6, 'day').startOf('day').toDate(), today.endOf('day').toDate()]
      }
    },
    {
      text: '最近30天',
      value: () => {
        const today = dayjs()
        return [today.subtract(29, 'day').startOf('day').toDate(), today.endOf('day').toDate()]
      }
    },
    {
      text: '本月',
      value: () => {
        const today = dayjs()
        return [today.startOf('month').toDate(), today.endOf('month').toDate()]
      }
    },
    {
      text: '上月',
      value: () => {
        const lastMonth = dayjs().subtract(1, 'month')
        return [lastMonth.startOf('month').toDate(), lastMonth.endOf('month').toDate()]
      }
    }
  ]
}