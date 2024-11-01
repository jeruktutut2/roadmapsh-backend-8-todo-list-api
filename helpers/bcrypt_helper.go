package helpers

import "golang.org/x/crypto/bcrypt"

type BcryptHelper interface {
	GenerateFromPassword(password []byte, cost int) ([]byte, error)
	CompareHashAndPassword(hashedPassword []byte, password []byte) error
}

type BcryptHelperImplementation struct {
}

func NewBcryptHelper() BcryptHelper {
	return &BcryptHelperImplementation{}
}

func (helper *BcryptHelperImplementation) GenerateFromPassword(password []byte, cost int) ([]byte, error) {
	return bcrypt.GenerateFromPassword(password, cost)
}

func (helper *BcryptHelperImplementation) CompareHashAndPassword(hashedPassword []byte, password []byte) error {
	return bcrypt.CompareHashAndPassword(hashedPassword, password)
}
