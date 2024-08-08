package user

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/yux77yux/blog-backend/internal/model"
	"github.com/yux77yux/blog-backend/utils/aliyun"
	"github.com/yux77yux/blog-backend/utils/log_utils"
	"github.com/yux77yux/blog-backend/utils/redis_utils"
	"github.com/yux77yux/blog-backend/utils/user_utils"
)

func jsonDecoder(w http.ResponseWriter, r *http.Request, target interface{}) {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
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
}

func SignIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)

		return
	}

	var user model.UsernameAndPassword
	jsonDecoder(w, r, &user)

	w.Header().Set("Content-Type", "application/json")

	result, err := user_utils.SignIn(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := map[string]string{"err": err.Error()}
		json.NewEncoder(w).Encode(response)
		log.Println("Error:user SignIn: ", err)
		log_utils.Logger.Printf("Error:user SignIn: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func SignUp(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)

		return
	}

	var user model.UsernameAndPassword

	jsonDecoder(w, r, &user)

	w.Header().Set("Content-Type", "application/json")

	err := user_utils.AddUser(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := map[string]string{"err": err.Error()}
		json.NewEncoder(w).Encode(response)
		log.Println("Error:user AddUser: ", err)
		log_utils.Logger.Printf("Error:user AddUser: %v", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func SignOut(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)

		return
	}

	idStr := r.FormValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id format", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = redis_utils.SetUserOnline(int32(id), false)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := map[string]string{"err": err.Error()}
		json.NewEncoder(w).Encode(response)
		log.Println("Error:user SignOut: ", err)
		log_utils.Logger.Printf("Error:user SignOut: %v", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	response := map[string]string{"success": "Sign out successful"}
	json.NewEncoder(w).Encode(response)
}

func FetchUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)

		return
	}

	uid := r.URL.Query().Get("uid")

	if uid == "" {
		http.Error(w, "UID is required", http.StatusBadRequest)
		return
	}

	time.Sleep(500 * time.Millisecond)

	w.Header().Set("Content-Type", "application/json")

	var currentUser *model.UserIncidental
	currentUser, err := redis_utils.GetUserFromRedis(uid)

	if err != nil {
		currentUser, err = user_utils.FetchUser(uid)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response := map[string]string{"err": err.Error()}
			json.NewEncoder(w).Encode(response)
			log.Println("Error:user FetchUser: ", err)
			log_utils.Logger.Printf("Error:user FetchUser: %v", err)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(currentUser)
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
	uid := strconv.Itoa(id + 100000000)
	id32 := int32(id)

	var file_name string
	var profile_path string
	file, fileHeader, err := r.FormFile("profile")
	if err != nil {
		log.Println("Error retrieving file:", err)
		log_utils.Logger.Printf("Error: retrieving file %v", err)

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
			log.Println("Error:user UploadFile: uploading file:", err)
			log_utils.Logger.Printf("Error:user UploadFile: uploading file %v", err)

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

	go (func() {
		err = user_utils.UpdateProfile(&modify_info)
		if err != nil {
			log.Println("Error:user UpdateProfile: ", err)
			log_utils.Logger.Printf("Error:user UpdateProfile: %v", err)
		}
	})()

	err = redis_utils.ModifyUserField(uid, "Profile", modify_info.Profile)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := map[string]string{"err": err.Error()}
		json.NewEncoder(w).Encode(response)
		log.Println("Error:user UpdateProfile: ", err)
		log_utils.Logger.Printf("Error:user UpdateProfile: %v", err)
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

	jsonDecoder(w, r, &modify_info)

	uid := strconv.Itoa(int(modify_info.Id + 100000000))

	go (func() {
		err := user_utils.UpdateName(&modify_info)
		if err != nil {
			log.Println("Error:user UpdateName mysql: ", err)
			log_utils.Logger.Printf("Error:user UpdateName mysql: %v", err)
		}
	})()

	err := redis_utils.ModifyUserField(uid, "Name", modify_info.Name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := map[string]string{"err": err.Error()}
		json.NewEncoder(w).Encode(response)
		log.Println("Error:user UpdateName: ", err)
		log_utils.Logger.Printf("Error:user UpdateName: %v", err)
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

	jsonDecoder(w, r, &modify_info)

	uid := strconv.Itoa(int(modify_info.Id + 100000000))

	go (func() {
		err := user_utils.UpdateBio(&modify_info)
		if err != nil {
			log.Println("Error:user UpdateBio mysql: ", err)
			log_utils.Logger.Printf("Error:user UpdateBio mysql: %v", err)
		}
	})()

	err := redis_utils.ModifyUserField(uid, "Bio", modify_info.Bio)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := map[string]string{"err": err.Error()}
		json.NewEncoder(w).Encode(response)
		log.Println("Error:user UpdateBio: ", err)
		log_utils.Logger.Printf("Error:user UpdateBio: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	response := map[string]string{"success": "OK!"}
	json.NewEncoder(w).Encode(response)
}
