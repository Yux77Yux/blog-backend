package api

import (
	"github.com/yux77yux/blog-backend/config"
	"github.com/yux77yux/blog-backend/internal/handlers/user"
	"net/http"
)

func UserHandlers(mux *http.ServeMux) {
	mux.Handle("/api/user/sign-in", config.CorsMiddleware(http.HandlerFunc(user.SignIn)))
	mux.Handle("/api/user/token-sign-in", config.CorsMiddleware(http.HandlerFunc(user.AutoSignIn)))
	mux.Handle("/api/user/sign-up", config.CorsMiddleware(http.HandlerFunc(user.SignUp)))
	mux.Handle("/api/user/sign-out", config.CorsMiddleware(http.HandlerFunc(user.SignOut)))
	mux.Handle("/api/user/fetch-user", config.CorsMiddleware(http.HandlerFunc(user.FetchUser)))
	mux.Handle("/api/user/update-profile", config.CorsMiddleware(http.HandlerFunc(user.UpdateProfile)))
	mux.Handle("/api/user/update-name", config.CorsMiddleware(http.HandlerFunc(user.UpdateName)))
	mux.Handle("/api/user/update-bio", config.CorsMiddleware(http.HandlerFunc(user.UpdateBio)))
}
