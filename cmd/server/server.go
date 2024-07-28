package server

import (
	"github.com/yux77yux/blog-backend/internal/handlers/articles"
	"github.com/yux77yux/blog-backend/internal/handlers/user"
	"log"
	"net/http"
)

func Server() {
	user.UserHandlers()

	articles.ArticlesHandlers()

	if err := http.ListenAndServe(":3001", nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
