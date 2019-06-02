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
	fmt.Printf("%s", GenerateKey())
}

func GenerateKey() string {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(fmt.Errorf("Error generating P256 ECDSA key: %v", err))
	}
	if key == nil {
		panic(fmt.Errorf("Could not generate P256 ECDSA key"))
	}

	x509Encoded, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		panic(fmt.Errorf("Error encoding P256 ECDSA key: %v", err))
	}
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})

	return string(pemEncoded)
}
