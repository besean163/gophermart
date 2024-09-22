package handlers

import (
	"errors"
	"net/http"

	"github.com/besean163/gophermart/internal/entities"
	"github.com/go-chi/chi/v5"
)

var (
	ErrCannotGetUser = errors.New("can't get user by context")
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

func NewHandlers(
	authService AuthService,
	loyaltyService LoyaltyService,
	hashSecret string,
) Handler {
	h := Handler{
		Router:         chi.NewRouter(),
		AuthService:    authService,
		LoyaltyService: loyaltyService,
		HashSecret:     hashSecret,
	}

	h.mount()
	return h
}

func (handler Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.Router.ServeHTTP(w, r)
}

func (handler Handler) mount() {
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

func getRequestUser(r *http.Request) (*entities.User, error) {
	user, ok := r.Context().Value(userKeyContext("user")).(entities.User)
	if !ok {
		return nil, ErrCannotGetUser

	}
	return &user, nil
}
