#!/bin/sh
# Script to generate self-signed SSL certificate inside container
# This runs during Docker initialization

set -e

CERT_DIR="/etc/nginx/certs"
CERT_FILE="$CERT_DIR/server.crt"
KEY_FILE="$CERT_DIR/server.key"
CONFIG_FILE="$CERT_DIR/san.cnf"

# Create directory if not exists
mkdir -p "$CERT_DIR"

# Check if certificate already exists
if [ ! -f "$CERT_FILE" ] || [ ! -f "$KEY_FILE" ]; then
    echo "Generating self-signed SSL certificate with SAN..."
    
    # Create config with Subject Alternative Names
    cat > "$CONFIG_FILE" << EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
C = RU
ST = Moscow
L = Moscow
O = Warehouse
CN = localhost

[v3_req]
subjectAltName = DNS:localhost,DNS:127.0.0.1,IP:127.0.0.1,DNS:192.168.0.101,IP:192.168.0.101
EOF
    
    # Generate key
    openssl genrsa -out "$KEY_FILE" 2048 2>/dev/null
    
    # Generate certificate with SAN
    openssl req -new -x509 -key "$KEY_FILE" -out "$CERT_FILE" \
        -days 365 -config "$CONFIG_FILE" 2>/dev/null
    
    echo "Certificate generated successfully!"
    echo "Certificate: $CERT_FILE"
    echo "Key: $KEY_FILE"
    echo "SAN: localhost, 127.0.0.1, 192.168.0.101"
else
    echo "Using existing certificate"
fi

# Set proper permissions
chmod 644 "$CERT_FILE"
chmod 600 "$KEY_FILE"

echo "SSL setup complete"
