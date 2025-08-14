# CLAUDE.md

Guidance for Claude Code (claude.ai/code) when working with this repository.

## Project Overview

OpenPenPal - Campus handwritten letter platform with digital tracking. Microservices architecture: Next.js frontend + Go/Python/Java backend services.

## Development Commands

```bash
# Quick Start
./startup/quick-start.sh demo --auto-open      # Demo mode (recommended)
./startup/quick-start.sh development --auto-open # All services
./startup/check-status.sh                       # Service status
./startup/stop-all.sh                          # Stop all
./startup/force-cleanup.sh                     # Force cleanup

# Frontend (cd frontend)
npm run dev/build/lint/lint:fix/type-check/test/test:e2e

# Backend (cd backend)
go run main.go / go mod tidy

# Testing
./scripts/test-apis.sh              # API tests
./test-kimi/run_tests.sh           # Integration
./startup/tests/test-permissions.sh # Permissions
```

## Architecture

### Services & Ports
- Frontend: Next.js 14 + TypeScript (3000)
- Backend: Go + Gin (8080)
- Write: Python/FastAPI (8001)
- Courier: Go (8002)
- Admin: Java/Spring Boot (8003)
- OCR: Python (8004)
- Gateway: Go (8000)

### Key Components
- Auth: JWT + roles (admin/courier/senior_courier/coordinator)
- Database: PostgreSQL (required)
- Real-time: WebSocket
- Storage: Local uploads + QR codes
- Shared: `/shared/go/pkg/` modules

### Critical: Four-Level Courier System (CORE)

**Hierarchy**:
1. **L4 城市总代**: City-wide control, creates L3 (Zone: BEIJING)
2. **L3 校级信使**: School distribution, creates L2 (Zone: BJDX)
3. **L2 片区信使**: Zone management, creates L1 (Zone: District)
4. **L1 楼栋信使**: Direct delivery (Zone: BJDX-A-101)

**Features**:
- Smart assignment (location + load balancing)
- QR scan workflow (collected→in_transit→delivered)
- Performance-based promotion
- Real-time WebSocket tracking
- Gamification + leaderboards

**Batch Generation Powers (L3/L4 CRITICAL)**:
- **L3 校级信使**: School-level batch generation, manages campus codes (AABBCC format)
- **L4 城市总代**: City-wide batch generation, cross-school operations
- **Signal Code System**: Complete batch generation via `GenerateCodeBatch` API
- **Permission Matrix**: Hierarchical inheritance (L4 inherits all L3 powers)
- **Hidden UI**: Batch functions exist but UI entry points are not obvious
- **Core APIs**: POST `/api/signal-codes/batch`, POST `/api/signal-codes/assign`

**Key Files**:
- `services/courier-service/internal/services/hierarchy.go`
- `frontend/src/components/courier/CourierPermissionGuard.tsx`
- `services/courier-service/internal/models/courier.go`
- **Batch Generation System (L3/L4)**:
  - `services/courier-service/internal/services/signal_code_service.go` (BatchGenerate API)
  - `services/courier-service/internal/handlers/signal_code_handler.go` (Batch endpoints)
  - `services/courier-service/internal/services/postal_management.go` (L3/L4 permissions)
  - `services/courier-service/internal/models/signal_code.go` (Batch models)

### Database
Entities: User, Letter, Courier, Museum. GORM + PostgreSQL (required, no SQLite).

## File Structure

**Backend**: main.go, internal/{config,handlers,middleware,models,services}/
**Frontend**: src/{app,components,hooks,lib,stores,types}/

## Environment Setup

### PostgreSQL (Required)
```bash
# Start DB
brew services start postgresql  # macOS
sudo systemctl start postgresql # Linux

# Setup
createdb openpenpal
export DATABASE_URL="postgres://$(whoami):password@localhost:5432/openpenpal"
export DB_TYPE="postgres"

# Migrate
cd backend && go run main.go migrate
```

**Note**: macOS uses system username (`whoami`), Linux may need 'postgres'

### Test Accounts
- admin/admin123 (super_admin)
- alice/secret123 (student) - Updated password due to 8+ char requirement
- courier_level[1-4]/secret123 (L1-L4 courier) - Updated passwords

### Common Issues
- Ports: `./startup/force-cleanup.sh`
- Permissions: Check middleware
- DB: Ensure PostgreSQL running
- Auth: Frontend must query DB, no hardcoding
- Password Reset: Use `cd backend && go run cmd/admin/reset_passwords.go -user=username -password=newpass`
- React Hooks Error: Fixed conditional hook calls, ensure consistent component rendering

## Development Principles

### Architecture (SOTA)
1. Microservices with clean separation
2. Shared libraries in `/shared/go/pkg/`
3. 4-level RBAC
4. WebSocket real-time
5. Multi-layer testing

### Git
- `main`: Production only
- Features: `feature/description`
- Commits: `feat/fix/docs: message`

### Configuration
- Backend: `internal/config/config.go`
- Frontend: `src/lib/api.ts`
- Use env vars, no hardcoding

### Standards
- Go: gofmt
- TS: ESLint + strict
- DB: Consistent GORM, snake_case JSON fields
- API: Shared response format
- Files: snake_case.go, PascalCase.tsx, kebab-case.ts
- Field Naming: Backend uses snake_case, Frontend matches exactly (no camelCase conversion)

### Courier System Verification

**Key Files**: services/courier-service/, role_compatibility.go, CourierPermissionGuard.tsx

**Testing**:
```bash
./startup/tests/test-permissions.sh
cd services/courier-service && ./test_apis.sh
curl -X GET "http://localhost:8002/api/v1/courier/hierarchy/level/2"

# Test L3/L4 Batch Generation Powers
curl -X POST "http://localhost:8002/api/signal-codes/batch" \
  -H "Authorization: Bearer $L3_TOKEN" \
  -d '{"batch_no":"B001","school_id":"BJDX","quantity":100}'
  
curl -X POST "http://localhost:8002/api/signal-codes/assign" \
  -H "Authorization: Bearer $L4_TOKEN" \
  -d '{"codes":["PK5F3D","PK5F3E"],"assignee_id":"courier123"}'
```

**Hierarchy Rules**:
- L4→L3→L2→L1 creation chain
- Task flow: Available→Accepted→Collected→InTransit→Delivered
- Zone-based permissions
- Performance promotions

**Endpoints** (8002): /hierarchy, /tasks, /scan, /leaderboard

## OP Code System (CRITICAL)

**Format**: AABBCC (6 digits)
- AA: School (PK=北大, QH=清华, BD=北交大)
- BB: Area (5F=5号楼, 3D=3食堂, 2G=2号门)
- CC: Point (3D=303室, 1A=1层A区, 12=12号桌)

Example: PK5F3D = 北大5号楼303室

**Features**:
- Unified 6-digit encoding
- Privacy control (PK5F** hides last 2)
- Hierarchical permissions
- Reuses SignalCode infrastructure

**Models**: SignalCode (repurposed), Letter (+OPCode fields), Courier (+ManagedOPCodePrefix)

### OP Code API & Services

**Services**: opcode_service.go (Apply/Assign/Search/Validate/Stats/Migrate)
**Handlers**: opcode_handler.go (Privacy-aware endpoints)

**Endpoints**:
```bash
# Public
GET /api/v1/opcode/:code
GET /api/v1/opcode/validate

# Protected  
POST /api/v1/opcode/apply
GET /api/v1/opcode/search
GET /api/v1/opcode/stats/:school_code

# Admin
POST /api/v1/opcode/admin/applications/:id/review
```

**Privacy**: Full/Partial(PK5F**)/Public
**Permissions**: L1 limited, L2+ prefix access, Admin full
**Migration**: Zone→OPCode mapping (BEIJING→BJ, BJDX→BD)

**Validation**: 6 chars uppercase alphanumeric, unique, hierarchical

### OP Code Integration Status (✅ Complete)

**1. Letter Service**: RecipientOPCode/SenderOPCode fields, QR with OP data
**2. Courier Tasks**: Pickup/Delivery/CurrentOPCode, prefix permissions, geographic routing
**3. Museum**: OriginOPCode for provenance
**4. QR Enhancement**: JSON format with OP Code validation
**Architecture**: OPCode Service → Letter/Courier/Museum/Notification services
**Tables**: signal_codes (repurposed), letters, courier_tasks, museum_items (all with OP fields)

## FSD Barcode System (Enhanced LetterCode)

**Principle**: Enhanced existing LetterCode instead of creating new models

**Enhanced LetterCode Model**:
- Original fields preserved (ID, LetterID, Code, QRCodeURL, etc)
- FSD additions: Status, RecipientCode, EnvelopeID, scan tracking
- Status lifecycle: unactivated→bound→in_transit→delivered
- Lifecycle methods: IsValidTransition(), IsActive(), CanBeBound()

### FSD Service Integration

**Request Models**: BindBarcodeRequest, UpdateBarcodeStatusRequest, EnvelopeWithBarcodeResponse

**Service Methods**:
- BindBarcodeToEnvelope() - FSD 6.2
- UpdateBarcodeStatus() - FSD 6.3
- GetBarcodeStatus()
- ValidateBarcodeOperation()

**Three-Way Binding**: LetterCode ↔ Envelope ↔ OP Code
**Process**: Generate→Bind→Associate→Scan→Deliver

### FSD Courier Integration

**Enhanced Models**: ScanRequest/Response with FSD fields (barcode, OP codes, validation)

**Task Service Methods**:
- UpdateTaskStatus() - Enhanced scanning
- validateOPCodePermission() - Level-based access
- getNextAction() - Smart recommendations
- calculateEstimatedDelivery() - Time estimates

**OP Code Permissions**:
- L4: Anywhere
- L3: Same school
- L2: Same school+area
- L1: Same 4-digit prefix

### FSD Endpoints

**Letter Barcode** (8080):
- POST /api/v1/letters/barcode/bind
- PUT /api/v1/letters/barcode/:code/status
- GET /api/v1/letters/barcode/:code/status
- POST /api/v1/letters/barcode/:code/validate

**Courier Scan** (8002):
- POST /api/v1/courier/scan/:code
- GET /api/v1/courier/scan/history/:id
- POST /api/v1/courier/barcode/:code/validate-access

**Lifecycle Test**: Bind→Scan→Update→Query

### FSD Benefits & Status

**✅ ACHIEVED**:
- 8-digit barcode + lifecycle management
- OP Code integration + envelope binding
- 4-level courier validation
- Real-time tracking + smart recommendations
- Backward compatible

**🔧 ELEGANT**: Enhanced existing models, no duplication

**INTEGRATION COMPLETE**: All systems integrated with FSD compliance

# Test QR code scanning with OP Code validation  
curl -X POST "http://localhost:8080/api/v1/courier/scan" \
  -H "Authorization: Bearer $COURIER_TOKEN" \
  -d '{"qr_data":"...","current_op_code":"PK5F01"}'
```

**Integration Points**:
- ✅ Letter creation/delivery uses OP Code for addressing
- ✅ Courier task assignment based on OP Code prefixes  
- ✅ Museum entries reference OP Code locations
- ✅ QR codes contain structured OP Code data for location tracking
- ✅ Permission system validates courier access by OP Code areas
- ✅ Geographic analytics and reporting by OP Code regions

### OP Code Implementation Details

**Models**: OPCodeApplication, OPCodeRequest, OPCodeAssignRequest, OPCodeSearchRequest, OPCodeStats
**Types**: dormitory/shop/box/club, pending/approved/rejected
**Utils**: Generate/Parse/Validate/FormatOPCode
**Service**: Apply/Assign/Get/Search/Stats/ValidateAccess/Migrate
**Handlers**: User endpoints + Admin review

**Status**: ✅ Complete - Models, Service, Handler, Routes, Validation, Migration

**Test**: Use provided curl commands with proper auth tokens

## SOTA Enhancements (State-of-the-Art)

### React Optimization Utilities
- **Location**: `frontend/src/lib/utils/react-optimizer.ts`
- **Features**: Smart memoization, virtual scrolling, performance monitoring, lazy loading
- **Usage**: `useDebouncedValue`, `useThrottledCallback`, `useOptimizedState`, `smartMemo`

### Enhanced API Client
- **Location**: `frontend/src/lib/utils/enhanced-api-client.ts`  
- **Features**: Circuit breaker pattern, request deduplication, intelligent caching
- **Benefits**: Improved reliability, reduced redundant requests, better UX

### Error Handling
- **Enhanced Error Boundary**: `frontend/src/components/error-boundary/enhanced-error-boundary.tsx`
- **Performance Monitor**: `frontend/src/lib/utils/performance-monitor.ts`
- **Cache Manager**: `frontend/src/lib/utils/cache-manager.ts`

### Authentication System
- **Enhanced Provider**: `frontend/src/app/providers/auth-provider-enhanced.tsx`
- **Debug Tools**: Development-only auth debugging widget
- **Security**: CSRF protection, token rotation, secure storage

## Recent Fixes

### React Hooks Error Resolution
- **Issue**: "Rendered more hooks than during the previous render" 
- **Fix**: Consistent hook execution, proper useCallback usage, cleanup handling
- **Location**: `auth-provider-enhanced.tsx:138-152`

### TypeScript Consistency
- **Issue**: Field naming mismatch (camelCase ↔ snake_case)
- **Fix**: Updated all frontend types to match backend JSON exactly
- **Affected**: User types, Letter types, API responses, state management

### Database Connection
- **Issue**: Connection string parsing error
- **Fix**: Use `config.DatabaseName` instead of `config.DatabaseURL`
- **Location**: `backend/internal/config/database.go:45`

```