package middlewares

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/spf13/viper"
)

type GooglePublicKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
	Use string `json:"use"`
	Alg string `json:"alg"`
}

type GooglePublicKeysResponse struct {
	Keys []GooglePublicKey `json:"keys"`
}

func fetchGooglePublicKey() ([]GooglePublicKey, error) {
	response, err := http.Get("https://www.googleapis.com/oauth2/v3/certs")
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// to JSON
	var googlePublicKeys GooglePublicKeysResponse
	err = json.Unmarshal(body, &googlePublicKeys)
	if err != nil {
		return nil, err
	}

	return googlePublicKeys.Keys, nil
}

func parsePublicKey(nStr, eStr string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return nil, err
	}

	e := 0
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}

	pubKey := &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: e,
	}
	return pubKey, nil
}

/**
 * AuthenticateHandler
 * @Description: OAuth2.0 Authentication Middleware
 * @param c
 */
func AuthenticateHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Authorization") == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// Split the Authorization header
		authHeader := strings.Split(c.GetHeader("Authorization"), " ")
		if len(authHeader) != 2 || authHeader[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// verify if jwt signature public key is valid
		token, err := jwt.Parse(authHeader[1], func(token *jwt.Token) (interface{}, error) {
			// Check if the signing method is valid
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			kid := token.Header["kid"].(string)
			if _, ok := token.Header["kid"]; !ok {
				return nil, fmt.Errorf("kid not found")
			}

			// Check if the kid is valid
			googlePublicKeys, err := fetchGooglePublicKey()
			if err != nil {
				return nil, err
			}

			for _, key := range googlePublicKeys {
				if key.Kid == kid {
					pubKey, err := parsePublicKey(key.N, key.E)
					if err != nil {
						return nil, err
					}

					return pubKey, nil
				}
			}

			return nil, fmt.Errorf("kid not found")
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// validate the claims
		// aud, iss
		if claims["aud"] != viper.GetString("OAUTH2.CLIENT_ID") || claims["iss"] != "https://accounts.google.com" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		c.Set("email", claims["email"])
		c.Set("username", claims["name"])

		c.Next()
	}
}
