package auth

import(
	"time"
	"testing"
	"github.com/google/uuid"
)



func TestCreateToken(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"
	expiresIn := time.Hour

	token, err := MakeJWT(userID, secret, expiresIn)
	if err !=nil {
    	    t.Fatalf("Make JWT() failed: %v", err)
	}

	validateUserID, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWT failed: %v", err)
	}

	if validateUserID != userID {
    	    t.Errorf("expected %v, got %v", userID, validateUserID)
	}

}



func TestExpiredToken(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"
	expiresIn := time.Nanosecond

	token, err := MakeJWT(userID, secret, expiresIn)
	if err != nil {
	    t.Fatalf("MakeJWT() failed: %v", err)
	}

	time.Sleep(2 * time.Millisecond)

	validateUserID, err := ValidateJWT(token, secret)
	if err == nil {
	    t.Errorf("expected error for expired token, got none (id=%v)", validateUserID)
	}

}




func TestWrongSecretToken(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"
	expiresIn := time.Hour

	token, err := MakeJWT(userID, secret, expiresIn)
	if err !=nil {
    	    t.Fatalf("MakeJWT() failed: %v", err)
	}

	secret = "failed-secret-test"

	validateUserID, err := ValidateJWT(token, secret)
	if err == nil {
    	    t.Fatalf("Expected error with wrong secret, got none (id=%v)", validateUserID)
	}

}


