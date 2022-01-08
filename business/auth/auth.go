package auth

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
)

type KeyLookup interface {
	PrivateKey(kid string) (*rsa.PrivateKey, error)
	PublicKey(kid string) (*rsa.PublicKey, error)
}

type Auth struct {
	activeKID string
	keyLookup KeyLookup
	method    jwt.SigningMethod
	keyFunc   func(t *jwt.Token) (interface{}, error)
	parser    *jwt.Parser
}

func New(activeKID string, keyLookup KeyLookup) (*Auth, error) {
	_, err := keyLookup.PrivateKey(activeKID)
	if err != nil {
		return nil, errors.New("active KID does not exist in store")
	}

	method := jwt.GetSigningMethod("RS256")
	if method == nil {
		return nil, errors.New("configuring algorithm RS256")
	}

	keyFunc := func(t *jwt.Token) (interface{}, error) {
		kid, ok := t.Header["kid"]
		if !ok {
			return nil, errors.New("missing key id (kid) in token header")
		}
		kidID, ok := kid.(string)
		if !ok {
			return nil, errors.New("user token key id (kid) must be string")
		}

		return keyLookup.PublicKey(kidID)
	}
	parser := jwt.NewParser(jwt.WithValidMethods([]string{"RS256"}))

	a := Auth{
		activeKID: activeKID,
		keyLookup: keyLookup,
		method:    method,
		keyFunc:   keyFunc,
		parser:    parser,
	}
	return &a, nil
}

func (a *Auth) GenerateToken(claims Claims) (string, error) {
	token := jwt.NewWithClaims(a.method, claims)
	token.Header["kid"] = a.activeKID

	privateKey, err := a.keyLookup.PrivateKey(a.activeKID)
	if err != nil {
		return "", errors.New("kid lookup failed")
	}

	str, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}

	return str, nil
}

func (a *Auth) ValidateToken(tokenStr string) (Claims, error) {
	var claims Claims
	token, err := a.parser.ParseWithClaims(tokenStr, &claims, a.keyFunc)
	if err != nil {
		return Claims{}, fmt.Errorf("parsing token: %w", err)
	}

	if !token.Valid {
		return Claims{}, errors.New("invalid token")
	}

	return claims, nil
}