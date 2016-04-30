#!/bin/bash

rands() {
    cat /dev/urandom | tr -dc 'a-zA-Z' | fold -w ${1:-16} | head -n1
}

apt-get update
apt-get install -y net-tools supervisor openssl

IP=$(ifconfig eth0 | awk '/inet addr/{print substr($2,6)}')
BLOCK=$(rands)
HASH=$(rands)
QUIMBY_JWT_PRIV=/var/lib/quimby/certs/jwt/private_key.pem
QUIMBY_JWT_PUB=/var/lib/quimby/certs/jwt/public_key.pem
QUIMBY_TLS_KEYS=/var/lib/quimby/certs/tls
QUIMBY_TLS_KEY=/var/lib/quimby/certs/tls/key.pem
QUIMBY_TLS_CERT=/var/lib/quimby/certs/tls/cert.pem

mkdir -p /etc/quimby
mkdir -p /var/lib/quimby/certs/jwt
mkdir -p /var/lib/quimby/certs/tls

echo "export QUIMBY_DB=/var/lib/quimby/quimby.db
export QUIMBY_INTERFACE=$IP
export QUIMBY_PORT=443
export QUIMBY_INTERNAL_PORT=8989
export QUIMBY_HOST=http://$IP
export QUIMBY_BLOCK_KEY=$BLOCK
export QUIMBY_HASH_KEY= $HASH
export QUIMBY_JWT_PRIV=$QUIMBY_JWT_PRIV
export QUIMBY_JWT_PUB=$QUIMBY_JWT_PUB
export QUIMBY_TLS_KEY=$QUIMBY_TLS_KEY
export QUIMBY_TLS_CERT=$QUIMBY_TLS_CERT
export QUIMBY_USER=quimby" > /etc/quimby/quimby.conf

echo "environment=QUIMBY_DB='/var/lib/quimby/quimby.db',
  QUIMBY_INTERFACE='$IP',
  QUIMBY_PORT='443',
  QUIMBY_INTERNAL_PORT='8989',
  QUIMBY_HOST='http://$IP',
  QUIMBY_BLOCK_KEY='$BLOCK',
  QUIMBY_HASH_KEY=' $HASH',
  QUIMBY_JWT_PRIV='$QUIMBY_JWT_PRIV',
  QUIMBY_JWT_PUB='$QUIMBY_JWT_PUB',
  QUIMBY_TLS_KEY='$QUIMBY_TLS_KEY',
  QUIMBY_TLS_CERT='$QUIMBY_TLS_CERT',
  QUIMBY_USER='quimby'
command=/usr/local/bin/quimby serve" > /etc/supervisor/conf.d/quimby.conf

# generate keys for JWT
openssl genrsa -out $QUIMBY_JWT_PRIV 2048
openssl rsa -in $QUIMBY_JWT_PRIV -pubout -out $QUIMBY_JWT_PUB

# generate keys for TLS
quimby cert --domain localhost --path $QUIMBY_TLS_KEYS
