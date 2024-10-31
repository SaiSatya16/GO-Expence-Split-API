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

type GroupHandler struct {
	groupRepo *repository.GroupRepository
}

func NewGroupHandler(groupRepo *repository.GroupRepository) *GroupHandler {
	return &GroupHandler{groupRepo: groupRepo}
}

func (h *GroupHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	var input models.GroupCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	if err := input.Validate(); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	group, err := h.groupRepo.Create(&input, userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "error creating group")
		return
	}

	response.JSON(w, http.StatusCreated, group)
}

func (h *GroupHandler) GetUserGroups(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	groups, err := h.groupRepo.GetUserGroups(userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "error fetching groups")
		return
	}

	response.JSON(w, http.StatusOK, groups)
}

func (h *GroupHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	groupID, err := strconv.Atoi(params["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid group ID")
		return
	}

	group, err := h.groupRepo.GetByID(groupID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "group not found")
		return
	}

	response.JSON(w, http.StatusOK, group)
}
