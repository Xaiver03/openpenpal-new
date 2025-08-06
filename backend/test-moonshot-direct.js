/**
 * Direct Moonshot API Test
 * Tests the API key directly without going through the backend
 */

const https = require('https');

async function testMoonshotAPI() {
  console.log('🌙 Testing Moonshot API directly...\n');
  
  // First, get the API key from database
  const { execSync } = require('child_process');
  const result = execSync(`psql -U $USER -d openpenpal -t -c "SELECT api_key FROM ai_configs WHERE provider='moonshot' LIMIT 1"`).toString().trim();
  
  if (!result) {
    console.error('❌ No Moonshot API key found in database');
    return;
  }
  
  const apiKey = result;
  console.log('✅ Found API key in database\n');
  
  // Test the API directly
  const data = JSON.stringify({
    model: 'moonshot-v1-8k',
    messages: [
      {
        role: 'system',
        content: '你是一个友好的助手。请用简短的一句话回答。'
      },
      {
        role: 'user',
        content: '你好，请用一句话介绍你自己。'
      }
    ],
    temperature: 0.7,
    max_tokens: 100,
    stream: false
  });
  
  const options = {
    hostname: 'api.moonshot.cn',
    port: 443,
    path: '/v1/chat/completions',
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${apiKey}`,
      'Content-Length': Buffer.byteLength(data)
    }
  };
  
  return new Promise((resolve, reject) => {
    const req = https.request(options, (res) => {
      let responseData = '';
      
      res.on('data', (chunk) => {
        responseData += chunk;
      });
      
      res.on('end', () => {
        console.log(`📡 Response Status: ${res.statusCode}`);
        console.log(`📡 Response Headers:`, res.headers);
        console.log('\n📄 Response Body:');
        
        try {
          const parsed = JSON.parse(responseData);
          console.log(JSON.stringify(parsed, null, 2));
          
          if (res.statusCode === 200 && parsed.choices && parsed.choices[0]) {
            console.log('\n✅ Moonshot API is working!');
            console.log(`🤖 AI Response: "${parsed.choices[0].message.content}"`);
          } else if (res.statusCode === 401) {
            console.log('\n❌ API Key is invalid or expired');
          } else if (res.statusCode === 429) {
            console.log('\n❌ Rate limit exceeded or quota exhausted');
          } else {
            console.log('\n❌ API call failed');
          }
        } catch (e) {
          console.log(responseData);
          console.log('\n❌ Failed to parse response');
        }
        
        resolve();
      });
    });
    
    req.on('error', (error) => {
      console.error('❌ Request error:', error);
      reject(error);
    });
    
    req.write(data);
    req.end();
  });
}

// Run the test
testMoonshotAPI().catch(console.error);