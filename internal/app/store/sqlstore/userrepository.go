package sqlstore

import (
	"github.com/nekitkas/restAPI/internal/app/model"
)

type UserRepository struct {
	store *Store
}

func (r *UserRepository) Create(user *model.User) error {
	if err := user.BeforeCreate(); err != nil {
		return err
	}

	if err := r.store.db.QueryRow(
		"INSERT INTO users (email, encrypted_password) VALUES ($1, $2) RETURNING id",
		user.Email,
		user.EncryptedPassword,
	).Scan(&user.ID); err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) Find(id int) (*model.User, error) {
	user := &model.User{}
	if err := r.store.db.QueryRow(
		"SELECT id,email,encrypted_password FROM users WHERE id = $1",
		id).Scan(
		&user.ID,
		&user.Email,
		&user.EncryptedPassword,
	); err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	user := &model.User{}
	if err := r.store.db.QueryRow(
		"SELECT id,email,encrypted_password FROM users WHERE email = $1",
		email).Scan(
		&user.ID,
		&user.Email,
		&user.EncryptedPassword,
	); err != nil {
		return nil, err
	}
	return user, nil
}
