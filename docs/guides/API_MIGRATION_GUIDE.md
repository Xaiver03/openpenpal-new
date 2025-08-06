# API Migration Guide - OpenPenPal

## Overview

This guide helps migrate frontend code to use the corrected API endpoints that align with the backend implementation.

## 1. Letter Service Migration

### Old API Calls → New API Calls

```typescript
// ❌ OLD: Using drafts endpoint
await apiClient.post('/letters/drafts', data)
// ✅ NEW: Use letters endpoint directly
await apiClient.post('/letters/', data)

// ❌ OLD: Update draft
await apiClient.put(`/letters/drafts/${id}`, data)
// ✅ NEW: Update letter
await apiClient.put(`/letters/${id}`, data)

// ❌ OLD: Get drafts
await apiClient.get('/letters/drafts')
// ✅ NEW: Get letters with status filter
await apiClient.get('/letters/?status=draft')

// ❌ OLD: Publish letter (not implemented)
await apiClient.post(`/letters/${id}/publish`)
// ✅ NEW: Update letter status
await apiClient.put(`/letters/${id}`, { status: 'published' })

// ❌ OLD: Reply with wrong endpoint
await apiClient.post(`/letters/read/${code}/reply`, data)
// ✅ NEW: Use replies endpoint
await apiClient.post('/letters/replies', {
  parent_letter_id: originalLetterId,
  ...data
})
```

### Removed Endpoints (Not Implemented in Backend)

These endpoints should be removed from frontend code:
- `/api/v1/letters/:id/like` - Like functionality not implemented
- `/api/v1/letters/:id/share` - Share functionality not implemented
- `/api/v1/letters/templates` - Templates not implemented
- `/api/v1/letters/search` - Use list with search param instead
- `/api/v1/letters/popular` - Use list with sort_by param
- `/api/v1/letters/recommended` - Not implemented
- `/api/v1/letters/batch` - Batch operations not implemented
- `/api/v1/letters/export` - Export not implemented
- `/api/v1/letters/auto-save` - Use regular update instead

## 2. Museum Service Migration

### Import Changes

```typescript
// ❌ OLD: Import from non-existent file
import { MuseumEntry } from '../../types/museum' // This file didn't exist

// ✅ NEW: Import from new types file
import { MuseumEntry } from '../../types/museum' // Now properly defined
```

### API Call Updates

```typescript
// ❌ OLD: Museum endpoints in letter service
await letterService.contributeToMuseum(data)

// ✅ NEW: Use museum service
await museumService.submitToMuseum(data)
```

### Removed Museum Endpoints

These endpoints are not implemented in backend:
- `/api/v1/museum/popular` - Use entries with sort
- `/api/v1/museum/exhibitions/:id` - Individual exhibition not implemented
- `/api/v1/museum/tags` - Tags included in entries
- `/api/v1/museum/entries/:id/interact` - Interactions not implemented
- `/api/v1/museum/entries/:id/react` - Reactions not implemented
- `/api/v1/museum/entries/:id/withdraw` - Withdrawal not implemented
- `/api/v1/museum/my-submissions` - Use entries with user filter

## 3. Response Format Standardization

### Old Response Format
```typescript
interface OldResponse {
  success: boolean
  data: any
  message: string
}
```

### New Standardized Format
```typescript
interface NewResponse {
  code: number        // 0 for success, error code otherwise
  message: string     // Success or error message
  data: any          // Response data
  timestamp: string  // ISO timestamp
}
```

### Response Handler Update

```typescript
// Update api-client.ts to handle both formats during migration
const handleResponse = async (response: Response) => {
  const data = await response.json()
  
  // Handle new format
  if ('code' in data) {
    if (data.code === 0) {
      return { success: true, data: data.data, message: data.message }
    } else {
      throw new Error(data.message)
    }
  }
  
  // Handle old format (for backward compatibility)
  if ('success' in data) {
    return data
  }
  
  // Assume success if neither format
  return { success: true, data, message: 'Success' }
}
```

## 4. Authentication Updates

### Headers Standardization

```typescript
// Ensure all API calls use consistent auth headers
const headers = {
  'Authorization': `Bearer ${token}`,
  'Content-Type': 'application/json',
}
```

## 5. Component Updates

### Letter Writing Component

```typescript
// OLD: Using draft-specific logic
const saveDraft = async () => {
  await letterService.createDraft(data)
}

// NEW: Using unified letter creation
const saveLetter = async () => {
  await letterService.createLetter({
    ...data,
    status: 'draft' // Explicitly set status
  })
}
```

### Museum Submission Component

```typescript
// OLD: Using letter service for museum
const submitToMuseum = async () => {
  await letterService.contributeToMuseum(letterId)
}

// NEW: Using proper museum service
const submitToMuseum = async () => {
  await museumService.submitToMuseum({
    letter_id: letterId,
    display_preference: 'anonymous',
    submission_reason: 'Share my story'
  })
}
```

## 6. Testing Checklist

After migration, test these key flows:

1. **Letter Creation Flow**
   - [ ] Create new letter
   - [ ] Save as draft
   - [ ] Update draft
   - [ ] Publish letter
   - [ ] Generate QR code

2. **Letter Reading Flow**
   - [ ] Read by QR code
   - [ ] Mark as read
   - [ ] Create reply
   - [ ] View thread

3. **Museum Flow**
   - [ ] View museum entries
   - [ ] Submit to museum
   - [ ] View exhibitions
   - [ ] Admin approval (if admin)

4. **Error Handling**
   - [ ] Network errors show properly
   - [ ] Auth errors redirect to login
   - [ ] Validation errors display correctly

## 7. Gradual Migration Strategy

1. **Phase 1**: Update type definitions
   - Add museum types
   - Update letter types

2. **Phase 2**: Create new service files
   - Create letter-service-fixed.ts
   - Create museum-fixed.ts
   - Keep old files for reference

3. **Phase 3**: Update components
   - Start with less critical components
   - Update imports gradually
   - Test each component

4. **Phase 4**: Remove old code
   - Delete deprecated service files
   - Remove unused API calls
   - Clean up imports

## 8. Common Pitfalls

1. **Don't forget to update imports** - Many files may import from old services
2. **Check error handling** - New format may break existing error handlers
3. **Update tests** - API mocks need to match new endpoints
4. **Check loading states** - Response format changes may affect loading logic
5. **Verify auth flow** - Ensure tokens are passed correctly

## 9. Need Help?

If you encounter issues during migration:
1. Check backend route definitions in `backend/main.go`
2. Verify response format in browser DevTools
3. Check server logs for unhandled routes
4. Refer to API_CONSISTENCY_REPORT.md for detailed analysis