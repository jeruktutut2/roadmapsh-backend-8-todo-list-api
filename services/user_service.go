package services

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"
	"todo-list-api/helpers"
	modelentities "todo-list-api/models/entities"
	modelrequests "todo-list-api/models/requests"
	"todo-list-api/repositories"
	"todo-list-api/utils"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, registerRequest modelrequests.RegisterRequest) (httpCode int, accessToken string, refreshToken string, response interface{})
	Login(ctx context.Context, loginRequest modelrequests.LoginRequest) (httpCode int, accessToken string, refreshToken string, response interface{})
	RefreshToken(ctx context.Context, refreshToken string) (httpCode int, accessToken string, response interface{})
}

type UserServiceImplementation struct {
	PostgresUtil   utils.PostgresUtil
	Validate       *validator.Validate
	UserRepository repositories.UserRepository
	BcryptHelper   helpers.BcryptHelper
	JwtHelper      helpers.JwtHelper
}

func NewUserService(postgresUtil utils.PostgresUtil, validate *validator.Validate, userRepository repositories.UserRepository, bcryptHelper helpers.BcryptHelper, jwtHelper helpers.JwtHelper) UserService {
	return &UserServiceImplementation{
		PostgresUtil:   postgresUtil,
		Validate:       validate,
		UserRepository: userRepository,
		BcryptHelper:   bcryptHelper,
		JwtHelper:      jwtHelper,
	}
}

func (service *UserServiceImplementation) Register(ctx context.Context, registerRequest modelrequests.RegisterRequest) (httpCode int, accessToken string, refreshToken string, response interface{}) {
	var err error
	err = service.Validate.Struct(registerRequest)
	if err != nil {
		httpCode = http.StatusBadRequest
		response = helpers.ToResponse(err.Error())
		return
	}

	tx, err := service.PostgresUtil.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}

	defer func() {
		errCommitOrRollback := service.PostgresUtil.CommitOrRollback(tx, ctx, err)
		if errCommitOrRollback != nil {
			httpCode = http.StatusInternalServerError
			response = helpers.ToResponse(errCommitOrRollback.Error())
		}
	}()

	var user modelentities.User
	user.Name = pgtype.Text{Valid: true, String: registerRequest.Name}
	user.Email = pgtype.Text{Valid: true, String: registerRequest.Email}
	passwordByte, err := service.BcryptHelper.GenerateFromPassword([]byte(registerRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}
	user.Password = pgtype.Text{Valid: true, String: string(passwordByte)}
	lastInsertedId, err := service.UserRepository.Create(tx, ctx, user)
	if err != nil {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}
	user.Id = pgtype.Int4{Valid: true, Int32: int32(lastInsertedId)}

	jwtAccessTokenTimeEnv := os.Getenv("JWT_ACCESS_TOKEN_TIME")
	jwtAccessTokenTime, err := strconv.Atoi(jwtAccessTokenTimeEnv)
	if err != nil {
		return
	}
	accessToken, err = service.JwtHelper.GenerateAccessToken(int(user.Id.Int32), user.Name.String, user.Email.String, jwtAccessTokenTime, os.Getenv("JWT_SECRET"))
	if err != nil {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}
	jwtRefreshTokenTimeEnv := os.Getenv("JWT_REFRESH_TOKEN_TIME")
	jwtRefreshTokenTime, err := strconv.Atoi(jwtRefreshTokenTimeEnv)
	if err != nil {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}
	refreshToken, err = service.JwtHelper.GenerateRefreshToken(int(user.Id.Int32), jwtRefreshTokenTime, os.Getenv("JWT_SECRET"))
	if err != nil {
		httpCode = http.StatusInternalServerError
		accessToken = ""
		response = helpers.ToResponse(err.Error())
		return
	}

	rowsAffected, err := service.UserRepository.UpdateRefreshToken(tx, ctx, refreshToken, int(user.Id.Int32))
	if err != nil {
		accessToken = ""
		refreshToken = ""
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}
	if rowsAffected != 1 {
		accessToken = ""
		refreshToken = ""
		httpCode = http.StatusInternalServerError
		err = errors.New("rows affected not one")
		response = helpers.ToResponse(err.Error())
		return
	}

	httpCode = http.StatusCreated
	response = helpers.ToResponse("successfully registered")
	return
}

func (service *UserServiceImplementation) Login(ctx context.Context, loginRequest modelrequests.LoginRequest) (httpCode int, accessToken string, refreshToken string, response interface{}) {
	err := service.Validate.Struct(loginRequest)
	if err != nil {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}

	tx, err := service.PostgresUtil.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}

	defer func() {
		errCommitOrRollback := service.PostgresUtil.CommitOrRollback(tx, ctx, err)
		if errCommitOrRollback != nil {
			httpCode = http.StatusInternalServerError
			response = helpers.ToResponse(errCommitOrRollback.Error())
		}
	}()

	user, err := service.UserRepository.FindByEmail(tx, ctx, loginRequest.Email)
	if err != nil && err != pgx.ErrNoRows {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	} else if err != nil && err == pgx.ErrNoRows {
		httpCode = http.StatusBadRequest
		response = helpers.ToResponse("wrong email or password")
		return
	}

	err = service.BcryptHelper.CompareHashAndPassword([]byte(user.Password.String), []byte(loginRequest.Password))
	if err != nil {
		httpCode = http.StatusBadRequest
		response = helpers.ToResponse("wrong email or password")
		return
	}

	jwtAccessTokenTimeEnv := os.Getenv("JWT_ACCESS_TOKEN_TIME")
	jwtAccessTokenTime, err := strconv.Atoi(jwtAccessTokenTimeEnv)
	if err != nil {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}
	accessToken, err = service.JwtHelper.GenerateAccessToken(int(user.Id.Int32), user.Name.String, user.Email.String, jwtAccessTokenTime, os.Getenv("JWT_SECRET"))
	if err != nil {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}

	jwtRefreshTokenTimeEnv := os.Getenv("JWT_REFRESH_TOKEN_TIME")
	jwtRefreshTokenTime, err := strconv.Atoi(jwtRefreshTokenTimeEnv)
	if err != nil {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}
	refreshToken, err = service.JwtHelper.GenerateRefreshToken(int(user.Id.Int32), jwtRefreshTokenTime, os.Getenv("JWT_SECRET"))
	if err != nil {
		accessToken = ""
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}

	rowsAffected, err := service.UserRepository.UpdateRefreshToken(tx, ctx, refreshToken, int(user.Id.Int32))
	if err != nil {
		accessToken = ""
		refreshToken = ""
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}
	if rowsAffected != 1 {
		accessToken = ""
		refreshToken = ""
		httpCode = http.StatusInternalServerError
		err = errors.New("rows affected not one")
		response = helpers.ToResponse(err.Error())
		return
	}

	httpCode = http.StatusOK
	response = helpers.ToResponse("successfully login")
	return
}

func (service *UserServiceImplementation) RefreshToken(ctx context.Context, refreshToken string) (httpCode int, accessToken string, response interface{}) {
	user, err := service.UserRepository.FindByRefreshToken(service.PostgresUtil.GetPool(), ctx, refreshToken)
	if err != nil && err != pgx.ErrNoRows {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	} else if err != nil && err == pgx.ErrNoRows {
		httpCode = http.StatusBadRequest
		response = helpers.ToResponse("cannot find user by resfresh token")
		return
	}

	jwtAccessTokenTimeEnv := os.Getenv("JWT_ACCESS_TOKEN_TIME")
	jwtAccessTokenTime, err := strconv.Atoi(jwtAccessTokenTimeEnv)
	if err != nil {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}
	accessToken, err = service.JwtHelper.GenerateAccessToken(int(user.Id.Int32), user.Name.String, user.Email.String, jwtAccessTokenTime, os.Getenv("JWT_SECRET"))
	if err != nil {
		httpCode = http.StatusInternalServerError
		response = helpers.ToResponse(err.Error())
		return
	}
	httpCode = http.StatusOK
	response = helpers.ToResponse("successfully refresh token")
	return
}
