#!/bin/bash
set -e

# Configuration
DIR="certs"
mkdir -p $DIR

# Generate CA (Certificate Authority)
echo ">>> Generating CA..."
openssl req -new -x509 -days 365 -nodes \
    -out $DIR/ca.crt \
    -keyout $DIR/ca.key \
    -subj "/C=ID/ST=Jakarta/L=Jakarta/O=HelixCorp/OU=Platform/CN=HelixRootCA"

# Generate Server Certificate
echo ">>> Generating Server Certs (svc-product)..."
openssl req -new -nodes \
    -out $DIR/server.csr \
    -keyout $DIR/server.key \
    -subj "/C=ID/ST=Jakarta/L=Jakarta/O=HelixCorp/OU=Backend/CN=localhost"

# Sign Server Cert with CA
openssl x509 -req -in $DIR/server.csr \
    -CA $DIR/ca.crt -CAkey $DIR/ca.key -CAcreateserial \
    -out $DIR/server.crt -days 365 \
    -extfile <(printf "subjectAltName=DNS:localhost,IP:127.0.0.1")

# Generate Client Certificate (For testing/curl)
echo ">>> Generating Client Certs..."
openssl req -new -nodes \
    -out $DIR/client.csr \
    -keyout $DIR/client.key \
    -subj "/C=ID/ST=Jakarta/L=Jakarta/O=HelixCorp/OU=Client/CN=authorized-client"

# Sign Client Cert with CA
openssl x509 -req -in $DIR/client.csr \
    -CA $DIR/ca.crt -CAkey $DIR/ca.key -CAcreateserial \
    -out $DIR/client.crt -days 365

# Cleanup CSR
rm $DIR/*.csr

echo ">>> Done! Certificates are in '$DIR/'"
echo "To test mTLS:"
echo "  curl --cert $DIR/client.crt --key $DIR/client.key --cacert $DIR/ca.crt https://localhost:30000/v1/products"