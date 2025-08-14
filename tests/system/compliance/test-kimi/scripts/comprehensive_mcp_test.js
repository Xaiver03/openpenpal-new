// 完整的 OpenPenPal MCP 浏览器测试
const { chromium } = require('playwright');

async function comprehensiveTest() {
    console.log('🚀 OpenPenPal 完整功能测试开始...');
    console.log('时间:', new Date().toLocaleString());
    console.log('='.repeat(60));
    
    const browser = await chromium.launch({ 
        headless: false,
        slowMo: 300  // 减慢操作以便观察
    });
    
    const context = await browser.newContext({
        viewport: { width: 1280, height: 720 }
    });
    
    const page = await context.newPage();
    
    const results = {
        homepage: { status: false, details: [], time: 0 },
        navigation: { status: false, details: [], time: 0 },
        writePage: { status: false, details: [], time: 0 },
        readPage: { status: false, details: [], time: 0 },
        responsiveDesign: { status: false, details: [], time: 0 },
        performance: { status: false, details: [], time: 0 }
    };
    
    try {
        // 测试 1: 首页功能测试
        console.log('📖 [1/6] 首页功能测试...');
        const startTime = Date.now();
        
        await page.goto('http://localhost:3000');
        await page.waitForLoadState('networkidle');
        const loadTime = Date.now() - startTime;
        
        // 检查页面基本元素
        const title = await page.title();
        const logo = await page.locator('text=OpenPenPal').first().isVisible();
        const heroSection = await page.locator('h1').first().isVisible();
        const mainCTA = await page.locator('text=写信去').first().isVisible();
        
        results.homepage.time = loadTime;
        results.homepage.details.push(`页面标题: ${title}`);
        results.homepage.details.push(`加载时间: ${loadTime}ms`);
        results.homepage.details.push(`Logo显示: ${logo ? '✅' : '❌'}`);
        results.homepage.details.push(`主标题显示: ${heroSection ? '✅' : '❌'}`);
        results.homepage.details.push(`主要CTA显示: ${mainCTA ? '✅' : '❌'}`);
        
        if (logo && heroSection && mainCTA) {
            results.homepage.status = true;
            console.log('   ✅ 首页基本功能正常');
        } else {
            console.log('   ❌ 首页关键元素缺失');
        }
        
        // 测试 2: 导航功能测试
        console.log('🧭 [2/6] 导航功能测试...');
        const navStartTime = Date.now();
        
        const navTests = [
            { name: '写信去', url: '/write' },
            { name: '写作广场', url: '/plaza' },
            { name: '信件博物馆', url: '/museum' },
            { name: '信封商城', url: '/shop' }
        ];
        
        let navSuccess = 0;
        for (const nav of navTests) {
            try {
                await page.goto('http://localhost:3000');
                await page.waitForLoadState('networkidle');
                
                const navLink = page.locator(`nav a:has-text("${nav.name}")`);
                await navLink.click();
                await page.waitForLoadState('networkidle');
                
                const currentUrl = page.url();
                const isCorrectPage = currentUrl.includes(nav.url);
                
                results.navigation.details.push(`${nav.name}: ${isCorrectPage ? '✅' : '❌'} (${currentUrl})`);
                
                if (isCorrectPage) {
                    navSuccess++;
                    console.log(`   ✅ ${nav.name} 导航成功`);
                } else {
                    console.log(`   ❌ ${nav.name} 导航失败 - URL: ${currentUrl}`);
                }
            } catch (error) {
                results.navigation.details.push(`${nav.name}: ❌ 错误: ${error.message}`);
                console.log(`   ❌ ${nav.name} 导航异常: ${error.message.substring(0, 50)}...`);
            }
        }
        
        results.navigation.time = Date.now() - navStartTime;
        results.navigation.status = navSuccess === navTests.length;
        
        // 测试 3: 写信页面功能
        console.log('✍️ [3/6] 写信页面功能测试...');
        const writeStartTime = Date.now();
        
        await page.goto('http://localhost:3000/write');
        await page.waitForLoadState('networkidle');
        
        // 检查写信页面元素
        const titleInput = await page.locator('input[placeholder*="标题"], input[placeholder*="title"]').isVisible();
        const contentTextarea = await page.locator('textarea').isVisible();
        const saveButton = await page.locator('text=保存草稿').isVisible();
        const generateButton = await page.locator('text=生成编号贴纸, text=生成编号').isVisible();
        
        results.writePage.details.push(`标题输入框: ${titleInput ? '✅' : '❌'}`);
        results.writePage.details.push(`内容文本域: ${contentTextarea ? '✅' : '❌'}`);
        results.writePage.details.push(`保存按钮: ${saveButton ? '✅' : '❌'}`);
        results.writePage.details.push(`生成按钮: ${generateButton ? '✅' : '❌'}`);
        
        if (titleInput && contentTextarea) {
            // 尝试填写内容
            try {
                await page.fill('input[placeholder*="标题"], input[placeholder*="title"]', 'MCP测试信件');
                await page.fill('textarea', '这是通过MCP浏览器自动化工具生成的测试信件。');
                results.writePage.details.push('内容填写: ✅');
                console.log('   ✅ 表单填写测试成功');
            } catch (error) {
                results.writePage.details.push(`内容填写: ❌ ${error.message}`);
                console.log('   ❌ 表单填写测试失败');
            }
        }
        
        results.writePage.time = Date.now() - writeStartTime;
        results.writePage.status = titleInput && contentTextarea;
        
        // 测试 4: 阅读页面功能
        console.log('📖 [4/6] 阅读页面功能测试...');
        const readStartTime = Date.now();
        
        await page.goto('http://localhost:3000/read/OP1K2L3M4N5O');
        await page.waitForLoadState('networkidle');
        
        const letterTitle = await page.locator('h1').isVisible();
        const letterContent = await page.locator('div:has-text("亲爱的朋友")').isVisible();
        const replyButton = await page.locator('text=回信').isVisible();
        const shareButton = await page.locator('text=分享').isVisible();
        
        results.readPage.details.push(`信件标题: ${letterTitle ? '✅' : '❌'}`);
        results.readPage.details.push(`信件内容: ${letterContent ? '✅' : '❌'}`);
        results.readPage.details.push(`回信按钮: ${replyButton ? '✅' : '❌'}`);
        results.readPage.details.push(`分享按钮: ${shareButton ? '✅' : '❌'}`);
        
        // 测试回信功能
        if (replyButton) {
            try {
                await page.click('text=回信');
                await page.waitForLoadState('networkidle');
                
                const currentUrl = page.url();
                const isReplyMode = currentUrl.includes('reply_to=');
                results.readPage.details.push(`回信跳转: ${isReplyMode ? '✅' : '❌'}`);
                
                if (isReplyMode) {
                    console.log('   ✅ 回信功能测试成功');
                } else {
                    console.log('   ❌ 回信功能异常');
                }
            } catch (error) {
                results.readPage.details.push(`回信功能: ❌ ${error.message}`);
            }
        }
        
        results.readPage.time = Date.now() - readStartTime;
        results.readPage.status = letterTitle && letterContent && replyButton;
        
        // 测试 5: 响应式设计
        console.log('📱 [5/6] 响应式设计测试...');
        const responsiveStartTime = Date.now();
        
        await page.goto('http://localhost:3000');
        await page.waitForLoadState('networkidle');
        
        // 桌面端测试
        await page.setViewportSize({ width: 1280, height: 720 });
        await page.waitForTimeout(1000);
        const desktopNav = await page.locator('nav').isVisible();
        
        // 平板端测试
        await page.setViewportSize({ width: 768, height: 1024 });
        await page.waitForTimeout(1000);
        const tabletLayout = await page.locator('nav').isVisible();
        
        // 移动端测试
        await page.setViewportSize({ width: 375, height: 667 });
        await page.waitForTimeout(1000);
        const mobileMenu = await page.locator('button').first().isVisible();
        
        results.responsiveDesign.details.push(`桌面端导航: ${desktopNav ? '✅' : '❌'}`);
        results.responsiveDesign.details.push(`平板端布局: ${tabletLayout ? '✅' : '❌'}`);
        results.responsiveDesign.details.push(`移动端菜单: ${mobileMenu ? '✅' : '❌'}`);
        
        results.responsiveDesign.time = Date.now() - responsiveStartTime;
        results.responsiveDesign.status = desktopNav && tabletLayout && mobileMenu;
        
        // 恢复桌面端视口
        await page.setViewportSize({ width: 1280, height: 720 });
        
        // 测试 6: 性能测试
        console.log('⚡ [6/6] 性能测试...');
        const perfStartTime = Date.now();
        
        await page.goto('http://localhost:3000');
        await page.waitForLoadState('networkidle');
        
        const performanceMetrics = await page.evaluate(() => {
            const navigation = performance.getEntriesByType('navigation')[0];
            const paint = performance.getEntriesByType('paint');
            
            return {
                loadTime: navigation.loadEventEnd - navigation.loadEventStart,
                domContentLoaded: navigation.domContentLoadedEventEnd - navigation.domContentLoadedEventStart,
                firstPaint: paint.find(p => p.name === 'first-paint')?.startTime || 0,
                firstContentfulPaint: paint.find(p => p.name === 'first-contentful-paint')?.startTime || 0
            };
        });
        
        results.performance.details.push(`页面加载: ${performanceMetrics.loadTime.toFixed(0)}ms`);
        results.performance.details.push(`DOM解析: ${performanceMetrics.domContentLoaded.toFixed(0)}ms`);
        results.performance.details.push(`首次绘制: ${performanceMetrics.firstPaint.toFixed(0)}ms`);
        results.performance.details.push(`首次内容绘制: ${performanceMetrics.firstContentfulPaint.toFixed(0)}ms`);
        
        results.performance.time = Date.now() - perfStartTime;
        results.performance.status = performanceMetrics.loadTime < 3000;
        
        // 最终截图
        await page.screenshot({ path: 'openpenpal-final-test.png', fullPage: true });
        
    } catch (error) {
        console.error('❌ 测试过程中发生严重错误:', error);
    } finally {
        await browser.close();
        
        // 生成详细测试报告
        console.log('\n📋 详细测试报告');
        console.log('='.repeat(60));
        
        const totalTests = Object.keys(results).length;
        const passedTests = Object.values(results).filter(r => r.status).length;
        const passRate = ((passedTests / totalTests) * 100).toFixed(1);
        
        console.log(`测试总数: ${totalTests}`);
        console.log(`通过测试: ${passedTests}`);
        console.log(`失败测试: ${totalTests - passedTests}`);
        console.log(`通过率: ${passRate}%`);
        console.log('');
        
        Object.entries(results).forEach(([testName, result]) => {
            const statusIcon = result.status ? '✅' : '❌';
            const testDisplayName = {
                homepage: '首页功能',
                navigation: '导航功能', 
                writePage: '写信页面',
                readPage: '阅读页面',
                responsiveDesign: '响应式设计',
                performance: '性能表现'
            }[testName];
            
            console.log(`${statusIcon} ${testDisplayName} (${result.time}ms)`);
            result.details.forEach(detail => {
                console.log(`   ${detail}`);
            });
            console.log('');
        });
        
        // 问题和建议
        console.log('🎯 测试总结:');
        if (passRate >= 80) {
            console.log('🎉 应用整体功能良好，可以投入使用！');
        } else if (passRate >= 60) {
            console.log('⚠️  应用基本功能正常，但存在部分问题需要修复。');
        } else {
            console.log('❌ 应用存在较多问题，建议全面检查和修复。');
        }
        
        // 具体建议
        console.log('\n💡 优化建议:');
        if (!results.performance.status) {
            console.log('- 优化页面加载性能，目标3秒内完成');
        }
        if (!results.responsiveDesign.status) {
            console.log('- 改进移动端响应式设计');
        }
        if (!results.navigation.status) {
            console.log('- 修复导航路由问题');
        }
        if (!results.writePage.status) {
            console.log('- 检查写信页面表单元素');
        }
        if (!results.readPage.status) {
            console.log('- 验证阅读页面数据加载');
        }
        
        console.log('\n📸 截图文件: openpenpal-final-test.png');
        console.log('🏁 测试完成时间:', new Date().toLocaleString());
    }
}

// 运行测试
comprehensiveTest().catch(console.error);