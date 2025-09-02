package auth

import (
	"fmt"
	"time"
	"errors"
	"strings"
	"net/http"
	"crypto/rand"
	"encoding/hex"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
)


func HashPassword(password string) (string, error) {
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
	    return "", fmt.Errorf("Unable to hash password: %w", err)
	}
	return string(hashedPwd), nil
}




func CheckPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
	    return fmt.Errorf("Unable to verify hash: %w", err)
	}

	return nil

}



func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	now := time.Now().UTC()
	expires := now.Add(expiresIn)



	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
	    Issuer:	"chirpy",
	    IssuedAt:	jwt.NewNumericDate(now),
	    ExpiresAt:	jwt.NewNumericDate(expires),
	    Subject:	userID.String(),
    	})

	return token.SignedString([]byte(tokenSecret))

}



func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}

	tok, err := jwt.ParseWithClaims(
            tokenString,
            claims,
            func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}
                return []byte(tokenSecret), nil
            },
	)

	if err != nil {
            return uuid.Nil, err
    	}
	
	if !tok.Valid {
	    return uuid.Nil, fmt.Errorf("invalid token")
	}

    	userIDString, err := claims.GetSubject()
    	if err != nil {
            return uuid.Nil, err
    	}

    	userID, err := uuid.Parse(userIDString)
    	if err != nil {
            return uuid.Nil, err
    	}

    	return userID, nil

}



func GetBearerToken(headers http.Header) (string, error) {
	val := headers.Get("Authorization")
	if !strings.HasPrefix(val, "Bearer ") {
	    return "", errors.New("missing or invalid authorization header")
	}

	token := strings.TrimSpace(strings.TrimPrefix(val, "Bearer "))
	
	if token == "" {
	    return "", errors.New("missing or invalid authorization header")
	}

	return token, nil
}



func MakeRefreshToken() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
	    newErr := fmt.Errorf("Unable to create refresh token: %v", err)
	    return "", newErr
	}
	
	encodedStr := hex.EncodeToString(key)

	return encodedStr, nil

}



func GetAPIKey(h http.Header) (string, error) {
	auth := h.Get("Authorization")
	if auth == "" {
	    return "", errors.New("missing or invalid authorization header")
	}

	const prefix = "ApiKey "

	if !strings.HasPrefix(auth, prefix) {
	    return "", errors.New("missing or invalid authorization header")
	}

	key := strings.TrimSpace(strings.TrimPrefix(auth, prefix))
	if key == "" {
	    return "", errors.New("missing or invalid authorization header")
	}

	return key, nil

}












