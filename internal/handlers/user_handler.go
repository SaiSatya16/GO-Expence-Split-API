package handlers

import (
	"encoding/json"
	"expense-sharing-api/internal/models"
	"expense-sharing-api/internal/repository"
	"expense-sharing-api/pkg/auth"
	"expense-sharing-api/pkg/hash"
	"expense-sharing-api/pkg/response"
	"net/http"
)

type UserHandler struct {
	userRepo *repository.UserRepository
}

func NewUserHandler(userRepo *repository.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input models.UserRegister
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	if err := input.Validate(); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Hash password
	passwordHash, err := hash.HashPassword(input.Password)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "error processing request")
		return
	}

	// Create user
	user, err := h.userRepo.Create(&input, passwordHash)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "error creating user")
		return
	}

	// Generate token
	token, err := auth.GenerateToken(user.UserID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "error generating token")
		return
	}

	response.JSON(w, http.StatusCreated, map[string]interface{}{
		"user":  user,
		"token": token,
	})
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input models.UserLogin
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	user, err := h.userRepo.GetByEmail(input.Email)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if !hash.CheckPassword(input.Password, user.PasswordHash) {
		response.Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := auth.GenerateToken(user.UserID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "error generating token")
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"user":  user,
		"token": token,
	})
}
