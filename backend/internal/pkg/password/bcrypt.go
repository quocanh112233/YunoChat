package password

import (
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

// DummyHash is an invalid bcrypt hash to prevent timing attacks.
// It costs exactly the same to "compare" against this hash.
// This hash was generated from a random string and cannot be reversed.
var DummyHash []byte

func init() {
	// Giả lập DummyHash 1 lần lúc startup để không phải generate lại mỗi lần call
	hash, err := bcrypt.GenerateFromPassword([]byte("timing_attack_prevention_dummy_password"), bcryptCost)
	if err != nil {
		panic("failed to generate dummy hash: " + err.Error())
	}
	DummyHash = hash
}

// Hash passwords using bcrypt algorithm with the predefined cost.
func Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	return string(bytes), err
}

// Compare a hashed password with a plain text password.
func Compare(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// PreventTimingAttack performs a dummy bcrypt check to artificially delay response
// for preventing email enumeration attacks via timing.
func PreventTimingAttack(password string) {
	_ = bcrypt.CompareHashAndPassword(DummyHash, []byte(password))
}
