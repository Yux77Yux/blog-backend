package user

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/yux77yux/blog-backend/internal/model"
	"github.com/yux77yux/blog-backend/utils"
)

func SignIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}

	var user model.UsernameAndPassword

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		if err == io.EOF {
			http.Error(w, "Empty request body", http.StatusBadRequest)
		} else if syntaxErr, ok := err.(*json.SyntaxError); ok {
			http.Error(w, fmt.Sprintf("Syntax error at byte offset %d", syntaxErr.Offset), http.StatusBadRequest)
		} else if unmarshalErr, ok := err.(*json.UnmarshalTypeError); ok {
			http.Error(w, fmt.Sprintf("Incorrect type for field %s", unmarshalErr.Field), http.StatusBadRequest)
		} else {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")

	result, err := utils.SignIn(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := map[string]string{"err": err.Error()}
		json.NewEncoder(w).Encode(response)
		log.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func SignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}

	var user model.UsernameAndPassword

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		if err == io.EOF {
			http.Error(w, "Empty request body", http.StatusBadRequest)
		} else if syntaxErr, ok := err.(*json.SyntaxError); ok {
			http.Error(w, fmt.Sprintf("Syntax error at byte offset %d", syntaxErr.Offset), http.StatusBadRequest)
		} else if unmarshalErr, ok := err.(*json.UnmarshalTypeError); ok {
			http.Error(w, fmt.Sprintf("Incorrect type for field %s", unmarshalErr.Field), http.StatusBadRequest)
		} else {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = utils.AddUser(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := map[string]string{"err": err.Error()}
		json.NewEncoder(w).Encode(response)
		log.Println(err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func SignOut(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}

	idStr := r.FormValue("id")

	// 将 idStr 转换为 int 类型
	id, err := strconv.Atoi(idStr)
	if err != nil {
		// 处理转换错误
		http.Error(w, "Invalid id format", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = utils.SignOut(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := map[string]string{"err": err.Error()}
		json.NewEncoder(w).Encode(response)
		log.Println(err)
		return
	}
	w.WriteHeader(http.StatusOK)
	response := map[string]string{"success": "Sign out successful"}
	json.NewEncoder(w).Encode(response)
}

func FetchUserDetail(w http.ResponseWriter, r *http.Request) {

}
