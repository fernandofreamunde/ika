package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	pw := "testPassword123!"
	hashed, err := HashPassword(pw)
	if err != nil {
		t.Fatalf("hashing password failed: %v", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashed), []byte(pw))
	if err != nil {
		t.Fatalf("hash does not match password failed: %v", err)
	}
}

func TestMakingJwt(t *testing.T) {
	userId := uuid.New()
	secret := "topSecret"
	expiresIn := time.Duration(5 * time.Minute)

	token, err := MakeJWT(userId, secret, expiresIn)
	if err != nil {
		t.Fatalf("Failed to make JWT: %v", err)
	}

	if token == "" {
		t.Fatalf("Expected Token to be a string and is empty!")
	}
}

func TestValidatingJwt(t *testing.T) {
	userId := uuid.New()
	secret := "topSecret"
	expiresIn := time.Duration(5 * time.Minute)

	token, err := MakeJWT(userId, secret, expiresIn)
	if err != nil {
		t.Fatalf("Failed to make JWT: %v", err)
	}

	id, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("Failed to validate JWT: %v", err)
	}

	if userId != id {
		t.Fatal("returning userId is different from userID used to create the token")
	}
}

func TestValidatingJwtWithWrongSecretFails(t *testing.T) {
	userId := uuid.New()
	secret := "topSecret"
	expiresIn := time.Duration(5 * time.Minute)

	token, err := MakeJWT(userId, secret, expiresIn)
	if err != nil {
		t.Fatalf("Failed to make JWT: %v", err)
	}

	_, err = ValidateJWT(token, "notTheSameSecret")
	if err == nil {
		t.Fatalf("Token was Validated With a different secret")
	}
}

func TestValidatingExpieredJwtFails(t *testing.T) {
	userId := uuid.New()
	secret := "topSecret"
	expiresIn := time.Duration(-5 * time.Minute)

	token, err := MakeJWT(userId, secret, expiresIn)
	if err != nil {
		t.Fatalf("Failed to make JWT: %v", err)
	}

	_, err = ValidateJWT(token, secret)
	if err == nil {
		t.Fatalf("Expiered Token was Validated.")
	}
}

func TestGetBearerToken(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer MY_FAKE_TOKEN")

	tokenString, err := GetBearerToken(headers)
	if err != nil {
		t.Fatal("Failed to retrieve token from header")
	}

	expected := "MY_FAKE_TOKEN"
	if tokenString != expected {
		t.Fatalf("Returned a different token than provided! expected: '%s' provided: '%s'", expected, tokenString)
	}
}

func TestGetBearerTokenFailsWhenHeaderMissing(t *testing.T) {
	headers := http.Header{}

	_, err := GetBearerToken(headers)
	if err == nil {
		t.Fatal("Failed to throw error when header not set")
	}
}

func TestGetBearerTokenFailsWhenHeaderHasNoBearer(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "MY_FAKE_TOKEN")

	_, err := GetBearerToken(headers)
	if err == nil {
		t.Fatal("Failed to throw error when 'Bearer ' is not present in the header")
	}
}
