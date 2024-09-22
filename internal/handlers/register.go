package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/besean163/gophermart/internal/entities"
	"github.com/besean163/gophermart/internal/logger"
	"go.uber.org/zap"
)

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
