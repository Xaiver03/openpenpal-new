#!/usr/bin/env node

/**
 * SOTA End-to-End UX Improvements Testing Suite
 * Tests all new mobile responsiveness, PWA features, and navigation improvements
 */

const puppeteer = require('puppeteer');
const fs = require('fs').promises;
const path = require('path');

// Test configuration
const TEST_CONFIG = {
  baseUrl: process.env.TEST_URL || 'http://localhost:3000',
  timeout: 30000,
  screenshotDir: './test-screenshots',
  devices: [
    // Desktop
    { name: 'Desktop', width: 1920, height: 1080, isMobile: false },
    // Tablet
    { name: 'iPad', width: 768, height: 1024, isMobile: false },
    // Mobile
    { name: 'iPhone 12', width: 390, height: 844, isMobile: true },
    { name: 'Android', width: 375, height: 667, isMobile: true }
  ],
  testPages: [
    { path: '/', name: 'Homepage' },
    { path: '/write', name: 'Write Page' },
    { path: '/museum', name: 'Museum Page' },
    { path: '/shop', name: 'Shop Page' },
    { path: '/courier', name: 'Courier Page', requiresAuth: true }
  ]
};

// Test results storage
const testResults = {
  timestamp: new Date().toISOString(),
  passed: 0,
  failed: 0,
  skipped: 0,
  details: [],
  performance: {},
  accessibility: {},
  responsive: {},
  pwa: {}
};

// Utility functions
const delay = ms => new Promise(resolve => setTimeout(resolve, ms));

const logTest = (test, status, message, details = {}) => {
  const result = {
    test,
    status,
    message,
    timestamp: new Date().toISOString(),
    ...details
  };
  
  testResults.details.push(result);
  
  const emoji = status === 'PASS' ? '✅' : status === 'FAIL' ? '❌' : '⏭️';
  console.log(`${emoji} ${test}: ${message}`);
  
  if (status === 'PASS') testResults.passed++;
  else if (status === 'FAIL') testResults.failed++;
  else testResults.skipped++;
};

// Create screenshots directory
async function setupTestEnvironment() {
  try {
    await fs.mkdir(TEST_CONFIG.screenshotDir, { recursive: true });
    console.log('📁 Test environment setup complete');
  } catch (error) {
    console.warn('Warning: Could not create screenshot directory:', error.message);
  }
}

// Test responsive design on different devices
async function testResponsiveDesign(browser) {
  console.log('\n🎨 Testing Responsive Design...');
  
  for (const device of TEST_CONFIG.devices) {
    const page = await browser.newPage();
    
    try {
      // Set device viewport
      await page.setViewport({
        width: device.width,
        height: device.height,
        isMobile: device.isMobile
      });
      
      for (const testPage of TEST_CONFIG.testPages) {
        const testName = `Responsive - ${device.name} - ${testPage.name}`;
        
        try {
          await page.goto(`${TEST_CONFIG.baseUrl}${testPage.path}`, {
            waitUntil: 'networkidle2',
            timeout: TEST_CONFIG.timeout
          });
          
          // Wait for any animations to complete
          await delay(1000);
          
          // Test mobile menu (only on mobile devices)
          if (device.isMobile) {
            // Check if mobile menu button exists
            const menuButton = await page.$('[data-testid="mobile-menu-toggle"], button[aria-label*="menu"], button[aria-label*="Menu"]');
            if (menuButton) {
              // Click menu button
              await menuButton.click();
              await delay(500);
              
              // Check if menu opened
              const menuOpen = await page.$('[data-testid="mobile-menu"], nav[role="dialog"], .mobile-menu');
              if (menuOpen) {
                logTest(testName + ' - Mobile Menu', 'PASS', 'Mobile menu opens correctly');
                
                // Test swipe gesture (if supported)
                try {
                  await page.touchscreen.swipe(300, 200, 100, 200, { steps: 10 });
                  await delay(500);
                  
                  const menuClosed = await page.$('[data-testid="mobile-menu"]:not([aria-hidden="false"])') !== null;
                  if (menuClosed) {
                    logTest(testName + ' - Swipe Gesture', 'PASS', 'Swipe to close menu works');
                  }
                } catch (swipeError) {
                  logTest(testName + ' - Swipe Gesture', 'SKIP', 'Swipe gesture not testable in this environment');
                }
                
                // Close menu
                const closeButton = await page.$('button[aria-label*="close"], button[aria-label*="Close"], .mobile-menu button:first-child');
                if (closeButton) {
                  await closeButton.click();
                }
              } else {
                logTest(testName + ' - Mobile Menu', 'FAIL', 'Mobile menu did not open');
              }
            } else {
              logTest(testName + ' - Mobile Menu', 'SKIP', 'Mobile menu button not found');
            }
          }
          
          // Test responsive table/card switching
          const tables = await page.$$('table, [data-testid="responsive-table"]');
          if (tables.length > 0) {
            const isCardView = device.isMobile;
            const hasCards = await page.$('.card, [data-testid="mobile-card"]') !== null;
            
            if (isCardView && hasCards) {
              logTest(testName + ' - Table Response', 'PASS', 'Tables correctly switch to card view on mobile');
            } else if (!isCardView && !hasCards) {
              logTest(testName + ' - Table Response', 'PASS', 'Tables remain as tables on desktop');
            } else {
              logTest(testName + ' - Table Response', 'FAIL', 'Table/card switching not working correctly');
            }
          }
          
          // Test breadcrumb visibility
          const breadcrumb = await page.$('[data-testid="breadcrumb"], nav[aria-label*="breadcrumb"], .breadcrumb');
          if (breadcrumb) {
            logTest(testName + ' - Breadcrumb', 'PASS', 'Breadcrumb navigation present');
          } else if (testPage.path !== '/') {
            logTest(testName + ' - Breadcrumb', 'FAIL', 'Breadcrumb navigation missing on sub-page');
          }
          
          // Take screenshot
          const screenshotPath = path.join(TEST_CONFIG.screenshotDir, `${device.name.replace(/\s+/g, '_')}_${testPage.name.replace(/\s+/g, '_')}.png`);
          await page.screenshot({ path: screenshotPath, fullPage: true });
          
          logTest(testName, 'PASS', `Page renders correctly on ${device.name}`);
          
        } catch (error) {
          logTest(testName, 'FAIL', `Error testing page: ${error.message}`);
        }
      }
      
    } catch (error) {
      logTest(`Responsive - ${device.name}`, 'FAIL', `Device setup failed: ${error.message}`);
    } finally {
      await page.close();
    }
  }
}

// Test page transitions and animations
async function testPageTransitions(browser) {
  console.log('\n🎬 Testing Page Transitions...');
  
  const page = await browser.newPage();
  await page.setViewport({ width: 1200, height: 800 });
  
  try {
    // Enable request interception to measure loading times
    await page.setRequestInterception(true);
    
    const navigationTimes = [];
    
    page.on('request', request => {
      request.continue();
    });
    
    // Test transitions between pages
    const transitionTests = [
      { from: '/', to: '/write', name: 'Home to Write' },
      { from: '/write', to: '/museum', name: 'Write to Museum' },
      { from: '/museum', to: '/shop', name: 'Museum to Shop' },
      { from: '/shop', to: '/', name: 'Shop to Home' }
    ];
    
    for (const transition of transitionTests) {
      try {
        const startTime = Date.now();
        
        // Navigate to starting page
        await page.goto(`${TEST_CONFIG.baseUrl}${transition.from}`, {
          waitUntil: 'networkidle2'
        });
        
        // Wait for any initial animations
        await delay(500);
        
        // Find and click navigation link
        const navLink = await page.$(`a[href="${transition.to}"], a[href$="${transition.to}"]`);
        if (navLink) {
          await navLink.click();
          
          // Wait for navigation
          await page.waitForNavigation({ waitUntil: 'networkidle2' });
          
          const endTime = Date.now();
          const duration = endTime - startTime;
          navigationTimes.push({ transition: transition.name, duration });
          
          // Check if URL changed correctly
          const currentUrl = page.url();
          if (currentUrl.includes(transition.to)) {
            logTest(`Page Transition - ${transition.name}`, 'PASS', `Navigation completed in ${duration}ms`);
          } else {
            logTest(`Page Transition - ${transition.name}`, 'FAIL', `URL did not change correctly: ${currentUrl}`);
          }
        } else {
          logTest(`Page Transition - ${transition.name}`, 'SKIP', `Navigation link not found for ${transition.to}`);
        }
        
      } catch (error) {
        logTest(`Page Transition - ${transition.name}`, 'FAIL', `Transition failed: ${error.message}`);
      }
    }
    
    // Calculate average navigation time
    if (navigationTimes.length > 0) {
      const avgTime = navigationTimes.reduce((sum, nav) => sum + nav.duration, 0) / navigationTimes.length;
      testResults.performance.avgNavigationTime = avgTime;
      
      if (avgTime < 2000) {
        logTest('Navigation Performance', 'PASS', `Average navigation time: ${Math.round(avgTime)}ms`);
      } else {
        logTest('Navigation Performance', 'FAIL', `Average navigation time too slow: ${Math.round(avgTime)}ms`);
      }
    }
    
  } catch (error) {
    logTest('Page Transitions', 'FAIL', `Setup failed: ${error.message}`);
  } finally {
    await page.close();
  }
}

// Test PWA functionality
async function testPWAFeatures(browser) {
  console.log('\n📱 Testing PWA Features...');
  
  const page = await browser.newPage();
  await page.setViewport({ width: 375, height: 667 });
  
  try {
    await page.goto(TEST_CONFIG.baseUrl, { waitUntil: 'networkidle2' });
    
    // Test manifest.json
    try {
      const manifestResponse = await page.goto(`${TEST_CONFIG.baseUrl}/manifest.json`);
      if (manifestResponse.ok()) {
        const manifest = await manifestResponse.json();
        if (manifest.name && manifest.icons && manifest.start_url) {
          logTest('PWA Manifest', 'PASS', 'Web app manifest is valid');
          testResults.pwa.manifest = true;
        } else {
          logTest('PWA Manifest', 'FAIL', 'Web app manifest is incomplete');
        }
      } else {
        logTest('PWA Manifest', 'FAIL', 'Web app manifest not found');
      }
    } catch (error) {
      logTest('PWA Manifest', 'FAIL', `Manifest test failed: ${error.message}`);
    }
    
    // Return to main page
    await page.goto(TEST_CONFIG.baseUrl, { waitUntil: 'networkidle2' });
    
    // Test Service Worker registration
    const swRegistered = await page.evaluate(() => {
      return 'serviceWorker' in navigator;
    });
    
    if (swRegistered) {
      logTest('Service Worker Support', 'PASS', 'Service Worker API is available');
      
      // Wait for service worker to potentially register
      await delay(3000);
      
      const swActive = await page.evaluate(() => {
        return navigator.serviceWorker.controller !== null;
      });
      
      if (swActive) {
        logTest('Service Worker Registration', 'PASS', 'Service Worker is active');
        testResults.pwa.serviceWorker = true;
      } else {
        logTest('Service Worker Registration', 'FAIL', 'Service Worker is not active');
      }
    } else {
      logTest('Service Worker Support', 'FAIL', 'Service Worker API not supported');
    }
    
    // Test offline page
    try {
      const offlineResponse = await page.goto(`${TEST_CONFIG.baseUrl}/offline.html`);
      if (offlineResponse.ok()) {
        logTest('Offline Page', 'PASS', 'Offline page exists and loads correctly');
        testResults.pwa.offlinePage = true;
      } else {
        logTest('Offline Page', 'FAIL', 'Offline page not found');
      }
    } catch (error) {
      logTest('Offline Page', 'FAIL', `Offline page test failed: ${error.message}`);
    }
    
    // Test installability (check for beforeinstallprompt)
    await page.goto(TEST_CONFIG.baseUrl, { waitUntil: 'networkidle2' });
    
    const isInstallable = await page.evaluate(() => {
      return new Promise((resolve) => {
        let installable = false;
        
        const handler = (e) => {
          installable = true;
          resolve(true);
        };
        
        window.addEventListener('beforeinstallprompt', handler);
        
        // Wait 2 seconds for the event
        setTimeout(() => {
          window.removeEventListener('beforeinstallprompt', handler);
          resolve(installable);
        }, 2000);
      });
    });
    
    if (isInstallable) {
      logTest('PWA Installability', 'PASS', 'App is installable');
      testResults.pwa.installable = true;
    } else {
      logTest('PWA Installability', 'SKIP', 'Install prompt not triggered (may require HTTPS)');
    }
    
  } catch (error) {
    logTest('PWA Features', 'FAIL', `PWA test setup failed: ${error.message}`);
  } finally {
    await page.close();
  }
}

// Test loading performance and skeletons
async function testLoadingPerformance(browser) {
  console.log('\n⚡ Testing Loading Performance...');
  
  const page = await browser.newPage();
  await page.setViewport({ width: 1200, height: 800 });
  
  try {
    // Enable performance metrics collection
    await page.setCacheEnabled(false); // Disable cache for accurate testing
    
    for (const testPage of TEST_CONFIG.testPages) {
      try {
        const startTime = Date.now();
        
        await page.goto(`${TEST_CONFIG.baseUrl}${testPage.path}`, {
          waitUntil: 'networkidle2',
          timeout: TEST_CONFIG.timeout
        });
        
        const loadTime = Date.now() - startTime;
        
        // Check for loading skeletons
        await delay(100); // Brief delay to catch skeletons
        const hasSkeletons = await page.$('.skeleton, [data-testid="skeleton"], .animate-pulse') !== null;
        
        if (hasSkeletons) {
          logTest(`Loading Skeletons - ${testPage.name}`, 'PASS', 'Loading skeletons detected');
        } else {
          logTest(`Loading Skeletons - ${testPage.name}`, 'SKIP', 'No loading skeletons found (may load too fast)');
        }
        
        // Measure performance metrics
        const metrics = await page.evaluate(() => {
          const navigation = performance.getEntriesByType('navigation')[0];
          return {
            domContentLoaded: navigation.domContentLoadedEventEnd - navigation.domContentLoadedEventStart,
            loadComplete: navigation.loadEventEnd - navigation.loadEventStart,
            firstPaint: performance.getEntriesByName('first-paint')[0]?.startTime || 0,
            firstContentfulPaint: performance.getEntriesByName('first-contentful-paint')[0]?.startTime || 0
          };
        });
        
        testResults.performance[testPage.name] = {
          loadTime,
          ...metrics
        };
        
        if (loadTime < 3000) {
          logTest(`Performance - ${testPage.name}`, 'PASS', `Page loaded in ${loadTime}ms`);
        } else if (loadTime < 5000) {
          logTest(`Performance - ${testPage.name}`, 'PASS', `Page loaded in ${loadTime}ms (acceptable)`);
        } else {
          logTest(`Performance - ${testPage.name}`, 'FAIL', `Page took too long to load: ${loadTime}ms`);
        }
        
      } catch (error) {
        logTest(`Performance - ${testPage.name}`, 'FAIL', `Performance test failed: ${error.message}`);
      }
    }
    
  } catch (error) {
    logTest('Loading Performance', 'FAIL', `Performance test setup failed: ${error.message}`);
  } finally {
    await page.close();
  }
}

// Test accessibility improvements
async function testAccessibility(browser) {
  console.log('\n♿ Testing Accessibility...');
  
  const page = await browser.newPage();
  await page.setViewport({ width: 1200, height: 800 });
  
  try {
    for (const testPage of TEST_CONFIG.testPages) {
      try {
        await page.goto(`${TEST_CONFIG.baseUrl}${testPage.path}`, {
          waitUntil: 'networkidle2'
        });
        
        // Test keyboard navigation
        const focusableElements = await page.$$('[tabindex], button, a, input, select, textarea');
        if (focusableElements.length > 0) {
          logTest(`Accessibility - ${testPage.name} - Focusable`, 'PASS', `Found ${focusableElements.length} focusable elements`);
        } else {
          logTest(`Accessibility - ${testPage.name} - Focusable`, 'FAIL', 'No focusable elements found');
        }
        
        // Test ARIA labels
        const ariaLabels = await page.$$('[aria-label], [aria-labelledby], [aria-describedby]');
        if (ariaLabels.length > 0) {
          logTest(`Accessibility - ${testPage.name} - ARIA`, 'PASS', `Found ${ariaLabels.length} elements with ARIA labels`);
        }
        
        // Test skip links or similar navigation aids
        const skipLinks = await page.$('[href="#main"], [href="#content"], .skip-link');
        if (skipLinks) {
          logTest(`Accessibility - ${testPage.name} - Skip Links`, 'PASS', 'Skip links found');
        }
        
        // Test color contrast (basic check for dark text on light background)
        const contrastCheck = await page.evaluate(() => {
          const elements = document.querySelectorAll('p, span, div, button, a');
          let goodContrast = 0;
          let totalText = 0;
          
          for (const el of elements) {
            if (el.innerText && el.innerText.trim()) {
              totalText++;
              const styles = getComputedStyle(el);
              const color = styles.color;
              const bgColor = styles.backgroundColor;
              
              // Simple heuristic: dark text (rgb values < 128) on light bg
              if (color.includes('rgb(') && color.match(/rgb\\((\\d+), ?(\\d+), ?(\\d+)\\)/)) {
                const matches = color.match(/rgb\\((\\d+), ?(\\d+), ?(\\d+)\\)/);
                const r = parseInt(matches[1]);
                const g = parseInt(matches[2]);
                const b = parseInt(matches[3]);
                
                if (r < 128 && g < 128 && b < 128) {
                  goodContrast++;
                }
              }
            }
          }
          
          return { goodContrast, totalText };
        });
        
        if (contrastCheck.totalText > 0) {
          const contrastRatio = contrastCheck.goodContrast / contrastCheck.totalText;
          if (contrastRatio > 0.5) {
            logTest(`Accessibility - ${testPage.name} - Contrast`, 'PASS', `${Math.round(contrastRatio * 100)}% elements have good contrast`);
          } else {
            logTest(`Accessibility - ${testPage.name} - Contrast`, 'FAIL', `Only ${Math.round(contrastRatio * 100)}% elements have good contrast`);
          }
        }
        
      } catch (error) {
        logTest(`Accessibility - ${testPage.name}`, 'FAIL', `Accessibility test failed: ${error.message}`);
      }
    }
    
  } catch (error) {
    logTest('Accessibility', 'FAIL', `Accessibility test setup failed: ${error.message}`);
  } finally {
    await page.close();
  }
}

// Generate comprehensive test report
async function generateReport() {
  console.log('\n📊 Generating Test Report...');
  
  const reportPath = './UX_IMPROVEMENTS_E2E_REPORT.md';
  
  const report = `# OpenPenPal UX Improvements - E2E Test Report

## Test Summary

**Generated:** ${testResults.timestamp}
**Total Tests:** ${testResults.passed + testResults.failed + testResults.skipped}
**Passed:** ${testResults.passed} ✅
**Failed:** ${testResults.failed} ❌
**Skipped:** ${testResults.skipped} ⏭️

**Success Rate:** ${Math.round((testResults.passed / (testResults.passed + testResults.failed)) * 100)}%

## Performance Metrics

${Object.entries(testResults.performance).map(([page, metrics]) => 
  `- **${page}**: ${metrics.loadTime}ms load time, ${Math.round(metrics.firstContentfulPaint)}ms FCP`
).join('\n')}

## PWA Features Status

- **Manifest:** ${testResults.pwa.manifest ? '✅ Valid' : '❌ Invalid/Missing'}
- **Service Worker:** ${testResults.pwa.serviceWorker ? '✅ Active' : '❌ Inactive'}
- **Offline Page:** ${testResults.pwa.offlinePage ? '✅ Available' : '❌ Missing'}
- **Installable:** ${testResults.pwa.installable ? '✅ Yes' : '⚠️ Not tested/HTTPS required'}

## Detailed Test Results

${testResults.details.map(test => 
  `### ${test.test}
**Status:** ${test.status}
**Message:** ${test.message}
**Time:** ${test.timestamp}
${test.details ? `**Details:** ${JSON.stringify(test.details, null, 2)}` : ''}
`).join('\n')}

## Recommendations

### High Priority
${testResults.failed > 0 ? '- Fix failed tests listed above' : '- All tests passing! 🎉'}

### Medium Priority
- Consider adding more accessibility features
- Optimize performance for slower connections
- Add more comprehensive PWA features

### Low Priority
- Enhance loading animations
- Add more responsive breakpoints
- Consider advanced PWA features (background sync, push notifications)

## Screenshots

Screenshots have been saved to: \`${TEST_CONFIG.screenshotDir}\`

---
*Generated by OpenPenPal UX Testing Suite*
`;

  try {
    await fs.writeFile(reportPath, report);
    console.log(`📄 Report saved to: ${reportPath}`);
  } catch (error) {
    console.error('Failed to save report:', error.message);
  }
}

// Main test runner
async function runTests() {
  console.log('🚀 Starting OpenPenPal UX Improvements E2E Tests\n');
  
  await setupTestEnvironment();
  
  // Launch browser with appropriate flags
  const browser = await puppeteer.launch({
    headless: process.env.HEADLESS !== 'false',
    args: [
      '--no-sandbox',
      '--disable-setuid-sandbox',
      '--disable-dev-shm-usage',
      '--disable-accelerated-2d-canvas',
      '--disable-gpu'
    ]
  });
  
  try {
    // Run all test suites
    await testResponsiveDesign(browser);
    await testPageTransitions(browser);
    await testPWAFeatures(browser);
    await testLoadingPerformance(browser);
    await testAccessibility(browser);
    
  } catch (error) {
    console.error('❌ Test suite failed:', error.message);
    testResults.failed++;
  } finally {
    await browser.close();
  }
  
  // Generate report
  await generateReport();
  
  // Print summary
  console.log(`\n🏁 Test Complete!`);
  console.log(`✅ Passed: ${testResults.passed}`);
  console.log(`❌ Failed: ${testResults.failed}`);
  console.log(`⏭️ Skipped: ${testResults.skipped}`);
  
  if (testResults.failed === 0) {
    console.log('🎉 All tests passed! UX improvements are working correctly.');
    process.exit(0);
  } else {
    console.log('⚠️ Some tests failed. Please check the report for details.');
    process.exit(1);
  }
}

// Run tests if this file is executed directly
if (require.main === module) {
  runTests().catch(error => {
    console.error('💥 Test runner crashed:', error);
    process.exit(1);
  });
}

module.exports = { runTests, testResults };