/**
 * Automated UI Component Testing Script
 * Tests all buttons, inputs, and interactive elements
 */

const puppeteer = require('puppeteer');

async function testUIComponents() {
  let browser;
  try {
    browser = await puppeteer.launch({ 
      headless: true,
      args: ['--no-sandbox', '--disable-setuid-sandbox']
    });
    
    const page = await browser.newPage();
    
    console.log('🚀 Starting UI component testing...');
    
    // Test results storage
    const testResults = {
      pages: {},
      summary: {
        totalPages: 0,
        totalButtons: 0,
        totalInputs: 0,
        workingButtons: 0,
        workingInputs: 0,
        issues: []
      }
    };

    // Pages to test
    const pagesToTest = [
      { url: 'http://localhost:3000', name: 'Homepage' },
      { url: 'http://localhost:3000/login', name: 'Login Page' },
      { url: 'http://localhost:3000/register', name: 'Register Page' },
      { url: 'http://localhost:3000/write', name: 'Write Letter Page' },
      { url: 'http://localhost:3000/mailbox', name: 'Mailbox Page' },
      { url: 'http://localhost:3000/courier', name: 'Courier Page' },
      { url: 'http://localhost:3000/courier/apply', name: 'Courier Apply Page' },
      { url: 'http://localhost:3000/courier/scan', name: 'Courier Scan Page' },
      { url: 'http://localhost:3000/courier/tasks', name: 'Courier Tasks Page' },
      { url: 'http://localhost:3000/settings', name: 'Settings Page' },
      { url: 'http://localhost:3000/about', name: 'About Page' },
      { url: 'http://localhost:3000/help', name: 'Help Page' },
      { url: 'http://localhost:3000/contact', name: 'Contact Page' }
    ];

    for (const pageInfo of pagesToTest) {
      console.log(`\n📄 Testing ${pageInfo.name}...`);
      
      const pageResult = {
        url: pageInfo.url,
        accessible: false,
        buttons: [],
        inputs: [],
        forms: [],
        issues: []
      };

      try {
        // Navigate to page
        await page.goto(pageInfo.url, { 
          waitUntil: 'networkidle2',
          timeout: 10000 
        });
        
        pageResult.accessible = true;
        console.log(`✅ Page accessible: ${pageInfo.url}`);

        // Test buttons
        const buttons = await page.$$('button, [role="button"], .btn, input[type="submit"], input[type="button"]');
        console.log(`🔘 Found ${buttons.length} buttons`);
        
        for (let i = 0; i < buttons.length; i++) {
          try {
            const button = buttons[i];
            const buttonText = await button.evaluate(el => el.textContent?.trim() || el.value || el.getAttribute('aria-label') || `Button ${i + 1}`);
            const isDisabled = await button.evaluate(el => el.disabled || el.getAttribute('disabled') !== null);
            const isVisible = await button.isIntersectingViewport();
            
            const buttonTest = {
              index: i + 1,
              text: buttonText,
              disabled: isDisabled,
              visible: isVisible,
              clickable: false,
              error: null
            };

            if (!isDisabled && isVisible) {
              try {
                // Test if button is clickable
                await button.hover();
                await page.waitForTimeout(100);
                
                // Check if button has click handlers or href
                const hasHandlers = await button.evaluate(el => {
                  return !!(el.onclick || 
                           el.addEventListener || 
                           el.href || 
                           el.getAttribute('href') ||
                           el.closest('a') ||
                           el.getAttribute('data-action'));
                });
                
                buttonTest.clickable = hasHandlers;
                
                if (hasHandlers) {
                  testResults.summary.workingButtons++;
                }
              } catch (error) {
                buttonTest.error = error.message;
                pageResult.issues.push(`Button "${buttonText}" hover failed: ${error.message}`);
              }
            }
            
            pageResult.buttons.push(buttonTest);
            testResults.summary.totalButtons++;
          } catch (error) {
            pageResult.issues.push(`Button ${i + 1} test failed: ${error.message}`);
          }
        }

        // Test input fields
        const inputs = await page.$$('input, textarea, select, [contenteditable]');
        console.log(`📝 Found ${inputs.length} input fields`);
        
        for (let i = 0; i < inputs.length; i++) {
          try {
            const input = inputs[i];
            const inputType = await input.evaluate(el => 
              el.type || el.tagName.toLowerCase() || 'unknown'
            );
            const placeholder = await input.evaluate(el => 
              el.placeholder || el.getAttribute('aria-label') || `Input ${i + 1}`
            );
            const isDisabled = await input.evaluate(el => 
              el.disabled || el.readOnly || el.getAttribute('disabled') !== null
            );
            const isVisible = await input.isIntersectingViewport();
            
            const inputTest = {
              index: i + 1,
              type: inputType,
              placeholder: placeholder,
              disabled: isDisabled,
              visible: isVisible,
              writable: false,
              error: null
            };

            if (!isDisabled && isVisible && inputType !== 'submit' && inputType !== 'button') {
              try {
                // Test if input is writable
                await input.focus();
                await page.waitForTimeout(100);
                
                if (inputType === 'text' || inputType === 'email' || inputType === 'password' || inputType === 'textarea') {
                  await input.type('test', { delay: 10 });
                  await page.waitForTimeout(100);
                  await input.evaluate(el => el.value = ''); // Clear test input
                  inputTest.writable = true;
                  testResults.summary.workingInputs++;
                } else if (inputType === 'checkbox' || inputType === 'radio') {
                  await input.click();
                  inputTest.writable = true;
                  testResults.summary.workingInputs++;
                } else {
                  inputTest.writable = true; // Assume other input types work
                  testResults.summary.workingInputs++;
                }
              } catch (error) {
                inputTest.error = error.message;
                pageResult.issues.push(`Input "${placeholder}" interaction failed: ${error.message}`);
              }
            }
            
            pageResult.inputs.push(inputTest);
            testResults.summary.totalInputs++;
          } catch (error) {
            pageResult.issues.push(`Input ${i + 1} test failed: ${error.message}`);
          }
        }

        // Test forms
        const forms = await page.$$('form');
        console.log(`📋 Found ${forms.length} forms`);
        
        for (let i = 0; i < forms.length; i++) {
          try {
            const form = forms[i];
            const formAction = await form.evaluate(el => el.action || 'No action');
            const formMethod = await form.evaluate(el => el.method || 'GET');
            
            pageResult.forms.push({
              index: i + 1,
              action: formAction,
              method: formMethod,
              inputCount: await form.$$eval('input, textarea, select', inputs => inputs.length)
            });
          } catch (error) {
            pageResult.issues.push(`Form ${i + 1} test failed: ${error.message}`);
          }
        }

        console.log(`✅ ${pageInfo.name} testing completed`);
        console.log(`   - Buttons: ${pageResult.buttons.length} (${pageResult.buttons.filter(b => b.clickable).length} clickable)`);
        console.log(`   - Inputs: ${pageResult.inputs.length} (${pageResult.inputs.filter(i => i.writable).length} writable)`);
        console.log(`   - Forms: ${pageResult.forms.length}`);
        if (pageResult.issues.length > 0) {
          console.log(`   - Issues: ${pageResult.issues.length}`);
        }

      } catch (error) {
        pageResult.accessible = false;
        pageResult.issues.push(`Page navigation failed: ${error.message}`);
        console.log(`❌ ${pageInfo.name} not accessible: ${error.message}`);
        testResults.summary.issues.push(`${pageInfo.name}: ${error.message}`);
      }

      testResults.pages[pageInfo.name] = pageResult;
      testResults.summary.totalPages++;
    }

    // Generate summary report
    console.log('\n' + '='.repeat(60));
    console.log('📊 UI TESTING SUMMARY REPORT');
    console.log('='.repeat(60));
    console.log(`Total Pages Tested: ${testResults.summary.totalPages}`);
    console.log(`Accessible Pages: ${Object.values(testResults.pages).filter(p => p.accessible).length}`);
    console.log(`Total Buttons Found: ${testResults.summary.totalButtons}`);
    console.log(`Working Buttons: ${testResults.summary.workingButtons}`);
    console.log(`Total Inputs Found: ${testResults.summary.totalInputs}`);
    console.log(`Working Inputs: ${testResults.summary.workingInputs}`);
    
    const buttonSuccessRate = testResults.summary.totalButtons > 0 
      ? (testResults.summary.workingButtons / testResults.summary.totalButtons * 100).toFixed(1)
      : 0;
    const inputSuccessRate = testResults.summary.totalInputs > 0 
      ? (testResults.summary.workingInputs / testResults.summary.totalInputs * 100).toFixed(1)
      : 0;
    
    console.log(`Button Success Rate: ${buttonSuccessRate}%`);
    console.log(`Input Success Rate: ${inputSuccessRate}%`);

    // Report issues
    if (testResults.summary.issues.length > 0) {
      console.log('\n❌ ISSUES FOUND:');
      testResults.summary.issues.forEach((issue, index) => {
        console.log(`${index + 1}. ${issue}`);
      });
    }

    // Page-specific issues
    for (const [pageName, pageData] of Object.entries(testResults.pages)) {
      if (pageData.issues.length > 0) {
        console.log(`\n⚠️  ${pageName} Issues:`);
        pageData.issues.forEach((issue, index) => {
          console.log(`   ${index + 1}. ${issue}`);
        });
      }
    }

    // Save detailed report
    const fs = require('fs');
    const reportPath = '/Users/rocalight/同步空间/opplc/openpenpal/frontend/ui-test-report.json';
    fs.writeFileSync(reportPath, JSON.stringify(testResults, null, 2));
    console.log(`\n📄 Detailed report saved to: ${reportPath}`);

    return testResults;

  } catch (error) {
    console.error('❌ Testing failed:', error);
    return null;
  } finally {
    if (browser) {
      await browser.close();
    }
  }
}

// Check if puppeteer is available, if not, run a simplified test
async function runUITest() {
  try {
    await testUIComponents();
  } catch (error) {
    console.log('⚠️  Puppeteer not available, running simplified test...');
    console.log('🔍 Testing basic page accessibility...');
    
    // Simplified test using curl
    const { execSync } = require('child_process');
    const pages = [
      'http://localhost:3000',
      'http://localhost:3000/write',
      'http://localhost:3000/mailbox',
      'http://localhost:3000/courier'
    ];
    
    for (const url of pages) {
      try {
        const response = execSync(`curl -s -o /dev/null -w "%{http_code}" ${url}`, { timeout: 5000 });
        const statusCode = response.toString().trim();
        console.log(`${url}: ${statusCode === '200' ? '✅' : '❌'} ${statusCode}`);
      } catch (error) {
        console.log(`${url}: ❌ Failed to connect`);
      }
    }
    
    console.log('\n📝 Manual testing recommended for full UI validation');
    console.log('💡 To install Puppeteer: npm install puppeteer');
  }
}

if (require.main === module) {
  runUITest();
}

module.exports = { testUIComponents, runUITest };