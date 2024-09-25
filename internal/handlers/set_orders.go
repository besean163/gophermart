package handlers

import (
	"io"
	"net/http"

	"github.com/EClaesson/go-luhn"
	"github.com/besean163/gophermart/internal/entities"
)

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
		}

		w.WriteHeader(http.StatusConflict)
		return
	}

	order := entities.NewOrder(numOrder, user.ID)
	err = handler.LoyaltyService.SaveOrder(*order)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
