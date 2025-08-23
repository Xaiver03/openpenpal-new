# CSP (Content Security Policy) Comprehensive Analysis Report

## Executive Summary

The OpenPenPal project has CSP configurations that are causing frequent issues and blocking legitimate functionality. After thorough analysis, I've identified several critical problems and conflicts that need resolution.

## Current CSP Configuration Analysis

### 1. Configuration Location
- **Primary Config**: `/frontend/src/lib/security/https-config.ts`
- **Middleware**: `/frontend/src/middleware-security.ts` and `/frontend/src/middleware.ts`
- **Applied via**: SecurityHeaders class in middleware chain

### 2. Current CSP Directives (Development)

```javascript
{
  'default-src': ["'self'"],
  'script-src': ["'self'", "'unsafe-inline'", "'unsafe-eval'", "https://cdn.jsdelivr.net", "http://localhost:*"],
  'style-src': ["'self'", "'unsafe-inline'", "https://fonts.googleapis.com"],
  'font-src': ["'self'", "https://fonts.gstatic.com", "data:"],
  'img-src': ["'self'", "data:", "https:", "http:", "blob:"],
  'connect-src': ["'self'", "ws://localhost:*", "wss://localhost:*", "ws://*", "wss://*", "http://localhost:*", "https://*"],
  'media-src': ["'self'"],
  'object-src': ["'none'", "data:"],
  'base-uri': ["'self'"],
  'form-action': ["'self'"]
}
```

### 3. Current CSP Directives (Production)

```javascript
{
  'default-src': ["'self'"],
  'script-src': ["'self'", "'unsafe-inline'", "'unsafe-eval'", "https://cdn.jsdelivr.net", "https://unpkg.com"],
  'style-src': ["'self'", "'unsafe-inline'", "https://fonts.googleapis.com"],
  'font-src': ["'self'", "https://fonts.gstatic.com", "data:"],
  'img-src': ["'self'", "data:", "https:", "blob:"],
  'connect-src': ["'self'", "ws://localhost:*", "wss://localhost:*", "ws://*", "wss://*", "https://*"],
  'media-src': ["'self'"],
  'object-src': ["'none'", "data:"],
  'frame-ancestors': ["'none'"],
  'base-uri': ["'self'"],
  'form-action': ["'self'"],
  'upgrade-insecure-requests': []
}
```

## Identified Issues and Conflicts

### 1. **Overly Permissive Directives**
- **`'unsafe-inline'` and `'unsafe-eval'`**: These defeat the purpose of CSP by allowing inline scripts and eval()
- **`ws://*` and `wss://*`**: Allows WebSocket connections to ANY domain
- **`https://*`**: Allows connections to ANY HTTPS endpoint
- **Risk**: These broad permissions make CSP ineffective against XSS attacks

### 2. **Next.js Integration Issues**
- Next.js requires `'unsafe-inline'` for hydration scripts
- Next.js uses inline styles for styled-jsx
- Service worker needs proper CSP headers
- **Missing**: worker-src directive for service workers
- **Missing**: manifest-src for PWA manifest

### 3. **WebSocket Configuration Problems**
- WebSocket URLs are constructed dynamically based on protocol
- Development uses `ws://localhost:8080`
- Production should use `wss://` but CSP allows both `ws://` and `wss://`
- **Security Risk**: Allowing unencrypted WebSocket in production

### 4. **API Endpoint Conflicts**
- Multiple API base URLs configured:
  - Gateway: `http://localhost:8080/api/v1`
  - Backend: `http://localhost:8080/api/v1`
  - Microservices: Various ports (8001-8004)
- CSP allows all `http://localhost:*` which is too broad

### 5. **Missing Critical Directives**
- **worker-src**: Not defined, may block service workers
- **manifest-src**: Not defined, may block PWA manifest
- **child-src**: Not defined, defaults to frame-src
- **frame-src**: Not defined, defaults to child-src

### 6. **Object-src Conflict**
- Currently set to `['none', 'data:']` which is contradictory
- Should be either `'none'` OR include `data:` for SVG support
- This causes SVG data URLs to potentially be blocked

### 7. **Report-Only Mode Issues**
- Development has `reportOnly: false` but no reporting endpoint
- Production doesn't use report-only for testing
- No CSP violation reporting mechanism

### 8. **Nonce Implementation**
- Nonce generation code exists but is disabled (`nonce: false`)
- Without nonce, must use `'unsafe-inline'` which is insecure
- Nonce would allow removing `'unsafe-inline'` while supporting Next.js

### 9. **Cross-Origin Headers Conflicts**
- COEP: `require-corp` may block external resources
- COOP: `same-origin` may affect OAuth flows
- CORP: `same-origin` may block legitimate cross-origin requests

## Specific Functionality Blocked

### 1. **Potential Service Worker Issues**
- Missing `worker-src` directive
- Service worker at `/sw.js` may be blocked in strict environments

### 2. **Dynamic Script Loading**
- Next.js dynamic imports may be affected
- Chunk loading could fail without proper CSP

### 3. **Third-Party Integrations**
- Google Fonts are allowed but no actual usage found
- CDN resources (jsdelivr, unpkg) allowed but risky

### 4. **WebSocket Connections**
- Too permissive, allowing any WebSocket connection
- Should be restricted to specific domains

## Recommendations

### 1. **Implement Nonce-Based CSP**
```javascript
// Enable nonce in configuration
csp: {
  enabled: true,
  nonce: true, // Enable this
  directives: {
    'script-src': ["'self'", "'nonce-{NONCE}'", "https://cdn.jsdelivr.net"],
    'style-src': ["'self'", "'nonce-{NONCE}'", "https://fonts.googleapis.com"]
  }
}
```

### 2. **Restrict Connect-src**
```javascript
'connect-src': [
  "'self'",
  "http://localhost:8080", // Main backend
  "http://localhost:8001", // Write service
  "http://localhost:8002", // Courier service
  "http://localhost:8003", // Admin service
  "http://localhost:8004", // OCR service
  "ws://localhost:8080",    // WebSocket (dev only)
  "wss://your-domain.com"   // WebSocket (prod only)
]
```

### 3. **Add Missing Directives**
```javascript
'worker-src': ["'self'", "blob:"],
'child-src': ["'self'", "blob:"],
'manifest-src': ["'self'"],
'frame-src': ["'none'"],
```

### 4. **Fix Object-src**
```javascript
'object-src': ["'none'"], // Remove data: if not needed
// OR
'object-src': ["'self'", "data:"], // If SVG data URLs are required
```

### 5. **Implement CSP Reporting**
```javascript
// Add report-uri or report-to directive
'report-uri': ["/api/security/csp-report"],
// OR use Reporting API
'report-to': ["csp-endpoint"]
```

### 6. **Environment-Specific CSP**
- Development: More permissive for debugging
- Staging: Test stricter CSP with report-only
- Production: Strict CSP with violation reporting

### 7. **Remove Unsafe Directives (Long-term)**
- Migrate away from `'unsafe-inline'` using nonces
- Remove `'unsafe-eval'` by updating code
- Implement strict CSP gradually

### 8. **Update Cross-Origin Policies**
```javascript
headers: {
  'Cross-Origin-Embedder-Policy': 'credentialless', // Instead of require-corp
  'Cross-Origin-Opener-Policy': 'same-origin-allow-popups', // For OAuth
  'Cross-Origin-Resource-Policy': 'cross-origin' // For CDN resources
}
```

## Implementation Priority

### High Priority (Immediate)
1. Fix object-src directive conflict
2. Add worker-src for service worker support
3. Restrict connect-src to specific endpoints
4. Add CSP violation reporting

### Medium Priority (Next Sprint)
1. Implement nonce-based CSP
2. Remove wildcard WebSocket permissions
3. Add manifest-src for PWA
4. Test with report-only mode

### Low Priority (Future)
1. Remove 'unsafe-inline' completely
2. Remove 'unsafe-eval' 
3. Implement strict CSP
4. Add integrity checks for external scripts

## Testing Strategy

1. **Enable Report-Only Mode**
   - Monitor violations without breaking functionality
   - Collect data on what needs to be allowed

2. **Gradual Rollout**
   - Start with report-only in production
   - Fix violations based on reports
   - Enable enforcement once stable

3. **Automated Testing**
   - Add CSP header validation tests
   - Test all features with strict CSP
   - Monitor for console errors

## Conclusion

The current CSP configuration is too permissive to provide meaningful security benefits while still causing occasional blocking issues. The recommended approach is to:

1. Fix immediate conflicts (object-src, worker-src)
2. Implement nonce-based CSP to remove unsafe-inline
3. Gradually tighten policies based on actual usage
4. Monitor and adjust based on violation reports

This will result in a CSP that actually provides security benefits without breaking legitimate functionality.