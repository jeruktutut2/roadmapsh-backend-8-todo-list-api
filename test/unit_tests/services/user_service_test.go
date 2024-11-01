package services_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	modelentities "todo-list-api/models/entities"
	modelrequests "todo-list-api/models/requests"
	"todo-list-api/services"
	mockhelpers "todo-list-api/test/unit_tests/helpers/mocks"
	mockrepositories "todo-list-api/test/unit_tests/repositories/mocks"
	mockutils "todo-list-api/test/unit_tests/utils/mocks"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
)

type UserServiceTestSuite struct {
	suite.Suite
	ctx                   context.Context
	options               pgx.TxOptions
	tx                    pgx.Tx
	pool                  *pgxpool.Pool
	errInternalServer     error
	errRowsAffectedNotOne error
	registerRequest       modelrequests.RegisterRequest
	loginRequest          modelrequests.LoginRequest
	user                  modelentities.User
	postgresUtilMock      *mockutils.PostgresUtilMock
	validate              *validator.Validate
	userRepositoryMock    *mockrepositories.UserRepositoryMock
	bcryptHelperMock      *mockhelpers.BcryptHelperMock
	jwtHelperMock         *mockhelpers.JwtHelperMock
	pgxTxMock             *mockutils.PgxTxMock
	userService           services.UserService
}

func TestUserTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

func (sut *UserServiceTestSuite) SetupSuite() {
	sut.T().Log("SetupSuite")
	sut.ctx = context.Background()
	sut.errInternalServer = errors.New("internal server error")
	sut.errRowsAffectedNotOne = errors.New("rows affected not one")
}

func (sut *UserServiceTestSuite) SetupTest() {
	sut.T().Log("SetupTest")
	sut.registerRequest = modelrequests.RegisterRequest{
		Name:     "John Doe",
		Email:    "john@doe.com",
		Password: "password",
	}
	sut.loginRequest = modelrequests.LoginRequest{
		Email:    "john@doe.com",
		Password: "password",
	}
	sut.user = modelentities.User{
		Id:           pgtype.Int4{Valid: true, Int32: 1},
		Name:         pgtype.Text{Valid: true, String: "John Doe"},
		Email:        pgtype.Text{Valid: true, String: "john@doe.com"},
		Password:     pgtype.Text{Valid: true, String: "password"},
		RefreshToken: pgtype.Text{Valid: true, String: "refresh-token"},
	}
	sut.pool = &pgxpool.Pool{}
	sut.postgresUtilMock = new(mockutils.PostgresUtilMock)
	sut.validate = validator.New()
	sut.userRepositoryMock = new(mockrepositories.UserRepositoryMock)
	sut.bcryptHelperMock = new(mockhelpers.BcryptHelperMock)
	sut.jwtHelperMock = new(mockhelpers.JwtHelperMock)
	sut.pgxTxMock = new(mockutils.PgxTxMock)
	sut.userService = services.NewUserService(sut.postgresUtilMock, sut.validate, sut.userRepositoryMock, sut.bcryptHelperMock, sut.jwtHelperMock)
}

func (sut *UserServiceTestSuite) BeforeTest(suiteName, testName string) {
	sut.T().Log("BeforeTest: " + suiteName + " " + testName)
}

func (sut *UserServiceTestSuite) Test01RegisterValidationError() {
	sut.T().Log("Test01RegisterValidationError")
	sut.registerRequest = modelrequests.RegisterRequest{}
	httpCode, accessToken, refreshToken, response := sut.userService.Register(sut.ctx, sut.registerRequest)
	sut.Equal(httpCode, 400)
	sut.Equal(accessToken, "")
	sut.Equal(refreshToken, "")
	sut.NotEqual(response, nil)
}

func (sut *UserServiceTestSuite) Test02RegisterBeginTxError() {
	sut.T().Log("Test02RegisterBeginTxError")
	sut.postgresUtilMock.Mock.On("BeginTx", sut.ctx, sut.options).Return(sut.pgxTxMock, sut.errInternalServer)
	httpCode, accessToken, refreshToken, response := sut.userService.Register(sut.ctx, sut.registerRequest)
	sut.Equal(httpCode, 500)
	sut.Equal(accessToken, "")
	sut.Equal(refreshToken, "")
	sut.NotEqual(response, nil)
}

func (sut *UserServiceTestSuite) Test03RegisterGenerateFromPasswordError() {
	sut.T().Log("Test03RegisterGenerateFromPasswordError")
	sut.postgresUtilMock.Mock.On("BeginTx", sut.ctx, sut.options).Return(sut.pgxTxMock, nil)
	// should return []uint8{}, or there will be error on commit or rollback or the error will not be going to commit or rollback
	sut.bcryptHelperMock.Mock.On("GenerateFromPassword", []byte(sut.registerRequest.Password), bcrypt.DefaultCost).Return([]uint8{}, sut.errInternalServer)
	sut.postgresUtilMock.Mock.On("CommitOrRollback", sut.pgxTxMock, sut.ctx, sut.errInternalServer).Return(nil)
	httpCode, accessToken, refreshToken, response := sut.userService.Register(sut.ctx, sut.registerRequest)
	sut.Equal(httpCode, 500)
	sut.Equal(accessToken, "")
	sut.Equal(refreshToken, "")
	sut.NotEqual(response, nil)
}

func (sut *UserServiceTestSuite) Test04RegisterGenerateFromPasswordErrorCommitOrRollbackError() {
	sut.T().Log("Test04RegisterGenerateFromPasswordErrorCommitOrRollbackError")
	sut.postgresUtilMock.Mock.On("BeginTx", sut.ctx, sut.options).Return(sut.pgxTxMock, nil)
	sut.bcryptHelperMock.Mock.On("GenerateFromPassword", []byte(sut.registerRequest.Password), bcrypt.DefaultCost).Return([]uint8{}, sut.errInternalServer)
	sut.postgresUtilMock.Mock.On("CommitOrRollback", sut.pgxTxMock, sut.ctx, sut.errInternalServer).Return(sut.errInternalServer)
	httpCode, accessToken, refreshToken, response := sut.userService.Register(sut.ctx, sut.registerRequest)
	sut.Equal(httpCode, 500)
	sut.Equal(accessToken, "")
	sut.Equal(refreshToken, "")
	sut.NotEqual(response, nil)
}

func (sut *UserServiceTestSuite) Test05RegisterUserRepositoryCreateError() {
	sut.T().Log("Test05RegisterUserRepositoryCreateError")
	sut.postgresUtilMock.Mock.On("BeginTx", sut.ctx, sut.options).Return(sut.pgxTxMock, nil)
	passwordByte := []byte{112, 97, 115, 115, 119, 111, 114, 100}
	sut.bcryptHelperMock.Mock.On("GenerateFromPassword", []byte(sut.registerRequest.Password), bcrypt.DefaultCost).Return(passwordByte, nil)
	sut.user.Id = pgtype.Int4{Valid: false, Int32: 0}
	sut.user.RefreshToken = pgtype.Text{Valid: false, String: ""}
	var lastInsertedId int
	sut.userRepositoryMock.Mock.On("Create", sut.pgxTxMock, sut.ctx, sut.user).Return(lastInsertedId, sut.errInternalServer)
	sut.postgresUtilMock.Mock.On("CommitOrRollback", sut.pgxTxMock, sut.ctx, sut.errInternalServer).Return(nil)
	httpCode, accessToken, refreshToken, response := sut.userService.Register(sut.ctx, sut.registerRequest)
	sut.Equal(httpCode, 500)
	sut.Equal(accessToken, "")
	sut.Equal(refreshToken, "")
	sut.NotEqual(response, nil)
}

func (sut *UserServiceTestSuite) Test06RegisterAccessTokenError() {
	sut.T().Log("Test06RegisterAccessTokenError")
	sut.postgresUtilMock.Mock.On("BeginTx", sut.ctx, sut.options).Return(sut.pgxTxMock, nil)
	passwordByte := []byte{112, 97, 115, 115, 119, 111, 114, 100}
	sut.bcryptHelperMock.Mock.On("GenerateFromPassword", []byte(sut.registerRequest.Password), bcrypt.DefaultCost).Return(passwordByte, nil)
	sut.user.Id = pgtype.Int4{Valid: false, Int32: 0}
	sut.user.RefreshToken = pgtype.Text{Valid: false, String: ""}
	sut.userRepositoryMock.Mock.On("Create", sut.pgxTxMock, sut.ctx, sut.user).Return(1, nil)
	jwtAccessTokenTime := 15
	sut.user.Id = pgtype.Int4{Valid: true, Int32: 1}
	var accessToken string
	sut.jwtHelperMock.Mock.On("GenerateAccessToken", int(sut.user.Id.Int32), sut.user.Name.String, sut.user.Email.String, jwtAccessTokenTime, "secret").Return(accessToken, sut.errInternalServer)
	sut.postgresUtilMock.Mock.On("CommitOrRollback", sut.pgxTxMock, sut.ctx, sut.errInternalServer).Return(nil)
	httpCode, accessToken, refreshToken, response := sut.userService.Register(sut.ctx, sut.registerRequest)
	sut.Equal(httpCode, 500)
	sut.Equal(accessToken, "")
	sut.Equal(refreshToken, "")
	sut.NotEqual(response, nil)
}

func (sut *UserServiceTestSuite) Test07RegisterGenerateRefreshTokenError() {
	sut.T().Log("Test07RegisterGenerateRefreshTokenError")
	sut.postgresUtilMock.Mock.On("BeginTx", sut.ctx, sut.options).Return(sut.pgxTxMock, nil)
	passwordByte := []byte{112, 97, 115, 115, 119, 111, 114, 100}
	sut.bcryptHelperMock.Mock.On("GenerateFromPassword", []byte(sut.registerRequest.Password), bcrypt.DefaultCost).Return(passwordByte, nil)
	sut.user.Id = pgtype.Int4{Valid: false, Int32: 0}
	sut.user.RefreshToken = pgtype.Text{Valid: false, String: ""}
	sut.userRepositoryMock.Mock.On("Create", sut.pgxTxMock, sut.ctx, sut.user).Return(1, nil)
	jwtAccessTokenTime := 15
	sut.user.Id = pgtype.Int4{Valid: true, Int32: 1}
	var accessToken string
	accessToken = "accessToken"
	sut.jwtHelperMock.Mock.On("GenerateAccessToken", int(sut.user.Id.Int32), sut.user.Name.String, sut.user.Email.String, jwtAccessTokenTime, "secret").Return(accessToken, nil)
	jwtRefreshTokenTime := 1
	var refreshToken string
	sut.jwtHelperMock.Mock.On("GenerateRefreshToken", int(sut.user.Id.Int32), jwtRefreshTokenTime, "secret").Return(refreshToken, sut.errInternalServer)
	sut.postgresUtilMock.Mock.On("CommitOrRollback", sut.pgxTxMock, sut.ctx, sut.errInternalServer).Return(nil)
	httpCode, accessToken, refreshToken, response := sut.userService.Register(sut.ctx, sut.registerRequest)
	sut.Equal(httpCode, 500)
	sut.Equal(accessToken, "")
	sut.Equal(refreshToken, "")
	sut.NotEqual(response, nil)
}

func (sut *UserServiceTestSuite) Test08RegisterUpdateRefreshTokenError() {
	sut.T().Log("Test08RegisterUpdateRefreshTokenError")
	sut.postgresUtilMock.Mock.On("BeginTx", sut.ctx, sut.options).Return(sut.pgxTxMock, nil)
	passwordByte := []byte{112, 97, 115, 115, 119, 111, 114, 100}
	sut.bcryptHelperMock.Mock.On("GenerateFromPassword", []byte(sut.registerRequest.Password), bcrypt.DefaultCost).Return(passwordByte, nil)
	sut.user.Id = pgtype.Int4{Valid: false, Int32: 0}
	sut.user.RefreshToken = pgtype.Text{Valid: false, String: ""}
	sut.userRepositoryMock.Mock.On("Create", sut.pgxTxMock, sut.ctx, sut.user).Return(1, nil)
	jwtAccessTokenTime := 15
	sut.user.Id = pgtype.Int4{Valid: true, Int32: 1}
	sut.jwtHelperMock.Mock.On("GenerateAccessToken", int(sut.user.Id.Int32), sut.user.Name.String, sut.user.Email.String, jwtAccessTokenTime, "secret").Return("accessToken", nil)
	jwtRefreshTokenTime := 1
	sut.jwtHelperMock.Mock.On("GenerateRefreshToken", int(sut.user.Id.Int32), jwtRefreshTokenTime, "secret").Return("refreshToken", nil)
	var rowsAffected int64
	sut.userRepositoryMock.Mock.On("UpdateRefreshToken", sut.pgxTxMock, sut.ctx, "refreshToken", int(sut.user.Id.Int32)).Return(rowsAffected, sut.errInternalServer)
	sut.postgresUtilMock.Mock.On("CommitOrRollback", sut.pgxTxMock, sut.ctx, sut.errInternalServer).Return(nil)
	httpCode, accessToken, refreshToken, response := sut.userService.Register(sut.ctx, sut.registerRequest)
	sut.Equal(httpCode, 500)
	sut.Equal(accessToken, "")
	sut.Equal(refreshToken, "")
	sut.NotEqual(response, nil)
}

func (sut *UserServiceTestSuite) Test09RegisterUpdateRefreshTokenRowsAffectedNotOne() {
	sut.T().Log("Test09RegisterUpdateRefreshTokenRowsAffectedNotOne")
	sut.postgresUtilMock.Mock.On("BeginTx", sut.ctx, sut.options).Return(sut.pgxTxMock, nil)
	passwordByte := []byte{112, 97, 115, 115, 119, 111, 114, 100}
	sut.bcryptHelperMock.Mock.On("GenerateFromPassword", []byte(sut.registerRequest.Password), bcrypt.DefaultCost).Return(passwordByte, nil)
	sut.user.Id = pgtype.Int4{Valid: false, Int32: 0}
	sut.user.RefreshToken = pgtype.Text{Valid: false, String: ""}
	sut.userRepositoryMock.Mock.On("Create", sut.pgxTxMock, sut.ctx, sut.user).Return(1, nil)
	jwtAccessTokenTime := 15
	sut.user.Id = pgtype.Int4{Valid: true, Int32: 1}
	sut.jwtHelperMock.Mock.On("GenerateAccessToken", int(sut.user.Id.Int32), sut.user.Name.String, sut.user.Email.String, jwtAccessTokenTime, "secret").Return("accessToken", nil)
	jwtRefreshTokenTime := 1
	sut.jwtHelperMock.Mock.On("GenerateRefreshToken", int(sut.user.Id.Int32), jwtRefreshTokenTime, "secret").Return("refreshToken", nil)
	var rowsAffected int64
	rowsAffected = 0
	sut.userRepositoryMock.Mock.On("UpdateRefreshToken", sut.pgxTxMock, sut.ctx, "refreshToken", int(sut.user.Id.Int32)).Return(rowsAffected, nil)
	sut.postgresUtilMock.Mock.On("CommitOrRollback", sut.pgxTxMock, sut.ctx, sut.errRowsAffectedNotOne).Return(nil)
	httpCode, accessToken, refreshToken, response := sut.userService.Register(sut.ctx, sut.registerRequest)
	sut.Equal(httpCode, 500)
	sut.Equal(accessToken, "")
	sut.Equal(refreshToken, "")
	sut.NotEqual(response, nil)
}

func (sut *UserServiceTestSuite) Test10RegisterSuccess() {
	sut.T().Log("Test10RegisterSuccess")
	sut.postgresUtilMock.Mock.On("BeginTx", sut.ctx, sut.options).Return(sut.pgxTxMock, nil)
	passwordByte := []byte{112, 97, 115, 115, 119, 111, 114, 100}
	sut.bcryptHelperMock.Mock.On("GenerateFromPassword", []byte(sut.registerRequest.Password), bcrypt.DefaultCost).Return(passwordByte, nil)
	sut.user.Id = pgtype.Int4{Valid: false, Int32: 0}
	sut.user.RefreshToken = pgtype.Text{Valid: false, String: ""}
	sut.userRepositoryMock.Mock.On("Create", sut.pgxTxMock, sut.ctx, sut.user).Return(1, nil)
	jwtAccessTokenTime := 15
	sut.user.Id = pgtype.Int4{Valid: true, Int32: 1}
	sut.jwtHelperMock.Mock.On("GenerateAccessToken", int(sut.user.Id.Int32), sut.user.Name.String, sut.user.Email.String, jwtAccessTokenTime, "secret").Return("accessToken", nil)
	jwtRefreshTokenTime := 1
	sut.jwtHelperMock.Mock.On("GenerateRefreshToken", int(sut.user.Id.Int32), jwtRefreshTokenTime, "secret").Return("refreshToken", nil)
	var rowsAffected int64
	rowsAffected = 1
	sut.userRepositoryMock.Mock.On("UpdateRefreshToken", sut.pgxTxMock, sut.ctx, "refreshToken", int(sut.user.Id.Int32)).Return(rowsAffected, nil)
	sut.postgresUtilMock.Mock.On("CommitOrRollback", sut.pgxTxMock, sut.ctx, nil).Return(nil)
	httpCode, accessToken, refreshToken, response := sut.userService.Register(sut.ctx, sut.registerRequest)
	sut.Equal(httpCode, http.StatusCreated)
	sut.NotEqual(accessToken, "")
	sut.NotEqual(refreshToken, "")
	sut.NotEqual(response, nil)
}

func (sut *UserServiceTestSuite) Test11LoginValidationError() {
	sut.T().Log("Test11LoginValidationError")
	sut.loginRequest = modelrequests.LoginRequest{}
	httpCode, accessToken, refreshToken, response := sut.userService.Login(sut.ctx, sut.loginRequest)
	sut.Equal(httpCode, http.StatusInternalServerError)
	sut.Equal(accessToken, "")
	sut.Equal(refreshToken, "")
	sut.NotEqual(response, nil)
}

func (sut *UserServiceTestSuite) Test12LoginBeginTxError() {
	sut.T().Log("Test12LoginBeginTxError")
	sut.postgresUtilMock.Mock.On("BeginTx", sut.ctx, sut.options).Return(sut.pgxTxMock, sut.errInternalServer)
	httpCode, accessToken, refreshToken, response := sut.userService.Login(sut.ctx, sut.loginRequest)
	sut.Equal(httpCode, http.StatusInternalServerError)
	sut.Equal(accessToken, "")
	sut.Equal(refreshToken, "")
	sut.NotEqual(response, nil)
}

func (sut *UserServiceTestSuite) Test13LoginFindByEmailError() {
	sut.T().Log("Test13LoginFindByEmailError")
	sut.postgresUtilMock.Mock.On("BeginTx", sut.ctx, sut.options).Return(sut.pgxTxMock, nil)
	var user modelentities.User
	sut.userRepositoryMock.Mock.On("FindByEmail", sut.pgxTxMock, sut.ctx, sut.loginRequest.Email).Return(user, sut.errInternalServer)
	sut.postgresUtilMock.Mock.On("CommitOrRollback", sut.pgxTxMock, sut.ctx, sut.errInternalServer).Return(nil)
	httpCode, accessToken, refreshToken, response := sut.userService.Login(sut.ctx, sut.loginRequest)
	sut.Equal(httpCode, http.StatusInternalServerError)
	sut.Equal(accessToken, "")
	sut.Equal(refreshToken, "")
	sut.NotEqual(response, nil)
}

func (sut *UserServiceTestSuite) Test14LoginFindByEmailWrongEmailOrPassword() {
	sut.T().Log("Test14LoginFindByEmailWrongEmailOrPassword")
	sut.postgresUtilMock.Mock.On("BeginTx", sut.ctx, sut.options).Return(sut.pgxTxMock, nil)
	var user modelentities.User
	sut.userRepositoryMock.Mock.On("FindByEmail", sut.pgxTxMock, sut.ctx, sut.loginRequest.Email).Return(user, pgx.ErrNoRows)
	sut.postgresUtilMock.Mock.On("CommitOrRollback", sut.pgxTxMock, sut.ctx, pgx.ErrNoRows).Return(nil)
	httpCode, accessToken, refreshToken, response := sut.userService.Login(sut.ctx, sut.loginRequest)
	sut.Equal(httpCode, http.StatusBadRequest)
	sut.Equal(accessToken, "")
	sut.Equal(refreshToken, "")
	sut.NotEqual(response, nil)
}

func (sut *UserServiceTestSuite) Test15LoginCompareHashAndPassword() {
	sut.T().Log("Test15LoginCompareHashAndPassword")
	sut.postgresUtilMock.Mock.On("BeginTx", sut.ctx, sut.options).Return(sut.pgxTxMock, nil)
	sut.userRepositoryMock.Mock.On("FindByEmail", sut.pgxTxMock, sut.ctx, sut.loginRequest.Email).Return(sut.user, nil)
	sut.bcryptHelperMock.Mock.On("CompareHashAndPassword", []byte(sut.user.Password.String), []byte(sut.loginRequest.Password)).Return(sut.errInternalServer)
	sut.postgresUtilMock.Mock.On("CommitOrRollback", sut.pgxTxMock, sut.ctx, sut.errInternalServer).Return(nil)
	httpCode, accessToken, refreshToken, response := sut.userService.Login(sut.ctx, sut.loginRequest)
	sut.Equal(httpCode, http.StatusBadRequest)
	sut.Equal(accessToken, "")
	sut.Equal(refreshToken, "")
	sut.NotEqual(response, nil)
}

func (sut *UserServiceTestSuite) Test16LoginGenerateAccessTokenError() {
	sut.T().Log("Test16LoginGenerateAccessTokenError")
	sut.postgresUtilMock.Mock.On("BeginTx", sut.ctx, sut.options).Return(sut.pgxTxMock, nil)
	sut.userRepositoryMock.Mock.On("FindByEmail", sut.pgxTxMock, sut.ctx, sut.loginRequest.Email).Return(sut.user, nil)
	sut.bcryptHelperMock.Mock.On("CompareHashAndPassword", []byte(sut.user.Password.String), []byte(sut.loginRequest.Password)).Return(nil)
	jwtAccessTokenTime := 15
	sut.jwtHelperMock.Mock.On("GenerateAccessToken", int(sut.user.Id.Int32), sut.user.Name.String, sut.user.Email.String, jwtAccessTokenTime, "secret").Return("", sut.errInternalServer)
	sut.postgresUtilMock.Mock.On("CommitOrRollback", sut.pgxTxMock, sut.ctx, sut.errInternalServer).Return(nil)
	httpCode, accessToken, refreshToken, response := sut.userService.Login(sut.ctx, sut.loginRequest)
	sut.Equal(httpCode, http.StatusInternalServerError)
	sut.Equal(accessToken, "")
	sut.Equal(refreshToken, "")
	sut.NotEqual(response, nil)
}

func (sut *UserServiceTestSuite) Test17LoginGenerateRefreshTokenError() {
	sut.T().Log("Test17LoginGenerateRefreshTokenError")
	sut.postgresUtilMock.Mock.On("BeginTx", sut.ctx, sut.options).Return(sut.pgxTxMock, nil)
	sut.userRepositoryMock.Mock.On("FindByEmail", sut.pgxTxMock, sut.ctx, sut.loginRequest.Email).Return(sut.user, nil)
	sut.bcryptHelperMock.Mock.On("CompareHashAndPassword", []byte(sut.user.Password.String), []byte(sut.loginRequest.Password)).Return(nil)
	jwtAccessTokenTime := 15
	sut.jwtHelperMock.Mock.On("GenerateAccessToken", int(sut.user.Id.Int32), sut.user.Name.String, sut.user.Email.String, jwtAccessTokenTime, "secret").Return("accessToken", nil)
	jwtRefreshTokenTime := 1
	sut.jwtHelperMock.Mock.On("GenerateRefreshToken", int(sut.user.Id.Int32), jwtRefreshTokenTime, "secret").Return("", sut.errInternalServer)
	sut.postgresUtilMock.Mock.On("CommitOrRollback", sut.pgxTxMock, sut.ctx, sut.errInternalServer).Return(nil)
	httpCode, accessToken, refreshToken, response := sut.userService.Login(sut.ctx, sut.loginRequest)
	sut.Equal(httpCode, http.StatusInternalServerError)
	sut.Equal(accessToken, "")
	sut.Equal(refreshToken, "")
	sut.NotEqual(response, nil)
}

func (sut *UserServiceTestSuite) Test18LoginUpdateRefreshTokenError() {
	sut.T().Log("Test18LoginUpdateRefreshTokenError")
	sut.postgresUtilMock.Mock.On("BeginTx", sut.ctx, sut.options).Return(sut.pgxTxMock, nil)
	sut.userRepositoryMock.Mock.On("FindByEmail", sut.pgxTxMock, sut.ctx, sut.loginRequest.Email).Return(sut.user, nil)
	sut.bcryptHelperMock.Mock.On("CompareHashAndPassword", []byte(sut.user.Password.String), []byte(sut.loginRequest.Password)).Return(nil)
	jwtAccessTokenTime := 15
	sut.jwtHelperMock.Mock.On("GenerateAccessToken", int(sut.user.Id.Int32), sut.user.Name.String, sut.user.Email.String, jwtAccessTokenTime, "secret").Return("accessToken", nil)
	jwtRefreshTokenTime := 1
	sut.jwtHelperMock.Mock.On("GenerateRefreshToken", int(sut.user.Id.Int32), jwtRefreshTokenTime, "secret").Return("refreshToken", nil)
	var rowsAffected int64
	sut.userRepositoryMock.Mock.On("UpdateRefreshToken", sut.pgxTxMock, sut.ctx, "refreshToken", int(sut.user.Id.Int32)).Return(rowsAffected, sut.errInternalServer)
	sut.postgresUtilMock.Mock.On("CommitOrRollback", sut.pgxTxMock, sut.ctx, sut.errInternalServer).Return(nil)
	httpCode, accessToken, refreshToken, response := sut.userService.Login(sut.ctx, sut.loginRequest)
	sut.Equal(httpCode, http.StatusInternalServerError)
	sut.Equal(accessToken, "")
	sut.Equal(refreshToken, "")
	sut.NotEqual(response, nil)
}

func (sut *UserServiceTestSuite) Test19LoginUpdateRefreshTokenRowsAffected() {
	sut.T().Log("Test19LoginUpdateRefreshTokenRowsAffected")
	sut.postgresUtilMock.Mock.On("BeginTx", sut.ctx, sut.options).Return(sut.pgxTxMock, nil)
	sut.userRepositoryMock.Mock.On("FindByEmail", sut.pgxTxMock, sut.ctx, sut.loginRequest.Email).Return(sut.user, nil)
	sut.bcryptHelperMock.Mock.On("CompareHashAndPassword", []byte(sut.user.Password.String), []byte(sut.loginRequest.Password)).Return(nil)
	jwtAccessTokenTime := 15
	sut.jwtHelperMock.Mock.On("GenerateAccessToken", int(sut.user.Id.Int32), sut.user.Name.String, sut.user.Email.String, jwtAccessTokenTime, "secret").Return("accessToken", nil)
	jwtRefreshTokenTime := 1
	sut.jwtHelperMock.Mock.On("GenerateRefreshToken", int(sut.user.Id.Int32), jwtRefreshTokenTime, "secret").Return("refreshToken", nil)
	var rowsAffected int64
	sut.userRepositoryMock.Mock.On("UpdateRefreshToken", sut.pgxTxMock, sut.ctx, "refreshToken", int(sut.user.Id.Int32)).Return(rowsAffected, nil)
	sut.postgresUtilMock.Mock.On("CommitOrRollback", sut.pgxTxMock, sut.ctx, sut.errRowsAffectedNotOne).Return(nil)
	httpCode, accessToken, refreshToken, response := sut.userService.Login(sut.ctx, sut.loginRequest)
	sut.Equal(httpCode, http.StatusInternalServerError)
	sut.Equal(accessToken, "")
	sut.Equal(refreshToken, "")
	sut.NotEqual(response, nil)
}

func (sut *UserServiceTestSuite) Test20LoginSuccess() {
	sut.T().Log("Test20LoginSuccess")
	sut.postgresUtilMock.Mock.On("BeginTx", sut.ctx, sut.options).Return(sut.pgxTxMock, nil)
	sut.userRepositoryMock.Mock.On("FindByEmail", sut.pgxTxMock, sut.ctx, sut.loginRequest.Email).Return(sut.user, nil)
	sut.bcryptHelperMock.Mock.On("CompareHashAndPassword", []byte(sut.user.Password.String), []byte(sut.loginRequest.Password)).Return(nil)
	jwtAccessTokenTime := 15
	sut.jwtHelperMock.Mock.On("GenerateAccessToken", int(sut.user.Id.Int32), sut.user.Name.String, sut.user.Email.String, jwtAccessTokenTime, "secret").Return("accessToken", nil)
	jwtRefreshTokenTime := 1
	sut.jwtHelperMock.Mock.On("GenerateRefreshToken", int(sut.user.Id.Int32), jwtRefreshTokenTime, "secret").Return("refreshToken", nil)
	var rowsAffected int64
	rowsAffected = 1
	sut.userRepositoryMock.Mock.On("UpdateRefreshToken", sut.pgxTxMock, sut.ctx, "refreshToken", int(sut.user.Id.Int32)).Return(rowsAffected, nil)
	sut.postgresUtilMock.Mock.On("CommitOrRollback", sut.pgxTxMock, sut.ctx, nil).Return(nil)
	httpCode, accessToken, refreshToken, response := sut.userService.Login(sut.ctx, sut.loginRequest)
	sut.Equal(httpCode, http.StatusOK)
	sut.Equal(accessToken, "accessToken")
	sut.Equal(refreshToken, "refreshToken")
	sut.NotEqual(response, nil)
}

func (sut *UserServiceTestSuite) Test21RefreshTokenFindByRefreshTokenError() {
	sut.T().Log("Test21RefreshTokenFindByRefreshTokenError")
	sut.postgresUtilMock.Mock.On("GetPool").Return(sut.pool)
	refreshToken := "refreshToken"
	var user modelentities.User
	sut.userRepositoryMock.Mock.On("FindByRefreshToken", sut.pool, sut.ctx, refreshToken).Return(user, sut.errInternalServer)
	httpCode, accessToken, response := sut.userService.RefreshToken(sut.ctx, refreshToken)
	sut.Equal(httpCode, http.StatusInternalServerError)
	sut.Equal(accessToken, "")
	sut.NotEqual(response, nil)
}

func (sut *UserServiceTestSuite) Test22RefreshTokenFindByRefreshTokenRowsAffectedNotOne() {
	sut.T().Log("Test22RefreshTokenFindByRefreshTokenRowsAffectedNotOne")
	sut.postgresUtilMock.Mock.On("GetPool").Return(sut.pool)
	refreshToken := "refreshToken"
	var user modelentities.User
	sut.userRepositoryMock.Mock.On("FindByRefreshToken", sut.pool, sut.ctx, refreshToken).Return(user, pgx.ErrNoRows)
	httpCode, accessToken, response := sut.userService.RefreshToken(sut.ctx, refreshToken)
	sut.Equal(httpCode, http.StatusBadRequest)
	sut.Equal(accessToken, "")
	sut.NotEqual(response, nil)
}

func (sut *UserServiceTestSuite) Test23RefreshTokenGenerateAccessTokenError() {
	sut.T().Log("Test23RefreshTokenGenerateAccessTokenError")
	sut.postgresUtilMock.Mock.On("GetPool").Return(sut.pool)
	refreshToken := "refreshToken"
	sut.userRepositoryMock.Mock.On("FindByRefreshToken", sut.pool, sut.ctx, refreshToken).Return(sut.user, nil)
	jwtAccessTokenTime := 15
	sut.jwtHelperMock.Mock.On("GenerateAccessToken", int(sut.user.Id.Int32), sut.user.Name.String, sut.user.Email.String, jwtAccessTokenTime, "secret").Return("", sut.errInternalServer)
	httpCode, accessToken, response := sut.userService.RefreshToken(sut.ctx, refreshToken)
	sut.Equal(httpCode, http.StatusInternalServerError)
	sut.Equal(accessToken, "")
	sut.NotEqual(response, nil)
}

func (sut *UserServiceTestSuite) Test24RefreshTokenSuccess() {
	sut.T().Log("Test24RefreshTokenSuccess")
	sut.postgresUtilMock.Mock.On("GetPool").Return(sut.pool)
	refreshToken := "refreshToken"
	sut.userRepositoryMock.Mock.On("FindByRefreshToken", sut.pool, sut.ctx, refreshToken).Return(sut.user, nil)
	jwtAccessTokenTime := 15
	sut.jwtHelperMock.Mock.On("GenerateAccessToken", int(sut.user.Id.Int32), sut.user.Name.String, sut.user.Email.String, jwtAccessTokenTime, "secret").Return("accessToken", nil)
	httpCode, accessToken, response := sut.userService.RefreshToken(sut.ctx, refreshToken)
	sut.Equal(httpCode, http.StatusOK)
	sut.Equal(accessToken, "accessToken")
	sut.NotEqual(response, nil)
}

func (sut *UserServiceTestSuite) AfterTest(suiteName, testName string) {
	sut.T().Log("AfterTest: " + suiteName + " " + testName)
}

func (sut *UserServiceTestSuite) TearDownTest() {
	sut.T().Log("TearDownTest")
}

func (sut *UserServiceTestSuite) TearDownSuite() {
	sut.T().Log("TearDownSuite")
}
