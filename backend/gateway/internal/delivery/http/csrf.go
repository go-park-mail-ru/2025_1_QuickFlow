package http

import (
	"net/http"

	"quickflow/gateway/utils"
)

type CSRFHandler struct{}

func NewCSRFHandler() *CSRFHandler {
	return &CSRFHandler{}
}

// GetCSRF godoc
// @Summary Get CSRF token
// @Description Get CSRF token
// @Tags CSRF
// @Accept json
// @Produce json
// @Success 200 {object} forms.CSRFResponse "CSRF token"
// @Failure 500 {object} forms.ErrorForm "Server error"
// @Router /api/csrf [get]
func (c *CSRFHandler) GetCSRF(w http.ResponseWriter, r *http.Request) {
	token, err := utils.GenerateCSRFToken()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    token,
		HttpOnly: true,
		Secure:   true,
	})

	w.Header().Set("X-CSRF-Token", token)
}
