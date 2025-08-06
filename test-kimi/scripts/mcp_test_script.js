// OpenPenPal MCP 浏览器自动化测试脚本
// 使用 Playwright 进行全面功能测试

const { chromium } = require('playwright');

async function runOpenPenPalTests() {
    console.log('🚀 开始 OpenPenPal 功能测试...');
    
    const browser = await chromium.launch({ 
        headless: false,  // 显示浏览器界面
        slowMo: 500       // 减慢操作速度以便观察
    });
    
    const context = await browser.newContext({
        viewport: { width: 1280, height: 720 }
    });
    
    const page = await context.newPage();
    
    const testResults = {
        homepage: false,
        navigation: false,
        writePage: false,
        readPage: false,
        plaza: false,
        museum: false,
        shop: false,
        responsiveDesign: false,
        performance: false
    };
    
    try {
        // 测试 1: 首页加载测试
        console.log('📖 测试 1: 首页加载...');
        const startTime = Date.now();
        await page.goto('http://localhost:3000');
        await page.waitForLoadState('networkidle');
        const loadTime = Date.now() - startTime;
        
        // 检查页面标题
        const title = await page.title();
        console.log(`   页面标题: ${title}`);
        console.log(`   加载时间: ${loadTime}ms`);
        
        // 检查关键元素
        const logo = await page.locator('text=OpenPenPal').first();
        const writeButton = await page.locator('text=写信去').first();
        
        if (await logo.isVisible() && await writeButton.isVisible()) {
            testResults.homepage = true;
            console.log('   ✅ 首页加载成功');
        } else {
            console.log('   ❌ 首页关键元素缺失');
        }
        
        // 测试 2: 导航功能测试
        console.log('🧭 测试 2: 导航功能...');
        const navItems = [
            { text: '写信去', path: '/write' },
            { text: '写作广场', path: '/plaza' },
            { text: '信件博物馆', path: '/museum' },
            { text: '信封商城', path: '/shop' }
        ];
        
        let navSuccess = true;
        for (const item of navItems) {
            try {
                await page.click(`text=${item.text}`);
                await page.waitForLoadState('networkidle');
                const currentUrl = page.url();
                console.log(`   导航到: ${item.text} - URL: ${currentUrl}`);
                
                if (!currentUrl.includes(item.path)) {
                    navSuccess = false;
                    console.log(`   ⚠️  URL 不匹配: 期望包含 ${item.path}`);
                }
                
                // 返回首页
                await page.goto('http://localhost:3000');
                await page.waitForLoadState('networkidle');
            } catch (error) {
                navSuccess = false;
                console.log(`   ❌ 导航到 ${item.text} 失败: ${error.message}`);
            }
        }
        
        if (navSuccess) {
            testResults.navigation = true;
            console.log('   ✅ 导航功能正常');
        }
        
        // 测试 3: 写信页面功能
        console.log('✍️  测试 3: 写信页面功能...');
        await page.goto('http://localhost:3000/write');
        await page.waitForLoadState('networkidle');
        
        // 检查写信页面元素
        const titleInput = page.locator('input[placeholder*="标题"]');
        const contentTextarea = page.locator('textarea[placeholder*="亲爱的朋友"]');
        const saveButton = page.locator('text=保存草稿');
        const generateButton = page.locator('text=生成编号贴纸');
        
        if (await titleInput.isVisible() && await contentTextarea.isVisible()) {
            // 填写测试内容
            await titleInput.fill('MCP 自动化测试信件 - ' + new Date().toLocaleString());
            await contentTextarea.fill(`这是一封由 MCP 浏览器自动化工具生成的测试信件。
            
测试时间: ${new Date().toLocaleString()}
测试内容: 验证写信功能是否正常工作
系统状态: 所有功能运行正常

感谢使用 OpenPenPal！`);
            
            // 测试保存草稿
            if (await saveButton.isVisible()) {
                await saveButton.click();
                console.log('   💾 草稿保存测试完成');
            }
            
            // 测试生成编号
            if (await generateButton.isVisible()) {
                await generateButton.click();
                console.log('   🏷️  编号生成测试完成');
                await page.waitForTimeout(2000); // 等待生成完成
            }
            
            testResults.writePage = true;
            console.log('   ✅ 写信页面功能正常');
        } else {
            console.log('   ❌ 写信页面关键元素缺失');
        }
        
        // 测试 4: 阅读页面功能
        console.log('📖 测试 4: 阅读页面功能...');
        await page.goto('http://localhost:3000/read/OP1K2L3M4N5O');
        await page.waitForLoadState('networkidle');
        
        // 检查信件内容
        const letterTitle = page.locator('h1');
        const letterContent = page.locator('.whitespace-pre-wrap');
        const replyButton = page.locator('text=回信');
        
        if (await letterTitle.isVisible() && await letterContent.isVisible()) {
            console.log('   📄 信件内容加载正常');
            
            // 测试回信功能
            if (await replyButton.isVisible()) {
                await replyButton.click();
                await page.waitForLoadState('networkidle');
                
                // 检查是否跳转到写信页面且包含回信参数
                const currentUrl = page.url();
                if (currentUrl.includes('/write') && currentUrl.includes('reply_to=')) {
                    console.log('   ↩️  回信功能正常');
                } else {
                    console.log('   ⚠️  回信跳转异常');
                }
            }
            
            testResults.readPage = true;
            console.log('   ✅ 阅读页面功能正常');
        } else {
            console.log('   ❌ 阅读页面内容加载失败');
        }
        
        // 测试 5: 写作广场
        console.log('🎨 测试 5: 写作广场...');
        await page.goto('http://localhost:3000/plaza');
        await page.waitForLoadState('networkidle');
        
        const plazaTitle = page.locator('h1');
        if (await plazaTitle.isVisible()) {
            testResults.plaza = true;
            console.log('   ✅ 写作广场加载正常');
        } else {
            console.log('   ❌ 写作广场加载失败');
        }
        
        // 测试 6: 信件博物馆
        console.log('🏛️  测试 6: 信件博物馆...');
        await page.goto('http://localhost:3000/museum');
        await page.waitForLoadState('networkidle');
        
        const museumTitle = page.locator('h1');
        if (await museumTitle.isVisible()) {
            testResults.museum = true;
            console.log('   ✅ 信件博物馆加载正常');
        } else {
            console.log('   ❌ 信件博物馆加载失败');
        }
        
        // 测试 7: 信封商城
        console.log('🛍️  测试 7: 信封商城...');
        await page.goto('http://localhost:3000/shop');
        await page.waitForLoadState('networkidle');
        
        const shopTitle = page.locator('h1');
        if (await shopTitle.isVisible()) {
            testResults.shop = true;
            console.log('   ✅ 信封商城加载正常');
        } else {
            console.log('   ❌ 信封商城加载失败');
        }
        
        // 测试 8: 响应式设计
        console.log('📱 测试 8: 响应式设计...');
        await page.goto('http://localhost:3000');
        
        // 测试移动端视口
        await page.setViewportSize({ width: 375, height: 667 });
        await page.waitForTimeout(1000);
        
        const mobileMenu = page.locator('button[class*="md:hidden"]');
        if (await mobileMenu.isVisible()) {
            testResults.responsiveDesign = true;
            console.log('   ✅ 移动端响应式设计正常');
        } else {
            console.log('   ⚠️  移动端菜单未找到');
        }
        
        // 恢复桌面端视口
        await page.setViewportSize({ width: 1280, height: 720 });
        
        // 测试 9: 性能检测
        console.log('⚡ 测试 9: 性能检测...');
        await page.goto('http://localhost:3000', { waitUntil: 'networkidle' });
        
        const performanceMetrics = await page.evaluate(() => {
            const navigation = performance.getEntriesByType('navigation')[0];
            return {
                loadTime: navigation.loadEventEnd - navigation.loadEventStart,
                domContentLoaded: navigation.domContentLoadedEventEnd - navigation.domContentLoadedEventStart,
                firstPaint: performance.getEntriesByName('first-paint')[0]?.startTime || 0,
                firstContentfulPaint: performance.getEntriesByName('first-contentful-paint')[0]?.startTime || 0
            };
        });
        
        console.log('   📊 性能指标:');
        console.log(`      页面加载时间: ${performanceMetrics.loadTime.toFixed(2)}ms`);
        console.log(`      DOM 解析时间: ${performanceMetrics.domContentLoaded.toFixed(2)}ms`);
        console.log(`      首次绘制: ${performanceMetrics.firstPaint.toFixed(2)}ms`);
        console.log(`      首次内容绘制: ${performanceMetrics.firstContentfulPaint.toFixed(2)}ms`);
        
        if (performanceMetrics.loadTime < 3000) {
            testResults.performance = true;
            console.log('   ✅ 性能表现良好');
        } else {
            console.log('   ⚠️  页面加载时间偏长');
        }
        
    } catch (error) {
        console.error('❌ 测试过程中发生错误:', error);
    } finally {
        // 生成测试报告
        console.log('\n📋 测试报告:');
        console.log('='.repeat(50));
        
        const passedTests = Object.values(testResults).filter(Boolean).length;
        const totalTests = Object.keys(testResults).length;
        
        console.log(`总测试项: ${totalTests}`);
        console.log(`通过测试: ${passedTests}`);
        console.log(`测试通过率: ${((passedTests / totalTests) * 100).toFixed(1)}%`);
        console.log('');
        
        Object.entries(testResults).forEach(([test, passed]) => {
            const status = passed ? '✅ 通过' : '❌ 失败';
            const testName = {
                homepage: '首页加载',
                navigation: '导航功能',
                writePage: '写信页面',
                readPage: '阅读页面',
                plaza: '写作广场',
                museum: '信件博物馆',
                shop: '信封商城',
                responsiveDesign: '响应式设计',
                performance: '性能表现'
            }[test] || test;
            
            console.log(`${status} ${testName}`);
        });
        
        console.log('\n🎯 测试建议:');
        if (!testResults.performance) {
            console.log('- 考虑优化页面加载性能');
        }
        if (!testResults.responsiveDesign) {
            console.log('- 检查移动端响应式设计');
        }
        if (passedTests === totalTests) {
            console.log('🎉 所有测试通过！应用功能完整，可以投入使用。');
        } else {
            console.log('⚠️  部分测试未通过，建议修复后重新测试。');
        }
        
        await browser.close();
        console.log('\n🏁 测试完成！');
    }
}

// 运行测试
runOpenPenPalTests().catch(console.error);