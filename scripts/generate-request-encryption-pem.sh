#!/bin/sh

set -eu

usage() {
  cat <<'EOF'
Usage:
  generate-request-encryption-pem.sh [output-dir] [key-name] [bits]

Arguments:
  output-dir  Output directory. Default: ./secrets/request-encryption
  key-name    File name prefix. Default: request-encryption
  bits        RSA key size, at least 2048. Default: 3072

Example:
  ./scripts/generate-request-encryption-pem.sh
  ./scripts/generate-request-encryption-pem.sh /etc/wklive/secrets admin-api 4096

Outputs:
  <output-dir>/<key-name>-private.pem  PKCS#8 RSA private key
  <output-dir>/<key-name>-public.pem   SPKI RSA public key
EOF
}

case "${1:-}" in
  -h|--help)
    usage
    exit 0
    ;;
esac

output_dir=${1:-./secrets/request-encryption}
key_name=${2:-request-encryption}
bits=${3:-3072}

case "$bits" in
  ''|*[!0-9]*)
    echo "error: bits must be an integer" >&2
    exit 1
    ;;
esac

if [ "$bits" -lt 2048 ]; then
  echo "error: bits must be at least 2048" >&2
  exit 1
fi

case "$key_name" in
  ''|*[!A-Za-z0-9._-]*)
    echo "error: key-name may contain only letters, numbers, dot, underscore, and hyphen" >&2
    exit 1
    ;;
esac

if ! command -v openssl >/dev/null 2>&1; then
  echo "error: openssl is required" >&2
  echo "macOS: brew install openssl" >&2
  echo "Debian/Ubuntu: sudo apt-get install openssl" >&2
  echo "RHEL/Rocky/AlmaLinux: sudo dnf install openssl" >&2
  exit 1
fi

private_key="$output_dir/$key_name-private.pem"
public_key="$output_dir/$key_name-public.pem"

if [ -e "$private_key" ] || [ -e "$public_key" ]; then
  echo "error: key file already exists; refusing to overwrite" >&2
  echo "private: $private_key" >&2
  echo "public:  $public_key" >&2
  exit 1
fi

umask 077
mkdir -p "$output_dir"

openssl genpkey \
  -algorithm RSA \
  -pkeyopt "rsa_keygen_bits:$bits" \
  -out "$private_key"

openssl pkey \
  -in "$private_key" \
  -pubout \
  -out "$public_key"

chmod 600 "$private_key"
chmod 644 "$public_key"

echo "RSA key pair generated successfully."
echo "Private key: $private_key"
echo "Public key:  $public_key"
echo "RSA bits:    $bits"
echo
echo "Set RequestEncryption.RSAPrivateKeyPath to:"
echo "  $private_key"
