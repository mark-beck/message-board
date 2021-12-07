# Authentification Server

### generate Token Ecdsa keys
```shell
openssl ecparam -name prime256v1 -genkey -noout -out cert/jwt.private.key
openssl pkcs8 -topk8 -nocrypt -in cert/jwt.private.key -out cert/jwt.private.pem
openssl ec -in cert/jwt.private.pem -pubout -out cert/jwt.public.pem
rm /cert/jwt.private.key
```

### generate ssl cert and keys

Generate the root cert:
```shell
openssl genrsa -des3 -out root.key 2048
openssl req -x509 -new -nodes -key root.key -sha256 -days 3650 -out root.pem -subj '/CN=My Root'
```
Create the signed SSL cert:
```shell
openssl genrsa -out sslprivate.key 2048
openssl req -new -key sslprivate.key -out sslprivate.csr -subj '/CN=www.domain.com'
```

now create a file named ```sslprivate.ext``` with the following content:
```
authorityKeyIdentifier=keyid,issuer 
basicConstraints=CA:FALSE
keyUsage=digitalSignature
subjectAltName=@alt_names 
[alt_names]
DNS.1=domain.com
DNS.2=www.domain.com
```
Now run: 
```shell
openssl x509 -req -in sslprivate.csr -CA root.pem -CAkey root.key -CAcreateserial -out sslprivate.crt -sha256 -days 365 -extfile sslprivate.ext
```