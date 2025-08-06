-- OpenPenPal PostgreSQL 设置脚本

-- 创建用户（如果不存在）
DO
$do$
BEGIN
   IF NOT EXISTS (
      SELECT FROM pg_catalog.pg_user
      WHERE  usename = 'openpenpal') THEN
      CREATE USER openpenpal WITH PASSWORD 'openpenpal123';
   END IF;
END
$do$;

-- 创建数据库（如果不存在）
SELECT 'CREATE DATABASE openpenpal OWNER openpenpal'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'openpenpal')\gexec

-- 授予权限
GRANT ALL PRIVILEGES ON DATABASE openpenpal TO openpenpal;

-- 连接到 openpenpal 数据库并设置权限
\c openpenpal

-- 授予 schema 权限
GRANT ALL ON SCHEMA public TO openpenpal;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO openpenpal;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO openpenpal;