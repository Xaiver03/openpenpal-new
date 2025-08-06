import { NextRequest, NextResponse } from 'next/server'
import { API_VERSIONS, CURRENT_VERSION, APIVersionManager } from '@/lib/api/versioning'

// OpenAPI/Swagger Documentation
const openAPISpec = {
  openapi: '3.0.0',
  info: {
    title: 'OpenPenPal API',
    description: 'State-of-the-Art API for Modern Letter Delivery Platform',
    version: '2.0.0',
    contact: {
      name: 'OpenPenPal Team',
      email: 'api@openpenpal.org'
    },
    license: {
      name: 'MIT',
      url: 'https://opensource.org/licenses/MIT'
    }
  },
  servers: [
    {
      url: 'http://localhost:3000/api',
      description: 'Development server'
    },
    {
      url: 'https://api.openpenpal.org',
      description: 'Production server'
    }
  ],
  paths: {
    '/v2/auth/login': {
      post: {
        summary: 'User login',
        tags: ['Authentication'],
        requestBody: {
          required: true,
          content: {
            'application/json': {
              schema: {
                type: 'object',
                properties: {
                  email: { type: 'string', format: 'email' },
                  password: { type: 'string', minLength: 6 }
                },
                required: ['email', 'password']
              }
            }
          }
        },
        responses: {
          '200': {
            description: 'Login successful',
            content: {
              'application/json': {
                schema: { $ref: '#/components/schemas/AuthResponse' }
              }
            }
          },
          '401': {
            description: 'Invalid credentials',
            content: {
              'application/json': {
                schema: { $ref: '#/components/schemas/ErrorResponse' }
              }
            }
          }
        }
      }
    },
    '/v2/letters': {
      get: {
        summary: 'Get letters',
        tags: ['Letters'],
        parameters: [
          {
            name: 'status',
            in: 'query',
            schema: {
              type: 'string',
              enum: ['DRAFT', 'GENERATED', 'COLLECTED', 'IN_TRANSIT', 'DELIVERED', 'FAILED']
            }
          },
          {
            name: 'page',
            in: 'query',
            schema: { type: 'integer', minimum: 1, default: 1 }
          },
          {
            name: 'limit',
            in: 'query',
            schema: { type: 'integer', minimum: 1, maximum: 100, default: 20 }
          }
        ],
        responses: {
          '200': {
            description: 'Letters retrieved successfully',
            content: {
              'application/json': {
                schema: {
                  allOf: [
                    { $ref: '#/components/schemas/BaseResponse' },
                    {
                      type: 'object',
                      properties: {
                        data: {
                          type: 'array',
                          items: { $ref: '#/components/schemas/Letter' }
                        }
                      }
                    }
                  ]
                }
              }
            }
          }
        }
      },
      post: {
        summary: 'Create a new letter',
        tags: ['Letters'],
        security: [{ bearerAuth: [] }],
        requestBody: {
          required: true,
          content: {
            'application/json': {
              schema: { $ref: '#/components/schemas/CreateLetterRequest' }
            }
          }
        },
        responses: {
          '201': {
            description: 'Letter created successfully',
            content: {
              'application/json': {
                schema: {
                  allOf: [
                    { $ref: '#/components/schemas/BaseResponse' },
                    {
                      type: 'object',
                      properties: {
                        data: { $ref: '#/components/schemas/Letter' }
                      }
                    }
                  ]
                }
              }
            }
          }
        }
      }
    },
    '/v2/courier/tasks': {
      get: {
        summary: 'Get courier tasks',
        tags: ['Courier'],
        security: [{ bearerAuth: [] }],
        parameters: [
          {
            name: 'status',
            in: 'query',
            schema: {
              type: 'string',
              enum: ['PENDING', 'COLLECTED', 'IN_TRANSIT', 'DELIVERED', 'FAILED']
            }
          },
          {
            name: 'priority',
            in: 'query',
            schema: {
              type: 'string',
              enum: ['NORMAL', 'URGENT']
            }
          }
        ],
        responses: {
          '200': {
            description: 'Tasks retrieved successfully',
            content: {
              'application/json': {
                schema: {
                  allOf: [
                    { $ref: '#/components/schemas/BaseResponse' },
                    {
                      type: 'object',
                      properties: {
                        data: {
                          type: 'array',
                          items: { $ref: '#/components/schemas/CourierTask' }
                        }
                      }
                    }
                  ]
                }
              }
            }
          }
        }
      }
    },
    '/v2/graphql': {
      post: {
        summary: 'GraphQL endpoint',
        tags: ['GraphQL'],
        description: 'Execute GraphQL queries and mutations',
        requestBody: {
          required: true,
          content: {
            'application/json': {
              schema: {
                type: 'object',
                properties: {
                  query: { type: 'string' },
                  variables: { type: 'object' },
                  operationName: { type: 'string' }
                },
                required: ['query']
              }
            }
          }
        },
        responses: {
          '200': {
            description: 'GraphQL response',
            content: {
              'application/json': {
                schema: {
                  type: 'object',
                  properties: {
                    data: { type: 'object' },
                    errors: {
                      type: 'array',
                      items: {
                        type: 'object',
                        properties: {
                          message: { type: 'string' },
                          path: { type: 'array' },
                          extensions: { type: 'object' }
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  },
  components: {
    schemas: {
      BaseResponse: {
        type: 'object',
        properties: {
          success: { type: 'boolean' },
          meta: {
            type: 'object',
            properties: {
              version: { type: 'string' },
              timestamp: { type: 'string', format: 'date-time' },
              requestId: { type: 'string' },
              deprecationWarning: { type: 'string' },
              pagination: {
                type: 'object',
                properties: {
                  page: { type: 'integer' },
                  limit: { type: 'integer' },
                  total: { type: 'integer' },
                  hasNext: { type: 'boolean' }
                }
              }
            },
            required: ['version', 'timestamp', 'requestId']
          }
        },
        required: ['success', 'meta']
      },
      ErrorResponse: {
        allOf: [
          { $ref: '#/components/schemas/BaseResponse' },
          {
            type: 'object',
            properties: {
              error: {
                type: 'object',
                properties: {
                  code: { type: 'string' },
                  message: { type: 'string' },
                  details: { type: 'object' }
                },
                required: ['code', 'message']
              }
            },
            required: ['error']
          }
        ]
      },
      User: {
        type: 'object',
        properties: {
          id: { type: 'string' },
          username: { type: 'string' },
          email: { type: 'string', format: 'email' },
          nickname: { type: 'string' },
          role: {
            type: 'string',
            enum: ['USER', 'COURIER', 'SENIOR_COURIER', 'COURIER_COORDINATOR', 'SCHOOL_ADMIN', 'PLATFORM_ADMIN', 'SUPER_ADMIN']
          },
          schoolCode: { type: 'string' },
          isActive: { type: 'boolean' },
          createdAt: { type: 'string', format: 'date-time' },
          updatedAt: { type: 'string', format: 'date-time' }
        },
        required: ['id', 'username', 'email', 'nickname', 'role', 'schoolCode', 'isActive']
      },
      Letter: {
        type: 'object',
        properties: {
          id: { type: 'string' },
          title: { type: 'string' },
          content: { type: 'string' },
          status: {
            type: 'string',
            enum: ['DRAFT', 'GENERATED', 'COLLECTED', 'IN_TRANSIT', 'DELIVERED', 'FAILED']
          },
          style: {
            type: 'string',
            enum: ['CLASSIC', 'MODERN', 'HANDWRITTEN', 'ELEGANT']
          },
          code: { type: 'string', nullable: true },
          schoolCode: { type: 'string' },
          createdAt: { type: 'string', format: 'date-time' },
          updatedAt: { type: 'string', format: 'date-time' }
        },
        required: ['id', 'title', 'content', 'status', 'style', 'schoolCode']
      },
      CreateLetterRequest: {
        type: 'object',
        properties: {
          title: { type: 'string', minLength: 1, maxLength: 200 },
          content: { type: 'string', minLength: 1, maxLength: 10000 },
          style: {
            type: 'string',
            enum: ['CLASSIC', 'MODERN', 'HANDWRITTEN', 'ELEGANT'],
            default: 'CLASSIC'
          }
        },
        required: ['title', 'content']
      },
      CourierTask: {
        type: 'object',
        properties: {
          id: { type: 'string' },
          letterCode: { type: 'string' },
          senderName: { type: 'string' },
          senderPhone: { type: 'string', nullable: true },
          recipientHint: { type: 'string' },
          targetLocation: { type: 'string' },
          currentLocation: { type: 'string', nullable: true },
          priority: {
            type: 'string',
            enum: ['NORMAL', 'URGENT']
          },
          status: {
            type: 'string',
            enum: ['PENDING', 'COLLECTED', 'IN_TRANSIT', 'DELIVERED', 'FAILED']
          },
          estimatedTime: { type: 'integer' },
          distance: { type: 'number' },
          reward: { type: 'integer' },
          deadline: { type: 'string', format: 'date-time', nullable: true },
          instructions: { type: 'string', nullable: true },
          createdAt: { type: 'string', format: 'date-time' },
          updatedAt: { type: 'string', format: 'date-time' }
        },
        required: ['id', 'letterCode', 'senderName', 'recipientHint', 'targetLocation', 'priority', 'status', 'estimatedTime', 'distance', 'reward']
      },
      AuthResponse: {
        allOf: [
          { $ref: '#/components/schemas/BaseResponse' },
          {
            type: 'object',
            properties: {
              data: {
                type: 'object',
                properties: {
                  token: { type: 'string' },
                  user: { $ref: '#/components/schemas/User' }
                },
                required: ['token', 'user']
              }
            },
            required: ['data']
          }
        ]
      }
    },
    securitySchemes: {
      bearerAuth: {
        type: 'http',
        scheme: 'bearer',
        bearerFormat: 'JWT'
      }
    }
  },
  tags: [
    {
      name: 'Authentication',
      description: 'User authentication and authorization'
    },
    {
      name: 'Letters',
      description: 'Letter creation, management, and tracking'
    },
    {
      name: 'Courier',
      description: 'Courier task management and delivery operations'
    },
    {
      name: 'GraphQL',
      description: 'GraphQL endpoint for advanced queries'
    }
  ]
}

export async function GET(request: NextRequest) {
  const { searchParams } = new URL(request.url)
  const format = searchParams.get('format') || 'html'
  const version = searchParams.get('version') || CURRENT_VERSION

  if (format === 'json') {
    return NextResponse.json(openAPISpec)
  }

  // Generate HTML documentation
  const docsHTML = `
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>OpenPenPal API Documentation</title>
  <style>
    body {
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
      margin: 0;
      padding: 0;
      background: #f8fafc;
      color: #1a202c;
    }
    .header {
      background: linear-gradient(135deg, #d97706 0%, #92400e 100%);
      color: white;
      padding: 2rem 0;
      text-align: center;
    }
    .container {
      max-width: 1200px;
      margin: 0 auto;
      padding: 2rem;
    }
    .version-badge {
      display: inline-block;
      background: rgba(255,255,255,0.2);
      padding: 0.25rem 0.75rem;
      border-radius: 9999px;
      font-size: 0.875rem;
      margin-bottom: 1rem;
    }
    .title {
      font-size: 3rem;
      font-weight: 700;
      margin: 0;
    }
    .subtitle {
      font-size: 1.25rem;
      margin: 0.5rem 0 0 0;
      opacity: 0.9;
    }
    .section {
      background: white;
      border-radius: 8px;
      padding: 2rem;
      margin: 2rem 0;
      box-shadow: 0 1px 3px rgba(0,0,0,0.1);
    }
    .section-title {
      font-size: 1.5rem;
      font-weight: 600;
      margin-bottom: 1rem;
      color: #d97706;
    }
    .version-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
      gap: 1rem;
      margin: 1rem 0;
    }
    .version-card {
      border: 1px solid #e2e8f0;
      border-radius: 8px;
      padding: 1.5rem;
      background: #f7fafc;
    }
    .version-card.current {
      border-color: #d97706;
      background: #fffbf5;
    }
    .version-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 1rem;
    }
    .version-name {
      font-size: 1.25rem;
      font-weight: 600;
    }
    .status-badge {
      padding: 0.25rem 0.5rem;
      border-radius: 4px;
      font-size: 0.75rem;
      font-weight: 500;
      text-transform: uppercase;
    }
    .status-stable {
      background: #d1fae5;
      color: #065f46;
    }
    .status-deprecated {
      background: #fee2e2;
      color: #991b1b;
    }
    .changelog {
      list-style: none;
      padding: 0;
    }
    .changelog li {
      padding: 0.25rem 0;
      position: relative;
      padding-left: 1rem;
    }
    .changelog li:before {
      content: '•';
      color: #d97706;
      position: absolute;
      left: 0;
    }
    .endpoints {
      margin: 2rem 0;
    }
    .endpoint {
      border: 1px solid #e2e8f0;
      border-radius: 8px;
      margin: 1rem 0;
      overflow: hidden;
    }
    .endpoint-header {
      background: #f7fafc;
      padding: 1rem;
      border-bottom: 1px solid #e2e8f0;
      display: flex;
      align-items: center;
      gap: 1rem;
    }
    .method {
      padding: 0.25rem 0.5rem;
      border-radius: 4px;
      font-size: 0.75rem;
      font-weight: 600;
      text-transform: uppercase;
    }
    .method-get { background: #dbeafe; color: #1e40af; }
    .method-post { background: #dcfce7; color: #166534; }
    .method-put { background: #fef3c7; color: #92400e; }
    .method-delete { background: #fee2e2; color: #991b1b; }
    .endpoint-path {
      font-family: 'Monaco', 'Menlo', 'Courier New', monospace;
      font-weight: 500;
    }
    .endpoint-body {
      padding: 1rem;
    }
    .quick-links {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
      gap: 1rem;
      margin: 2rem 0;
    }
    .quick-link {
      display: block;
      background: white;
      border: 1px solid #e2e8f0;
      border-radius: 8px;
      padding: 1.5rem;
      text-decoration: none;
      color: inherit;
      transition: all 0.2s;
    }
    .quick-link:hover {
      border-color: #d97706;
      transform: translateY(-2px);
      box-shadow: 0 4px 12px rgba(0,0,0,0.1);
    }
    .quick-link-title {
      font-weight: 600;
      margin-bottom: 0.5rem;
      color: #d97706;
    }
    .quick-link-desc {
      color: #64748b;
      font-size: 0.875rem;
    }
  </style>
</head>
<body>
  <div class="header">
    <div class="container">
      <div class="version-badge">API v${version}</div>
      <h1 class="title">OpenPenPal API</h1>
      <p class="subtitle">State-of-the-Art Letter Delivery Platform API</p>
    </div>
  </div>

  <div class="container">
    <!-- Quick Links -->
    <div class="quick-links">
      <a href="/api/graphql" class="quick-link">
        <div class="quick-link-title">🚀 GraphQL Playground</div>
        <div class="quick-link-desc">Interactive GraphQL API explorer</div>
      </a>
      <a href="/api/docs?format=json" class="quick-link">
        <div class="quick-link-title">📋 OpenAPI Spec</div>
        <div class="quick-link-desc">Machine-readable API specification</div>
      </a>
      <a href="https://github.com/Xaiver03/openpenpal-lc" class="quick-link">
        <div class="quick-link-title">📖 Source Code</div>
        <div class="quick-link-desc">View the complete codebase on GitHub</div>
      </a>
      <a href="#migration" class="quick-link">
        <div class="quick-link-title">🔄 Migration Guide</div>
        <div class="quick-link-desc">Upgrade between API versions</div>
      </a>
    </div>

    <!-- API Versions -->
    <div class="section">
      <h2 class="section-title">📦 API Versions</h2>
      <div class="version-grid">
        ${Object.entries(API_VERSIONS).map(([key, versionInfo]) => `
          <div class="version-card ${key === CURRENT_VERSION ? 'current' : ''}">
            <div class="version-header">
              <div class="version-name">${key} (${versionInfo.version})</div>
              <div class="status-badge status-${versionInfo.status}">${versionInfo.status}</div>
            </div>
            <p><strong>Released:</strong> ${versionInfo.releaseDate}</p>
            ${versionInfo.supportUntil ? `<p><strong>Support until:</strong> ${versionInfo.supportUntil}</p>` : ''}
            <div>
              <strong>Changes:</strong>
              <ul class="changelog">
                ${versionInfo.changelog.map(change => `<li>${change}</li>`).join('')}
              </ul>
            </div>
            ${versionInfo.breakingChanges.length > 0 ? `
              <div>
                <strong style="color: #dc2626;">Breaking Changes:</strong>
                <ul class="changelog">
                  ${versionInfo.breakingChanges.map(change => `<li>${change}</li>`).join('')}
                </ul>
              </div>
            ` : ''}
          </div>
        `).join('')}
      </div>
    </div>

    <!-- Key Endpoints -->
    <div class="section">
      <h2 class="section-title">🔗 Key Endpoints</h2>
      <div class="endpoints">
        <div class="endpoint">
          <div class="endpoint-header">
            <span class="method method-post">POST</span>
            <span class="endpoint-path">/api/v2/auth/login</span>
          </div>
          <div class="endpoint-body">
            <p>Authenticate user and receive JWT token</p>
            <p><strong>Body:</strong> { "email": "user@example.com", "password": "password" }</p>
          </div>
        </div>

        <div class="endpoint">
          <div class="endpoint-header">
            <span class="method method-get">GET</span>
            <span class="endpoint-path">/api/v2/letters</span>
          </div>
          <div class="endpoint-body">
            <p>Get letters with optional filtering</p>
            <p><strong>Query params:</strong> status, page, limit</p>
          </div>
        </div>

        <div class="endpoint">
          <div class="endpoint-header">
            <span class="method method-post">POST</span>
            <span class="endpoint-path">/api/v2/letters</span>
          </div>
          <div class="endpoint-body">
            <p>Create a new letter</p>
            <p><strong>Body:</strong> { "title": "Letter Title", "content": "Letter content", "style": "CLASSIC" }</p>
          </div>
        </div>

        <div class="endpoint">
          <div class="endpoint-header">
            <span class="method method-get">GET</span>
            <span class="endpoint-path">/api/v2/courier/tasks</span>
          </div>
          <div class="endpoint-body">
            <p>Get courier tasks (requires courier role)</p>
            <p><strong>Query params:</strong> status, priority, page, limit</p>
          </div>
        </div>

        <div class="endpoint">
          <div class="endpoint-header">
            <span class="method method-post">POST</span>
            <span class="endpoint-path">/api/v2/graphql</span>
          </div>
          <div class="endpoint-body">
            <p>Execute GraphQL queries and mutations</p>
            <p><strong>Body:</strong> { "query": "query { systemStats { totalUsers } }" }</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Authentication -->
    <div class="section">
      <h2 class="section-title">🔐 Authentication</h2>
      <p>The API uses JWT (JSON Web Tokens) for authentication. Include the token in the Authorization header:</p>
      <pre style="background: #f7fafc; padding: 1rem; border-radius: 4px; border: 1px solid #e2e8f0;">Authorization: Bearer YOUR_JWT_TOKEN</pre>
    </div>

    <!-- Rate Limiting -->
    <div class="section">
      <h2 class="section-title">⚡ Rate Limiting</h2>
      <p>API requests are rate-limited per version:</p>
      <ul>
        <li><strong>v1:</strong> 100 requests per hour</li>
        <li><strong>v2:</strong> 1,000 requests per hour</li>
      </ul>
      <p>Rate limit headers are included in all responses.</p>
    </div>

    <!-- Migration Guide -->
    <div class="section" id="migration">
      <h2 class="section-title">🔄 Migration Guide</h2>
      <h3>Migrating from v1 to v2</h3>
      <ol>
        ${APIVersionManager.getMigrationGuide('v1', 'v2').map(step => `<li>${step}</li>`).join('')}
      </ol>
    </div>

    <!-- Error Handling -->
    <div class="section">
      <h2 class="section-title">❌ Error Handling</h2>
      <p>All API responses follow a consistent format:</p>
      <pre style="background: #f7fafc; padding: 1rem; border-radius: 4px; border: 1px solid #e2e8f0;">{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": { ... }
  },
  "meta": {
    "version": "v2",
    "timestamp": "2024-07-20T10:30:00Z",
    "requestId": "req_1234567890"
  }
}</pre>
    </div>
  </div>
</body>
</html>
  `

  return new Response(docsHTML, {
    headers: { 'Content-Type': 'text/html' }
  })
}