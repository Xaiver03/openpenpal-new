import { NextRequest, NextResponse } from 'next/server'

// Mock data for demo
const mockData = {
  users: [
    { id: '1', username: 'alice', nickname: '爱丽丝', role: 'USER', schoolCode: 'PKU001', isActive: true },
    { id: '2', username: 'courier1', nickname: '信使小王', role: 'COURIER', schoolCode: 'PKU001', isActive: true }
  ],
  letters: [
    { id: '1', title: '给朋友的问候信', status: 'GENERATED', code: 'OP1K2L3M4N5O', userId: '1' }
  ],
  tasks: [
    { id: '1', letterCode: 'OP1K2L3M4N5O', senderName: '爱丽丝', status: 'PENDING', reward: 10 }
  ]
}

// Simple GraphQL-like query parser
function parseQuery(query: string) {
  if (query.includes('systemStats')) {
    return {
      systemStats: {
        totalUsers: mockData.users.length,
        totalLetters: mockData.letters.length,
        totalCouriers: mockData.users.filter(u => u.role === 'COURIER').length,
        totalDeliveries: mockData.letters.filter(l => l.status === 'DELIVERED').length,
        activeUsers: mockData.users.filter(u => u.isActive).length,
        activeCouriers: mockData.users.filter(u => u.role === 'COURIER' && u.isActive).length,
        averageDeliveryTime: 24.5,
        deliverySuccessRate: 98.2
      }
    }
  }
  
  if (query.includes('users')) {
    return { users: mockData.users }
  }
  
  if (query.includes('letters')) {
    return { letters: mockData.letters }
  }
  
  if (query.includes('courierTasks')) {
    return { courierTasks: mockData.tasks }
  }
  
  return { error: 'Unknown query' }
}

export async function GET() {
  // Return GraphQL playground in development
  if (process.env.NODE_ENV === 'development') {
    const playgroundHTML = `
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>OpenPenPal GraphQL API</title>
  <style>
    body { 
      margin: 0; 
      font-family: system-ui, -apple-system, sans-serif;
      background: #f7f8fa;
      padding: 40px;
    }
    .container {
      max-width: 1000px;
      margin: 0 auto;
      background: white;
      border-radius: 12px;
      padding: 40px;
      box-shadow: 0 4px 24px rgba(0,0,0,0.1);
    }
    .header {
      text-align: center;
      margin-bottom: 40px;
    }
    .title {
      font-size: 2.5rem;
      font-weight: 700;
      color: #d97706;
      margin-bottom: 8px;
    }
    .subtitle {
      color: #6b7280;
      font-size: 1.1rem;
    }
    .section {
      margin: 30px 0;
      padding: 20px;
      background: #f9fafb;
      border-radius: 8px;
      border-left: 4px solid #d97706;
    }
    .query-box {
      background: #1f2937;
      color: #e5e7eb;
      border-radius: 8px;
      padding: 20px;
      margin: 15px 0;
      font-family: 'Monaco', 'Menlo', 'Courier New', monospace;
      font-size: 14px;
      overflow-x: auto;
      line-height: 1.5;
    }
    .btn {
      background: #d97706;
      color: white;
      border: none;
      padding: 12px 24px;
      border-radius: 6px;
      cursor: pointer;
      font-size: 14px;
      font-weight: 600;
      margin: 10px 5px;
      transition: background 0.2s;
    }
    .btn:hover { 
      background: #b45309; 
    }
    .result {
      margin-top: 20px;
      padding: 20px;
      background: #f0f9ff;
      border: 1px solid #0ea5e9;
      border-radius: 8px;
    }
    .error {
      background: #fef2f2;
      border-color: #f87171;
      color: #dc2626;
    }
    .grid {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 20px;
      margin: 20px 0;
    }
    @media (max-width: 768px) {
      .grid { grid-template-columns: 1fr; }
      .container { padding: 20px; margin: 20px; }
    }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h1 class="title">🚀 OpenPenPal GraphQL API</h1>
      <p class="subtitle">State-of-the-Art GraphQL API for Modern Letter Delivery</p>
    </div>
    
    <div class="section">
      <h3>📊 System Stats Query</h3>
      <p>Get real-time platform statistics:</p>
      <div class="query-box">query {
  systemStats {
    totalUsers
    totalLetters
    totalCouriers
    averageDeliveryTime
    deliverySuccessRate
  }
}</div>
      <button class="btn" onclick="testQuery('systemStats')">Test System Stats</button>
    </div>

    <div class="grid">
      <div class="section">
        <h3>👥 Users Query</h3>
        <div class="query-box">query {
  users {
    id
    nickname
    role
    schoolCode
  }
}</div>
        <button class="btn" onclick="testQuery('users')">Test Users</button>
      </div>

      <div class="section">
        <h3>✉️ Letters Query</h3>
        <div class="query-box">query {
  letters {
    id
    title
    status
    code
  }
}</div>
        <button class="btn" onclick="testQuery('letters')">Test Letters</button>
      </div>
    </div>

    <div class="section">
      <h3>🚚 Courier Tasks Query</h3>
      <div class="query-box">query {
  courierTasks {
    id
    letterCode
    senderName
    status
    reward
  }
}</div>
      <button class="btn" onclick="testQuery('courierTasks')">Test Tasks</button>
    </div>

    <div id="result"></div>
  </div>

  <script>
    async function testQuery(type) {
      const queries = {
        systemStats: \`query { systemStats { totalUsers totalLetters totalCouriers averageDeliveryTime deliverySuccessRate } }\`,
        users: \`query { users { id nickname role schoolCode } }\`,
        letters: \`query { letters { id title status code } }\`,
        courierTasks: \`query { courierTasks { id letterCode senderName status reward } }\`
      };
      
      const query = queries[type];
      const resultDiv = document.getElementById('result');
      
      try {
        resultDiv.innerHTML = '<div class="result">🔄 Loading...</div>';
        
        const response = await fetch('/api/graphql', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ query })
        });
        
        const result = await response.json();
        
        resultDiv.innerHTML = \`
          <div class="result">
            <h4>✅ Query Result:</h4>
            <pre style="background: #1f2937; color: #e5e7eb; padding: 15px; border-radius: 6px; overflow: auto; margin: 10px 0;">\${JSON.stringify(result, null, 2)}</pre>
          </div>
        \`;
      } catch (error) {
        resultDiv.innerHTML = \`
          <div class="result error">
            <h4>❌ Error:</h4>
            <pre>\${error.message}</pre>
          </div>
        \`;
      }
    }
  </script>
</body>
</html>
    `
    
    return new Response(playgroundHTML, {
      headers: { 'Content-Type': 'text/html' }
    })
  }
  
  return NextResponse.json({ 
    message: 'OpenPenPal GraphQL API is ready',
    version: '1.0.0',
    endpoint: '/api/graphql'
  })
}

export async function POST(request: NextRequest) {
  try {
    const body = await request.json()
    const { query, variables = {} } = body

    if (!query) {
      return NextResponse.json(
        { error: 'Query is required' },
        { status: 400 }
      )
    }

    // Parse and execute the query
    const result = parseQuery(query)
    
    return NextResponse.json({ data: result })
  } catch (error) {
    console.error('GraphQL Error:', error)
    return NextResponse.json(
      { error: 'Internal server error', details: error instanceof Error ? error.message : 'Unknown error' },
      { status: 500 }
    )
  }
}