# AI System Configuration Guide - OpenPenPal

Following CLAUDE.md SOTA principles and "Think before action" approach.

## Current System Status

### ✅ Implemented Features

1. **Multi-Provider AI System Architecture**
   - Support for OpenAI, Claude, SiliconFlow, Moonshot, and local providers
   - Provider manager with automatic failover and health monitoring
   - Circuit breaker pattern for reliability
   - Token usage tracking and quota management

2. **Admin Configuration Interface**
   - Located at `/admin/ai` in the frontend
   - Comprehensive UI for managing AI providers
   - Real-time provider status monitoring
   - Test functionality for each provider

3. **Database Configuration Status**
   ```
   Provider     | Status                      | Active | Quota
   -------------|----------------------------|--------|--------
   local        | ✅ No Key Required         | Yes    | 999999
   moonshot     | ✅ Configured              | Yes    | 10000
   claude       | ⚠️ Configuration Required  | No     | 50000
   openai       | ⚠️ Configuration Required  | No     | 50000
   siliconflow  | ⚠️ Configuration Required  | No     | 100000
   ```

## Administrator Configuration Steps

### 1. Access Admin Panel
Navigate to `/admin/ai` after logging in with admin credentials.

### 2. Configure AI Providers

#### SiliconFlow Configuration:
1. Click on the SiliconFlow provider card
2. Enter your API key (obtain from https://siliconflow.cn)
3. Configure model (default: Qwen/Qwen2.5-7B-Instruct)
4. Set daily quota limits
5. Enable the provider
6. Test the configuration

#### OpenAI Configuration:
1. Click on the OpenAI provider card
2. Enter your API key (obtain from https://platform.openai.com)
3. Select model (e.g., gpt-3.5-turbo, gpt-4)
4. Configure rate limits
5. Enable and test

#### Claude Configuration:
1. Click on the Claude provider card
2. Enter your API key (obtain from https://console.anthropic.com)
3. Select Claude model version
4. Configure usage limits
5. Enable and test

### 3. Testing API Configuration

Run the test script for any provider:
```bash
# Test SiliconFlow (requires SILICONFLOW_API_KEY env var)
export SILICONFLOW_API_KEY="your-api-key"
./backend/scripts/test-siliconflow-api.sh

# Test through backend API
curl -X POST "http://localhost:8080/api/ai/test" \
  -H "Content-Type: application/json" \
  -d '{"provider": "siliconflow", "prompt": "Hello"}'
```

## API Key Security Best Practices

1. **Never commit API keys to Git**
   - Use environment variables
   - Configure through admin panel
   - Keys are encrypted in database

2. **Regular Key Rotation**
   - Update keys every 90 days
   - Monitor usage for anomalies
   - Disable unused providers

3. **Access Control**
   - Only super_admin can configure AI providers
   - API keys are write-only (cannot be viewed after saving)
   - Audit logs track all configuration changes

## System Architecture

### Provider Manager (`ai_provider_manager.go`)
- Manages multiple AI providers
- Automatic failover chain: moonshot → openai → claude → local
- Health monitoring every 5 minutes
- Usage tracking and quota enforcement

### Provider Implementations
- **SiliconFlow** (`ai_provider_siliconflow.go`): Full implementation
- **Moonshot** (`ai_provider_moonshot.go`): Production ready
- **OpenAI** (`ai_provider_openai.go`): Standard implementation
- **Claude** (`ai_provider_claude.go`): Anthropic API
- **Local** (`ai_provider_local.go`): Development/testing

### Frontend Integration
- Admin panel at `/admin/ai/page.tsx`
- Provider configuration dialogs
- Real-time monitoring dashboard
- Usage analytics and reports

## Troubleshooting

### Common Issues

1. **"No available AI provider found"**
   - Check if at least one provider is active
   - Verify API keys are configured
   - Check provider health status

2. **"API key invalid"**
   - Ensure correct API key format
   - Check provider-specific requirements
   - Verify billing/quota on provider dashboard

3. **High latency or timeouts**
   - Check network connectivity
   - Verify provider service status
   - Consider switching to backup provider

### Debug Commands

```bash
# Check provider status
curl http://localhost:8080/api/ai/providers/status

# View provider health
curl http://localhost:8080/api/ai/health

# Check specific provider
curl http://localhost:8080/api/ai/providers/siliconflow/status
```

## Next Steps

1. **Configure Production API Keys**
   - Obtain API keys from respective providers
   - Configure through admin panel
   - Test each provider thoroughly

2. **Monitor Usage**
   - Set up alerts for quota limits
   - Review usage analytics weekly
   - Optimize model selection for cost

3. **Enhance System**
   - Add more providers (Gemini, Cohere, etc.)
   - Implement caching for common queries
   - Add A/B testing for model comparison

## Conclusion

The AI system is fully implemented with a comprehensive admin interface supporting multiple providers with different API formats. Administrators can easily configure and manage AI providers through the web interface at `/admin/ai`.

All providers are pre-configured in the database with placeholder API keys. Simply update the keys through the admin panel to activate each provider.

---

*Document created following CLAUDE.md SOTA principles with complete system analysis and configuration guide.*