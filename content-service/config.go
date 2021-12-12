package main

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"
)

type Config struct {
	token_secret string
}

func (self Config) load() (secrets Secrets) {
	publicKeyFile, err := os.Open(self.token_secret)
	if err != nil {
		log.Fatalf("cant find tokensecret %v\n", err)
	}

	pemfileinfo, _ := publicKeyFile.Stat()
	var size int64 = pemfileinfo.Size()
	pembytes := make([]byte, size)

	buffer := bufio.NewReader(publicKeyFile)
	_, err = buffer.Read(pembytes)

	data, _ := pem.Decode([]byte(pembytes))

	publicKeyFile.Close()

	publicKeyImported, err := x509.ParsePKIXPublicKey(data.Bytes)
	if err != nil {
		log.Fatalf("cant parse key %v\n", err)
	}

	public_ecdsa_key, ok := publicKeyImported.(*ecdsa.PublicKey)
	if !ok {
		log.Fatalf("not of type ecdsa %T\n", publicKeyImported)
	}

	return Secrets{
		token_public: public_ecdsa_key,
	}
}

type Secrets struct {
	token_public *ecdsa.PublicKey
}
