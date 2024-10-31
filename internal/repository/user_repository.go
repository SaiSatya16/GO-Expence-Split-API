package repository

import (
	"expense-sharing-api/internal/models"

	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.UserRegister, passwordHash string) (*models.User, error) {
	query := `
        INSERT INTO users (email, full_name, password_hash)
        VALUES (?, ?, ?)
        RETURNING user_id, email, full_name, created_at`

	var created models.User
	err := r.db.QueryRowx(query, user.Email, user.FullName, passwordHash).StructScan(&created)
	if err != nil {
		return nil, err
	}

	return &created, nil
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	query := `SELECT * FROM users WHERE email = ?`
	err := r.db.Get(&user, query, email)
	return &user, err
}

func (r *UserRepository) GetByID(userID int) (*models.User, error) {
	var user models.User
	query := `SELECT * FROM users WHERE user_id = ?`
	err := r.db.Get(&user, query, userID)
	return &user, err
}
