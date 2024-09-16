package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/EClaesson/go-luhn"
	"github.com/besean163/gophermart/internal/entities"
	"github.com/besean163/gophermart/internal/logger"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type AuthService interface {
	GetUser(login string) *entities.User
	SaveUser(entities.User) error
	BuildUserToken(entities.User) (string, error)
	GetUserByToken(token string) (*entities.User, error)
}

type LoyaltyService interface {
	GetOrder(orderID string) *entities.Order
	GetUserOrders(userID int) []*entities.Order
	GetUserWithdrawals(userID int) []*entities.Withdrawn
	GetUserBalance(userID int) entities.Balance
	SaveOrder(entities.Order) error
	SaveWithdrawn(entities.Withdrawn) error
}

type JobService interface {
	GetNotCalcOrders() []*entities.Order
	SaveOrder(entities.Order) error
}

type Handler struct {
	Router         *chi.Mux
	AuthService    AuthService
	LoyaltyService LoyaltyService
	HashSecret     string
}

func New(
	authService AuthService,
	loyaltyService LoyaltyService,
	hashSecret string,
) Handler {
	return Handler{
		Router:         chi.NewRouter(),
		AuthService:    authService,
		LoyaltyService: loyaltyService,
		HashSecret:     hashSecret,
	}
}

func (handler Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.Router.ServeHTTP(w, r)
}

func (handler Handler) Mount() {
	handler.Router.Route("/api/user", func(r chi.Router) {
		r.Post("/login", handler.Login)
		r.Post("/register", handler.Register)
		r.Group(func(r chi.Router) {
			r.Use(handler.AuthMiddleware)
			r.Get("/orders", handler.GetOrders)
			r.Get("/withdrawals", handler.GetBalanceHistory)
			r.Post("/orders", handler.SetOrders)
			r.Route("/balance", func(r chi.Router) {
				r.Get("/", handler.GetBalance)
				r.Post("/withdraw", handler.ChangeBalance)
			})
		})
	})
}

func (handler Handler) Login(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var inputUser entities.User
	err = json.Unmarshal(body, &inputUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = inputUser.Validate()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	inputUser.HashingPassword()
	existUser := handler.AuthService.GetUser(inputUser.Login)
	if existUser == nil || existUser.Password != inputUser.Password {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token, err := handler.AuthService.BuildUserToken(*existUser)
	if err != nil {
		logger.Get().Warn("can't set token", zap.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", token)
}

func (handler Handler) Register(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var inputUser entities.User
	err = json.Unmarshal(body, &inputUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = inputUser.Validate()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	existUser := handler.AuthService.GetUser(inputUser.Login)
	if existUser != nil {
		w.WriteHeader(http.StatusConflict)
		return
	}

	inputUser.HashingPassword()
	handler.AuthService.SaveUser(inputUser)

	token, err := handler.AuthService.BuildUserToken(inputUser)
	if err != nil {
		logger.Get().Warn("can't set token", zap.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", token)
}

func (handler Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	user, err := getRequestUser(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	balance := handler.LoyaltyService.GetUserBalance(user.ID)
	body, err := json.Marshal(balance)
	if err != nil {
		logger.Get().Warn("json marshal error", zap.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func (handler Handler) GetBalanceHistory(w http.ResponseWriter, r *http.Request) {
	user, err := getRequestUser(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	withdrawns := handler.LoyaltyService.GetUserWithdrawals(user.ID)
	if withdrawns == nil {
		logger.Get().Warn("nil withdrawns", zap.String("error", "nil withdrawns"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(withdrawns) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	body, err := json.Marshal(withdrawns)
	if err != nil {
		logger.Get().Warn("json marshal error", zap.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func (handler Handler) ChangeBalance(w http.ResponseWriter, r *http.Request) {
	user, err := getRequestUser(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if r.Body == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var withdrawn entities.Withdrawn
	err = json.Unmarshal(body, &withdrawn)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if withdrawn.Sum <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	isValidLuna, err := luhn.IsValid(withdrawn.OrderNumber)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !isValidLuna {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	balance := handler.LoyaltyService.GetUserBalance(user.ID)
	if balance.Current < withdrawn.Sum {
		w.WriteHeader(http.StatusPaymentRequired)
		return
	}

	withdrawn.UserID = user.ID
	withdrawn.ProccesedAt = time.Now()
	err = handler.LoyaltyService.SaveWithdrawn(withdrawn)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (handler Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	user, err := getRequestUser(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	orders := handler.LoyaltyService.GetUserOrders(user.ID)
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	body, err := json.Marshal(orders)
	if err != nil {
		logger.Get().Warn("json marshal error", zap.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func (handler Handler) SetOrders(w http.ResponseWriter, r *http.Request) {
	user, err := getRequestUser(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	numOrder := string(body)

	isValidLuna, err := luhn.IsValid(numOrder)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !isValidLuna {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	existOrder := handler.LoyaltyService.GetOrder(numOrder)
	if existOrder != nil {
		if existOrder.UserID == user.ID {
			w.WriteHeader(http.StatusOK)
			return
		} else {
			w.WriteHeader(http.StatusConflict)
			return
		}
	}

	order := entities.NewOrder(numOrder, user.ID)
	err = handler.LoyaltyService.SaveOrder(*order)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func getRequestUser(r *http.Request) (*entities.User, error) {
	user, ok := r.Context().Value(userKeyContext("user")).(entities.User)
	if !ok {
		return nil, errors.New("can't get user by context")

	}
	return &user, nil
}
