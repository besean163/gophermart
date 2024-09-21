package entities

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
)

type User struct {
	ID       int `gorm:"primarykey"`
	Login    string
	Password string
}

func (user User) Validate() error {
	if user.Password == "" {
		return errors.New("empty password")
	}
	return nil
}

func (user *User) HashingPassword() error {

	h := md5.New()
	_, err := h.Write([]byte(user.Password))
	if err != nil {
		return err
	}

	user.Password = hex.EncodeToString(h.Sum(nil))
	return nil
}
