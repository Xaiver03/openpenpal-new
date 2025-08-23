import { NextRequest, NextResponse } from 'next/server'

/**
 * CSP Violation Report Endpoint
 * Handles Content Security Policy violation reports
 */
export async function POST(request: NextRequest) {
  try {
    const report = await request.json()
    
    // Log CSP violation for monitoring
    console.warn('CSP Violation Report:', {
      timestamp: new Date().toISOString(),
      documentUri: report['csp-report']?.['document-uri'],
      violatedDirective: report['csp-report']?.['violated-directive'],
      effectiveDirective: report['csp-report']?.['effective-directive'],
      blockedUri: report['csp-report']?.['blocked-uri'],
      sourceFile: report['csp-report']?.['source-file'],
      lineNumber: report['csp-report']?.['line-number'],
      columnNumber: report['csp-report']?.['column-number'],
      userAgent: request.headers.get('user-agent'),
      ip: request.headers.get('x-forwarded-for') || request.headers.get('x-real-ip')
    })
    
    // In production, you would:
    // 1. Store in database for analysis
    // 2. Send to monitoring service (Sentry, etc.)
    // 3. Alert on critical violations
    
    return NextResponse.json({ success: true }, { status: 204 })
  } catch (error) {
    console.error('Failed to process CSP report:', error)
    return NextResponse.json({ error: 'Invalid report' }, { status: 400 })
  }
}

// CSP reports should not require authentication
export const runtime = 'edge'