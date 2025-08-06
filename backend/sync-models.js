const fs = require('fs');
const path = require('path');

// Script to generate synchronized TypeScript interfaces from Go models

class ModelSynchronizer {
  constructor() {
    this.goModels = {};
    this.tsInterfaces = [];
  }

  // Parse Go struct and extract fields
  parseGoStruct(content, structName) {
    const structRegex = new RegExp(`type\\s+${structName}\\s+struct\\s*{([^}]+)}`, 's');
    const match = content.match(structRegex);
    
    if (!match) return null;
    
    const fields = [];
    const fieldRegex = /(\w+)\s+([^\s`]+)\s*`[^`]*json:"([^",]+)([^`]*)`/g;
    let fieldMatch;
    
    while ((fieldMatch = fieldRegex.exec(match[1])) !== null) {
      const goName = fieldMatch[1];
      const goType = fieldMatch[2];
      const jsonName = fieldMatch[3];
      const tags = fieldMatch[4];
      
      if (jsonName !== '-') {
        fields.push({
          goName,
          goType,
          jsonName,
          optional: tags.includes('omitempty'),
          tsType: this.goTypeToTS(goType)
        });
      }
    }
    
    return fields;
  }

  // Convert Go type to TypeScript type
  goTypeToTS(goType) {
    const typeMap = {
      'string': 'string',
      'int': 'number',
      'int32': 'number',
      'int64': 'number',
      'float32': 'number',
      'float64': 'number',
      'bool': 'boolean',
      'time.Time': 'string', // ISO date string
      '*time.Time': 'string | null',
      '[]string': 'string[]',
      '[]int': 'number[]',
      'uuid.UUID': 'string',
      '*string': 'string | null',
      '*int': 'number | null',
    };
    
    // Handle pointer types
    if (goType.startsWith('*')) {
      const baseType = goType.substring(1);
      const tsType = typeMap[baseType] || this.goTypeToTS(baseType);
      return `${tsType} | null`;
    }
    
    // Handle slice types
    if (goType.startsWith('[]')) {
      const baseType = goType.substring(2);
      const tsType = typeMap[baseType] || baseType;
      return `${tsType}[]`;
    }
    
    // Handle custom types
    return typeMap[goType] || goType;
  }

  // Generate TypeScript interface
  generateTSInterface(modelName, fields) {
    let interfaceStr = `export interface ${modelName} {\n`;
    
    for (const field of fields) {
      const optional = field.optional ? '?' : '';
      interfaceStr += `  ${field.jsonName}${optional}: ${field.tsType};\n`;
    }
    
    interfaceStr += '}\n';
    return interfaceStr;
  }

  // Generate synchronized models
  async generateSyncedModels() {
    console.log('🔄 Generating synchronized TypeScript models from Go structs...\n');
    
    // Critical models to sync
    const modelsToSync = [
      { go: 'User', file: 'internal/models/user.go' },
      { go: 'Letter', file: 'internal/models/letter.go' },
      { go: 'Courier', file: 'internal/models/courier.go' },
      { go: 'AIConfig', file: 'internal/models/ai.go' },
      { go: 'MuseumItem', file: 'internal/models/museum.go' },
      { go: 'MuseumEntry', file: 'internal/models/museum.go' }
    ];
    
    let output = `// Auto-generated TypeScript interfaces from Go models
// Generated on: ${new Date().toISOString()}
// DO NOT EDIT MANUALLY - Use sync-models.js to regenerate

`;
    
    for (const model of modelsToSync) {
      if (fs.existsSync(model.file)) {
        console.log(`Processing ${model.go} from ${model.file}...`);
        const content = fs.readFileSync(model.file, 'utf8');
        const fields = this.parseGoStruct(content, model.go);
        
        if (fields) {
          const tsInterface = this.generateTSInterface(model.go, fields);
          output += tsInterface + '\n';
          console.log(`✅ Generated interface for ${model.go} with ${fields.length} fields`);
        } else {
          console.log(`⚠️  Could not parse ${model.go}`);
        }
      } else {
        console.log(`❌ File not found: ${model.file}`);
      }
    }
    
    // Write synchronized models
    const outputPath = '../frontend/src/types/models-sync.ts';
    fs.writeFileSync(outputPath, output);
    console.log(`\n✅ Synchronized models written to ${outputPath}`);
    
    // Generate model mapping utilities
    this.generateMappingUtils();
  }

  // Generate utility functions for model mapping
  generateMappingUtils() {
    const utils = `// Utility functions for model field mapping

// Convert snake_case to camelCase
export function snakeToCamel<T extends Record<string, any>>(obj: T): any {
  if (Array.isArray(obj)) {
    return obj.map(item => snakeToCamel(item));
  }
  
  if (obj !== null && typeof obj === 'object') {
    return Object.keys(obj).reduce((result, key) => {
      const camelKey = key.replace(/_([a-z])/g, (_, letter) => letter.toUpperCase());
      result[camelKey] = snakeToCamel(obj[key]);
      return result;
    }, {} as any);
  }
  
  return obj;
}

// Convert camelCase to snake_case
export function camelToSnake<T extends Record<string, any>>(obj: T): any {
  if (Array.isArray(obj)) {
    return obj.map(item => camelToSnake(item));
  }
  
  if (obj !== null && typeof obj === 'object') {
    return Object.keys(obj).reduce((result, key) => {
      const snakeKey = key.replace(/([A-Z])/g, '_$1').toLowerCase();
      result[snakeKey] = camelToSnake(obj[key]);
      return result;
    }, {} as any);
  }
  
  return obj;
}

// Type-safe model mappers
export const ModelMappers = {
  user: {
    fromAPI: (data: any): User => snakeToCamel(data),
    toAPI: (user: User): any => camelToSnake(user)
  },
  letter: {
    fromAPI: (data: any): Letter => snakeToCamel(data),
    toAPI: (letter: Letter): any => camelToSnake(letter)
  },
  courier: {
    fromAPI: (data: any): Courier => snakeToCamel(data),
    toAPI: (courier: Courier): any => camelToSnake(courier)
  }
};
`;
    
    const utilsPath = '../frontend/src/utils/model-mappers.ts';
    fs.writeFileSync(utilsPath, utils);
    console.log(`✅ Model mapping utilities written to ${utilsPath}`);
  }
}

// Create API consistency checker
class APIConsistencyFixer {
  async generateAPIClient() {
    console.log('\n🔧 Generating consistent API client...\n');
    
    const apiClient = `// Consistent API client with proper error handling and route mapping
import { ModelMappers } from '@/utils/model-mappers';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

class APIClient {
  private token: string | null = null;
  
  setToken(token: string | null) {
    this.token = token;
  }
  
  private async request<T>(
    path: string,
    options: RequestInit = {}
  ): Promise<T> {
    // Fix path inconsistencies
    const fixedPath = this.fixAPIPath(path);
    
    const headers = {
      'Content-Type': 'application/json',
      ...(this.token && { Authorization: \`Bearer \${this.token}\` }),
      ...options.headers,
    };
    
    const response = await fetch(\`\${API_BASE}\${fixedPath}\`, {
      ...options,
      headers,
    });
    
    if (!response.ok) {
      throw new Error(\`API error: \${response.status}\`);
    }
    
    const data = await response.json();
    return data;
  }
  
  private fixAPIPath(path: string): string {
    // Map frontend paths to backend paths
    const pathMap: Record<string, string> = {
      '/api/auth/login': '/api/v1/auth/login',
      '/api/auth/register': '/api/v1/auth/register',
      '/api/auth/logout': '/api/v1/auth/logout',
      '/api/auth/me': '/api/v1/users/me',
      '/api/auth/refresh': '/api/v1/auth/refresh',
    };
    
    return pathMap[path] || path;
  }
  
  // Auth methods
  auth = {
    login: async (credentials: LoginRequest) => {
      const response = await this.request<LoginResponse>('/api/auth/login', {
        method: 'POST',
        body: JSON.stringify(credentials),
      });
      return response;
    },
    
    register: async (data: RegisterRequest) => {
      const response = await this.request<RegisterResponse>('/api/auth/register', {
        method: 'POST',
        body: JSON.stringify(data),
      });
      return response;
    },
    
    getCurrentUser: async () => {
      const response = await this.request<UserResponse>('/api/auth/me');
      return {
        ...response,
        data: ModelMappers.user.fromAPI(response.data)
      };
    },
  };
  
  // Letter methods
  letters = {
    create: async (letter: Partial<Letter>) => {
      const response = await this.request<LetterResponse>('/api/v1/letters', {
        method: 'POST',
        body: JSON.stringify(ModelMappers.letter.toAPI(letter)),
      });
      return {
        ...response,
        data: ModelMappers.letter.fromAPI(response.data)
      };
    },
    
    get: async (id: string) => {
      const response = await this.request<LetterResponse>(\`/api/v1/letters/\${id}\`);
      return {
        ...response,
        data: ModelMappers.letter.fromAPI(response.data)
      };
    },
  };
}

export const apiClient = new APIClient();
`;
    
    const clientPath = '../frontend/src/lib/api-client-fixed.ts';
    fs.writeFileSync(clientPath, apiClient);
    console.log(`✅ Fixed API client written to ${clientPath}`);
  }
}

// Run synchronization
async function main() {
  const synchronizer = new ModelSynchronizer();
  await synchronizer.generateSyncedModels();
  
  const apiFixer = new APIConsistencyFixer();
  await apiFixer.generateAPIClient();
  
  console.log('\n✅ Model and API synchronization complete!');
  console.log('\nNext steps:');
  console.log('1. Review generated files in frontend/src/types/models-sync.ts');
  console.log('2. Update imports to use synchronized models');
  console.log('3. Use ModelMappers for API data transformation');
  console.log('4. Replace api-client.ts with api-client-fixed.ts');
}

main().catch(console.error);