const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

class DeepConsistencyChecker {
  constructor() {
    this.report = {
      critical: [],
      warnings: [],
      inconsistencies: [],
      suggestions: []
    };
  }

  // 1. Deep API Analysis
  async analyzeAPIs() {
    console.log('\n🔍 Deep API Analysis...\n');
    
    // Analyze frontend API calls
    const apiCalls = {};
    const frontendDir = '../frontend/src';
    
    // Scan all TypeScript/JavaScript files for API calls
    const files = execSync(`find ${frontendDir} -name "*.ts" -o -name "*.tsx" | grep -v node_modules`, { encoding: 'utf8' })
      .trim().split('\n').filter(f => f);
    
    for (const file of files) {
      if (fs.existsSync(file)) {
        const content = fs.readFileSync(file, 'utf8');
        
        // Find fetch calls
        const fetchCalls = content.matchAll(/fetch\s*\(\s*[`'"](\/api\/[^`'"]+)[`'"]/g);
        for (const match of fetchCalls) {
          const api = match[1];
          apiCalls[api] = apiCalls[api] || [];
          apiCalls[api].push(file.replace(frontendDir, ''));
        }
        
        // Find axios/api-client calls
        const apiClientCalls = content.matchAll(/api\.(get|post|put|delete|patch)\s*\(\s*[`'"](\/[^`'"]+)[`'"]/g);
        for (const match of apiClientCalls) {
          const api = `/api/v1${match[2]}`;
          apiCalls[api] = apiCalls[api] || [];
          apiCalls[api].push(file.replace(frontendDir, ''));
        }
      }
    }
    
    // Analyze backend routes
    const backendRoutes = {};
    const mainGo = fs.readFileSync('main.go', 'utf8');
    
    // Extract route groups
    const groups = mainGo.matchAll(/(\w+)\s*:=\s*(\w+)\.Group\s*\(\s*["']([^"']+)["']\)/g);
    const groupMap = {};
    for (const match of groups) {
      groupMap[match[1]] = match[3];
    }
    
    // Extract routes with their groups
    const routes = mainGo.matchAll(/(\w+)\.(GET|POST|PUT|DELETE|PATCH)\s*\(\s*["']([^"']+)["'],\s*(\w+)/g);
    for (const match of routes) {
      const group = groupMap[match[1]] || '';
      const method = match[2];
      const route = match[3];
      const handler = match[4];
      const fullRoute = `/api/v1${group}${route}`;
      
      backendRoutes[fullRoute] = {
        method,
        handler,
        group: match[1]
      };
    }
    
    // Compare frontend calls with backend routes
    console.log(`Frontend API calls found: ${Object.keys(apiCalls).length}`);
    console.log(`Backend routes defined: ${Object.keys(backendRoutes).length}`);
    
    // Find unmatched APIs
    for (const [api, files] of Object.entries(apiCalls)) {
      const normalizedApi = api.replace(/\/\d+/g, '/:id').replace(/\?.*$/, '');
      let matched = false;
      
      for (const route of Object.keys(backendRoutes)) {
        if (route === normalizedApi || route.includes(normalizedApi.replace('/api/v1', ''))) {
          matched = true;
          break;
        }
      }
      
      if (!matched) {
        this.report.critical.push({
          type: 'API_MISMATCH',
          message: `Frontend calls ${api} but backend doesn't implement it`,
          files: files
        });
      }
    }
  }

  // 2. Deep Model Analysis
  async analyzeModels() {
    console.log('\n🔍 Deep Model Analysis...\n');
    
    // Analyze TypeScript models
    const tsModels = {};
    const tsFiles = execSync('find ../frontend/src/types -name "*.ts"', { encoding: 'utf8' })
      .trim().split('\n').filter(f => f);
    
    for (const file of tsFiles) {
      if (fs.existsSync(file)) {
        const content = fs.readFileSync(file, 'utf8');
        const interfaces = content.matchAll(/export\s+(?:interface|type)\s+(\w+)(?:<[^>]+>)?\s*(?:=\s*)?{([^}]+)}/gs);
        
        for (const match of interfaces) {
          const modelName = match[1];
          const body = match[2];
          const fields = [];
          
          // Extract fields
          const fieldMatches = body.matchAll(/(\w+)\s*(\?)?:\s*([^;,\n]+)/g);
          for (const fieldMatch of fieldMatches) {
            fields.push({
              name: fieldMatch[1],
              optional: !!fieldMatch[2],
              type: fieldMatch[3].trim()
            });
          }
          
          tsModels[modelName] = {
            file: file.replace('../frontend/src/types/', ''),
            fields
          };
        }
      }
    }
    
    // Analyze Go models
    const goModels = {};
    const goFiles = execSync('find internal/models -name "*.go"', { encoding: 'utf8' })
      .trim().split('\n').filter(f => f);
    
    for (const file of goFiles) {
      if (fs.existsSync(file)) {
        const content = fs.readFileSync(file, 'utf8');
        const structs = content.matchAll(/type\s+(\w+)\s+struct\s*{([^}]+)}/gs);
        
        for (const match of structs) {
          const modelName = match[1];
          const body = match[2];
          const fields = [];
          
          // Extract fields with json tags
          const fieldMatches = body.matchAll(/(\w+)\s+([^`\s]+)\s*`[^`]*json:"([^",]+)[^`]*`/g);
          for (const fieldMatch of fieldMatches) {
            fields.push({
              goName: fieldMatch[1],
              type: fieldMatch[2],
              jsonName: fieldMatch[3]
            });
          }
          
          goModels[modelName] = {
            file: file.replace('internal/models/', ''),
            fields
          };
        }
      }
    }
    
    // Compare models
    const modelPairs = [
      { ts: 'User', go: 'User' },
      { ts: 'Letter', go: 'Letter' },
      { ts: 'Courier', go: 'Courier' },
      { ts: 'AIConfig', go: 'AIConfig' },
      { ts: 'MuseumItem', go: 'MuseumItem' }
    ];
    
    for (const pair of modelPairs) {
      const tsModel = tsModels[pair.ts];
      const goModel = goModels[pair.go];
      
      if (tsModel && goModel) {
        // Check field consistency
        for (const tsField of tsModel.fields) {
          const goField = goModel.fields.find(f => f.jsonName === tsField.name);
          if (!goField) {
            this.report.warnings.push({
              type: 'MODEL_FIELD_MISMATCH',
              message: `TypeScript ${pair.ts}.${tsField.name} has no corresponding Go field`,
              details: { tsFile: tsModel.file, goFile: goModel.file }
            });
          }
        }
        
        for (const goField of goModel.fields) {
          const tsField = tsModel.fields.find(f => f.name === goField.jsonName);
          if (!tsField && goField.jsonName !== '-') {
            this.report.warnings.push({
              type: 'MODEL_FIELD_MISMATCH',
              message: `Go ${pair.go}.${goField.goName} (json: ${goField.jsonName}) has no corresponding TypeScript field`,
              details: { tsFile: tsModel.file, goFile: goModel.file }
            });
          }
        }
      }
    }
    
    console.log(`TypeScript models analyzed: ${Object.keys(tsModels).length}`);
    console.log(`Go models analyzed: ${Object.keys(goModels).length}`);
  }

  // 3. Database Schema Analysis
  async analyzeDatabaseSchema() {
    console.log('\n🔍 Deep Database Schema Analysis...\n');
    
    try {
      // Get detailed schema information
      const schemaQuery = `
        SELECT 
          t.table_name,
          array_agg(
            json_build_object(
              'column', c.column_name,
              'type', c.data_type,
              'nullable', c.is_nullable,
              'default', c.column_default
            ) ORDER BY c.ordinal_position
          ) as columns
        FROM information_schema.tables t
        JOIN information_schema.columns c ON t.table_name = c.table_name
        WHERE t.table_schema = 'public' 
          AND t.table_type = 'BASE TABLE'
          AND t.table_name NOT LIKE '%gorm%'
        GROUP BY t.table_name
        ORDER BY t.table_name;
      `;
      
      const result = execSync(
        `psql -U rocalight -d openpenpal -t -A -F'|' -c "${schemaQuery}" 2>/dev/null || echo "DB_ERROR"`,
        { encoding: 'utf8' }
      );
      
      if (result.includes('DB_ERROR')) {
        this.report.critical.push({
          type: 'DATABASE_ERROR',
          message: 'Cannot connect to database for schema analysis'
        });
        return;
      }
      
      // Parse schema
      const schema = {};
      const lines = result.trim().split('\n');
      for (const line of lines) {
        const [table, columnsJson] = line.split('|');
        if (table && columnsJson) {
          schema[table] = JSON.parse(columnsJson);
        }
      }
      
      // Check critical tables
      const criticalTables = ['users', 'letters', 'couriers', 'ai_configs'];
      for (const table of criticalTables) {
        if (!schema[table]) {
          this.report.critical.push({
            type: 'MISSING_TABLE',
            message: `Critical table '${table}' is missing from database`
          });
        }
      }
      
      // Check for expected columns
      if (schema.users) {
        const expectedColumns = ['id', 'username', 'email', 'password', 'role'];
        for (const col of expectedColumns) {
          if (!schema.users.some(c => c.column === col)) {
            this.report.warnings.push({
              type: 'MISSING_COLUMN',
              message: `Expected column 'users.${col}' not found`
            });
          }
        }
      }
      
      console.log(`Database tables analyzed: ${Object.keys(schema).length}`);
      
    } catch (error) {
      this.report.critical.push({
        type: 'DATABASE_ERROR',
        message: `Database analysis failed: ${error.message}`
      });
    }
  }

  // 4. Business Logic Consistency
  async analyzeBusinessLogic() {
    console.log('\n🔍 Deep Business Logic Analysis...\n');
    
    // Check courier hierarchy implementation
    const hierarchyFile = '../services/courier-service/internal/services/hierarchy.go';
    if (fs.existsSync(hierarchyFile)) {
      const content = fs.readFileSync(hierarchyFile, 'utf8');
      
      // Check for level definitions
      const levels = [
        { name: 'Level1', pattern: /Level1|LEVEL_1|CourierLevel1/ },
        { name: 'Level2', pattern: /Level2|LEVEL_2|CourierLevel2/ },
        { name: 'Level3', pattern: /Level3|LEVEL_3|CourierLevel3/ },
        { name: 'Level4', pattern: /Level4|LEVEL_4|CourierLevel4/ }
      ];
      
      for (const level of levels) {
        if (!level.pattern.test(content)) {
          this.report.inconsistencies.push({
            type: 'COURIER_HIERARCHY',
            message: `Courier ${level.name} not properly implemented in hierarchy service`
          });
        }
      }
    }
    
    // Check OP Code implementation
    const opCodeFile = 'internal/services/opcode_service.go';
    if (fs.existsSync(opCodeFile)) {
      const content = fs.readFileSync(opCodeFile, 'utf8');
      
      // Check OP Code format validation
      if (!content.includes('regexp.MustCompile') || !content.includes('[A-Z]{2}')) {
        this.report.warnings.push({
          type: 'OPCODE_VALIDATION',
          message: 'OP Code format validation might be missing or incorrect'
        });
      }
    }
    
    // Check letter status flow
    const letterService = 'internal/services/letter_service.go';
    if (fs.existsSync(letterService)) {
      const content = fs.readFileSync(letterService, 'utf8');
      const statuses = ['draft', 'sent', 'collected', 'in_transit', 'delivered'];
      
      for (const status of statuses) {
        if (!content.includes(status)) {
          this.report.inconsistencies.push({
            type: 'LETTER_STATUS_FLOW',
            message: `Letter status '${status}' not found in letter service`
          });
        }
      }
    }
  }

  // 5. Authentication Flow Analysis
  async analyzeAuthFlow() {
    console.log('\n🔍 Deep Authentication Flow Analysis...\n');
    
    // Check JWT implementation consistency
    const jwtFiles = [
      { path: 'internal/middleware/auth.go', side: 'backend' },
      { path: '../frontend/src/lib/services/auth-service.ts', side: 'frontend' }
    ];
    
    for (const file of jwtFiles) {
      if (!fs.existsSync(file.path)) {
        this.report.critical.push({
          type: 'AUTH_MISSING',
          message: `Critical auth file missing: ${file.path}`,
          side: file.side
        });
      }
    }
    
    // Check token refresh implementation
    const tokenRefreshProvider = '../frontend/src/components/providers/token-refresh-provider.tsx';
    if (fs.existsSync(tokenRefreshProvider)) {
      const content = fs.readFileSync(tokenRefreshProvider, 'utf8');
      if (!content.includes('refreshToken') || !content.includes('401')) {
        this.report.warnings.push({
          type: 'TOKEN_REFRESH',
          message: 'Token refresh mechanism might not be properly implemented'
        });
      }
    }
    
    // Check CSRF protection
    const csrfMiddleware = 'internal/middleware/csrf.go';
    if (!fs.existsSync(csrfMiddleware)) {
      this.report.suggestions.push({
        type: 'SECURITY',
        message: 'Consider implementing CSRF protection middleware'
      });
    }
  }

  // Generate comprehensive report
  generateReport() {
    console.log('\n📊 Deep Consistency Analysis Report\n');
    console.log('='.repeat(60));
    
    const summary = {
      critical: this.report.critical.length,
      warnings: this.report.warnings.length,
      inconsistencies: this.report.inconsistencies.length,
      suggestions: this.report.suggestions.length
    };
    
    console.log('\nSummary:');
    console.log(`❌ Critical Issues: ${summary.critical}`);
    console.log(`⚠️  Warnings: ${summary.warnings}`);
    console.log(`🔄 Inconsistencies: ${summary.inconsistencies}`);
    console.log(`💡 Suggestions: ${summary.suggestions}`);
    
    if (this.report.critical.length > 0) {
      console.log('\n🚨 Critical Issues:');
      this.report.critical.forEach((issue, i) => {
        console.log(`\n${i + 1}. ${issue.type}: ${issue.message}`);
        if (issue.files) console.log(`   Files: ${issue.files.join(', ')}`);
        if (issue.details) console.log(`   Details:`, issue.details);
      });
    }
    
    if (this.report.warnings.length > 0) {
      console.log('\n⚠️  Warnings:');
      this.report.warnings.forEach((warning, i) => {
        console.log(`\n${i + 1}. ${warning.type}: ${warning.message}`);
        if (warning.details) console.log(`   Details:`, warning.details);
      });
    }
    
    // Save detailed report
    fs.writeFileSync('deep-consistency-report.json', JSON.stringify(this.report, null, 2));
    console.log('\n✅ Detailed report saved to deep-consistency-report.json');
  }

  async runAnalysis() {
    console.log('🚀 Starting Deep Consistency Analysis...\n');
    
    await this.analyzeAPIs();
    await this.analyzeModels();
    await this.analyzeDatabaseSchema();
    await this.analyzeBusinessLogic();
    await this.analyzeAuthFlow();
    
    this.generateReport();
  }
}

// Run the analysis
const checker = new DeepConsistencyChecker();
checker.runAnalysis().catch(console.error);