package handler

import (
	"auth_service/internal/service"
	"encoding/json"
	"net/http"
	"response"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) RegisterHandler(rw http.ResponseWriter, r *http.Request) {
	req := &registerRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(rw, "invalid request body", http.StatusBadRequest)
		return
	}

	err := h.authService.RegisterUser(r.Context(), req.Email, req.Password)
	if err != nil {
		response.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	response.SetContentJSON(rw)
	rw.WriteHeader(http.StatusCreated)
	res := &registerResponse{
		Message: "success register user",
	}
	_ = json.NewEncoder(rw).Encode(res)
}

func (h *AuthHandler) LoginHandler(rw http.ResponseWriter, r *http.Request) {
	req := &loginRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(rw, "invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.authService.LoginUser(r.Context(), req.Email, req.Password)
	if err != nil {
		response.Error(rw, err.Error(), http.StatusUnauthorized)
		return
	}

	response.SetContentJSON(rw)
	rw.WriteHeader(http.StatusOK)
	res := &loginResponse{
		Message: "success login user",
		Token:   token,
	}
	_ = json.NewEncoder(rw).Encode(res)
}
