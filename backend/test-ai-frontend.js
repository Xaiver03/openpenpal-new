const puppeteer = require('puppeteer');

async function testAIFrontend() {
  console.log('🧪 Testing AI Frontend Components...\n');
  
  const browser = await puppeteer.launch({ 
    headless: true,
    args: ['--no-sandbox', '--disable-setuid-sandbox'] 
  });
  
  try {
    const page = await browser.newPage();
    
    // Enable detailed logging
    page.on('console', msg => console.log(`[Browser] ${msg.text()}`));
    page.on('response', response => {
      if (response.url().includes('/ai/')) {
        console.log(`[API Response] ${response.url()} - Status: ${response.status()}`);
      }
    });
    
    // Step 1: Go to AI page
    console.log('1️⃣ Navigating to AI page...');
    await page.goto('http://localhost:3000/ai', { waitUntil: 'networkidle2' });
    await page.waitForTimeout(2000);
    
    // Step 2: Test Writing Inspiration tab
    console.log('\n2️⃣ Testing Writing Inspiration (云锦传驿)...');
    
    // Click the Writing Inspiration tab
    const tabs = await page.$$('[role="tab"]');
    if (tabs.length >= 2) {
      await tabs[1].click();
      await page.waitForTimeout(1000);
    }
    
    // Find and click the inspiration button
    const inspirationButton = await page.$('button:has-text("获取灵感")') || 
                             await page.$('button:has-text("生成写作灵感")') ||
                             await page.$('button:has-text("获取写作灵感")');
    
    if (inspirationButton) {
      console.log('   ✓ Found inspiration button');
      
      // Intercept the API call
      const responsePromise = page.waitForResponse(
        response => response.url().includes('/api/v1/ai/inspiration'),
        { timeout: 10000 }
      );
      
      await inspirationButton.click();
      console.log('   ✓ Clicked inspiration button');
      
      // Wait for API response
      const response = await responsePromise;
      const responseData = await response.json();
      
      console.log(`   ✓ API Response Status: ${response.status()}`);
      console.log(`   ✓ API Response:`, JSON.stringify(responseData, null, 2));
      
      // Wait for UI update
      await page.waitForTimeout(2000);
      
      // Check if inspiration is displayed
      const inspirationContent = await page.$$eval('*', elements => {
        return elements
          .filter(el => el.textContent && (
            el.textContent.includes('写一写') || 
            el.textContent.includes('日常') || 
            el.textContent.includes('感悟')
          ))
          .map(el => el.textContent)
          .slice(0, 3);
      });
      
      if (inspirationContent.length > 0) {
        console.log('   ✓ Inspiration displayed in UI:');
        inspirationContent.forEach(content => {
          console.log(`     - ${content.substring(0, 50)}...`);
        });
      }
      
      // Take screenshot
      await page.screenshot({ path: 'ai-inspiration-test.png' });
      console.log('   ✓ Screenshot saved: ai-inspiration-test.png');
      
    } else {
      console.log('   ❌ Could not find inspiration button');
    }
    
    console.log('\n✅ AI Frontend Test Complete!');
    
  } catch (error) {
    console.error('❌ Test failed:', error);
  } finally {
    await browser.close();
  }
}

// Run the test
testAIFrontend().catch(console.error);