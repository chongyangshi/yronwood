package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

func main() {
	privateKey, err := GenerateKey()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s", privateKey)
}

func GenerateKey() (string, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", fmt.Errorf("Error generating P256 ECDSA key: %v", err)
	}
	if key == nil {
		return "", fmt.Errorf("Could not generate P256 ECDSA key")
	}

	x509Encoded, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return "", fmt.Errorf("Error encoding P256 ECDSA key: %v", err)
	}
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})

	return string(pemEncoded), nil
}
