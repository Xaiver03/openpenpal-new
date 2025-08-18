#!/bin/bash

# Production SSL Setup Script for PostgreSQL
# This script helps configure SSL certificates for secure database connections

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SSL_DIR="/etc/ssl/postgresql"
CERT_VALIDITY_DAYS=365
ENVIRONMENT=${ENVIRONMENT:-"development"}

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        print_error "This script must be run as root for production setup"
        print_info "For development, you can run without root"
        return 1
    fi
    return 0
}

# Function to create SSL directory
create_ssl_directory() {
    if [[ -d "$SSL_DIR" ]]; then
        print_info "SSL directory already exists: $SSL_DIR"
    else
        print_info "Creating SSL directory: $SSL_DIR"
        mkdir -p "$SSL_DIR"
        chmod 700 "$SSL_DIR"
        print_success "SSL directory created"
    fi
}

# Function to generate self-signed certificates (development only)
generate_self_signed_certs() {
    print_warning "Generating self-signed certificates for DEVELOPMENT ONLY"
    print_warning "DO NOT use these certificates in production!"
    
    local dev_ssl_dir="./dev-ssl"
    mkdir -p "$dev_ssl_dir"
    
    # Generate CA private key
    openssl genrsa -out "$dev_ssl_dir/ca-key.pem" 2048
    
    # Generate CA certificate
    openssl req -new -x509 -days $CERT_VALIDITY_DAYS -key "$dev_ssl_dir/ca-key.pem" \
        -out "$dev_ssl_dir/ca-cert.pem" \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=OpenPenPal/CN=OpenPenPal-CA"
    
    # Generate server private key
    openssl genrsa -out "$dev_ssl_dir/server-key.pem" 2048
    
    # Generate server certificate request
    openssl req -new -key "$dev_ssl_dir/server-key.pem" \
        -out "$dev_ssl_dir/server-req.pem" \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=OpenPenPal/CN=localhost"
    
    # Sign server certificate
    openssl x509 -req -days $CERT_VALIDITY_DAYS -in "$dev_ssl_dir/server-req.pem" \
        -CA "$dev_ssl_dir/ca-cert.pem" -CAkey "$dev_ssl_dir/ca-key.pem" \
        -CAcreateserial -out "$dev_ssl_dir/server-cert.pem"
    
    # Generate client private key
    openssl genrsa -out "$dev_ssl_dir/client-key.pem" 2048
    
    # Generate client certificate request
    openssl req -new -key "$dev_ssl_dir/client-key.pem" \
        -out "$dev_ssl_dir/client-req.pem" \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=OpenPenPal/CN=openpenpal-client"
    
    # Sign client certificate
    openssl x509 -req -days $CERT_VALIDITY_DAYS -in "$dev_ssl_dir/client-req.pem" \
        -CA "$dev_ssl_dir/ca-cert.pem" -CAkey "$dev_ssl_dir/ca-key.pem" \
        -CAcreateserial -out "$dev_ssl_dir/client-cert.pem"
    
    # Clean up certificate requests
    rm -f "$dev_ssl_dir"/*-req.pem
    
    # Set permissions
    chmod 600 "$dev_ssl_dir"/*.pem
    
    print_success "Self-signed certificates generated in $dev_ssl_dir/"
    print_info "CA Certificate: $dev_ssl_dir/ca-cert.pem"
    print_info "Server Certificate: $dev_ssl_dir/server-cert.pem"
    print_info "Server Key: $dev_ssl_dir/server-key.pem"
    print_info "Client Certificate: $dev_ssl_dir/client-cert.pem"
    print_info "Client Key: $dev_ssl_dir/client-key.pem"
}

# Function to install production certificates
install_production_certs() {
    print_info "Installing production certificates"
    
    # Check if certificates are provided
    if [[ ! -f "$1" ]] || [[ ! -f "$2" ]] || [[ ! -f "$3" ]]; then
        print_error "Usage: $0 install <ca-cert> <client-cert> <client-key>"
        exit 1
    fi
    
    local ca_cert="$1"
    local client_cert="$2"
    local client_key="$3"
    
    # Create SSL directory if needed
    create_ssl_directory
    
    # Copy certificates
    print_info "Copying certificates to $SSL_DIR"
    cp "$ca_cert" "$SSL_DIR/ca-cert.pem"
    cp "$client_cert" "$SSL_DIR/client-cert.pem"
    cp "$client_key" "$SSL_DIR/client-key.pem"
    
    # Set permissions
    chmod 600 "$SSL_DIR"/*.pem
    chown postgres:postgres "$SSL_DIR"/*.pem 2>/dev/null || true
    
    print_success "Production certificates installed"
}

# Function to verify certificates
verify_certificates() {
    local cert_dir="${1:-$SSL_DIR}"
    
    print_info "Verifying certificates in $cert_dir"
    
    # Check CA certificate
    if [[ -f "$cert_dir/ca-cert.pem" ]]; then
        print_info "CA Certificate:"
        openssl x509 -in "$cert_dir/ca-cert.pem" -noout -subject -issuer -dates
        echo ""
    else
        print_warning "CA certificate not found"
    fi
    
    # Check client certificate
    if [[ -f "$cert_dir/client-cert.pem" ]]; then
        print_info "Client Certificate:"
        openssl x509 -in "$cert_dir/client-cert.pem" -noout -subject -issuer -dates
        
        # Verify certificate chain
        if [[ -f "$cert_dir/ca-cert.pem" ]]; then
            print_info "Verifying certificate chain..."
            if openssl verify -CAfile "$cert_dir/ca-cert.pem" "$cert_dir/client-cert.pem" > /dev/null 2>&1; then
                print_success "Certificate chain is valid"
            else
                print_error "Certificate chain verification failed"
            fi
        fi
        echo ""
    else
        print_warning "Client certificate not found"
    fi
    
    # Check for expiring certificates
    if [[ -f "$cert_dir/client-cert.pem" ]]; then
        local expiry_date=$(openssl x509 -in "$cert_dir/client-cert.pem" -noout -enddate | cut -d= -f2)
        local expiry_timestamp=$(date -d "$expiry_date" +%s)
        local current_timestamp=$(date +%s)
        local days_until_expiry=$(( ($expiry_timestamp - $current_timestamp) / 86400 ))
        
        if [[ $days_until_expiry -lt 30 ]]; then
            print_warning "Certificate expires in $days_until_expiry days!"
        else
            print_info "Certificate valid for $days_until_expiry more days"
        fi
    fi
}

# Function to test SSL connection
test_ssl_connection() {
    print_info "Testing SSL connection"
    
    # Get connection parameters
    local host="${DB_HOST:-localhost}"
    local port="${DB_PORT:-5432}"
    local user="${DB_USER:-postgres}"
    local dbname="${DB_NAME:-openpenpal}"
    local sslmode="${DB_SSLMODE:-require}"
    
    print_info "Connection parameters:"
    print_info "  Host: $host"
    print_info "  Port: $port"
    print_info "  User: $user"
    print_info "  Database: $dbname"
    print_info "  SSL Mode: $sslmode"
    
    # Build connection string
    local conn_string="postgresql://$user@$host:$port/$dbname?sslmode=$sslmode"
    
    if [[ "$sslmode" == "verify-ca" ]] || [[ "$sslmode" == "verify-full" ]]; then
        if [[ -f "$SSL_DIR/ca-cert.pem" ]]; then
            conn_string="${conn_string}&sslrootcert=$SSL_DIR/ca-cert.pem"
        fi
        if [[ -f "$SSL_DIR/client-cert.pem" ]]; then
            conn_string="${conn_string}&sslcert=$SSL_DIR/client-cert.pem"
        fi
        if [[ -f "$SSL_DIR/client-key.pem" ]]; then
            conn_string="${conn_string}&sslkey=$SSL_DIR/client-key.pem"
        fi
    fi
    
    print_info "Testing connection..."
    if command -v psql > /dev/null; then
        if psql "$conn_string" -c "SELECT 1" > /dev/null 2>&1; then
            print_success "SSL connection successful"
        else
            print_error "SSL connection failed"
            print_info "Try running with PGSSLMODE=$sslmode psql ..."
        fi
    else
        print_warning "psql not found, skipping connection test"
    fi
}

# Function to generate environment configuration
generate_env_config() {
    local env_file="${1:-.env.ssl}"
    local cert_dir="${2:-$SSL_DIR}"
    
    print_info "Generating environment configuration"
    
    cat > "$env_file" << EOF
# PostgreSQL SSL Configuration
# Generated on $(date)

# SSL Mode (disable, allow, prefer, require, verify-ca, verify-full)
DB_SSLMODE=${DB_SSLMODE:-require}

# SSL Certificate Paths
DB_SSL_ROOT_CERT=$cert_dir/ca-cert.pem
DB_SSL_CERT=$cert_dir/client-cert.pem
DB_SSL_KEY=$cert_dir/client-key.pem

# Connection Parameters
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=\${DB_PASSWORD}
DB_NAME=${DB_NAME:-openpenpal}
EOF
    
    print_success "Environment configuration written to $env_file"
    print_info "Source this file before running the application:"
    print_info "  source $env_file"
}

# Main menu
show_usage() {
    echo "PostgreSQL SSL Setup Script"
    echo ""
    echo "Usage: $0 [command] [options]"
    echo ""
    echo "Commands:"
    echo "  generate-dev      Generate self-signed certificates for development"
    echo "  install           Install production certificates"
    echo "  verify            Verify installed certificates"
    echo "  test              Test SSL connection"
    echo "  env               Generate environment configuration"
    echo "  help              Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 generate-dev"
    echo "  $0 install ca.pem client.pem client-key.pem"
    echo "  $0 verify"
    echo "  $0 test"
    echo "  $0 env .env.production"
}

# Main script logic
case "${1:-help}" in
    generate-dev)
        if [[ "$ENVIRONMENT" == "production" ]]; then
            print_error "Cannot generate self-signed certificates in production environment"
            exit 1
        fi
        generate_self_signed_certs
        ;;
    install)
        if ! check_root; then
            print_warning "Running without root - certificates will be installed locally"
            SSL_DIR="./ssl"
        fi
        install_production_certs "$2" "$3" "$4"
        ;;
    verify)
        verify_certificates "${2:-$SSL_DIR}"
        ;;
    test)
        test_ssl_connection
        ;;
    env)
        generate_env_config "$2" "${3:-$SSL_DIR}"
        ;;
    help|*)
        show_usage
        ;;
esac