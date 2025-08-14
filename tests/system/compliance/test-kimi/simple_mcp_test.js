// 简化的 MCP 浏览器测试
const { chromium } = require('playwright');

async function simpleTest() {
    console.log('🔍 快速检查 OpenPenPal 状态...');
    
    const browser = await chromium.launch({ headless: false });
    const page = await browser.newPage();
    
    try {
        // 测试首页
        console.log('📖 检查首页...');
        await page.goto('http://localhost:3000', { timeout: 10000 });
        await page.waitForLoadState('networkidle', { timeout: 5000 });
        
        const title = await page.title();
        console.log(`✅ 首页标题: ${title}`);
        
        // 检查导航链接
        console.log('🔗 检查导航链接...');
        const links = await page.locator('nav a').allTextContents();
        console.log('导航链接:', links);
        
        // 尝试访问写信页面
        console.log('✍️ 测试写信页面...');
        try {
            await page.goto('http://localhost:3000/write');
            await page.waitForLoadState('networkidle', { timeout: 5000 });
            console.log('✅ 写信页面可访问');
        } catch (e) {
            console.log('❌ 写信页面访问失败:', e.message);
        }
        
        // 检查页面元素
        console.log('🔍 检查页面元素...');
        const pageContent = await page.content();
        console.log('页面长度:', pageContent.length);
        
        // 简单截图
        await page.screenshot({ path: 'openpenpal-test.png' });
        console.log('📸 截图已保存: openpenpal-test.png');
        
    } catch (error) {
        console.error('❌ 测试失败:', error.message);
    } finally {
        await browser.close();
    }
}

simpleTest().catch(console.error);