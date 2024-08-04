package user

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/yux77yux/blog-backend/internal/model"
	"github.com/yux77yux/blog-backend/utils/aliyun"
	"github.com/yux77yux/blog-backend/utils/user_utils"
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

	result, err := userutils.SignIn(user)
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

	err = userutils.AddUser(user)
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

	err = userutils.SignOut(id)
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

func FetchUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}

	uid := r.URL.Query().Get("uid")

	if uid == "" {
		http.Error(w, "UID is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	result, err := userutils.FetchUser(uid)
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

func FetchLatestUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
	idStr := r.URL.Query().Get("id")

	// 将 idStr 转换为 int 类型
	id, err := strconv.Atoi(idStr)
	if err != nil {
		// 处理转换错误
		http.Error(w, "Invalid id format", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	result, err := userutils.FetchLatestUser(id)
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

func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	// Parse the form to retrieve the file and other fields
	err := r.ParseMultipartForm(10 << 21) // Limit your file size to 10 MB
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	idStr := r.FormValue("id")
	// 将 idStr 转换为 int 类型
	id, err := strconv.Atoi(idStr)
	if err != nil {
		// 处理转换错误
		http.Error(w, "Invalid id format", http.StatusBadRequest)
		return
	}
	id32 := int32(id)

	var file_name string
	var profile_path string
	file, fileHeader, err := r.FormFile("profile")
	if err != nil {
		log.Println("Error retrieving file:", err)
		profile_path = ""
		w.WriteHeader(http.StatusInternalServerError)
		response := map[string]string{"err": err.Error()}
		json.NewEncoder(w).Encode(response)
		return
	} else {
		defer file.Close()

		file_name = "images/profiles/" + fileHeader.Filename
		profile_path, err = aliyun.UploadFile(file, file_name)
		if err != nil {
			log.Println("Error uploading file:", err)
			profile_path = ""
			w.WriteHeader(http.StatusInternalServerError)
			response := map[string]string{"err": err.Error()}
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	modify_info := model.UserModifyProfile{
		Id:      id32,
		Profile: profile_path,
	}

	err = userutils.UpdateProfile(modify_info)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := map[string]string{"err": err.Error()}
		json.NewEncoder(w).Encode(response)
		log.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	response := map[string]string{"success": "OK!"}
	json.NewEncoder(w).Encode(response)
}

func UpdateName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var modify_info model.UserModifyName

	if err := json.NewDecoder(r.Body).Decode(&modify_info); err != nil {
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

	err := userutils.UpdateName(modify_info)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := map[string]string{"err": err.Error()}
		json.NewEncoder(w).Encode(response)
		log.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	response := map[string]string{"success": "OK!"}
	json.NewEncoder(w).Encode(response)
}

func UpdateBio(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var modify_info model.UserModifyBio

	if err := json.NewDecoder(r.Body).Decode(&modify_info); err != nil {
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

	err := userutils.UpdateBio(modify_info)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := map[string]string{"err": err.Error()}
		json.NewEncoder(w).Encode(response)
		log.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	response := map[string]string{"success": "OK!"}
	json.NewEncoder(w).Encode(response)
}
