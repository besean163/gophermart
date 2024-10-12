package authservice

import (
	"errors"
	"fmt"
	"time"

	"github.com/besean163/gophermart/internal/entities"
	"github.com/golang-jwt/jwt/v4"
)

var (
	ErrEmptyHashSecret = errors.New("empty hash secret")
)

type Service struct {
	repository  UserRepository
	tokenSecret string
	tokenExpire time.Duration
}

type claims struct {
	jwt.RegisteredClaims
	UserLogin string
}

func New(repository UserRepository, secret string, tokenExpire time.Duration) Service {
	return Service{
		repository:  repository,
		tokenSecret: secret,
		tokenExpire: tokenExpire,
	}
}

type UserRepository interface {
	SaveUser(entities.User) error
	GetUser(login string) *entities.User
}

func (service Service) SaveUser(user entities.User) error {
	return service.repository.SaveUser(user)
}

func (service Service) GetUser(login string) *entities.User {
	return service.repository.GetUser(login)
}

func (service Service) BuildUserToken(user entities.User) (string, error) {
	if service.tokenSecret == "" {
		return "", ErrEmptyHashSecret
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(service.tokenExpire)),
		},
		UserLogin: user.Login,
	})

	tokenString, err := token.SignedString([]byte(service.tokenSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func getUserLoginByToken(token string, secret string) (string, error) {
	if token == "" {
		return "", fmt.Errorf("empty token")
	}

	claims := &claims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return "nil", err
	}

	if claims.ExpiresAt.Time.Unix() < time.Now().Unix() {
		return "", fmt.Errorf("token expire")
	}

	return claims.UserLogin, nil
}

func (service Service) GetUserByToken(token string) (*entities.User, error) {
	userLogin, err := getUserLoginByToken(token, service.tokenSecret)
	if err != nil {
		return nil, err
	}

	user := service.repository.GetUser(userLogin)
	if user == nil {
		return nil, fmt.Errorf("user not exist")
	}

	return user, nil
}
