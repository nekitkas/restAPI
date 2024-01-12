package teststore

import (
	"errors"
	"github.com/nekitkas/restAPI/internal/app/model"
)

type UserRepository struct {
	store *Store
	users map[int]*model.User
}

func (r *UserRepository) Create(user *model.User) error {
	if err := user.BeforeCreate(); err != nil {
		return err
	}

	r.users[user.ID] = user

	return nil
}

func (r *UserRepository) Find(id int) (*model.User, error) {
	user, ok := r.users[id]
	if !ok {
		return nil, errors.New("not found")
	}

	return user, nil
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}

	return nil, errors.New("not found")
}
