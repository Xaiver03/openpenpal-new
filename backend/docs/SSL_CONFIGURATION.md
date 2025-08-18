# PostgreSQL SSL Configuration Guide

## Overview

This guide provides comprehensive instructions for configuring SSL/TLS connections to PostgreSQL databases in the OpenPenPal backend system.

## SSL Modes

PostgreSQL supports the following SSL modes:

| Mode | Description | Security Level |
|------|-------------|----------------|
| `disable` | No SSL connection | None |
| `allow` | Try non-SSL first, then SSL | Low |
| `prefer` | Try SSL first, then non-SSL (default) | Medium |
| `require` | Require SSL, no certificate validation | Medium-High |
| `verify-ca` | Require SSL, verify server certificate is signed by trusted CA | High |
| `verify-full` | Require SSL, verify server certificate and hostname match | Highest |

## Environment-Based Defaults

The system automatically selects appropriate SSL modes based on the environment:

- **Development**: `disable` (for convenience)
- **Test**: `disable` (for speed)
- **Staging**: `prefer` (balance between security and compatibility)
- **Production**: `require` or `verify-full` (maximum security)

## Configuration

### Environment Variables

```bash
# SSL Mode
export DB_SSLMODE="require"

# SSL Certificates (for verify-ca and verify-full modes)
export DB_SSL_ROOT_CERT="/path/to/ca-cert.pem"      # CA certificate
export DB_SSL_CERT="/path/to/client-cert.pem"       # Client certificate
export DB_SSL_KEY="/path/to/client-key.pem"         # Client private key
```

### Application Configuration

The SSL configuration is automatically loaded from environment variables:

```go
// config/config.go
type Config struct {
    DBSSLMode     string // SSL mode
    DBSSLCert     string // Client certificate path
    DBSSLKey      string // Client private key path
    DBSSLRootCert string // CA certificate path
}
```

### Connection String

The system automatically builds the appropriate connection string:

```go
// Basic connection
dsn := "host=localhost port=5432 user=postgres dbname=openpenpal sslmode=require"

// With certificates
dsn := "host=localhost port=5432 user=postgres dbname=openpenpal sslmode=verify-full sslrootcert=/path/to/ca.pem sslcert=/path/to/client.pem sslkey=/path/to/client.key"
```

## Production Setup

### 1. Generate SSL Certificates

For production environments, you'll need:

1. **CA Certificate**: The certificate authority that signed the server certificate
2. **Client Certificate**: (Optional) For mutual TLS authentication
3. **Client Private Key**: (Optional) Corresponding to the client certificate

### 2. Configure PostgreSQL Server

Ensure your PostgreSQL server is configured for SSL:

```conf
# postgresql.conf
ssl = on
ssl_cert_file = 'server.crt'
ssl_key_file = 'server.key'
ssl_ca_file = 'root.crt'

# pg_hba.conf
# Require SSL for all connections
hostssl all all 0.0.0.0/0 md5
```

### 3. Deploy Certificates

Place certificates in a secure location:

```bash
# Create certificate directory
sudo mkdir -p /etc/ssl/postgresql
sudo chmod 700 /etc/ssl/postgresql

# Copy certificates
sudo cp ca-cert.pem /etc/ssl/postgresql/
sudo cp client-cert.pem /etc/ssl/postgresql/
sudo cp client-key.pem /etc/ssl/postgresql/

# Set permissions
sudo chmod 600 /etc/ssl/postgresql/*.pem
sudo chown postgres:postgres /etc/ssl/postgresql/*.pem
```

### 4. Configure Application

Set environment variables for production:

```bash
# Production environment
export ENVIRONMENT="production"
export DB_SSLMODE="verify-full"
export DB_SSL_ROOT_CERT="/etc/ssl/postgresql/ca-cert.pem"
export DB_SSL_CERT="/etc/ssl/postgresql/client-cert.pem"
export DB_SSL_KEY="/etc/ssl/postgresql/client-key.pem"
```

## Migration Tool SSL Support

The unified migration tool supports SSL configuration:

```bash
# Basic SSL
go run cmd/unified-migration/main.go --ssl=require

# With certificates
go run cmd/unified-migration/main.go \
  --ssl=verify-full \
  --ssl-root-cert=/path/to/ca.pem \
  --ssl-cert=/path/to/client.pem \
  --ssl-key=/path/to/client.key
```

## Troubleshooting

### Common Issues

1. **Connection refused with SSL**
   - Check if PostgreSQL server has SSL enabled
   - Verify firewall allows SSL port (usually 5432)

2. **Certificate verification failed**
   - Ensure CA certificate is correct
   - Check certificate paths and permissions
   - Verify hostname matches certificate CN

3. **Permission denied on certificate files**
   - Ensure application user can read certificate files
   - Check file permissions (should be 600 or 644)

### Testing SSL Connection

Test SSL connection using psql:

```bash
# Test with SSL
psql "postgresql://user:pass@host:5432/db?sslmode=require"

# Test with certificates
psql "postgresql://user:pass@host:5432/db?sslmode=verify-full&sslrootcert=/path/to/ca.pem"
```

### SSL Health Check

Use the SSL health check utility:

```go
import "openpenpal-backend/internal/config"

// Check SSL configuration
sslConfig := config.NewSSLConfig("production")
health, err := config.CheckSSLHealth(sslConfig)
if err != nil {
    log.Printf("SSL health check failed: %v", err)
}

log.Printf("SSL Enabled: %v", health.Enabled)
log.Printf("Certificate Valid: %v", health.CertificateValid)
log.Printf("Certificate Expiry: %v", health.CertificateExpiry)
```

## Security Best Practices

1. **Always use SSL in production**
   - Minimum `require` mode
   - Prefer `verify-full` for maximum security

2. **Protect certificate files**
   - Store in secure location
   - Restrict file permissions (600)
   - Use secrets management for cloud deployments

3. **Monitor certificate expiry**
   - Set up alerts for certificates expiring within 30 days
   - Implement automatic certificate rotation

4. **Use strong passwords**
   - Even with SSL, use strong database passwords
   - Consider certificate-based authentication

5. **Regular security audits**
   - Review SSL configuration regularly
   - Update certificates before expiry
   - Monitor for SSL vulnerabilities

## Cloud Provider Specific

### AWS RDS

```bash
# Download RDS CA certificate
wget https://s3.amazonaws.com/rds-downloads/rds-ca-2019-root.pem

# Configure
export DB_SSLMODE="require"
export DB_SSL_ROOT_CERT="./rds-ca-2019-root.pem"
```

### Google Cloud SQL

```bash
# Use Cloud SQL Proxy for automatic SSL
./cloud_sql_proxy -instances=PROJECT:REGION:INSTANCE=tcp:5432
```

### Azure Database for PostgreSQL

```bash
# Download Azure CA certificate
wget https://www.digicert.com/CACerts/BaltimoreCyberTrustRoot.crt.pem

# Configure
export DB_SSLMODE="require"
export DB_SSL_ROOT_CERT="./BaltimoreCyberTrustRoot.crt.pem"
```

## Performance Considerations

1. **SSL Overhead**
   - SSL adds ~10-20% overhead
   - Use connection pooling to minimize handshake cost
   - Consider SSL session caching

2. **Connection Pooling**
   - Reuse SSL connections
   - Configure appropriate pool size
   - Monitor connection metrics

3. **Certificate Validation**
   - `verify-full` adds hostname verification overhead
   - Balance security needs with performance
   - Cache validated connections

## Monitoring

Monitor SSL connections:

```sql
-- Check SSL connections
SELECT pid, ssl, version, cipher, bits, compression, clientdn 
FROM pg_stat_ssl 
WHERE pid = pg_backend_pid();

-- Count SSL vs non-SSL connections
SELECT ssl, count(*) 
FROM pg_stat_ssl 
JOIN pg_stat_activity USING(pid) 
GROUP BY ssl;
```

## Conclusion

Proper SSL configuration is crucial for production database security. This guide provides comprehensive instructions for setting up and maintaining secure PostgreSQL connections in the OpenPenPal system.

For additional security considerations, refer to the PostgreSQL documentation on SSL support.