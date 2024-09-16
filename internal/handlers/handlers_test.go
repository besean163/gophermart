package handlers

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/besean163/gophermart/internal/entities"
	"github.com/besean163/gophermart/internal/handlers/mock"
	"github.com/besean163/gophermart/internal/logger"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRegisterUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	authService := mock.NewMockAuthService(ctrl)
	loyaltyService := mock.NewMockLoyaltyService(ctrl)

	authUser := entities.User{
		Login:    "login_ok",
		Password: getMD5Pass("password_ok"),
	}
	authUserToken := "token"
	authService.EXPECT().BuildUserToken(authUser).Return(authUserToken, nil)
	authService.EXPECT().GetUser("login_fail").Return(&entities.User{
		Login:    "login_fail",
		Password: "password_fail",
	})
	authService.EXPECT().GetUser("login_ok").Return(nil)
	authService.EXPECT().SaveUser(authUser).Return(nil)

	handler := New(authService, loyaltyService, "")
	handler.Mount()

	tests := []struct {
		name   string
		method string
		code   int
		body   string
	}{
		{
			name:   "correct input",
			method: http.MethodPost,
			code:   200,
			body:   `{"login":"login_ok","password":"password_ok"}`,
		},
		{
			name:   "empty input",
			method: http.MethodPost,
			code:   400,
			body:   ``,
		},
		{
			name:   "uncorrect input",
			method: http.MethodPost,
			code:   400,
			body:   `{"login": "login_only"`,
		},
		{
			name:   "exist input",
			method: http.MethodPost,
			code:   409,
			body:   `{"login":"login_fail","password":"password_fail"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := strings.NewReader(test.body)
			request, _ := http.NewRequest(test.method, "/api/user/register", body)
			rr := httptest.NewRecorder()

			handler.Router.ServeHTTP(rr, request)

			response := rr.Result()
			defer response.Body.Close()
			assert.Equal(t, test.code, response.StatusCode)
		})
	}
}

func TestLoginUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	authService := mock.NewMockAuthService(ctrl)
	loyaltyService := mock.NewMockLoyaltyService(ctrl)
	secret := "test_secret"

	authUser := entities.User{
		Login:    "login_ok",
		Password: getMD5Pass("password_ok"),
	}
	authUserToken := "token"
	authService.EXPECT().BuildUserToken(authUser).Return(authUserToken, nil)
	authService.EXPECT().GetUser("login_ok").Return(&authUser)
	authService.EXPECT().GetUser("login_fail").Return(&entities.User{
		Login:    "login_fail",
		Password: "password_fail",
	})

	handler := New(authService, loyaltyService, secret)
	handler.Mount()

	tests := []struct {
		name      string
		method    string
		code      int
		body      string
		setHeader bool
	}{
		{
			name:      "correct input",
			method:    http.MethodPost,
			code:      200,
			body:      `{"login":"login_ok","password":"password_ok"}`,
			setHeader: true,
		},
		{
			name:   "empty input",
			method: http.MethodPost,
			code:   400,
			body:   ``,
		},
		{
			name:   "uncorrect input",
			method: http.MethodPost,
			code:   400,
			body:   `{"login": "login_only"`,
		},
		{
			name:   "exist input",
			method: http.MethodPost,
			code:   401,
			body:   `{"login":"login_fail","password":"password_fail"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := strings.NewReader(test.body)
			request, _ := http.NewRequest(test.method, "/api/user/login", body)
			rr := httptest.NewRecorder()

			handler.Router.ServeHTTP(rr, request)

			response := rr.Result()
			defer response.Body.Close()
			assert.Equal(t, test.code, response.StatusCode)
			assert.Equal(t, test.setHeader, response.Header.Get("Authorization") != "")
		})
	}
}

func TestPostOrder(t *testing.T) {
	logger.Set()

	secret := "test_secret"
	authUser := entities.User{
		ID:       1,
		Login:    "login_auth",
		Password: getMD5Pass("password_ok"),
	}

	ctrl := gomock.NewController(t)
	authService := mock.NewMockAuthService(ctrl)
	// отдает авторизованого пользователя
	authUserToken := "token"
	authService.EXPECT().BuildUserToken(authUser).Return(authUserToken, nil).AnyTimes()
	authService.EXPECT().GetUserByToken(authUserToken).Return(&authUser, nil).AnyTimes()
	authService.EXPECT().GetUserByToken("").Return(nil, errors.New("token error")).AnyTimes()
	authService.EXPECT().GetUser("login_auth").Return(&authUser).AnyTimes()

	loyaltyService := mock.NewMockLoyaltyService(ctrl)
	// показывает что нет созданного заказа, для проверки сохранения нового
	loyaltyService.EXPECT().GetOrder("1111111").Return(nil)
	// показывает что есть уже заказ
	loyaltyService.EXPECT().GetOrder("1111111").Return(&entities.Order{
		Number: "1111111",
		UserID: authUser.ID,
		Status: entities.OrderStatusNew,
	})
	// показывает что есть уже заказ, но на другом пользователе
	loyaltyService.EXPECT().GetOrder("2222222").Return(&entities.Order{
		Number: "2222222",
		UserID: 2,
		Status: entities.OrderStatusNew,
	})
	loyaltyService.EXPECT().SaveOrder(gomock.Any()).Return(nil)

	handler := New(authService, loyaltyService, secret)

	handler.Mount()

	tests := []struct {
		name      string
		method    string
		code      int
		body      string
		authToken string
	}{
		{
			name:      "correct add",
			method:    http.MethodPost,
			code:      202,
			body:      `1111111`,
			authToken: authUserToken,
		},
		{
			name:      "already added by user",
			method:    http.MethodPost,
			code:      200,
			body:      `1111111`,
			authToken: authUserToken,
		},
		{
			name:      "uncorrect input",
			method:    http.MethodPost,
			code:      400,
			body:      `abc`,
			authToken: authUserToken,
		},
		{
			name:   "unauthorized user",
			method: http.MethodPost,
			code:   401,
			body:   `1111111`,
		},
		{
			name:      "fail luna check",
			method:    http.MethodPost,
			code:      422,
			body:      `1111112`,
			authToken: authUserToken,
		},
		{
			name:      "already added by other user",
			method:    http.MethodPost,
			code:      409,
			body:      `2222222`,
			authToken: authUserToken,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := strings.NewReader(test.body)
			request, _ := http.NewRequest(test.method, "/api/user/orders", body)
			if test.authToken != "" {
				request.Header.Set("Authorization", authUserToken)
			}
			rr := httptest.NewRecorder()

			handler.Router.ServeHTTP(rr, request)

			response := rr.Result()
			defer response.Body.Close()
			assert.Equal(t, test.code, response.StatusCode)
		})
	}
}

func TestGetOrders(t *testing.T) {
	logger.Set()

	secret := "test_secret"
	authUser := entities.User{
		ID:       1,
		Login:    "login_auth",
		Password: getMD5Pass("password_ok"),
	}

	ctrl := gomock.NewController(t)
	authService := mock.NewMockAuthService(ctrl)
	authUserToken := "token"
	authService.EXPECT().BuildUserToken(authUser).Return(authUserToken, nil).AnyTimes()
	authService.EXPECT().GetUserByToken(authUserToken).Return(&authUser, nil).AnyTimes()
	authService.EXPECT().GetUserByToken("").Return(nil, errors.New("token error")).AnyTimes()
	authService.EXPECT().GetUser("login_auth").Return(&authUser).AnyTimes()

	loyaltyService := mock.NewMockLoyaltyService(ctrl)
	loyaltyService.EXPECT().GetUserOrders(authUser.ID).Return([]*entities.Order{
		{
			Number: "1111111",
			UserID: authUser.ID,
			Status: entities.OrderStatusNew,
		},
	})
	loyaltyService.EXPECT().GetUserOrders(authUser.ID).Return([]*entities.Order{})

	handler := New(authService, loyaltyService, secret)

	handler.Mount()

	tests := []struct {
		name      string
		method    string
		code      int
		authToken string
	}{
		{
			name:      "full list",
			method:    http.MethodGet,
			code:      200,
			authToken: authUserToken,
		},
		{
			name:      "empty list",
			method:    http.MethodGet,
			code:      204,
			authToken: authUserToken,
		},
		{
			name:   "unauthorized user",
			method: http.MethodGet,
			code:   401,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request, _ := http.NewRequest(test.method, "/api/user/orders", nil)
			if test.authToken != "" {
				request.Header.Set("Authorization", authUserToken)
			}
			rr := httptest.NewRecorder()

			handler.Router.ServeHTTP(rr, request)

			response := rr.Result()
			defer response.Body.Close()
			assert.Equal(t, test.code, response.StatusCode)
		})
	}
}

func TestGetWithdrawns(t *testing.T) {
	logger.Set()
	testTime, _ := time.Parse(time.RFC3339, "2020-12-09T16:09:57+03:00")

	secret := "test_secret"
	authUser := entities.User{
		ID:       1,
		Login:    "login_auth",
		Password: getMD5Pass("password_ok"),
	}

	ctrl := gomock.NewController(t)
	authService := mock.NewMockAuthService(ctrl)
	authUserToken := "token"
	authService.EXPECT().BuildUserToken(authUser).Return(authUserToken, nil).AnyTimes()
	authService.EXPECT().GetUserByToken(authUserToken).Return(&authUser, nil).AnyTimes()
	authService.EXPECT().GetUserByToken("").Return(nil, errors.New("token error")).AnyTimes()
	authService.EXPECT().GetUser("login_auth").Return(&authUser).AnyTimes()

	loyaltyService := mock.NewMockLoyaltyService(ctrl)
	loyaltyService.EXPECT().GetUserWithdrawals(authUser.ID).Return([]*entities.Withdrawn{
		{
			ID:          1,
			UserID:      authUser.ID,
			OrderNumber: "1111111",
			Sum:         0,
			ProccesedAt: testTime,
		},
	})
	loyaltyService.EXPECT().GetUserWithdrawals(authUser.ID).Return([]*entities.Withdrawn{})

	handler := New(authService, loyaltyService, secret)
	handler.Mount()

	tests := []struct {
		name      string
		method    string
		code      int
		authToken string
		outBody   string
	}{
		{
			name:      "full list",
			method:    http.MethodGet,
			code:      200,
			authToken: authUserToken,
			outBody:   `[{"order":"1111111","sum":0,"processed_at":"2020-12-09T16:09:57+03:00"}]`,
		},
		{
			name:      "empty list",
			method:    http.MethodGet,
			code:      204,
			authToken: authUserToken,
		},
		{
			name:   "unauthorized user",
			method: http.MethodGet,
			code:   401,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request, _ := http.NewRequest(test.method, "/api/user/withdrawals", nil)
			if test.authToken != "" {
				request.Header.Set("Authorization", authUserToken)
			}
			rr := httptest.NewRecorder()

			handler.Router.ServeHTTP(rr, request)

			response := rr.Result()
			defer response.Body.Close()
			assert.Equal(t, test.code, response.StatusCode)
			if test.outBody != "" {
				assert.Equal(t, rr.Body.String(), test.outBody)
			}
		})
	}
}

func TestGetBalance(t *testing.T) {
	logger.Set()

	secret := "test_secret"
	authUser := entities.User{
		ID:       1,
		Login:    "login_auth",
		Password: getMD5Pass("password_ok"),
	}

	ctrl := gomock.NewController(t)
	authService := mock.NewMockAuthService(ctrl)
	authUserToken := "token"
	authService.EXPECT().BuildUserToken(authUser).Return(authUserToken, nil).AnyTimes()
	authService.EXPECT().GetUserByToken(authUserToken).Return(&authUser, nil).AnyTimes()
	authService.EXPECT().GetUserByToken("").Return(nil, errors.New("token error")).AnyTimes()
	authService.EXPECT().GetUser("login_auth").Return(&authUser).AnyTimes()

	loyaltyService := mock.NewMockLoyaltyService(ctrl)
	loyaltyService.EXPECT().GetUserBalance(authUser.ID).Return(entities.Balance{
		Current:   100,
		Withdrawn: 50,
	})

	handler := New(authService, loyaltyService, secret)

	handler.Mount()

	tests := []struct {
		name      string
		method    string
		code      int
		authToken string
		outBody   string
	}{
		{
			name:      "full list",
			method:    http.MethodGet,
			code:      200,
			authToken: authUserToken,
			outBody:   `{"current":100,"withdrawn":50}`,
		},
		{
			name:   "unauthorized user",
			method: http.MethodGet,
			code:   401,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request, _ := http.NewRequest(test.method, "/api/user/balance/", nil)
			if test.authToken != "" {
				request.Header.Set("Authorization", authUserToken)
			}
			rr := httptest.NewRecorder()

			handler.Router.ServeHTTP(rr, request)

			response := rr.Result()
			defer response.Body.Close()
			assert.Equal(t, test.code, response.StatusCode)

			if test.outBody != "" {
				assert.Equal(t, rr.Body.String(), test.outBody)
			}
		})
	}
}

func TestSaveWithdrawn(t *testing.T) {
	logger.Set()

	secret := "test_secret"
	authUser := entities.User{
		ID:       1,
		Login:    "login_auth",
		Password: getMD5Pass("password_ok"),
	}

	ctrl := gomock.NewController(t)
	authService := mock.NewMockAuthService(ctrl)
	authUserToken := "token"
	authService.EXPECT().BuildUserToken(authUser).Return(authUserToken, nil).AnyTimes()
	authService.EXPECT().GetUserByToken(authUserToken).Return(&authUser, nil).AnyTimes()
	authService.EXPECT().GetUserByToken("").Return(nil, errors.New("token error")).AnyTimes()
	authService.EXPECT().GetUser("login_auth").Return(&authUser).AnyTimes()

	loyaltyService := mock.NewMockLoyaltyService(ctrl)
	loyaltyService.EXPECT().GetUserBalance(authUser.ID).Return(entities.Balance{
		Current:   15,
		Withdrawn: 0,
	}).AnyTimes()
	loyaltyService.EXPECT().SaveWithdrawn(gomock.Any()).Return(nil).AnyTimes()

	handler := New(authService, loyaltyService, secret)

	handler.Mount()

	tests := []struct {
		name      string
		method    string
		code      int
		authToken string
		inBody    string
	}{
		{
			name:   "unauthorized user",
			method: http.MethodGet,
			code:   401,
		},
		{
			name:      "correct input",
			method:    http.MethodPost,
			code:      200,
			authToken: authUserToken,
			inBody:    `{"order":"1111111","sum":10}`,
		},
		{
			name:      "not enough balance",
			method:    http.MethodPost,
			code:      402,
			authToken: authUserToken,
			inBody:    `{"order":"1111111","sum":20}`,
		},
		{
			name:      "not correct order number",
			method:    http.MethodPost,
			code:      422,
			authToken: authUserToken,
			inBody:    `{"order":"1111112","sum":10}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := strings.NewReader(test.inBody)
			request, _ := http.NewRequest(test.method, "/api/user/balance/withdraw", body)
			if test.authToken != "" {
				request.Header.Set("Authorization", authUserToken)
			}
			rr := httptest.NewRecorder()

			handler.Router.ServeHTTP(rr, request)

			response := rr.Result()
			defer response.Body.Close()
			assert.Equal(t, test.code, response.StatusCode)
		})
	}
}

func getMD5Pass(p string) string {
	h := md5.New()
	h.Write([]byte(p))

	return hex.EncodeToString(h.Sum(nil))
}
