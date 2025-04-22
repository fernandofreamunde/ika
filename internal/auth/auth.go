package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fernandofreamunde/ika/internal/db"
	"github.com/fernandofreamunde/ika/internal/user"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type LoginResponse struct {
	User         user.User `json:"user"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

func HashPassword(pw string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)

	return string(bytes), err
}

func CheckPasswordHash(hash, pw string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw))
}

func AuthenticateUser(u db.User, ctx context.Context, dbq func() *db.Queries) (LoginResponse, error) {

	expiresIn := 60 * 60
	// TODO: .env APP_SECRET
	jwt, _ := MakeJWT(u.ID, "IneedAnAppSecret", time.Duration(expiresIn)*time.Second)
	refreshToken, _ := MakeRefreshToken()

	_, err := dbq().CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		Token:     refreshToken,
		UpdatedAt: time.Now(),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Duration(60 * 24 * time.Hour)),
		UserID:    uuid.NullUUID{UUID: u.ID, Valid: true},
	})

	if err != nil {
		return LoginResponse{}, fmt.Errorf("Could not create refresh token.")
	}

	return LoginResponse{
		User: user.User{
			ID:        u.ID,
			Email:     u.Email,
			Nickname:  u.Nickname,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		},
		Token:        jwt,
		RefreshToken: refreshToken,
	}, nil
}

func RefreshJWT(h http.Header, ctx context.Context, q func() *db.Queries) (string, error) {

	tokenString, _ := GetBearerToken(h)
	token, err := q().GetRefreshToken(ctx, tokenString)

	if err != nil {
		return "", fmt.Errorf("Unauthorized.")
	}

	if token.ExpiresAt.Before(time.Now()) || token.RevokedAt.Valid {
		return "", fmt.Errorf("Unauthorized.")
	}

	expiresIn := 60 * 60
	jwt, err := MakeJWT(token.UserID.UUID, "IneedAnAppSecret", time.Duration(expiresIn)*time.Second)
	if err != nil {
		return "", fmt.Errorf("Could not create JWT.")
	}
	return jwt, nil
}

func RvokeRefreshToken(h http.Header, ctx context.Context, q func() *db.Queries) {

	tokenString, _ := GetBearerToken(h)
	token, _ := q().GetRefreshToken(ctx, tokenString)

	q().RevokeRefreshToken(ctx, token.Token)
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	})

	return t.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {

	t, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpcted signing method: %v", token.Header["alg"])
		}
		return []byte(tokenSecret), nil
	})

	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to parse token: %v", err)
	}

	if !t.Valid {
		return uuid.UUID{}, fmt.Errorf("invalid token")
	}

	claims, ok := t.Claims.(*jwt.RegisteredClaims)

	if !ok {
		return uuid.UUID{}, fmt.Errorf("invalid claims")
	}

	userId, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("invalid uuid")
	}

	return userId, nil
}

func GetBearerToken(headers http.Header) (string, error) {

	header := headers.Get("Authorization")

	if header == "" {
		return "", fmt.Errorf("Authorization header not set")
	}

	if !strings.Contains(header, "Bearer ") {
		return "", fmt.Errorf("Authorization header is not 'Bearer' token")
	}
	return header[7:], nil
}

func GetApiKey(headers http.Header) (string, error) {

	header := headers.Get("Authorization")

	if header == "" {
		return "", fmt.Errorf("Authorization header not set")
	}

	if !strings.Contains(header, "ApiKey ") {
		return "", fmt.Errorf("Authorization header does not contain 'ApiKey'")
	}
	return header[7:], nil
}

func MakeRefreshToken() (string, error) {

	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	t := hex.EncodeToString([]byte(key))
	return t, nil
}
