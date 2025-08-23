# OpenPenPal - Campus Handwritten Letter Platform

**Core Philosophy**: Git version management, strictly prohibit rewriting simplified versions when functionality is abnormal, Think before action, SOTA principles, cautious deletion, continuously optimize user experience, prohibit simplifying problems and skipping problems, prohibit hardcoding data.

## Tech Stack
- Frontend: Next.js 14, TypeScript, Tailwind CSS, React 18
- Backend: Go (Gin), Python (FastAPI), Java (Spring Boot), PostgreSQL 15  
- Testing: Jest, React Testing Library, Go testing, Python pytest
- Architecture: Microservices + WebSocket + JWT Authentication + 4-Level Courier System

## Common Commands
- ./startup/quick-start.sh demo --auto-open: Start demo mode (recommended)
- ./startup/quick-start.sh development --auto-open: Start all services
- ./startup/check-status.sh: Check service status
- ./startup/stop-all.sh: Stop all services
- ./startup/force-cleanup.sh: Force cleanup ports
- npm run dev: Start frontend dev server (cd frontend)
- go run main.go: Start backend service (cd backend)
- npm run type-check: Run TypeScript type checking
- ./scripts/test-apis.sh: Run API tests
- ./test-kimi/run_tests.sh: Run integration tests

## Coding Standards
- Use strict TypeScript mode, avoid any type
- Go code follows gofmt standard formatting
- File naming: snake_case.go, PascalCase.tsx, kebab-case.ts
- API field naming: Backend uses snake_case, frontend matches exactly (no camelCase conversion)
- Database fields: GORM + snake_case JSON fields
- Imports: Prefer destructuring imports import { foo } from 'bar'
- Configuration: Use environment variables, prohibit hardcoding

## Workflow
- Run type-check after each modification to verify TypeScript
- Git branch management: main for production branch, feature/description for feature branches
- Commit format: feat/fix/docs: message
- Think before action: Deep analysis of problems before coding implementation
- SOTA principle: Pursue state-of-the-art technical implementation, focus on performance and user experience
- Cautious deletion: Fully understand code purpose and dependencies before deletion
- Ensure all checks pass before PR (type checking, testing, code standards)

## Architecture Design

### Service Architecture Overview
- **Main Backend**: Go + Gin (8080) - Contains 60+ service modules
- **5 Independent Microservices**: Courier(8002), Write(8001), Admin(8003), OCR(8004), Gateway(8000)
- **Frontend**: Next.js 14 + TypeScript (3000)
- **Database**: PostgreSQL 15 (required) + Redis caching
- **Real-time**: WebSocket + Event-driven architecture

### 📚 Complete Architecture Documentation
For detailed architecture information, see: [docs/product/structure/](docs/product/structure/)
- Complete service catalog (60+ modules), diagrams, and business overview

## Core Business Systems

### Credit System (✅ Complete)
- **Activity System**: Smart scheduler with multiple activity types
- **Expiration System**: Tiered rules supporting 12 credit types 
- **Transfer System**: Secure transfers with role-based limits
- **Testing**: Use scripts in `./backend/scripts/test-credit-*.sh`

### 4-Level Courier System (Core Architecture)

**Hierarchy**: L4 (City) → L3 (School) → L2 (District) → L1 (Building)

**Core Features**:
- Smart assignment with location-based load balancing
- QR scanning workflow: Collected→InTransit→Delivered  
- Performance-based promotion and gamification
- Real-time WebSocket tracking
- Batch code generation (L3/L4 only)

**Key Components**: `services/courier-service/` and `/courier` frontend pages

### Database & Project Structure
- **Entities**: User, Letter, Courier, Museum, Credit System (24 tables), OP Code
- **ORM**: GORM with PostgreSQL 15 (SQLite not supported)
- **Structure**: See [Architecture Documentation](docs/product/structure/PRODUCT_ARCHITECTURE.md#💾-数据架构) for details

## Project Structure
- **Backend**: main.go, internal/{config,handlers,middleware,models,services}/
- **Frontend**: src/{app,components,hooks,lib,stores,types}/
- **Services**: courier-service/, write-service/, admin-service/, ocr-service/
- **Shared**: shared/go/pkg/ (shared Go modules)
- **Scripts**: startup/, scripts/, test-kimi/
- **Docs**: docs/ (product requirements and technical documentation)
- **Architecture Docs**: docs/product/structure/ (comprehensive architecture documentation)

## Environment Setup

### PostgreSQL (Required)
```bash
# Start database
brew services start postgresql  # macOS
sudo systemctl start postgresql # Linux

# Setup database
createdb openpenpal
export DATABASE_URL="postgres://$(whoami):password@localhost:5432/openpenpal"
export DB_TYPE="postgres"

# Database migration
cd backend && go run main.go migrate
```
**Note**: macOS uses system username (`whoami`), Linux may need 'postgres'

### Test Accounts
- admin/Admin123! (super_admin)
- alice/Secret123! (student) - Updated password meets security requirements
- courier_level[1-4]/Secret123! (L1-L4 courier) - Updated passwords meet security requirements

### Common Issue Troubleshooting
- Port conflicts: `./startup/force-cleanup.sh`
- Permission issues: Check middleware configuration
- Database: Ensure PostgreSQL is running
- Authentication: Frontend must query database, prohibit hardcoding
- Password reset: `cd backend && go run cmd/admin/reset_passwords.go -user=username -password=newpass`
- React Hooks errors: Fixed conditional hook calls, ensure consistent component rendering

## Development Principles and Standards

### SOTA Architecture Principles
1. Clear microservice separation
2. Shared libraries in `/shared/go/pkg/`
3. 4-level RBAC permission control
4. WebSocket real-time communication
5. Multi-layer testing strategy

### Git Version Management
- `main`: Only for production environment
- Feature branches: `feature/description`
- Commit format: `feat/fix/docs: message`
- **Think before action**: Deep analysis of problems before implementing solutions
- **Cautious deletion**: Fully understand code purpose and dependencies before deletion

### Configuration Management
- Backend configuration: `internal/config/config.go`
- Frontend configuration: `src/lib/api.ts`
- Use environment variables, prohibit hardcoding

### Development Standards
- Go: gofmt formatting
- TypeScript: ESLint + strict mode
- Database: Consistent GORM, snake_case JSON fields
- API: Unified response format
- File naming: snake_case.go, PascalCase.tsx, kebab-case.ts
- Field naming: Backend uses snake_case, frontend matches exactly (no camelCase conversion)

## Testing and Validation

### System Testing
- **API Tests**: `./scripts/test-apis.sh` and `./test-kimi/run_tests.sh`
- **Courier Tests**: `./startup/tests/test-permissions.sh`
- **Service Tests**: Each service has its own `test_apis.sh`

## OP Code & Barcode Systems

### OP Code System
- **Format**: AABBCC (6-digit) - School + Area + Location
- **Example**: PK5F3D = Peking University Building 5 Room 303
- **Integration**: 4-level courier system with privacy controls

### FSD Barcode System  
- **Lifecycle**: unactivated → bound → in_transit → delivered
- **Integration**: LetterCode ↔ Envelope ↔ OP Code binding
- **Features**: 8-digit tracking with courier validation

**📖 Details**: See [Architecture Documentation](docs/product/structure/PRODUCT_ARCHITECTURE.md)


## SOTA Enhancements (State-of-the-Art)

### Performance Optimizations
- **React Tools**: Smart memoization, virtual scrolling, performance monitoring
- **API Client**: Circuit breaker, request deduplication, smart caching  
- **Error Handling**: Enhanced error boundaries and recovery systems
- **Auth System**: CSRF protection, token rotation, secure storage

**Location**: `frontend/src/lib/utils/` and `frontend/src/components/`

---

**OpenPenPal** is a modern campus handwritten letter platform using microservice architecture, integrating advanced 4-level courier system, OP Code encoding, barcode tracking and innovative features.

*Follow "Think before action" and "SOTA principles" for optimal development.*