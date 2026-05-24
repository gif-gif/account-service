package security

import "testing"

func TestEncryptDecryptRoundTrip(t *testing.T) {
	codec, err := NewCredentialCodec("0123456789abcdef0123456789abcdef")
	if err != nil {
		t.Fatalf("NewCredentialCodec() error = %v", err)
	}

	ciphertext, err := codec.Encrypt("secret-token")
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}
	if ciphertext == "secret-token" {
		t.Fatal("ciphertext must not equal plaintext")
	}

	plaintext, err := codec.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}
	if plaintext != "secret-token" {
		t.Fatalf("plaintext = %q, want secret-token", plaintext)
	}
}

func TestDecryptRejectsWrongKey(t *testing.T) {
	codec, err := NewCredentialCodec("0123456789abcdef0123456789abcdef")
	if err != nil {
		t.Fatalf("NewCredentialCodec() error = %v", err)
	}
	wrongCodec, err := NewCredentialCodec("abcdef0123456789abcdef0123456789")
	if err != nil {
		t.Fatalf("NewCredentialCodec(wrong) error = %v", err)
	}

	ciphertext, err := codec.Encrypt("secret-token")
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	if _, err := wrongCodec.Decrypt(ciphertext); err == nil {
		t.Fatal("expected wrong key decrypt to fail")
	}
}

func TestAPIKeyGenerateHashAndVerify(t *testing.T) {
	key, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey() error = %v", err)
	}
	if len(key) < 40 {
		t.Fatalf("generated API key too short: %q", key)
	}

	hash := HashAPIKey(key)
	if hash == key {
		t.Fatal("hash must not equal plaintext API key")
	}
	if !VerifyAPIKey(key, hash) {
		t.Fatal("expected generated API key to verify")
	}
	if VerifyAPIKey("wrong-key", hash) {
		t.Fatal("expected wrong API key to fail verification")
	}
}

func TestPasswordHashAndVerify(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if hash == "correct horse battery staple" {
		t.Fatal("hash must not equal plaintext password")
	}
	if !VerifyPassword("correct horse battery staple", hash) {
		t.Fatal("expected password to verify")
	}
	if VerifyPassword("wrong password", hash) {
		t.Fatal("expected wrong password to fail verification")
	}
}
