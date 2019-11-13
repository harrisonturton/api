package main

import (
	"crypto/rsa"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"time"
)

type Claims struct {
	Username string
	*jwt.StandardClaims
}

var (
	publicKeyPath  = "/Users/harrisonturton/Documents/dev/go/src/github.com/harrisonturton/training-api/rsa.app.pub"
	privateKeyPath = "/Users/harrisonturton/Documents/dev/go/src/github.com/harrisonturton/training-api/rsa.app"

	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
)

func main() {
	rawPrivate, err := ioutil.ReadFile(privateKeyPath)
	fatal(err)
	rawPublic, err := ioutil.ReadFile(publicKeyPath)
	fatal(err)

	publicKey, err = jwt.ParseRSAPublicKeyFromPEM(rawPublic)
	fatal(err)
	privateKey, err = jwt.ParseRSAPrivateKeyFromPEM(rawPrivate)
	fatal(err)

	token := jwt.NewWithClaims(jwt.SigningMethodRS512, &Claims{
		"harrisonturton",
		&jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Second * 3).Unix(),
		},
	})
	fatal(err)
	fmt.Println("Sleeping")
	time.Sleep(time.Second * 5)
	fmt.Println("Wake up")

	tokenString, err := token.SignedString(privateKey)
	fatal(err)

	parsedToken, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})
	validationErr, ok := err.(*jwt.ValidationError)
	if !ok {
		fatal(err)
	}
	if ok && validationErr.Errors&(jwt.ValidationErrorExpired) != 0 {
		fmt.Printf("%v\n", validationErr.Errors)
		fmt.Println("Expired!")
	} else {
		fatal(err)
	}

	//} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {}

	fmt.Printf("%v\n", parsedToken.Claims)
	fmt.Println("Done!")
}

func fatal(err error) {
	if err != nil {
		panic(err)
	}
}
