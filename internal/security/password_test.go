package security

import "testing"

func TestPasswordHash(t *testing.T) {
	plainPassword := "password"
	hashedPassword, err := HashPassword(plainPassword)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	if isValid := ValidatePasswordHash(hashedPassword, plainPassword); !isValid {
		t.Fatalf("expected true got %v", isValid)
	}
}

func TestPassword_WrongPassword(t *testing.T) {
	plainPassword := "password"
	passwordHash, err := HashPassword(plainPassword)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	if isValid := ValidatePasswordHash(passwordHash, "wrong-password"); isValid {
		t.Fatalf("expected false got %v", isValid)
	} 

}

func TestPasswordHash_InvalidHash(t *testing.T) {
	plainPassword := "password"

	if isValid := ValidatePasswordHash("wrong-hash", plainPassword); isValid {
		t.Fatalf("expected false got %v", isValid)
	}
}
