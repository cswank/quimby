# Quimby


## TLS

### Server

    openssl req -x509 -newkey rsa:4096 -keyout server_key.pem -out server_cert.pem -nodes -days 365 -subj "/CN=localhost/O=Client\ Certificate"

### Client

Create:

    openssl req -newkey rsa:4096 -keyout craig_dev_key.pem -out craig_dev_csr.pem -nodes -days 365 -subj "/CN=Craig"

Sign it:

    openssl x509 -req -in craig_dev_csr.pem -CA server_cert.pem -CAkey server_key.pem -out craig_dev_cert.pem -set_serial 01 -days 365

Bundle it:

    openssl pkcs12 -export -clcerts -in craig_dev_cert.pem -inkey craig_dev_key.pem -out craig_dev.p12

Add the p12 file to your certificates in Keychain Access on macos (double click on the p12 file).

On iOS you need to email the p12 file to yourself, double-click to install it.  You then need to go to
Settings/General/About/Certificate Trust Settings and enable full trust for root certificates for the one
you just installed.
