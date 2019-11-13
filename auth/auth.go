package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/harrisonturton/api/config"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"io/ioutil"
	"net/http"
	"time"
)

// AccessTokenClaims are the claims held within a signed
// JWT access_token. It includes the refresh_token, which
// must be valid when refreshing the access_token.
type AccessTokenClaims struct {
	RefreshToken string
	*jwt.StandardClaims
}

// RefreshTokenClaims are the claims held in the refresh_token.
// A refresh token is longer-living, and helps us spread out
// database hits from every API request to once every 15 minutes.
type RefreshTokenClaims struct {
	Username string
	*jwt.StandardClaims
}

// Authenticator signs and verifies JWT tokens, and
// refreshes access_tokens and issues refresh_tokens.
type Authenticator struct {
	publicKey            *rsa.PublicKey
	privateKey           *rsa.PrivateKey
	refreshTokenLifespan int // in seconds
	accessTokenLifespan  int
}

func NewAuthenticator(config config.Auth) (*Authenticator, error) {
	rawPublicKey, err := ioutil.ReadFile(config.PublicKeyPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read public key file")
	}
	rawPrivateKey, err := ioutil.ReadFile(config.PrivateKeyPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read private key file")
	}
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(rawPublicKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse RSA public key from PEM")
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(rawPrivateKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse RSA private key from PEM")
	}
	return &Authenticator{publicKey, privateKey, config.RefreshTokenLifespan, config.AccessTokenLifespan}, nil
}

func HashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), 14)
}

func CheckPasswordHash(password string, hash []byte) bool {
	err := bcrypt.CompareHashAndPassword(hash, []byte(password))
	return err == nil
}

func GenerateToken(byteLength int) (string, error) {
	b := make([]byte, byteLength)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// NewRefreshKey will generate a new refresh_key, which verifies
// that the user was present at time of singing. This is a longer-lived
// key, and must be presented when refreshing an access_token, which is short-lived.
// This is stored in the database, and can be revoked. An attacker will only have
// the lifespan of the access_token to access API resources.
func (auth *Authenticator) NewRefreshToken(username string) (string, error) {
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodRS512, &RefreshTokenClaims{
		username,
		&jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Second * time.Duration(auth.refreshTokenLifespan)).Unix(),
		},
	})
	return refreshToken.SignedString(auth.privateKey)
}

func (auth *Authenticator) NewAccessToken(refreshToken string) (string, error) {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodRS512, &AccessTokenClaims{
		refreshToken,
		&jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Second * time.Duration(auth.accessTokenLifespan)).Unix(),
		},
	})
	return accessToken.SignedString(auth.privateKey)
}

func (auth *Authenticator) RefreshAccessToken(rawAccessToken string) (string, error) {
	accessTokenClaims, err := auth.parseAccessToken(rawAccessToken)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse access token")
	}
	refreshTokenClaims, err := auth.parseRefreshToken(accessTokenClaims.RefreshToken)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse refresh token")
	}
	newRefreshToken, err := auth.NewRefreshToken(refreshTokenClaims.Username)
	if err != nil {
		return "", errors.Wrap(err, "failed to create new refresh token")
	}
	return auth.NewAccessToken(newRefreshToken)
}

// parseAccessToken will attempt to parse an access_token. If it fails
// due to expiry, the error is ignored.
func (auth *Authenticator) parseAccessToken(rawAccessToken string) (*AccessTokenClaims, error) {
	var claims AccessTokenClaims
	_, err := jwt.ParseWithClaims(rawAccessToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return auth.publicKey, nil
	})
	// AccessToken might be expired, but that's OK if the refresh token
	// is still valid. Need to handle this case with error handling.
	validationErr, ok := err.(*jwt.ValidationError)
	// If parsing failed because of a reason that's not expiry
	if err != nil && (!ok || validationErr.Errors&jwt.ValidationErrorExpired == 0) {
		return nil, errors.Wrap(err, "failed to validate access token")
	}
	// Expired or still valid
	return &claims, nil
}

func (auth *Authenticator) VerifyAccessToken(rawAccessToken string) (AccessTokenClaims, error) {
	var claims AccessTokenClaims
	_, err := jwt.ParseWithClaims(rawAccessToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return auth.publicKey, nil
	})
	return claims, err
}

func (auth *Authenticator) parseRefreshToken(rawRefreshToken string) (*RefreshTokenClaims, error) {
	var claims RefreshTokenClaims
	_, err := jwt.ParseWithClaims(rawRefreshToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return auth.publicKey, nil
	})
	return &claims, err
}

// Secure is HTTP middleware that will block any requests without a valid token
func (auth *Authenticator) Secure(extractor func(r *http.Request) (string, error)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := extractor(r)
			if err != nil {
				fmt.Printf("Failed to extract: %v\n", err)
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
			if _, err := auth.VerifyAccessToken(token); err != nil {
				fmt.Printf("Failed to verify: %v\n", err)
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
