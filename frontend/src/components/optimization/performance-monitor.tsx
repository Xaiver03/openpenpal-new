'use client'

import { useEffect, useState } from 'react'
import { usePerformanceMetrics } from './performance-wrapper'

interface PerformanceData {
  lcp: number
  fid: number
  cls: number
  ttfb: number
  loadTime: number
  renderTime: number
}

export function PerformanceMonitor() {
  const { metrics, reportMetrics } = usePerformanceMetrics()
  const [showReport, setShowReport] = useState(false)

  useEffect(() => {
    // 在开发环境下自动显示性能报告
    if (process.env.NODE_ENV === 'development') {
      const timer = setTimeout(() => {
        reportMetrics()
        setShowReport(true)
      }, 3000)
      
      return () => clearTimeout(timer)
    }
  }, [reportMetrics])

  // 生产环境下不显示任何UI
  if (process.env.NODE_ENV === 'production') {
    return null
  }

  if (!showReport) {
    return null
  }

  const getScoreColor = (metric: string, value: number) => {
    switch (metric) {
      case 'lcp':
        return value <= 2500 ? 'text-green-600' : value <= 4000 ? 'text-yellow-600' : 'text-red-600'
      case 'fid':
        return value <= 100 ? 'text-green-600' : value <= 300 ? 'text-yellow-600' : 'text-red-600'
      case 'cls':
        return value <= 0.1 ? 'text-green-600' : value <= 0.25 ? 'text-yellow-600' : 'text-red-600'
      case 'ttfb':
        return value <= 600 ? 'text-green-600' : value <= 1200 ? 'text-yellow-600' : 'text-red-600'
      default:
        return 'text-gray-600'
    }
  }

  return (
    <div className="fixed bottom-4 right-4 bg-white border border-gray-200 rounded-lg shadow-lg p-4 max-w-xs z-50">
      <div className="flex items-center justify-between mb-3">
        <h3 className="text-sm font-semibold text-gray-900">性能监控</h3>
        <button
          onClick={() => setShowReport(false)}
          className="text-gray-400 hover:text-gray-600"
        >
          ✕
        </button>
      </div>
      
      <div className="space-y-2 text-xs">
        <div className="flex justify-between">
          <span>LCP:</span>
          <span className={getScoreColor('lcp', metrics.lcp)}>
            {metrics.lcp > 0 ? `${metrics.lcp.toFixed(0)}ms` : '-'}
          </span>
        </div>
        
        <div className="flex justify-between">
          <span>FID:</span>
          <span className={getScoreColor('fid', metrics.fid)}>
            {metrics.fid > 0 ? `${metrics.fid.toFixed(1)}ms` : '-'}
          </span>
        </div>
        
        <div className="flex justify-between">
          <span>CLS:</span>
          <span className={getScoreColor('cls', metrics.cls)}>
            {metrics.cls > 0 ? metrics.cls.toFixed(3) : '-'}
          </span>
        </div>
        
        <div className="flex justify-between">
          <span>TTFB:</span>
          <span className={getScoreColor('ttfb', metrics.ttfb)}>
            {metrics.ttfb > 0 ? `${metrics.ttfb.toFixed(0)}ms` : '-'}
          </span>
        </div>
        
        <div className="flex justify-between">
          <span>Load:</span>
          <span className="text-gray-600">
            {metrics.loadTime > 0 ? `${metrics.loadTime.toFixed(0)}ms` : '-'}
          </span>
        </div>
      </div>
      
      <button
        onClick={reportMetrics}
        className="mt-3 w-full text-xs bg-blue-500 text-white px-2 py-1 rounded hover:bg-blue-600"
      >
        更新报告
      </button>
    </div>
  )
}