package auth_test

import (
	"testing"

	"github.com/t0mer/go-certi/internal/auth"
)

func TestPasswordHashAndVerify(t *testing.T) {
	svc := auth.New("jwt-secret-for-test")

	hash, err := svc.HashPassword("hunter2")
	if err != nil {
		t.Fatal(err)
	}
	if hash == "hunter2" {
		t.Fatal("hash must not equal plaintext")
	}
	if !svc.CheckPassword("hunter2", hash) {
		t.Error("CheckPassword should return true for correct password")
	}
	if svc.CheckPassword("wrong", hash) {
		t.Error("CheckPassword should return false for wrong password")
	}
}

func TestAPITokenRoundTrip(t *testing.T) {
	svc := auth.New("secret")

	tok, err := svc.GenerateAPIToken()
	if err != nil {
		t.Fatal(err)
	}
	if len(tok) < 40 {
		t.Errorf("token too short: %q", tok)
	}
}

func TestJWTRoundTrip(t *testing.T) {
	svc := auth.New("test-secret")

	signed, err := svc.IssueJWT("admin")
	if err != nil {
		t.Fatal(err)
	}
	subject, err := svc.VerifyJWT(signed)
	if err != nil {
		t.Fatalf("VerifyJWT: %v", err)
	}
	if subject != "admin" {
		t.Errorf("subject = %q, want admin", subject)
	}
}

func TestJWTInvalid(t *testing.T) {
	svc := auth.New("secret")
	_, err := svc.VerifyJWT("not.a.token")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}
