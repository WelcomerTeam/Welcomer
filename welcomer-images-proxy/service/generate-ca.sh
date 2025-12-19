openssl genrsa -out welcomer-ca.key 4096

openssl req -x509 -new -nodes \
  -key welcomer-ca.key \
  -sha256 \
  -days 3650 \
  -out welcomer-ca.crt \
  -subj "/CN=Welcomer MITM CA" \
  -extensions v3_ca \
  -config <(
    cat <<'EOF'
[req]
distinguished_name=req
x509_extensions=v3_ca
prompt=no

[v3_ca]
basicConstraints=critical,CA:TRUE
keyUsage=critical,keyCertSign,cRLSign
subjectKeyIdentifier=hash
authorityKeyIdentifier=keyid:always
EOF
)
