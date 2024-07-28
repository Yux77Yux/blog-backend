package articles

import (
	"net/http"
)

func ArticlesHandlers() {
	http.HandleFunc("/api/articles/", func(w http.ResponseWriter, r *http.Request) {})
}
