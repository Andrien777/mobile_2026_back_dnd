package auth

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/ecdsafile"
)

const PrivateKey = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIN2dALnjdcZaIZg4QuA6Dw+kxiSW502kJfmBN3priIhPoAoGCCqGSM49
AwEHoUQDQgAE4pPyvrB9ghqkT1Llk0A42lixkugFd/TBdOp6wf69O9Nndnp4+HcR
s9SlG/8hjB2Hz42v4p3haKWv3uS1C6ahCQ==
-----END EC PRIVATE KEY-----`

const KeyID = `fake-key-id`
const FakeIssuer = "fake-issuer"
const FakeAudience = "example-users"
const PermissionsClaim = "perm"

type FakeAuthenticator struct {
	PrivateKey *ecdsa.PrivateKey
	KeySet     jwk.Set
}

var _ JWSValidator = (*FakeAuthenticator)(nil)

func NewFakeAuthenticator() (*FakeAuthenticator, error) {
	privKey, err := ecdsafile.LoadEcdsaPrivateKey([]byte(PrivateKey))
	if err != nil {
		return nil, fmt.Errorf("loading PEM private key: %w", err)
	}

	set := jwk.NewSet()
	pubKey := jwk.NewECDSAPublicKey()

	err = pubKey.FromRaw(&privKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("parsing jwk key: %w", err)
	}

	err = pubKey.Set(jwk.AlgorithmKey, jwa.ES256)
	if err != nil {
		return nil, fmt.Errorf("setting key algorithm: %w", err)
	}

	err = pubKey.Set(jwk.KeyIDKey, KeyID)
	if err != nil {
		return nil, fmt.Errorf("setting key ID: %w", err)
	}

	set.Add(pubKey)

	return &FakeAuthenticator{PrivateKey: privKey, KeySet: set}, nil
}

func (f *FakeAuthenticator) ValidateJWS(jwsString string) (jwt.Token, error) {
	return jwt.Parse([]byte(jwsString), jwt.WithKeySet(f.KeySet),
		jwt.WithAudience(FakeAudience), jwt.WithIssuer(FakeIssuer))
}

func (f *FakeAuthenticator) SignToken(t jwt.Token) ([]byte, error) {
	hdr := jws.NewHeaders()
	if err := hdr.Set(jws.AlgorithmKey, jwa.ES256); err != nil {
		return nil, fmt.Errorf("setting algorithm: %w", err)
	}
	if err := hdr.Set(jws.TypeKey, "JWT"); err != nil {
		return nil, fmt.Errorf("setting type: %w", err)
	}
	if err := hdr.Set(jws.KeyIDKey, KeyID); err != nil {
		return nil, fmt.Errorf("setting Key ID: %w", err)
	}
	return jwt.Sign(t, jwa.ES256, f.PrivateKey, jwt.WithHeaders(hdr))
}

func (f *FakeAuthenticator) CreateJWSWithClaims(claims []string, username string) ([]byte, error) {
	t := jwt.New()
	err := t.Set(jwt.IssuerKey, FakeIssuer)
	if err != nil {
		return nil, fmt.Errorf("setting issuer: %w", err)
	}
	err = t.Set(jwt.AudienceKey, FakeAudience)
	if err != nil {
		return nil, fmt.Errorf("setting audience: %w", err)
	}
	err = t.Set(PermissionsClaim, claims)
	if err != nil {
		return nil, fmt.Errorf("setting permissions: %w", err)
	}
	err = t.Set(jwt.SubjectKey, username)
	if err != nil {
		return nil, fmt.Errorf("setting username: %w", err)
	}
	return f.SignToken(t)
}
