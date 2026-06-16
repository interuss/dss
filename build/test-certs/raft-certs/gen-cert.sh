#!/bin/bash

# Default values
NODE_COUNT=3

# Help function
show_help() {
  echo "Usage: ./gen-cert.sh [options]"
  echo ""
  echo "Options:"
  echo "  -n, --nodes <num>   Number of nodes to generate (default: 3)"
  echo "  -c, --clean         Delete all .key, .crt, .csr, and .srl files"
  echo "  -h, --help          Show this help message"
  exit 0
}

# Cleanup function
clean_certs() {
    echo "Cleaning up all certificates and keys..."
    rm -f ./*.key ./*.crt ./*.csr ./*.srl
    echo "Done."
    exit 0
}

# Parse arguments
while [[ "$#" -gt 0 ]]; do
  case $1 in
    -n|--nodes) NODE_COUNT="$2"; shift ;;
    -c|--clean) clean_certs ;;
    -h|--help) show_help ;;
    *) echo "Unknown parameter: $1"; show_help ;;
  esac
  shift
done

echo "Creating CA and certificates for $NODE_COUNT nodes..."

# 1. Generate the Root CA
openssl req -x509 -newkey rsa:4096 -keyout ca.key -out ca.crt -days 3650 -sha256 -nodes -subj "/CN=root"

# 2. Loop to create node certificates
for ((i=1; i<=NODE_COUNT; i++))
do
    echo "Generating node$i..."

    # Create Request
    openssl req -out "node$i.csr" -newkey rsa:2048 -keyout "node$i.key" -nodes \
      -subj "/CN=node$i" \
      -addext "subjectAltName=DNS:node$i,DNS:localhost,IP:127.0.0.1"

    # Sign Certificate
    openssl x509 -req -in "node$i.csr" -CA ca.crt -CAkey ca.key \
      -out "node$i.crt" -days 365 -sha256 \
      -set_serial "$i" \
      -extfile <(echo "subjectAltName=DNS:node$i,DNS:localhost,IP:127.0.0.1")
done

# 3. Final Cleanup of temporary request files
rm -f ./*.csr
echo "Success! Generated $NODE_COUNT node certificates."