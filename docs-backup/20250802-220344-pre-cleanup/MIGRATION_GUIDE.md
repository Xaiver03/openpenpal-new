# 🛡️ Zero Breaking Changes Migration Guide

This guide provides step-by-step instructions to migrate from duplicated code to shared libraries **without breaking anything**.

## 🎯 Safety Principles

1. **Never delete original files** - Always keep backups
2. **Feature flags** - Use environment variables to switch implementations
3. **Gradual migration** - One service at a time
4. **Full rollback** - Instant revert capability
5. **Dual implementation** - Old and new code side-by-side

## 📋 Pre-Migration Checklist

- [ ] Git repository initialized
- [ ] All services tested and working
- [ ] Backup created (automatic via git)
- [ ] Environment variables documented

## 🔄 Migration Phases

### Phase 1: Shared Libraries (✅ COMPLETE)
- [x] Created shared/go/pkg/* libraries
- [x] Created shared/python/shared/* libraries  
- [x] Created unified scripts/ops.sh
- [x] Created shared Docker configurations

### Phase 2: Gradual Service Adoption (NEXT)

#### Backend Service Migration (First)
```bash
# Step 1: Test current implementation
cd /Users/rocalight/同步空间/opplc/openpenpal/backend
go test ./...

# Step 2: Add shared library import (non-breaking)
# Edit: backend/internal/handlers/letter_handler.go
# Add: import "github.com/openpenpal/shared/go/pkg/response"

# Step 3: Gradually replace response functions
# Old: sendJSONResponse(w, code, data)
# New: response.JSON(w, code, data)
# Both work side-by-side!
```

#### Courier Service Migration (Second)
```bash
# Step 1: Test current
cd /Users/rocalight/同步空间/opplc/openpenpal/services/courier-service
go test ./...

# Step 2: Add shared middleware
go get github.com/openpenpal/shared/go/pkg/middleware

# Step 3: Use shared middleware (optional)
# router.Use(middleware.AuthMiddleware)
# Old auth still works!
```

#### Python Services Migration (Third)
```bash
# Step 1: Test current
cd /Users/rocalight/同步空间/opplc/openpenpal/services/write-service
python -m pytest

# Step 2: Add shared library to requirements.txt
# Add: -e ../../shared/python

# Step 3: Import shared response
# from shared.response import APIResponse
```

### Phase 3: Docker Optimization (Last)
```bash
# Step 1: Test with new base Docker
# Use: shared/docker/base.Dockerfile
# Old Dockerfiles remain!

# Step 2: Validate all services
./scripts/ops.sh health
```

## 🚀 Quick Start Commands

### Test Current State
```bash
./scripts/ops.sh health          # Check all services
./scripts/ops.sh build backend   # Build specific service
```

### Safe Migration Commands
```bash
./scripts/ops.sh migrate backend     # Phase 2 for backend
./scripts/ops.sh status              # Check all services
```

### Rollback Commands
```bash
./scripts/ops.sh rollback backend    # Rollback specific service
git checkout main                    # Full project rollback
```

## 🔧 Environment Variables

### Feature Flags
```bash
# Enable shared libraries (default: false)
export USE_SHARED_LIBS=true

# Debug mode
export DEBUG_SHARED_LIBS=true

# Service-specific flags
export BACKEND_USE_SHARED=true
export COURIER_USE_SHARED=false
```

## 📊 Migration Progress

| Service | Status | Shared Libs | Old Code | Rollback Ready |
|---------|--------|-------------|----------|----------------|
| Backend | ⏳ Ready | ✅ Available | ✅ Preserved | ✅ Yes |
| Courier | ⏳ Ready | ✅ Available | ✅ Preserved | ✅ Yes |
| Gateway | ⏳ Ready | ✅ Available | ✅ Preserved | ✅ Yes |
| OCR | ⏳ Ready | ✅ Available | ✅ Preserved | ✅ Yes |
| Write | ⏳ Ready | ✅ Available | ✅ Preserved | ✅ Yes |

## 🚨 Safety Features

1. **Git Branch Protection**
   - All changes on separate branch
   - Main branch remains untouched
   - Easy rollback with `git checkout main`

2. **Service Isolation**
   - Each service migrates independently
   - No cross-service dependencies
   - Can rollback individual services

3. **Dual Implementation**
   - Old code continues to work
   - New code tested in parallel
   - Switch via environment variables

4. **Comprehensive Testing**
   - Health checks for all services
   - Integration tests preserved
   - Manual verification steps

## 🔄 Rollback Procedures

### Emergency Rollback
```bash
# 1. Stop all services
./scripts/ops.sh stop

# 2. Rollback to main branch
git checkout main
git reset --hard HEAD

# 3. Start services with original code
./scripts/ops.sh start
```

### Selective Rollback
```bash
# 1. Rollback specific service
./scripts/ops.sh rollback backend

# 2. Verify service works
./scripts/ops.sh health backend

# 3. Continue with other services
```

## ✅ Validation Checklist

Before proceeding to next phase:
- [ ] All services pass health checks
- [ ] Shared libraries compile without errors
- [ ] Original functionality preserved
- [ ] Rollback tested successfully
- [ ] Performance metrics unchanged

## 📞 Support

If any issues arise:
1. Check service logs: `./scripts/ops.sh logs [service]`
2. Run health check: `./scripts/ops.sh health [service]`
3. Rollback immediately: `./scripts/ops.sh rollback [service]`
4. Create GitHub issue with details