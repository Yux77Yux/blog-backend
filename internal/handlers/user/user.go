package user

import (
	"encoding/json"
	"github.com/yux77yux/blog-backend/internal/model"
	"net/http"
)

func signIn(w http.ResponseWriter, r *http.Request) {
	user := model.UsernameAndPassword{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
	}

	w.Header().Set("Content-Type", "application/json")

	response := map[string]string{"message": "Sign in successful"}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func signUp(w http.ResponseWriter, r *http.Request) {

}

func fetchUserDetail(w http.ResponseWriter, r *http.Request) {

}

func UserHandlers() {
	http.HandleFunc("/api/user/sign-in", signIn)
	http.HandleFunc("/api/user/sign-up", signUp)
	http.HandleFunc("/api/user/fetchUserDetail", fetchUserDetail)
}
