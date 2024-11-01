package mockhelpers

import (
	"github.com/stretchr/testify/mock"
)

type BcryptHelperMock struct {
	Mock mock.Mock
}

func (helper *BcryptHelperMock) GenerateFromPassword(password []byte, cost int) ([]byte, error) {
	arguments := helper.Mock.Called(password, cost)
	return arguments.Get(0).([]byte), arguments.Error(1)
}

func (helper *BcryptHelperMock) CompareHashAndPassword(hashedPassword []byte, password []byte) error {
	arguments := helper.Mock.Called(hashedPassword, password)
	return arguments.Error(0)
}
