package http

import (
	"net/http"

	"quickflow/utils"
)

func GetCSRF(w http.ResponseWriter, r *http.Request) {
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
