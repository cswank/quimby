package utils

import (
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

func SetupServer(domain, net string) {
	s := Setup{
		Interface: net,
		Domain:    domain,
	}
	f, err := ioutil.TempFile("", "")
	if err := f.Chmod(0777); err != nil {
		log.Fatal("couldn't run setup script", err)
	}
	if err != nil {
		log.Fatal("couldn't run setup script", err)
	}
	t := template.Must(template.New("script").Parse(script))
	if err := t.Execute(f, s); err != nil {
		log.Fatal("couldn't run setup script", err)
	}
	f.Close()

	cmd := exec.Command(f.Name())
	stdin, err := cmd.StdinPipe()

	if err != nil {
		log.Fatal("couldn't run setup script", err)
	}
	defer stdin.Close()

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	cmd.Wait()
	os.Remove(f.Name())
}

type Setup struct {
	Interface string
	Domain    string
}

const script = `#!/bin/bash

rands() {
    cat /dev/urandom | tr -dc 'a-zA-Z' | fold -w ${1:-16} | head -n1
}

apt-get update
apt-get install -y net-tools openssl

IP=$(ifconfig {{.Interface}} | awk '/inet addr/{print substr($2,6)}')
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
export QUIMBY_USER=quimby" > /etc/quimby/quimby.env

echo "[Unit]
Description=Gogadgets web interface
After=network.target

[Service]
Type=simple
PIDFile=/tmp/quimby.pid
Environment=QUIMBY_DB=/var/lib/quimby/quimby.db
Environment=QUIMBY_INTERFACE=$IP
Environment=QUIMBY_PORT=443
Environment=QUIMBY_INTERNAL_PORT=8989
Environment=QUIMBY_HOST=http://$IP
Environment=QUIMBY_BLOCK_KEY=$BLOCK
Environment=QUIMBY_HASH_KEY= $HASH
Environment=QUIMBY_JWT_PRIV=$QUIMBY_JWT_PRIV
Environment=QUIMBY_JWT_PUB=$QUIMBY_JWT_PUB
Environment=QUIMBY_TLS_KEY=$QUIMBY_TLS_KEY
Environment=QUIMBY_TLS_CERT=$QUIMBY_TLS_CERT
Environment=QUIMBY_USER=quimby
ExecStart=/usr/local/bin/quimby serve

[Install]
# multi-user.target corresponds to run level 3
# roughtly meaning wanted by system start
WantedBy    = multi-user.target
" > /etc/systemd/system/quimby.service

# generate keys for JWT
openssl genrsa -out $QUIMBY_JWT_PRIV 2048
openssl rsa -in $QUIMBY_JWT_PRIV -pubout -out $QUIMBY_JWT_PUB

# generate keys for TLS
quimby cert --domain {{domain}} --path $QUIMBY_TLS_KEYS
`
