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
