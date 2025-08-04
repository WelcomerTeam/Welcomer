PRIVATE_KEY_FILE="keys/private.pem"
PUBLIC_KEY_FILE="keys/public.pem"

set -e
mkdir -p keys

openssl genrsa -out "$PRIVATE_KEY_FILE" 2048
openssl rsa -in "$PRIVATE_KEY_FILE" -pubout -out "$PUBLIC_KEY_FILE"
