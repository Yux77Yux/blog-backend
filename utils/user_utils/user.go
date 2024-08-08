package user_utils

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"

	"github.com/go-sql-driver/mysql"
	"github.com/yux77yux/blog-backend/config"
	"github.com/yux77yux/blog-backend/internal/model"
	"github.com/yux77yux/blog-backend/utils/redis_utils"
)

func SignIn(user model.UsernameAndPassword) (*model.UserIncidental, error) {
	config.OpenDB()
	defer config.DB.Close()

	query := `
	SELECT 
	user.id
	FROM user
	WHERE user.username = ? 
	AND user.password = ? 
	`

	var id int32
	err := config.DB.QueryRow(query, user.Username, user.Password).Scan(
		&id,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user_utils SignIn QueryRow: user not found")
		}
		return nil, fmt.Errorf("user_utils SignIn QueryRow: user not found %v", err)
	}

	err = redis_utils.SetUserOnline(id, true)
	if err != nil {
		return nil, fmt.Errorf("user_utils SetUserOnline: setUserOnline failure %v", err)
	}

	uid := strconv.Itoa(int(id + 100000000))

	var currentUser *model.UserIncidental
	currentUser, err = redis_utils.GetUserFromRedis(uid)
	if err != nil {
		currentUser, err = FetchLatestUser(id)
		if err != nil {
			_ = redis_utils.SetUserOnline(id, false)
			return nil, fmt.Errorf("user_utils FetchLatestUser: fetchLatestUser failure %v", err)
		}

		go redis_utils.StoreUserInRedis(currentUser)
	}

	return currentUser, nil
}

func AddUser(user model.UsernameAndPassword) error {
	pattern := `^[A-Za-z!@#$%^&*()_+={}\[\]:;"'<>,.?/]+([A-Za-z0-9!@#$%^&*()_+={}\[\]:;"'<>,.?/]){7,}$`
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	matched := re.MatchString(user.Password)
	if !matched {
		return fmt.Errorf("user_utils AddUser: password or username not match the rule")
	}

	config.OpenDB()
	defer config.DB.Close()

	tx, err := config.DB.Begin()
	if err != nil {
		return err
	}

	query := `
	INSERT INTO user (username, password) VALUES
	(?, ?)
	`

	result, err := config.DB.Exec(query, user.Username, user.Password)
	if err != nil {
		tx.Rollback()
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return fmt.Errorf("user_utils AddUser Exec: 用户名已经存在")
		}
		return err
	}

	userID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("user_utils AddUser LastInsertId: %v", err)
	}
	uid := 100000000 + userID
	uidStr := strconv.FormatInt(uid, 10)
	query = `
	INSERT INTO user_incidental (uid, id, name, bio, profile, popularity) VALUES 
	(?, ?, '', '', '', 1.0)
	`
	_, err = tx.Exec(query, uidStr, userID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("user_utils AddUser Exec: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("user_utils AddUser Commit: %v", err)
	}

	return nil
}

func FetchUser(uid string) (*model.UserIncidental, error) {
	config.OpenDB()
	defer config.DB.Close()

	tx, err := config.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("user_utils FetchUser Begin: %v", err)
	}

	query := `
	SELECT 
	user_incidental.id AS id,
	user_incidental.uid AS uid,
	user_incidental.name AS name,
	user_incidental.bio AS bio,
	user_incidental.profile AS profile,
	user_incidental.popularity AS popularity

	FROM user_incidental
	WHERE user_incidental.uid = ? 
	`

	var currentUser model.UserIncidental
	err = tx.QueryRow(query, uid).Scan(
		&currentUser.Id,
		&currentUser.Uid,
		&currentUser.Name,
		&currentUser.Bio,
		&currentUser.Profile,
		&currentUser.Popularity,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user_utils FetchUser QueryRow ErrNoRows: %v", err)
		}
		return nil, fmt.Errorf("user_utils FetchUser QueryRow: %v", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("user_utils FetchUser Commit: %v", err)
	}

	currentUser.Status, err = redis_utils.GetUserOnline(currentUser.Id)
	if err != nil {
		return &currentUser, fmt.Errorf("user_utils FetchUser GetUserOnline: %v", err)
	}

	return &currentUser, nil
}

func FetchLatestUser(id int32) (*model.UserIncidental, error) {
	config.OpenDB()
	defer config.DB.Close()

	tx, err := config.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("user_utils FetchLatestUser Begin: %v", err)
	}

	query := `
	SELECT 
	user_incidental.id AS id,
	user_incidental.uid AS uid,
	user_incidental.name AS name,
	user_incidental.bio AS bio,
	user_incidental.profile AS profile,
	user_incidental.popularity AS popularity

	FROM user_incidental
	WHERE user_incidental.id = ? 
	`

	var currentUser model.UserIncidental
	err = tx.QueryRow(query, id).Scan(
		&currentUser.Id,
		&currentUser.Uid,
		&currentUser.Name,
		&currentUser.Bio,
		&currentUser.Profile,
		&currentUser.Popularity,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user_utils FetchLatestUser QueryRow ErrNoRows: %v", err)
		}
		return nil, fmt.Errorf("user_utils FetchLatestUser QueryRow: %v", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("user_utils FetchLatestUser Commit: %v", err)
	}

	currentUser.Status, err = redis_utils.GetUserOnline(id)
	if err != nil {
		return &currentUser, fmt.Errorf("user_utils FetchLatestUser GetUserOnline: %v", err)
	}

	return &currentUser, nil
}

func UpdateProfile(modify_info *model.UserModifyProfile) error {
	config.OpenDB()
	defer config.DB.Close()

	tx, err := config.DB.Begin()
	if err != nil {
		return fmt.Errorf("user_utils UpdateProfile Begin: %v", err)
	}

	query := `
	UPDATE user_incidental
	SET profile = ?
	WHERE user_incidental.id = ? 
	`

	_, err = tx.Exec(query, modify_info.Profile, modify_info.Id)

	if err != nil {
		return fmt.Errorf("user_utils UpdateProfile Exec: %v", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("user_utils UpdateProfile Commit: %v", err)
	}

	return nil
}

func UpdateName(modify_info *model.UserModifyName) error {
	config.OpenDB()
	defer config.DB.Close()

	tx, err := config.DB.Begin()
	if err != nil {
		return fmt.Errorf("user_utils UpdateName Begin: %v", err)
	}

	query := `
	UPDATE user_incidental
	SET name = ?
	WHERE user_incidental.id = ? 
	`

	_, err = tx.Exec(query, modify_info.Name, modify_info.Id)

	if err != nil {
		return fmt.Errorf("user_utils UpdateName Exec: %v", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("user_utils UpdateName Commit: %v", err)
	}

	return nil
}

func UpdateBio(modify_info *model.UserModifyBio) error {
	config.OpenDB()
	defer config.DB.Close()

	tx, err := config.DB.Begin()
	if err != nil {
		return fmt.Errorf("user_utils UpdateBio Begin: %v", err)
	}

	query := `
	UPDATE user_incidental
	SET bio = ?
	WHERE user_incidental.id = ? 
	`

	_, err = tx.Exec(query, modify_info.Bio, modify_info.Id)

	if err != nil {
		return fmt.Errorf("user_utils UpdateBio Exec: %v", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("user_utils UpdateBio Commit: %v", err)
	}

	return nil
}
