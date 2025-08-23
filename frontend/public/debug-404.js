// Debug script to catch all fetch requests and identify 404 sources
(function() {
    console.log('🔍 404 Debug Script Loaded');
    
    // Override fetch to log all requests
    const originalFetch = window.fetch;
    window.fetch = function(...args) {
        const url = args[0];
        console.log('📡 Fetch Request:', url, args);
        
        // Check if this is a relative path that would cause 404
        if (typeof url === 'string' && (url === 'public' || url === 'popular')) {
            console.error('🚨 Found problematic relative path request:', url);
            console.trace('Stack trace for relative path request');
        }
        
        return originalFetch.apply(this, args).catch(error => {
            console.error('❌ Fetch error for:', url, error);
            throw error;
        });
    };
    
    // Override XMLHttpRequest
    const originalXHROpen = XMLHttpRequest.prototype.open;
    XMLHttpRequest.prototype.open = function(method, url, ...args) {
        console.log('📡 XHR Request:', method, url);
        
        if (url === 'public' || url === 'popular') {
            console.error('🚨 Found problematic XHR request:', url);
            console.trace('Stack trace for XHR request');
        }
        
        return originalXHROpen.apply(this, [method, url, ...args]);
    };
    
    // Monitor all resource loads
    if (window.PerformanceObserver) {
        const observer = new PerformanceObserver((list) => {
            for (const entry of list.getEntries()) {
                if (entry.name.includes('public') || entry.name.includes('popular')) {
                    if (!entry.name.includes('/api/') && !entry.name.includes('http')) {
                        console.warn('🔍 Suspicious resource:', entry.name, entry);
                    }
                }
            }
        });
        observer.observe({ entryTypes: ['resource'] });
    }
    
    // Monitor dynamic script/link insertions
    const originalAppendChild = Element.prototype.appendChild;
    Element.prototype.appendChild = function(child) {
        if (child.tagName === 'SCRIPT' || child.tagName === 'LINK' || child.tagName === 'IMG') {
            const src = child.src || child.href;
            if (src && (src.endsWith('public') || src.endsWith('popular'))) {
                console.error('🚨 Dynamic element with problematic URL:', child.tagName, src);
                console.trace('Stack trace for dynamic element');
            }
        }
        return originalAppendChild.call(this, child);
    };
    
    console.log('✅ 404 Debug hooks installed');
})();