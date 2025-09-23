#!/bin/bash

# Generate SSL certificates for development/testing
# For production, use proper certificates from a trusted CA

set -e

SSL_DIR="nginx/ssl"
DOMAIN="localhost"

echo "Generating SSL certificates for $DOMAIN..."

# Create SSL directory if it doesn't exist
mkdir -p "$SSL_DIR"

# Generate private key
openssl genrsa -out "$SSL_DIR/key.pem" 2048

# Generate certificate signing request
openssl req -new -key "$SSL_DIR/key.pem" -out "$SSL_DIR/cert.csr" -subj "/C=US/ST=State/L=City/O=Organization/CN=$DOMAIN"

# Generate self-signed certificate
openssl x509 -req -days 365 -in "$SSL_DIR/cert.csr" -signkey "$SSL_DIR/key.pem" -out "$SSL_DIR/cert.pem"

# Set proper permissions
chmod 600 "$SSL_DIR/key.pem"
chmod 644 "$SSL_DIR/cert.pem"

# Clean up CSR file
rm "$SSL_DIR/cert.csr"

echo "SSL certificates generated successfully!"
echo "Certificate: $SSL_DIR/cert.pem"
echo "Private Key: $SSL_DIR/key.pem"
echo ""
echo "WARNING: These are self-signed certificates for development only."
echo "For production, obtain certificates from a trusted Certificate Authority."
echo ""
echo "To trust the certificate in your browser:"
echo "1. Open https://localhost in your browser"
echo "2. Click 'Advanced' and 'Proceed to localhost (unsafe)'"
echo "3. Or add the certificate to your system's trusted certificates"
