-- Fix Table Ownership Issues for OpenPenPal Database
-- This script changes ownership of all tables to the application user
-- Run as superuser (postgres or rocalight)

-- Change ownership of all tables in public schema
DO $$
DECLARE
    tbl RECORD;
    app_user TEXT := 'openpenpal_user';
BEGIN
    FOR tbl IN 
        SELECT tablename 
        FROM pg_tables 
        WHERE schemaname = 'public' 
        AND tableowner != app_user
    LOOP
        EXECUTE format('ALTER TABLE public.%I OWNER TO %I', tbl.tablename, app_user);
        RAISE NOTICE 'Changed ownership of table % to %', tbl.tablename, app_user;
    END LOOP;
END;
$$;

-- Change ownership of all sequences
DO $$
DECLARE
    seq RECORD;
    app_user TEXT := 'openpenpal_user';
BEGIN
    FOR seq IN 
        SELECT sequence_name 
        FROM information_schema.sequences 
        WHERE sequence_schema = 'public'
    LOOP
        EXECUTE format('ALTER SEQUENCE public.%I OWNER TO %I', seq.sequence_name, app_user);
        RAISE NOTICE 'Changed ownership of sequence % to %', seq.sequence_name, app_user;
    END LOOP;
END;
$$;

-- Change ownership of all views
DO $$
DECLARE
    v RECORD;
    app_user TEXT := 'openpenpal_user';
BEGIN
    FOR v IN 
        SELECT viewname 
        FROM pg_views 
        WHERE schemaname = 'public'
    LOOP
        EXECUTE format('ALTER VIEW public.%I OWNER TO %I', v.viewname, app_user);
        RAISE NOTICE 'Changed ownership of view % to %', v.viewname, app_user;
    END LOOP;
END;
$$;

-- Grant all privileges on schema to app user
GRANT ALL ON SCHEMA public TO openpenpal_user;

-- Grant all privileges on all tables
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO openpenpal_user;

-- Grant all privileges on all sequences
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO openpenpal_user;

-- Make sure app user can create objects in the future
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO openpenpal_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO openpenpal_user;

-- Verify ownership changes
SELECT 
    'Tables owned by ' || tableowner as ownership,
    COUNT(*) as count
FROM pg_tables 
WHERE schemaname = 'public'
GROUP BY tableowner;

SELECT 
    'Sequences owned by ' || sequence_schema as ownership,
    COUNT(*) as count
FROM information_schema.sequences
WHERE sequence_schema = 'public'
GROUP BY sequence_schema;