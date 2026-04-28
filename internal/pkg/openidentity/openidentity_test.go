package openidentity

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestVerifyUserTokenOK(t *testing.T) {
	token := signTestUserToken(t, jwt.SigningMethodHS256, map[string]any{
		"userId": "u_10001",
		"name":   "张三",
		"exp":    time.Now().Add(time.Hour).Unix(),
	}, "secret")

	claims, err := VerifyUserToken(token, "secret")
	if err != nil {
		t.Fatalf("expected token to verify: %v", err)
	}
	if claims.UserID != "u_10001" || claims.Name != "张三" {
		t.Fatalf("unexpected claims: %#v", claims)
	}
}

func TestVerifyUserTokenUsesJWTHeaderAlgorithm(t *testing.T) {
	token := signTestUserToken(t, jwt.SigningMethodHS384, map[string]any{
		"userId": "u_10001",
		"name":   "张三",
		"exp":    time.Now().Add(time.Hour).Unix(),
	}, "secret")

	claims, err := VerifyUserToken(token, "secret")
	if err != nil {
		t.Fatalf("expected HS384 token to verify from JWT header: %v", err)
	}
	if claims.UserID != "u_10001" || claims.Name != "张三" {
		t.Fatalf("unexpected claims: %#v", claims)
	}
}

func TestVerifyUserTokenRejectsInvalidSignature(t *testing.T) {
	token := signTestUserToken(t, jwt.SigningMethodHS256, map[string]any{
		"userId": "u_10001",
		"name":   "张三",
		"exp":    time.Now().Add(time.Hour).Unix(),
	}, "secret")

	if _, err := VerifyUserToken(token, "other-secret"); err == nil {
		t.Fatalf("expected invalid signature to fail")
	}
}

func TestVerifyUserTokenRejectsExpiredToken(t *testing.T) {
	token := signTestUserToken(t, jwt.SigningMethodHS256, map[string]any{
		"userId": "u_10001",
		"name":   "张三",
		"exp":    time.Now().Add(-time.Minute).Unix(),
	}, "secret")

	if _, err := VerifyUserToken(token, "secret"); err == nil {
		t.Fatalf("expected expired token to fail")
	}
}

func TestVerifyUserTokenRequiresUserIDAndName(t *testing.T) {
	tests := []map[string]any{
		{"name": "张三", "exp": time.Now().Add(time.Hour).Unix()},
		{"userId": "u_10001", "exp": time.Now().Add(time.Hour).Unix()},
	}
	for _, payload := range tests {
		token := signTestUserToken(t, jwt.SigningMethodHS256, payload, "secret")
		if _, err := VerifyUserToken(token, "secret"); err == nil {
			t.Fatalf("expected payload %#v to fail", payload)
		}
	}
}

func signTestUserToken(t *testing.T, method jwt.SigningMethod, payload map[string]any, secret string) string {
	t.Helper()
	token, err := jwt.NewWithClaims(method, jwt.MapClaims(payload)).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	return token
}
