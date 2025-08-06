# Implementation Guide - Missing API Endpoints

## Overview

This guide details how to integrate all the missing API endpoints that have been implemented to fix the frontend-backend inconsistencies.

## 1. Museum API Implementation

### Files Created:
- `backend/internal/handlers/museum_handler_extended.go` - Extended museum handler with all missing endpoints
- `backend/internal/services/museum_service_extended.go` - Extended museum service with business logic
- `backend/internal/models/museum_extended.go` - Additional models for museum features

### New Endpoints Implemented:

#### Public Endpoints:
```
GET  /api/v1/museum/popular              - Get popular museum entries
GET  /api/v1/museum/exhibitions/:id      - Get exhibition details
GET  /api/v1/museum/tags                 - Get popular tags
```

#### Protected Endpoints:
```
POST   /api/v1/museum/entries/:id/interact   - Record interaction (view, like, bookmark, share)
POST   /api/v1/museum/entries/:id/react      - Add reaction with comment
DELETE /api/v1/museum/entries/:id/withdraw   - Withdraw museum entry
GET    /api/v1/museum/my-submissions         - Get user's submissions
```

#### Admin Endpoints:
```
POST   /api/v1/admin/museum/entries/:id/moderate - Moderate museum entry
GET    /api/v1/admin/museum/entries/pending      - Get pending entries
POST   /api/v1/admin/museum/exhibitions          - Create exhibition
PUT    /api/v1/admin/museum/exhibitions/:id      - Update exhibition
DELETE /api/v1/admin/museum/exhibitions/:id      - Delete exhibition
POST   /api/v1/admin/museum/refresh-stats        - Refresh statistics
GET    /api/v1/admin/museum/analytics            - Get analytics data
```

### Integration Steps:

1. **Add routes to main.go** (use content from `museum_routes_patch.go`):
```go
// In public museum group (after line 174):
museum.GET("/popular", museumHandler.GetPopularMuseumEntries)
museum.GET("/exhibitions/:id", museumHandler.GetMuseumExhibitionByID)
museum.GET("/tags", museumHandler.GetMuseumTags)

// In protected museum group (after line 266):
museum.POST("/entries/:id/interact", museumHandler.InteractWithEntry)
museum.POST("/entries/:id/react", museumHandler.ReactToEntry)
museum.DELETE("/entries/:id/withdraw", museumHandler.WithdrawMuseumEntry)
museum.GET("/my-submissions", museumHandler.GetMySubmissions)

// In admin section (create new museum admin group):
museumAdmin := admin.Group("/museum")
{
    museumAdmin.POST("/entries/:id/moderate", museumHandler.ModerateMuseumEntry)
    museumAdmin.GET("/entries/pending", museumHandler.GetPendingMuseumEntries)
    museumAdmin.POST("/exhibitions", museumHandler.CreateMuseumExhibition)
    museumAdmin.PUT("/exhibitions/:id", museumHandler.UpdateMuseumExhibition)
    museumAdmin.DELETE("/exhibitions/:id", museumHandler.DeleteMuseumExhibition)
    museumAdmin.POST("/refresh-stats", museumHandler.RefreshMuseumStats)
    museumAdmin.GET("/analytics", museumHandler.GetMuseumAnalytics)
}
```

2. **Merge extended handler methods** into `museum_handler.go`
3. **Merge extended service methods** into `museum_service.go`
4. **Add new models** to database migrations

## 2. Letter API Implementation

### Files Created:
- `backend/internal/handlers/letter_handler_extended.go` - Extended letter handler
- `backend/internal/services/letter_service_extended.go` - Extended letter service
- `backend/internal/config/migration_extended.go` - Database migrations for new features

### New Endpoints Implemented:

```
GET    /api/v1/letters/drafts                  - Get draft list
POST   /api/v1/letters/:id/publish             - Publish letter
POST   /api/v1/letters/:id/like                - Like letter
POST   /api/v1/letters/:id/share               - Share letter
GET    /api/v1/letters/templates               - Get templates
GET    /api/v1/letters/templates/:id           - Get template by ID
POST   /api/v1/letters/search                  - Search letters
GET    /api/v1/letters/popular                 - Get popular letters
GET    /api/v1/letters/recommended             - Get recommended letters
POST   /api/v1/letters/batch                   - Batch operations
POST   /api/v1/letters/export                  - Export letters
POST   /api/v1/letters/auto-save               - Auto-save draft
POST   /api/v1/letters/writing-suggestions     - Get writing suggestions
```

### Integration Steps:

1. **Add routes to main.go** (use content from `letter_routes_patch.go`):
```go
// In protected letters group:
letters.GET("/drafts", letterHandler.GetDrafts)
letters.POST("/:id/publish", letterHandler.PublishLetter)
letters.POST("/:id/like", letterHandler.LikeLetter)
letters.POST("/:id/share", letterHandler.ShareLetter)
letters.GET("/templates", letterHandler.GetLetterTemplates)
letters.GET("/templates/:id", letterHandler.GetLetterTemplateByID)
letters.POST("/search", letterHandler.SearchLetters)
letters.GET("/popular", letterHandler.GetPopularLetters)
letters.GET("/recommended", letterHandler.GetRecommendedLetters)
letters.POST("/batch", letterHandler.BatchOperateLetters)
letters.POST("/export", letterHandler.ExportLetters)
letters.POST("/auto-save", letterHandler.AutoSaveDraft)
letters.POST("/writing-suggestions", letterHandler.GetWritingSuggestions)
```

2. **Merge extended handler methods** into `letter_handler.go`
3. **Merge extended service methods** into `letter_service.go`
4. **Run extended migrations** to add new tables

## 3. Database Migrations

### New Tables Required:
```sql
-- Museum extensions
CREATE TABLE museum_tags (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    category VARCHAR(50) DEFAULT 'general',
    usage_count INT DEFAULT 0,
    created_at TIMESTAMP
);

CREATE TABLE museum_interactions (
    id VARCHAR(36) PRIMARY KEY,
    entry_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    type VARCHAR(20) NOT NULL,
    created_at TIMESTAMP,
    INDEX idx_entry_id (entry_id),
    INDEX idx_user_id (user_id)
);

CREATE TABLE museum_reactions (
    id VARCHAR(36) PRIMARY KEY,
    entry_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    reaction_type VARCHAR(20) NOT NULL,
    comment TEXT,
    created_at TIMESTAMP,
    INDEX idx_entry_id (entry_id),
    INDEX idx_user_id (user_id)
);

CREATE TABLE museum_submissions (
    id VARCHAR(36) PRIMARY KEY,
    letter_id VARCHAR(36) NOT NULL,
    submitted_by VARCHAR(36) NOT NULL,
    display_preference VARCHAR(20) DEFAULT 'anonymous',
    pen_name VARCHAR(100),
    submission_reason TEXT,
    curator_notes TEXT,
    status VARCHAR(20) DEFAULT 'pending',
    submitted_at TIMESTAMP,
    reviewed_at TIMESTAMP,
    reviewed_by VARCHAR(36)
);

-- Letter extensions
CREATE TABLE letter_likes (
    id VARCHAR(36) PRIMARY KEY,
    letter_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    created_at TIMESTAMP,
    UNIQUE KEY unique_user_letter (user_id, letter_id)
);

CREATE TABLE letter_shares (
    id VARCHAR(36) PRIMARY KEY,
    letter_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    platform VARCHAR(50),
    created_at TIMESTAMP
);

CREATE TABLE letter_templates (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    category VARCHAR(50),
    preview_image VARCHAR(500),
    content_template TEXT,
    style_config JSON,
    is_premium BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    usage_count INT DEFAULT 0,
    rating DECIMAL(3,2),
    created_by VARCHAR(36),
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

### Migration Steps:

1. **Update database.go** to include new models in AutoMigrate
2. **Run migration command**: `go run cmd/migrate/main.go`
3. **Verify tables created**: Check database for new tables

## 4. Response Format Standardization

### Current Mixed Formats:
```go
// Old format
{
    "success": true,
    "data": {},
    "message": "Success"
}

// New standardized format
{
    "code": 0,
    "message": "Success",
    "data": {},
    "timestamp": "2024-01-01T00:00:00Z"
}
```

### Migration Strategy:

1. **Create response wrapper** that handles both formats
2. **Update handlers gradually** to use new format
3. **Update frontend API client** to handle both during transition

## 5. Path Alignment

### Issues to Fix:
- Frontend uses `/letters/drafts`, backend uses `/letters/?status=draft`
- Museum endpoints in letter service need to be moved
- Inconsistent parameter names

### Solution:
1. **Option A**: Update frontend to use correct paths (recommended)
2. **Option B**: Add route aliases in backend for compatibility

## 6. Testing Strategy

### API Testing:
```bash
# Test new museum endpoints
curl -X GET "http://localhost:8080/api/v1/museum/popular"
curl -X POST "http://localhost:8080/api/v1/museum/entries/123/interact" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"type": "like"}'

# Test new letter endpoints
curl -X GET "http://localhost:8080/api/v1/letters/drafts" \
  -H "Authorization: Bearer $TOKEN"
curl -X POST "http://localhost:8080/api/v1/letters/123/publish" \
  -H "Authorization: Bearer $TOKEN"
```

### Integration Testing:
1. Start backend with new routes
2. Update frontend to use new endpoints
3. Test each flow end-to-end

## 7. Deployment Checklist

- [ ] Merge all extended handlers into main handler files
- [ ] Merge all extended services into main service files
- [ ] Add all new routes to main.go
- [ ] Run database migrations
- [ ] Update frontend API services to use correct endpoints
- [ ] Test all new endpoints
- [ ] Update API documentation
- [ ] Deploy backend changes
- [ ] Deploy frontend changes

## 8. Rollback Plan

If issues occur:
1. Revert to previous backend version
2. Frontend will show errors but won't crash
3. Fix issues and redeploy

## Notes

- All new endpoints follow existing authentication and authorization patterns
- Error handling is consistent with existing code
- Database queries are optimized with proper indexes
- All endpoints support pagination where applicable