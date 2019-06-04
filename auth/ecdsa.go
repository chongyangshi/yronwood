package auth

// This authentication implementation is not sufficiently secure for critical
// access authentication, as this an application with lower levels of security
// requirements. It issues a signed token with fixed expiry time salted with
// a random payload, after authenticating the client initially with a shared
// secret. Asymmetric client authentication is explicitly not used, as this
// application's front-end should be usable on a generic device.

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/monzo/terrors"

	"github.com/icydoge/yronwood/config"
)

var (
	signingKey        *ecdsa.PrivateKey
	randomPayloadSize = 2048
)

type ECDSASignature struct {
	R, S *big.Int
}

func init() {
	block, _ := pem.Decode([]byte(config.ConfigAuthenticationSigningKey))
	x509Encoded := block.Bytes
	if block == nil {
		panic(fmt.Errorf("Could not decode ECDSA signing key from %s", config.ConfigAuthenticationSigningKey))
	}

	var err error
	signingKey, err = x509.ParseECPrivateKey(x509Encoded)
	if err != nil {
		panic(fmt.Errorf("Error reading ECDSA signing key: %v", err))
	}
}

func SignToken(validity time.Duration) (string, error) {
	randomPayload := make([]byte, randomPayloadSize)
	_, err := rand.Read(randomPayload)
	if err != nil {
		return "", terrors.Wrap(err, nil)
	}

	payloadHash := sha256.New()
	_, err = payloadHash.Write(randomPayload)
	if err != nil {
		return "", terrors.Wrap(err, nil)
	}
	payloadHash256 := hex.EncodeToString(payloadHash.Sum(nil))

	expiry := fmt.Sprintf("%d", time.Now().Add(validity).Unix())
	saltedExpiry := fmt.Sprintf("%s_%s", expiry, payloadHash256)

	tokenHash := sha256.New()
	_, err = tokenHash.Write([]byte(saltedExpiry))
	if err != nil {
		return "", terrors.Wrap(err, nil)
	}

	signature, err := signingKey.Sign(rand.Reader, tokenHash.Sum(nil), nil)
	if err != nil {
		return "", terrors.Wrap(err, nil)
	}

	token := fmt.Sprintf("%s_%s", saltedExpiry, hex.EncodeToString(signature))
	return token, nil
}

func VerifyToken(encodedToken string) (bool, error) {
	token, err := url.QueryUnescape(encodedToken)
	if err != nil {
		return false, terrors.BadRequest("invalid_token", fmt.Sprintf("Encoded authentication token is malformed: %v", err), nil)
	}

	components := strings.Split(token, "_")
	if len(components) != 3 {
		return false, terrors.BadRequest("invalid_token", "Authentication token is malformed", nil)
	}

	declaredExpiry, err := strconv.ParseInt(components[0], 10, 64)
	if err != nil {
		return false, terrors.BadRequest("invalid_token", fmt.Sprintf("Authentication token expiry is not a valid UNIX timestamp: %v", err), nil)
	}
	if time.Unix(declaredExpiry, 0).Before(time.Now()) {
		return false, terrors.Forbidden("token_expired", "Authentication token has expired", nil)
	}

	signature, err := hex.DecodeString(components[2])
	if err != nil {
		return false, terrors.BadRequest("invalid_token", fmt.Sprintf("Authentication token has invalid signature: %v", err), nil)
	}
	ecdsaSignature := &ECDSASignature{}
	_, err = asn1.Unmarshal(signature, ecdsaSignature)
	if err != nil {
		return false, terrors.BadRequest("invalid_token", fmt.Sprintf("Authentication token has unparsable signature: %v", err), nil)
	}

	saltedExpiry := []byte(fmt.Sprintf("%s_%s", components[0], components[1]))
	tokenHash := sha256.New()
	_, err = tokenHash.Write([]byte(saltedExpiry))
	if err != nil {
		return false, terrors.InternalService("", fmt.Sprintf("Error rehashing token: %v", err), nil)
	}

	validSignature := ecdsa.Verify(&signingKey.PublicKey, tokenHash.Sum(nil), ecdsaSignature.R, ecdsaSignature.S)
	return validSignature, nil
}
