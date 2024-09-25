package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/besean163/gophermart/internal/logger"
	"go.uber.org/zap"
)

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
