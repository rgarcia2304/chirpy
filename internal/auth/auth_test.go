package auth

import(
	"testing"
	"time"
	"github.com/google/uuid"
)

func TestJWTCreation(t *testing.T){
	userId := uuid.New()
	exp, err := time.ParseDuration("30m")
	if err != nil{
		t.Fatalf("failed to parse duration: %v", err)
	}

	jwtString, err := MakeJWT(userId, "hello", exp)
	if err != nil{
		t.Errorf(`Error creating the JWT Token`)
	}

	_, err = ValidateJWT(jwtString, "hello")

	
}

func TestJWT_WrongSecretFails(t *testing.T) {
	userID := uuid.New()

	tokenStr, err := MakeJWT(userID, "secret1", 30*time.Minute)
	if err != nil {
		t.Fatalf("MakeJWT error: %v", err)
	}

	got, err := ValidateJWT(tokenStr, "secret2")
	if err == nil {
		t.Fatalf("expected error, got nil (uuid=%v)", got)
	}
}

func TestJWT_ExpiredFails(t *testing.T) {
	userID := uuid.New()
	secret := "hello"

	tokenStr, err := MakeJWT(userID, secret, -1*time.Minute)
	if err != nil {
		t.Fatalf("MakeJWT error: %v", err)
	}

	_, err = ValidateJWT(tokenStr, secret)
	if err == nil {
		t.Fatalf("expected error for expired token, got nil")
	}
}
