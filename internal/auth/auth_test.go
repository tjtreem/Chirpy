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



func TestGetBearerToken(t *testing.T) {
	mk := func(v string) http.Header {
	    h := make(http.Header)
	    if v != "" {
		h.Set("Authorization", v)
	    }
	    return h
	}

	tests := [] struct{
	    name	string
	    hdr 	http.Header
	    token	string
	    wantErr	bool
	}{
	    {"no header", mk(""), "", true},
	    {"wrong scheme", mk("Basic abc123"), "", true},
	    {"bearer no token", mk("Bearer "), "", true},
	    {"bearer spaces only", mk("Bearer    "), "", true},
	    {"bearer tab", mk("Bearer\tabc123"), "abc123", false},
	    {"bearer spaced", mk("Bearer  abc123  "), "abc123", false},
	    {"bearer normal", mk("Bearer abc123"), "abc123", false},
	}

	for _, tc := range tests {
	    t.Run(tc.name, func(t *testing.T) {
		got, err := GetBearerToken(tc.hdr)
		if (err != nil) != tc.wantErr {
		    t.Fatalf("err=%v wantErr=%v", err, tc.wantErr)
		}
		if got != tc.token {
		    t.Fatalf("token got=%q want=%q", got, tc.token)
		}
	    })
	}
}





	
	












