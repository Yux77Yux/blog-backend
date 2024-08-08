package server

import (
	"github.com/yux77yux/blog-backend/api"
	"github.com/yux77yux/blog-backend/utils/log_utils"
	"log"
	"net/http"
)

func Server() {
	mux := http.NewServeMux()
	api.UserHandlers(mux)

	//"github.com/yux77yux/blog-backend/internal/handlers/articles"
	//articles.ArticlesHandlers()

	if err := http.ListenAndServe(":3001", mux); err != nil {
		log.Println(err)
		log_utils.Logger.Printf("port has already occupied or others: %v", err)
	}
}
