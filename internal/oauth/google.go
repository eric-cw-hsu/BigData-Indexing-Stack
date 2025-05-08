package oauth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
)

type GoogleTokenInfo struct {
	Email    string `json:"email"`
	Username string `json:"username"`
}

type googlePublicKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
	Use string `json:"use"`
	Alg string `json:"alg"`
}

type googlePublicKeysResponse struct {
	Keys []googlePublicKey `json:"keys"`
}

var googlePublicKeysCache []googlePublicKey
var googlePublicKeysCacheTime time.Time

func fetchGooglePublicKey(token string, expectedClientId string) ([]googlePublicKey, error) {
	if len(googlePublicKeysCache) > 0 && time.Since(googlePublicKeysCacheTime) < 1*time.Hour {
		return googlePublicKeysCache, nil
	}

	res, err := http.Get("https://www.googleapis.com/oauth2/v3/certs")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var googlePublicKeys googlePublicKeysResponse
	err = json.Unmarshal(body, &googlePublicKeys)
	if err != nil {
		return nil, err
	}

	googlePublicKeysCache = googlePublicKeys.Keys
	googlePublicKeysCacheTime = time.Now()

	return googlePublicKeys.Keys, nil
}

func parsePublicKey(key googlePublicKey) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		return nil, err
	}

	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	return &rsa.PublicKey{N: n, E: int(e.Int64())}, nil
}

func verifyGoogleToken(token string, expectedClientId string) (*GoogleTokenInfo, error) {
	googlePublicKeys, err := fetchGooglePublicKey(token, expectedClientId)
	if err != nil {
		return nil, err
	}

	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// check if the token is signed with RS256
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// get the kid from the token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("kid not found in token header")
		}

		// find the public key with the matching kid
		for _, key := range googlePublicKeys {
			if key.Kid == kid {
				publicKey, err := parsePublicKey(key)
				if err != nil {
					return nil, err
				}

				return publicKey, nil
			}
		}

		return nil, fmt.Errorf("public key not found")
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		email, ok := claims["email"].(string)
		if !ok {
			return nil, fmt.Errorf("email claim not found")
		}

		audience, ok := claims["aud"].(string)
		if !ok {
			return nil, fmt.Errorf("aud claim not found")
		}

		if audience != expectedClientId {
			return nil, fmt.Errorf("invalid audience")
		}

		if claims["iss"] != "accounts.google.com" && claims["iss"] != "https://accounts.google.com" {
			return nil, fmt.Errorf("invalid issuer")
		}

		username, ok := claims["name"].(string)
		if !ok {
			return nil, fmt.Errorf("username claim not found")
		}

		return &GoogleTokenInfo{Email: email, Username: username}, nil
	}

	return nil, fmt.Errorf("invalid token")
}
