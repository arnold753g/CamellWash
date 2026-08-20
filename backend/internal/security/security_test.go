package security

import "testing"

func TestPasswordPolicyAndHash(t *testing.T) {
	if ValidatePassword("too-short") == nil {
		t.Fatal("short password accepted")
	}
	hash, err := HashPassword("correct horse battery staple", 10)
	if err != nil {
		t.Fatal(err)
	}
	if !CheckPassword("correct horse battery staple", hash) || CheckPassword("wrong password", hash) {
		t.Fatal("password verification failed")
	}
}

func TestOpaqueSessionAndCSRF(t *testing.T) {
	token, err := RandomToken()
	if err != nil || len(token) != 64 {
		t.Fatalf("token=%q err=%v", token, err)
	}
	secret := "0123456789abcdef0123456789abcdef"
	csrf, err := CSRFToken(secret, token)
	if err != nil || !ValidCSRFToken(secret, token, csrf) || ValidCSRFToken(secret, token, csrf+"00") {
		t.Fatal("CSRF validation failed")
	}
}
