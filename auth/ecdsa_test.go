package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"
	"time"
)

func TestAdminToken(t *testing.T) {
	testKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Error generating P256 ECDSA key: %+v", err)
	}

	cachedSigningKey = testKey

	_, err = SignAdminToken(maxAdminTokenValidity + time.Hour*12)
	if err == nil {
		t.Fatal("Unexpected admin token signing success with validity too long")
	}

	adminToken, err := SignAdminToken(time.Hour * 4)
	if err != nil {
		t.Fatalf("Unexpected error signing admin token with valid validity: %+v", err)
	}

	if len(adminToken) == 0 {
		t.Fatal("Unexpected empty signing admin token")
	}

	verified, err := VerifyAdminToken(adminToken)
	if err != nil {
		t.Fatalf("Unexpected error verifying valid admin token: %+v", err)
	}

	if !verified {
		t.Fatal("Valid admin token cannot be verified")
	}

	verified, err = VerifyAdminToken("invalid")
	if verified || err == nil {
		t.Fatal("Verifying invalid token should trigger an error")
	}
}

func TestImageToken(t *testing.T) {
	testKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Error generating P256 ECDSA key: %+v", err)
	}

	cachedSigningKey = testKey

	testImagePath := "private/test000.jpg"

	_, err = SignImageToken(maxImageTokenValidity+time.Hour*12, testImagePath)
	if err == nil {
		t.Fatal("Unexpected image token signing success with validity too long")
	}

	_, err = SignImageToken(time.Hour*4, "")
	if err == nil {
		t.Fatal("Unexpected image token signing success with empty image path")
	}

	imageToken, err := SignImageToken(time.Hour*4, testImagePath)
	if err != nil {
		t.Fatalf("Unexpected error signing image token with valid validity: %+v", err)
	}

	if len(imageToken) == 0 {
		t.Fatal("Unexpected empty signing image token")
	}

	verified, err := VerifyImageToken(imageToken, testImagePath)
	if err != nil {
		t.Fatalf("Unexpected error verifying valid image token: %+v", err)
	}

	if !verified {
		t.Fatal("Valid image token cannot be verified")
	}

	verified, _ = VerifyAdminToken(imageToken)
	if verified {
		t.Fatal("Verifying image token with VerifyAdminToken should not succeed")
	}
}
