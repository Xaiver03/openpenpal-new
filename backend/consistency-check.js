const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

class ConsistencyChecker {
  constructor() {
    this.issues = [];
    this.warnings = [];
    this.info = [];
  }

  log(type, category, message) {
    const entry = { category, message, timestamp: new Date().toISOString() };
    switch(type) {
      case 'error': this.issues.push(entry); break;
      case 'warning': this.warnings.push(entry); break;
      case 'info': this.info.push(entry); break;
    }
    console.log(`[${type.toUpperCase()}] [${category}] ${message}`);
  }

  // 1. Check API Endpoints Consistency
  async checkAPIEndpoints() {
    console.log('\n🔍 Checking API Endpoints Consistency...\n');
    
    // Frontend API calls
    const frontendAPIs = new Set();
    const apiFiles = [
      '../frontend/src/lib/api-client.ts',
      '../frontend/src/lib/services/auth-service.ts',
      '../frontend/src/lib/services/letter-service.ts',
      '../frontend/src/lib/services/courier-service.ts',
      '../frontend/src/lib/services/ai-service.ts',
      '../frontend/src/lib/services/museum-service.ts',
    ];

    // Extract API calls from frontend
    for (const file of apiFiles) {
      if (fs.existsSync(file)) {
        const content = fs.readFileSync(file, 'utf8');
        const apiMatches = content.matchAll(/['"]\/api\/v1\/([^'"]+)['"]/g);
        for (const match of apiMatches) {
          frontendAPIs.add(`/api/v1/${match[1]}`);
        }
      }
    }

    // Backend routes
    const backendRoutes = new Set();
    const mainGo = fs.readFileSync('main.go', 'utf8');
    
    // Extract routes from main.go
    const routeMatches = mainGo.matchAll(/\.(GET|POST|PUT|DELETE|PATCH)\s*\(\s*["']([^"']+)["']/g);
    for (const match of routeMatches) {
      backendRoutes.add(match[2]);
    }

    // Check for mismatches
    console.log(`Frontend APIs found: ${frontendAPIs.size}`);
    console.log(`Backend routes found: ${backendRoutes.size}`);

    // APIs called by frontend but not in backend
    for (const api of frontendAPIs) {
      const baseAPI = api.replace(/\/:\w+/g, '/:id').replace(/\?.*$/, '');
      let found = false;
      for (const route of backendRoutes) {
        if (route.includes(baseAPI.replace('/api/v1', '')) || 
            baseAPI.includes(route.replace(/:\w+/g, ':id'))) {
          found = true;
          break;
        }
      }
      if (!found && !api.includes('undefined')) {
        this.log('error', 'API', `Frontend calls ${api} but backend doesn't implement it`);
      }
    }
  }

  // 2. Check Data Model Consistency
  async checkDataModels() {
    console.log('\n🔍 Checking Data Model Consistency...\n');
    
    // TypeScript interfaces
    const tsModels = {};
    const modelFiles = [
      '../frontend/src/types/user.ts',
      '../frontend/src/types/letter.ts',
      '../frontend/src/types/courier.ts',
      '../frontend/src/types/ai.ts',
      '../frontend/src/types/museum.ts',
    ];

    for (const file of modelFiles) {
      if (fs.existsSync(file)) {
        const content = fs.readFileSync(file, 'utf8');
        const interfaces = content.matchAll(/export\s+interface\s+(\w+)\s*{([^}]+)}/g);
        for (const match of interfaces) {
          const name = match[1];
          const fields = match[2].match(/(\w+)\s*[?:]?\s*([^;]+);/g) || [];
          tsModels[name] = fields.map(f => {
            const [, fieldName, fieldType] = f.match(/(\w+)\s*[?:]?\s*([^;]+);/) || [];
            return { name: fieldName, type: fieldType?.trim() };
          });
        }
      }
    }

    // Go models
    const goModels = {};
    const goModelFiles = execSync('find internal/models -name "*.go"', { encoding: 'utf8' })
      .trim().split('\n').filter(f => f);

    for (const file of goModelFiles) {
      if (fs.existsSync(file)) {
        const content = fs.readFileSync(file, 'utf8');
        const structs = content.matchAll(/type\s+(\w+)\s+struct\s*{([^}]+)}/g);
        for (const match of structs) {
          const name = match[1];
          const fields = match[2].match(/(\w+)\s+([^`]+)`[^`]+`/g) || [];
          goModels[name] = fields.map(f => {
            const [, fieldName, fieldType] = f.match(/(\w+)\s+([^`]+)`/) || [];
            return { name: fieldName, type: fieldType?.trim() };
          });
        }
      }
    }

    console.log(`TypeScript models found: ${Object.keys(tsModels).length}`);
    console.log(`Go models found: ${Object.keys(goModels).length}`);

    // Check for common models
    const commonModels = ['User', 'Letter', 'Courier', 'AIConfig'];
    for (const model of commonModels) {
      if (!goModels[model]) {
        this.log('warning', 'Model', `Go model ${model} not found`);
      }
    }
  }

  // 3. Check Authentication Consistency
  async checkAuthentication() {
    console.log('\n🔍 Checking Authentication Consistency...\n');
    
    // Check JWT implementation
    const jwtSecret = process.env.JWT_SECRET || 'check-.env-file';
    if (jwtSecret === 'check-.env-file') {
      this.log('warning', 'Auth', 'JWT_SECRET not found in environment');
    }

    // Check auth middleware usage
    const authMiddleware = fs.readFileSync('main.go', 'utf8');
    const protectedGroups = authMiddleware.match(/protected\s*:=\s*v1\.Group/g) || [];
    const publicGroups = authMiddleware.match(/public\s*:=\s*v1\.Group/g) || [];

    console.log(`Protected route groups: ${protectedGroups.length}`);
    console.log(`Public route groups: ${publicGroups.length}`);

    // Check frontend auth handling
    const authContext = fs.existsSync('../frontend/src/contexts/auth-context.tsx');
    const authService = fs.existsSync('../frontend/src/lib/services/auth-service.ts');

    if (!authContext) {
      this.log('error', 'Auth', 'Frontend auth context not found');
    }
    if (!authService) {
      this.log('error', 'Auth', 'Frontend auth service not found');
    }
  }

  // 4. Check Database Schema
  async checkDatabaseSchema() {
    console.log('\n🔍 Checking Database Schema...\n');
    
    try {
      // Get all tables
      const tables = execSync(`psql -U rocalight -d openpenpal -t -c "SELECT tablename FROM pg_tables WHERE schemaname = 'public';" 2>/dev/null || echo "DB_ERROR"`, { encoding: 'utf8' });
      
      if (tables.includes('DB_ERROR')) {
        this.log('error', 'Database', 'Cannot connect to database');
        return;
      }

      const tableList = tables.trim().split('\n').map(t => t.trim()).filter(t => t);
      console.log(`Database tables found: ${tableList.length}`);

      // Expected tables based on models
      const expectedTables = [
        'users', 'letters', 'couriers', 'ai_configs', 'ai_inspirations',
        'ai_usage_logs', 'museum_items', 'museum_entries', 'letter_templates',
        'signal_codes', 'envelope_styles', 'courier_tasks', 'scan_records'
      ];

      for (const table of expectedTables) {
        if (!tableList.some(t => t === table)) {
          this.log('warning', 'Database', `Expected table '${table}' not found`);
        }
      }

      // Check for orphaned tables
      for (const table of tableList) {
        if (!expectedTables.includes(table) && !table.includes('gorm') && !table.includes('migration')) {
          this.log('info', 'Database', `Additional table found: '${table}'`);
        }
      }

    } catch (error) {
      this.log('error', 'Database', `Failed to check schema: ${error.message}`);
    }
  }

  // 5. Check Business Logic Consistency
  async checkBusinessLogic() {
    console.log('\n🔍 Checking Business Logic Consistency...\n');
    
    // Check courier hierarchy
    const courierLevels = ['courier_level1', 'courier_level2', 'courier_level3', 'courier_level4'];
    const courierService = fs.readFileSync('../services/courier-service/internal/services/hierarchy.go', 'utf8');
    
    for (const level of courierLevels) {
      if (!courierService.includes(level)) {
        this.log('warning', 'Business', `Courier level ${level} not found in hierarchy service`);
      }
    }

    // Check OP Code format
    const opCodePattern = /[A-Z]{2}[A-Z0-9]{2}[A-Z0-9]{2}/;
    const opCodeService = fs.existsSync('internal/services/opcode_service.go');
    if (!opCodeService) {
      this.log('warning', 'Business', 'OP Code service not found');
    }

    // Check letter status flow
    const letterStatuses = ['draft', 'pending_payment', 'paid', 'collected', 'in_transit', 'delivered'];
    const letterModel = fs.readFileSync('internal/models/letter.go', 'utf8');
    
    for (const status of letterStatuses) {
      if (!letterModel.includes(status)) {
        this.log('info', 'Business', `Letter status '${status}' not found in model`);
      }
    }
  }

  // 6. Check Configuration Consistency
  async checkConfiguration() {
    console.log('\n🔍 Checking Configuration Consistency...\n');
    
    // Backend config
    const backendEnv = fs.existsSync('.env') ? fs.readFileSync('.env', 'utf8') : '';
    const requiredEnvVars = [
      'DATABASE_URL', 'JWT_SECRET', 'FRONTEND_URL', 'PORT',
      'MOONSHOT_API_KEY', 'AI_PROVIDER'
    ];

    for (const envVar of requiredEnvVars) {
      if (!backendEnv.includes(envVar)) {
        this.log('error', 'Config', `Required env var ${envVar} not found in backend .env`);
      }
    }

    // Frontend config
    const frontendEnv = fs.existsSync('../frontend/.env.local') ? 
      fs.readFileSync('../frontend/.env.local', 'utf8') : '';
    
    if (!frontendEnv.includes('NEXT_PUBLIC_API_URL')) {
      this.log('warning', 'Config', 'NEXT_PUBLIC_API_URL not found in frontend env');
    }

    // Service configs
    const services = [
      { name: 'admin-service', port: 8003 },
      { name: 'courier-service', port: 8002 },
      { name: 'write-service', port: 8001 },
      { name: 'ocr-service', port: 8004 }
    ];

    for (const service of services) {
      const configFile = `../services/${service.name}/config.yaml`;
      if (!fs.existsSync(configFile)) {
        this.log('info', 'Config', `Config file not found for ${service.name}`);
      }
    }
  }

  // Generate report
  generateReport() {
    console.log('\n📊 Consistency Check Report\n');
    console.log('='.repeat(50));
    
    const report = {
      timestamp: new Date().toISOString(),
      summary: {
        errors: this.issues.length,
        warnings: this.warnings.length,
        info: this.info.length
      },
      issues: this.issues,
      warnings: this.warnings,
      info: this.info
    };

    fs.writeFileSync('consistency-report.json', JSON.stringify(report, null, 2));
    
    console.log(`\n❌ Errors: ${this.issues.length}`);
    console.log(`⚠️  Warnings: ${this.warnings.length}`);
    console.log(`ℹ️  Info: ${this.info.length}`);
    
    if (this.issues.length > 0) {
      console.log('\n🚨 Critical Issues Found:');
      this.issues.forEach(issue => {
        console.log(`  - [${issue.category}] ${issue.message}`);
      });
    }
    
    console.log('\n✅ Full report saved to consistency-report.json');
  }

  async runAllChecks() {
    console.log('🚀 Starting Comprehensive Consistency Check...\n');
    
    await this.checkAPIEndpoints();
    await this.checkDataModels();
    await this.checkAuthentication();
    await this.checkDatabaseSchema();
    await this.checkBusinessLogic();
    await this.checkConfiguration();
    
    this.generateReport();
  }
}

// Run the checker
const checker = new ConsistencyChecker();
checker.runAllChecks().catch(console.error);