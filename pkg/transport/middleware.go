package transport

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"time"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			fmt.Println("no auth header")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		tokenString := authHeader[7:]
		token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return signingKey, nil
		})

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			fmt.Println(err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		fmt.Println("get claims")
		claims := token.Claims.(*UserClaims)

		ctx := context.WithValue(r.Context(), "claims", claims)

		// Continue request processing
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func generateToken(username int) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	fmt.Println("token")

	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = username
	claims["exp"] = time.Now().Add(expirationTime).Unix()

	signedToken, err := token.SignedString(signingKey)
	if err != nil {
		return "", err
	}

	fmt.Println("token")

	return signedToken, nil
}
