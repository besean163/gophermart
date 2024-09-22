package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/EClaesson/go-luhn"
	"github.com/besean163/gophermart/internal/entities"
)

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
