# PostgreSQL Index Optimization Guide

## Overview

This guide documents the PostgreSQL index optimization strategy for the OpenPenPal backend system. Proper indexing is crucial for query performance, especially as data volume grows.

## Index Strategy

### 1. Critical Performance Indexes

These indexes target the most frequent query patterns in the application:

#### Users Table
- **idx_users_school_role_active**: Composite index for school-based role queries
  - Columns: `school_code, role, is_active`
  - Partial: `WHERE is_active = true`
  - Use case: Finding active users by school and role

- **idx_users_created_at_desc**: Descending index for user listings
  - Columns: `created_at DESC`
  - Use case: Recent user registrations, paginated user lists

#### Letters Table
- **idx_letters_user_status_created**: Covering index for user letter queries
  - Columns: `user_id, status, created_at DESC`
  - Include: `title, style` (covering index)
  - Use case: User's letter list with minimal table lookups

- **idx_letters_recipient_status**: Recipient letter lookups
  - Columns: `recipient_op_code, status`
  - Partial: `WHERE status IN ('published', 'delivered')`
  - Use case: Finding letters for a specific recipient

- **idx_letters_deleted_at**: Soft delete optimization
  - Columns: `deleted_at`
  - Partial: `WHERE deleted_at IS NULL`
  - Use case: Filtering out soft-deleted records efficiently

#### Courier Tasks Table
- **idx_courier_tasks_courier_status**: Courier task list optimization
  - Columns: `courier_id, status, priority DESC, created_at`
  - Use case: Courier's task queue with priority sorting

- **idx_courier_tasks_pickup_delivery**: Geographic task queries
  - Columns: `pickup_op_code, delivery_op_code`
  - Partial: `WHERE status NOT IN ('completed', 'cancelled')`
  - Use case: Active task geographic distribution

### 2. Full-Text Search Indexes

GIN indexes for text search capabilities:

- **idx_letters_fulltext**: Letter content search
  ```sql
  CREATE INDEX idx_letters_fulltext ON letters 
  USING gin(to_tsvector('simple', title || ' ' || content));
  ```

- **idx_museum_items_fulltext**: Museum item search
  ```sql
  CREATE INDEX idx_museum_items_fulltext ON museum_items 
  USING gin(to_tsvector('simple', title || ' ' || description));
  ```

### 3. Foreign Key Indexes

All foreign key columns should have indexes for JOIN performance:
- `user_id` on all user-related tables
- `letter_id` on letter-related tables
- `courier_id` on courier tasks
- `op_code` references

## Implementation

### Using the Index Optimizer

1. **Analyze Current State**
   ```bash
   ./scripts/optimize-indexes.sh analyze
   ```

2. **Dry Run (Preview Changes)**
   ```bash
   go run cmd/tools/optimize-indexes/main.go --mode=create --dry-run --verbose
   ```

3. **Apply Optimizations**
   ```bash
   go run cmd/tools/optimize-indexes/main.go --mode=create --verbose
   ```

4. **Monitor Performance**
   ```bash
   ./scripts/optimize-indexes.sh monitor
   ```

### Migration Approach

The index optimization is implemented as a migration:

```go
// Apply migration
err := migrations.RegisterIndexMigration(db)

// Or manually
optimizer := config.NewIndexOptimizer(db, false, true)
err := optimizer.OptimizeAll()
```

### Phased Rollout

1. **Phase 1**: Critical performance indexes (immediate impact)
2. **Phase 2**: Full-text search indexes (feature enhancement)
3. **Phase 3**: Analyze and create additional indexes based on usage

## Best Practices

### 1. Index Design Principles

- **Selectivity**: Index columns with high selectivity (many distinct values)
- **Multi-column**: Order matters - most selective column first
- **Covering indexes**: Include frequently accessed columns to avoid table lookups
- **Partial indexes**: Use WHERE clauses for subset optimization

### 2. Index Maintenance

- **Regular REINDEX**: For heavily updated tables
  ```sql
  REINDEX TABLE CONCURRENTLY letters;
  ```

- **Monitor bloat**: Check index bloat regularly
  ```sql
  SELECT indexname, pg_size_pretty(pg_relation_size(indexrelid)) 
  FROM pg_stat_user_indexes 
  WHERE schemaname = 'public';
  ```

- **Update statistics**: Keep table statistics current
  ```sql
  ANALYZE letters;
  ```

### 3. Performance Monitoring

Key metrics to track:

1. **Index usage ratio**
   ```sql
   SELECT 
       tablename,
       100 * idx_scan / (seq_scan + idx_scan) as index_usage_percent
   FROM pg_stat_user_tables
   WHERE seq_scan + idx_scan > 0;
   ```

2. **Unused indexes**
   ```sql
   SELECT indexname 
   FROM pg_stat_user_indexes 
   WHERE idx_scan = 0 
   AND indexrelname NOT LIKE '%_pkey';
   ```

3. **Query performance**
   - Enable `pg_stat_statements`
   - Monitor slow query log
   - Use EXPLAIN ANALYZE

## Common Query Patterns

### 1. User Queries
```sql
-- Optimized by idx_users_school_role_active
SELECT * FROM users 
WHERE school_code = 'PKU001' 
  AND role = 'courier_level1' 
  AND is_active = true;
```

### 2. Letter Listings
```sql
-- Optimized by idx_letters_user_status_created (covering)
SELECT id, title, style, created_at 
FROM letters 
WHERE user_id = ? 
  AND status = 'published' 
ORDER BY created_at DESC 
LIMIT 20;
```

### 3. Courier Task Assignment
```sql
-- Optimized by idx_courier_tasks_pickup_delivery
SELECT * FROM courier_tasks 
WHERE pickup_op_code LIKE 'PK%' 
  AND status = 'pending' 
ORDER BY priority DESC, created_at ASC;
```

### 4. Full-Text Search
```sql
-- Optimized by GIN index
SELECT * FROM letters 
WHERE to_tsvector('simple', title || ' ' || content) 
  @@ to_tsquery('simple', 'campus & love');
```

## Troubleshooting

### High Sequential Scans
- Check for missing indexes on WHERE clause columns
- Consider partial indexes for common filters

### Slow JOINs
- Ensure foreign key columns are indexed
- Check JOIN order and statistics

### Index Not Used
- Update table statistics: `ANALYZE table_name;`
- Check selectivity - index may not be beneficial
- Verify query planner settings

### Index Bloat
- Regular REINDEX for heavily updated tables
- Consider partitioning for very large tables
- Monitor with pg_repack or similar tools

## Monitoring Scripts

### Check Index Health
```bash
# Overall index analysis
./scripts/optimize-indexes.sh analyze

# Missing indexes
./scripts/optimize-indexes.sh missing

# Unused indexes
./scripts/optimize-indexes.sh unused

# Real-time monitoring
./scripts/optimize-indexes.sh monitor
```

### Generate Reports
```bash
# Comprehensive report
./scripts/optimize-indexes.sh report

# Specific table analysis
go run cmd/tools/optimize-indexes/main.go --mode=analyze --table=letters
```

## Performance Impact

Expected improvements after optimization:

1. **Letter queries**: 50-80% faster for user letter lists
2. **Courier task assignment**: 60-90% improvement for geographic queries
3. **Search operations**: 10-100x faster with full-text indexes
4. **Soft delete filtering**: 40-60% improvement on large tables

## Conclusion

Proper index optimization is an ongoing process. Regular monitoring, analysis, and adjustment based on actual usage patterns will ensure optimal database performance as the application scales.