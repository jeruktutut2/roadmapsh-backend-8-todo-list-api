package mockhelpers

import "github.com/stretchr/testify/mock"

type JwtHelperMock struct {
	Mock mock.Mock
}

func (helper *JwtHelperMock) GenerateAccessToken(id int, name string, email string, jwtAccessTokenTime int, secret string) (accessToken string, err error) {
	arguments := helper.Mock.Called(id, name, email, jwtAccessTokenTime, secret)
	return arguments.Get(0).(string), arguments.Error(1)
}

func (helper *JwtHelperMock) GenerateRefreshToken(id int, jwtRefreshTokenTime int, secret string) (refreshToken string, err error) {
	arguments := helper.Mock.Called(id, jwtRefreshTokenTime, secret)
	return arguments.Get(0).(string), arguments.Error(1)
}
