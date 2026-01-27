package api

import "net/http"

func NewRouter(handler *AuthHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /login", handler.Login)
	mux.HandleFunc("POST /register", handler.Register)
	mux.HandleFunc("POST /refresh", handler.Refresh)
	mux.HandleFunc("POST /logout", handler.Logout)

	return mux
}
