package api

// Auth keys are signed by a trusted Minecraft server that has verified the user's identity.

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/golang-jwt/jwt/v4"

	"schemastash/global"

	"go.mongodb.org/mongo-driver/bson"

	"context"

	"fmt"
)

type PublicKey struct {
	Key string
	ID  string
}

func VerifyJWT(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		pubkeys := []PublicKey{}

		cursor, err := global.Mongo.Collection("public_keys").Find(context.TODO(), bson.D{{}})
		if err != nil {
			return nil, err
		}

		cursor.All(context.Background(), &pubkeys)

		for _, pubkeyPEM := range pubkeys {
			block, _ := pem.Decode([]byte(pubkeyPEM.Key))
			if block == nil {
				return nil, fmt.Errorf("no valid public keys found")
			}

			pubkey, err := x509.ParsePKIXPublicKey(block.Bytes)
			if err != nil {
				return nil, err
			}

			if _, ok := pubkey.(*rsa.PublicKey); ok {
				return pubkey, nil
			}
		}

		return nil, fmt.Errorf("no valid public keys found")
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
