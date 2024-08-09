package user

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/yux77yux/blog-backend/internal/model"
	"github.com/yux77yux/blog-backend/utils/aliyun"
	"github.com/yux77yux/blog-backend/utils/jwt_utils"
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

type UserResponse struct {
	UserInfo *model.UserIncidental `json:"userInfo"` // 用适当的类型替代 interface{}
	Token    string                `json:"token"`
}

func SignIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)

		return
	}

	var user model.UsernameAndPassword
	jsonDecoder(w, r, &user)

	w.Header().Set("Content-Type", "application/json")

	userInfo, err := user_utils.SignIn(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := map[string]string{"err": err.Error()}
		json.NewEncoder(w).Encode(response)
		log.Println("Error:user SignIn: ", err)
		log_utils.Logger.Printf("Error:user SignIn: %v", err)
		return
	}

	tokenString, err := jwt_utils.GenerateJWT(userInfo.Uid)
	if err != nil {
		log.Println("Error:user : ", err)
	}

	result := UserResponse{
		UserInfo: userInfo,
		Token:    tokenString,
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

func ToBlack(tokenString string) error {
	token, err := jwt_utils.ParseJWT(tokenString)
	if err != nil {
		return fmt.Errorf("ToBlack ParseJWT invalid: %v", err)
	}

	claims := token.Claims.(jwt.MapClaims)
	jti, ok := claims["jti"].(string)
	if !ok {
		return fmt.Errorf("ToBlack jti invalid")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return fmt.Errorf("ToBlack exp invalid")
	}

	err = redis_utils.AddJtiToBlacklist(jti, exp)
	if err != nil {
		return fmt.Errorf("ToBlack AddJtiToBlacklist invalid: %v", err)
	}

	return nil
}

func SignOut(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)

		return
	}

	uid := r.FormValue("uid")
	tokenString := r.Header.Get("Authorization")

	w.Header().Set("Content-Type", "application/json")

	err := redis_utils.SetUserOnline(uid, false)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := map[string]string{"err": err.Error()}
		json.NewEncoder(w).Encode(response)
		log.Println("Error:user SignOut: ", err)
		log_utils.Logger.Printf("Error:user SignOut: %v", err)
		return
	}

	err = ToBlack(tokenString)
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

	uid := r.FormValue("uid")

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
		Uid:     uid,
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

	go (func() {
		err := user_utils.UpdateName(&modify_info)
		if err != nil {
			log.Println("Error:user UpdateName mysql: ", err)
			log_utils.Logger.Printf("Error:user UpdateName mysql: %v", err)
		}
	})()

	err := redis_utils.ModifyUserField(modify_info.Uid, "Name", modify_info.Name)
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

	go (func() {
		err := user_utils.UpdateBio(&modify_info)
		if err != nil {
			log.Println("Error:user UpdateBio mysql: ", err)
			log_utils.Logger.Printf("Error:user UpdateBio mysql: %v", err)
		}
	})()

	err := redis_utils.ModifyUserField(modify_info.Uid, "Bio", modify_info.Bio)
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

func AutoSignIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		w.WriteHeader(http.StatusInternalServerError)
		response := map[string]string{"err": "Authorization header is missing"}
		json.NewEncoder(w).Encode(response)
		log.Println("Error: Authorization header is missing")
		return
	}

	token, err := jwt_utils.ParseJWT(tokenString)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := map[string]string{"err": err.Error()}
		json.NewEncoder(w).Encode(response)
		log.Println("Error: ParseJWT: ", err)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		response := map[string]string{"err": "invalid claims"}
		json.NewEncoder(w).Encode(response)
		log.Println("Error: Authorization header is missing")
		return
	}

	uid, ok := claims["sub"].(string)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		response := map[string]string{"err": "sub claim is missing or not a string"}
		json.NewEncoder(w).Encode(response)
		log.Println("Error: Authorization header is missing")
		return
	}

	time.Sleep(500 * time.Millisecond)

	var currentUser *model.UserIncidental
	currentUser, err = redis_utils.GetUserFromRedis(uid)

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
