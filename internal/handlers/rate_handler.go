package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"backend-assignment/internal/services"
	"backend-assignment/internal/utils"
)

type RateHandler struct {
	service *services.RateService
}

func NewRateHandler(service *services.RateService) *RateHandler {
	return &RateHandler{service: service}
}

func (h *RateHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	var raw map[string]json.RawMessage
	if err := utils.DecodeJSON(r.Body, &raw); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	userIDRaw, ok := raw["user_id"]
	if !ok {
		utils.WriteError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	var userID string
	if err := json.Unmarshal(userIDRaw, &userID); err != nil || strings.TrimSpace(userID) == "" {
		utils.WriteError(w, http.StatusBadRequest, "user_id is required")
		return
	}
	userID = strings.TrimSpace(userID)

	if _, ok := raw["payload"]; !ok {
		utils.WriteError(w, http.StatusBadRequest, "payload is required")
		return
	}

	if !h.service.Accept(userID) {
		utils.WriteError(w, http.StatusTooManyRequests, "rate limit exceeded")
		return
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]string{
		"message": "request accepted",
		"user_id": userID,
	})
}

func (h *RateHandler) Stats(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"users": h.service.Stats(),
	})
}
