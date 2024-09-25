package inmem

import "github.com/besean163/gophermart/internal/entities"

type Storage struct {
	Users []*entities.User
}

func New() *Storage {
	return &Storage{
		Users: make([]*entities.User, 0),
	}
}

func (storage *Storage) GetUser(login string) *entities.User {
	for _, user := range storage.Users {
		if user.Login == login {
			return user
		}
	}
	return nil
}

func (storage *Storage) SaveUser(user entities.User) error {
	storage.Users = append(storage.Users, &user)
	return nil
}
