import { NextRequest, NextResponse } from 'next/server'
import { getHealthMetrics, metricsCollector, SLOCalculator } from '@/lib/monitoring/metrics'
import { getCacheService } from '@/lib/cache/redis'

export async function GET(request: NextRequest) {
  const { searchParams } = new URL(request.url)
  const format = searchParams.get('format') || 'json'
  const detailed = searchParams.get('detailed') === 'true'

  try {
    // Collect health metrics
    const healthData = await getHealthMetrics()
    
    // Check cache health
    const cacheService = getCacheService()
    const cacheHealth = await cacheService.healthCheck()
    
    // Database health check (mock - implement actual check)
    const dbHealth = {
      connected: true,
      responseTime: 5, // ms
      activeConnections: 3
    }

    const health = {
      ...healthData,
      services: {
        cache: {
          status: cacheHealth.redis ? 'healthy' : 'degraded',
          details: cacheHealth
        },
        database: {
          status: dbHealth.connected ? 'healthy' : 'unhealthy',
          details: dbHealth
        },
        api: {
          status: 'healthy',
          version: '2.0.0'
        }
      }
    }

    // Determine overall status
    const serviceStatuses = Object.values(health.services).map(s => s.status)
    const overallStatus = serviceStatuses.some(s => s === 'unhealthy') 
      ? 'unhealthy' 
      : serviceStatuses.some(s => s === 'degraded') 
        ? 'degraded' 
        : 'healthy'

    const response = {
      status: overallStatus,
      timestamp: health.timestamp,
      uptime: health.uptime,
      version: '2.0.0',
      services: health.services,
      sli: health.slis,
      ...(health.violations.length > 0 && { violations: health.violations }),
      ...(detailed && { 
        metrics: health.metrics,
        environment: process.env.NODE_ENV 
      })
    }

    // Return as HTML for browser viewing
    if (format === 'html') {
      const htmlResponse = generateHealthHTML(response)
      return new Response(htmlResponse, {
        headers: { 'Content-Type': 'text/html' }
      })
    }

    // Return appropriate HTTP status
    const httpStatus = overallStatus === 'healthy' ? 200 : 
                      overallStatus === 'degraded' ? 200 : 503

    return NextResponse.json(response, { status: httpStatus })
    
  } catch (error) {
    console.error('Health check failed:', error)
    
    return NextResponse.json({
      status: 'unhealthy',
      timestamp: new Date().toISOString(),
      error: 'Health check failed',
      details: error instanceof Error ? error.message : 'Unknown error'
    }, { status: 503 })
  }
}

function generateHealthHTML(health: any): string {
  const statusColors = {
    healthy: '#10b981',
    degraded: '#f59e0b', 
    unhealthy: '#ef4444'
  }

  const statusColor = statusColors[health.status as keyof typeof statusColors] || '#6b7280'

  return `
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>OpenPenPal Health Dashboard</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
      background: #f8fafc;
      color: #1a202c;
      line-height: 1.6;
    }
    .header {
      background: linear-gradient(135deg, #d97706 0%, #92400e 100%);
      color: white;
      padding: 2rem 0;
      text-align: center;
    }
    .container {
      max-width: 1200px;
      margin: 0 auto;
      padding: 2rem;
    }
    .status-badge {
      display: inline-block;
      background: ${statusColor};
      color: white;
      padding: 0.5rem 1rem;
      border-radius: 9999px;
      font-weight: 600;
      font-size: 0.875rem;
      text-transform: uppercase;
      margin-bottom: 1rem;
    }
    .title {
      font-size: 2.5rem;
      font-weight: 700;
      margin: 0;
    }
    .subtitle {
      font-size: 1rem;
      margin: 0.5rem 0 0 0;
      opacity: 0.9;
    }
    .grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
      gap: 1.5rem;
      margin: 2rem 0;
    }
    .card {
      background: white;
      border-radius: 8px;
      padding: 1.5rem;
      box-shadow: 0 1px 3px rgba(0,0,0,0.1);
      border-left: 4px solid #e2e8f0;
    }
    .card.healthy { border-left-color: #10b981; }
    .card.degraded { border-left-color: #f59e0b; }
    .card.unhealthy { border-left-color: #ef4444; }
    .card-title {
      font-size: 1.25rem;
      font-weight: 600;
      margin-bottom: 1rem;
      display: flex;
      align-items: center;
      gap: 0.5rem;
    }
    .status-indicator {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      background: #10b981;
    }
    .status-indicator.degraded { background: #f59e0b; }
    .status-indicator.unhealthy { background: #ef4444; }
    .metric {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 0.5rem 0;
      border-bottom: 1px solid #f1f5f9;
    }
    .metric:last-child { border-bottom: none; }
    .metric-label { color: #64748b; }
    .metric-value { font-weight: 600; }
    .violations {
      background: #fef2f2;
      border: 1px solid #fecaca;
      border-radius: 6px;
      padding: 1rem;
      margin: 1rem 0;
    }
    .violations-title {
      color: #dc2626;
      font-weight: 600;
      margin-bottom: 0.5rem;
    }
    .violations ul {
      list-style: none;
      padding-left: 0;
    }
    .violations li {
      color: #dc2626;
      padding: 0.25rem 0;
    }
    .violations li:before {
      content: '⚠️ ';
      margin-right: 0.5rem;
    }
    .refresh-btn {
      background: #d97706;
      color: white;
      border: none;
      padding: 0.75rem 1.5rem;
      border-radius: 6px;
      cursor: pointer;
      font-weight: 600;
      margin: 1rem 0;
    }
    .refresh-btn:hover { background: #b45309; }
    .timestamp {
      color: #64748b;
      font-size: 0.875rem;
      text-align: center;
      margin-top: 2rem;
    }
    @media (max-width: 768px) {
      .grid { grid-template-columns: 1fr; }
      .container { padding: 1rem; }
    }
  </style>
</head>
<body>
  <div class="header">
    <div class="container">
      <div class="status-badge">${health.status}</div>
      <h1 class="title">System Health Dashboard</h1>
      <p class="subtitle">OpenPenPal Platform Monitoring</p>
    </div>
  </div>

  <div class="container">
    <button class="refresh-btn" onclick="window.location.reload()">🔄 Refresh Status</button>

    ${health.violations?.length > 0 ? `
    <div class="violations">
      <div class="violations-title">SLO Violations Detected</div>
      <ul>
        ${health.violations.map((v: string) => `<li>${v}</li>`).join('')}
      </ul>
    </div>
    ` : ''}

    <div class="grid">
      <!-- Overall Status -->
      <div class="card ${health.status}">
        <div class="card-title">
          <div class="status-indicator ${health.status}"></div>
          Overall Status
        </div>
        <div class="metric">
          <span class="metric-label">Status</span>
          <span class="metric-value">${health.status.toUpperCase()}</span>
        </div>
        <div class="metric">
          <span class="metric-label">Uptime</span>
          <span class="metric-value">${Math.floor(health.uptime / 3600)}h ${Math.floor((health.uptime % 3600) / 60)}m</span>
        </div>
        <div class="metric">
          <span class="metric-label">Version</span>
          <span class="metric-value">${health.version}</span>
        </div>
      </div>

      <!-- Services -->
      ${Object.entries(health.services).map(([name, service]: [string, any]) => `
      <div class="card ${service.status}">
        <div class="card-title">
          <div class="status-indicator ${service.status}"></div>
          ${name.charAt(0).toUpperCase() + name.slice(1)}
        </div>
        <div class="metric">
          <span class="metric-label">Status</span>
          <span class="metric-value">${service.status.toUpperCase()}</span>
        </div>
        ${service.details ? Object.entries(service.details).map(([key, value]: [string, any]) => `
        <div class="metric">
          <span class="metric-label">${key.replace(/([A-Z])/g, ' $1').toLowerCase()}</span>
          <span class="metric-value">${typeof value === 'boolean' ? (value ? 'Yes' : 'No') : value}</span>
        </div>
        `).join('') : ''}
      </div>
      `).join('')}

      <!-- SLI Metrics -->
      <div class="card">
        <div class="card-title">
          📊 Service Level Indicators
        </div>
        <div class="metric">
          <span class="metric-label">Uptime</span>
          <span class="metric-value">${health.sli.uptime.toFixed(2)}%</span>
        </div>
        <div class="metric">
          <span class="metric-label">Error Rate</span>
          <span class="metric-value">${health.sli.errorRate.toFixed(3)}%</span>
        </div>
        <div class="metric">
          <span class="metric-label">Response Time P95</span>
          <span class="metric-value">${health.sli.responseTime95.toFixed(0)}ms</span>
        </div>
        <div class="metric">
          <span class="metric-label">Response Time P99</span>
          <span class="metric-value">${health.sli.responseTime99.toFixed(0)}ms</span>
        </div>
        <div class="metric">
          <span class="metric-label">Delivery Success Rate</span>
          <span class="metric-value">${health.sli.deliverySuccessRate.toFixed(1)}%</span>
        </div>
      </div>
    </div>

    <div class="timestamp">
      Last updated: ${new Date(health.timestamp).toLocaleString()}
      <br>
      <a href="?format=json" style="color: #d97706; text-decoration: none;">View JSON</a> |
      <a href="?detailed=true" style="color: #d97706; text-decoration: none;">Detailed View</a> |
      <a href="/api/docs" style="color: #d97706; text-decoration: none;">API Docs</a>
    </div>
  </div>

  <script>
    // Auto-refresh every 30 seconds
    setTimeout(() => window.location.reload(), 30000);
  </script>
</body>
</html>
  `
}