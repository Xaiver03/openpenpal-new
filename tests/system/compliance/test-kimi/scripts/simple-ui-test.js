/**
 * Simple UI Testing - Basic Accessibility Check
 */

const { execSync } = require('child_process');

console.log('🔍 Testing basic page accessibility and UI components...\n');

const pages = [
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
  { url: 'http://localhost:3000/contact', name: 'Contact Page' },
  { url: 'http://localhost:3000/guide', name: 'Guide Page' },
  { url: 'http://localhost:3000/privacy', name: 'Privacy Page' },
  { url: 'http://localhost:3000/terms', name: 'Terms Page' }
];

const results = {
  accessible: [],
  notAccessible: [],
  totalButtons: 0,
  totalInputs: 0
};

for (const page of pages) {
  try {
    // Test page accessibility
    const statusCode = execSync(`curl -s -o /dev/null -w "%{http_code}" "${page.url}"`, { 
      timeout: 5000,
      encoding: 'utf8'
    }).trim();
    
    if (statusCode === '200') {
      console.log(`✅ ${page.name}: Accessible (${statusCode})`);
      results.accessible.push(page.name);
      
      // Count UI elements on accessible pages
      try {
        const pageContent = execSync(`curl -s "${page.url}"`, { 
          timeout: 5000,
          encoding: 'utf8'
        });
        
        // Count buttons and inputs using regex
        const buttonMatches = pageContent.match(/<button[^>]*>|<input[^>]*type=['"](?:button|submit)['"][^>]*>/gi) || [];
        const inputMatches = pageContent.match(/<input[^>]*>|<textarea[^>]*>/gi) || [];
        
        const buttonCount = buttonMatches.length;
        const inputCount = inputMatches.length;
        
        if (buttonCount > 0 || inputCount > 0) {
          console.log(`   📊 Found ${buttonCount} buttons, ${inputCount} inputs`);
        }
        
        results.totalButtons += buttonCount;
        results.totalInputs += inputCount;
        
      } catch (error) {
        console.log(`   ⚠️  Could not analyze page content`);
      }
      
    } else {
      console.log(`❌ ${page.name}: Not accessible (${statusCode})`);
      results.notAccessible.push(page.name);
    }
    
  } catch (error) {
    console.log(`❌ ${page.name}: Connection failed`);
    results.notAccessible.push(page.name);
  }
}

// Test API endpoints
console.log('\n🔗 Testing API endpoints...');
const apiEndpoints = [
  { url: 'http://localhost:3000/api/health', name: 'Health Check' },
  { url: 'http://localhost:3000/api/docs', name: 'API Documentation' },
  { url: 'http://localhost:3000/api/graphql', name: 'GraphQL Endpoint' }
];

for (const endpoint of apiEndpoints) {
  try {
    const statusCode = execSync(`curl -s -o /dev/null -w "%{http_code}" "${endpoint.url}"`, { 
      timeout: 5000,
      encoding: 'utf8'
    }).trim();
    
    console.log(`${statusCode === '200' ? '✅' : '❌'} ${endpoint.name}: ${statusCode}`);
  } catch (error) {
    console.log(`❌ ${endpoint.name}: Failed`);
  }
}

// Generate summary
console.log('\n' + '='.repeat(50));
console.log('📊 UI TESTING SUMMARY');
console.log('='.repeat(50));
console.log(`Total Pages Tested: ${pages.length}`);
console.log(`Accessible Pages: ${results.accessible.length}`);
console.log(`Not Accessible: ${results.notAccessible.length}`);
console.log(`Total Buttons Found: ${results.totalButtons}`);
console.log(`Total Inputs Found: ${results.totalInputs}`);

const accessibilityRate = (results.accessible.length / pages.length * 100).toFixed(1);
console.log(`Accessibility Rate: ${accessibilityRate}%`);

if (results.notAccessible.length > 0) {
  console.log('\n❌ Pages Not Accessible:');
  results.notAccessible.forEach((page, index) => {
    console.log(`${index + 1}. ${page}`);
  });
}

// Recommendations
console.log('\n💡 Recommendations:');
if (accessibilityRate < 80) {
  console.log('- Fix page routing and accessibility issues');
}
if (results.totalButtons < 10) {
  console.log('- Verify button elements are properly rendered');
}
if (results.totalInputs < 5) {
  console.log('- Check form input elements on key pages');
}

console.log('- For detailed testing, install Puppeteer: npm install puppeteer');
console.log('- Manual testing recommended for interaction verification');

// Save basic report
const fs = require('fs');
const report = {
  timestamp: new Date().toISOString(),
  summary: {
    totalPages: pages.length,
    accessiblePages: results.accessible.length,
    accessibilityRate: accessibilityRate + '%',
    totalButtons: results.totalButtons,
    totalInputs: results.totalInputs
  },
  accessible: results.accessible,
  notAccessible: results.notAccessible
};

fs.writeFileSync('ui-test-basic-report.json', JSON.stringify(report, null, 2));
console.log('\n📄 Basic report saved to: ui-test-basic-report.json');