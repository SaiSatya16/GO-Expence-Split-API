package handlers

import (
	"encoding/json"
	"expense-sharing-api/internal/middleware"
	"expense-sharing-api/internal/models"
	"expense-sharing-api/internal/repository"
	"expense-sharing-api/pkg/response"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type ExpenseHandler struct {
	expenseRepo *repository.ExpenseRepository
	groupRepo   *repository.GroupRepository
}

func NewExpenseHandler(expenseRepo *repository.ExpenseRepository, groupRepo *repository.GroupRepository) *ExpenseHandler {
	return &ExpenseHandler{
		expenseRepo: expenseRepo,
		groupRepo:   groupRepo,
	}
}

func (h *ExpenseHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	var input models.ExpenseCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	if err := input.Validate(); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Verify user is member of the group
	group, err := h.groupRepo.GetByID(input.GroupID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "group not found")
		return
	}

	isMember := false
	for _, member := range group.Members {
		if member.UserID == userID {
			isMember = true
			break
		}
	}

	if !isMember {
		response.Error(w, http.StatusForbidden, "user is not a member of this group")
		return
	}

	expense, err := h.expenseRepo.Create(&input, userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "error creating expense")
		return
	}

	response.JSON(w, http.StatusCreated, expense)
}

func (h *ExpenseHandler) GetGroupExpenses(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	groupID, err := strconv.Atoi(params["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid group ID")
		return
	}

	expenses, err := h.expenseRepo.GetGroupExpenses(groupID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "error fetching expenses")
		return
	}

	response.JSON(w, http.StatusOK, expenses)
}

func (h *ExpenseHandler) GetBalanceSheet(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)
	params := mux.Vars(r)
	groupID, err := strconv.Atoi(params["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid group ID")
		return
	}

	balances, err := h.expenseRepo.GetUserBalance(userID, groupID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "error fetching balance sheet")
		return
	}

	response.JSON(w, http.StatusOK, balances)
}
