package userutils

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"

	"github.com/go-sql-driver/mysql"
	"github.com/yux77yux/blog-backend/config"
	"github.com/yux77yux/blog-backend/internal/model"
)

func SignIn(user model.UsernameAndPassword) (model.UserIncidental, error) {
	db, err := sql.Open("mysql", config.ConnectionStr)
	if err != nil {
		return model.UserIncidental{}, err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return model.UserIncidental{}, err
	}

	query := `
	SELECT 
	user_incidental.id AS id,
	user_incidental.uid AS uid,
	user_incidental.name AS name,
	user_incidental.bio AS bio,
	user_incidental.profile AS profile,
	user_incidental.status AS status,
	user_incidental.popularity AS popularity

	FROM user
	INNER JOIN user_incidental ON user.id = user_incidental.id
	WHERE user.username = ? 
	AND user.password = ? 
	`

	var currentUser model.UserIncidental
	err = tx.QueryRow(query, user.Username, user.Password).Scan(
		&currentUser.Id,
		&currentUser.Uid,
		&currentUser.Name,
		&currentUser.Bio,
		&currentUser.Profile,
		&currentUser.Status,
		&currentUser.Popularity,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.UserIncidental{}, fmt.Errorf("user not found")
		}
		return model.UserIncidental{}, err
	}

	// 更新状态为 true
	updateQuery := `
    UPDATE user_incidental
    SET status = 1
    WHERE id = ?
    `
	_, err = tx.Exec(updateQuery, currentUser.Id)
	if err != nil {
		tx.Rollback()
		return model.UserIncidental{}, err
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return model.UserIncidental{}, err
	}

	currentUser.Status = true

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
		return fmt.Errorf("password or username not match the rule")
	}

	db, err := sql.Open("mysql", config.ConnectionStr)
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	query := `
	INSERT INTO user (username, password) VALUES
	(?, ?)
	`

	result, err := db.Exec(query, user.Username, user.Password)
	if err != nil {
		tx.Rollback()
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return fmt.Errorf("用户名已经存在")
		}
		return err
	}

	userID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return err
	}
	uid := 100000000 + userID
	uidStr := strconv.FormatInt(uid, 10)
	query = `
	INSERT INTO user_incidental (uid, id, name, bio, profile, status, popularity) VALUES 
	(?, ?, '', '', '', 0, 1.0)
	`
	_, err = tx.Exec(query, uidStr, userID)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func SignOut(id int) error {
	db, err := sql.Open("mysql", config.ConnectionStr)
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	query := `
	UPDATE user_incidental
	SET status = 0
	WHERE id = ? 
	`

	_, err = tx.Exec(query, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func FetchUser(uid string) (model.UserIncidental, error) {
	db, err := sql.Open("mysql", config.ConnectionStr)
	if err != nil {
		return model.UserIncidental{}, err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return model.UserIncidental{}, err
	}

	query := `
	SELECT 
	user_incidental.id AS id,
	user_incidental.uid AS uid,
	user_incidental.name AS name,
	user_incidental.bio AS bio,
	user_incidental.profile AS profile,
	user_incidental.status AS status,
	user_incidental.popularity AS popularity

	FROM user_incidental
	WHERE user_incidental.uid = ? 
	`

	var currentUser model.UserIncidental
	err = db.QueryRow(query, uid).Scan(
		&currentUser.Id,
		&currentUser.Uid,
		&currentUser.Name,
		&currentUser.Bio,
		&currentUser.Profile,
		&currentUser.Status,
		&currentUser.Popularity,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.UserIncidental{}, fmt.Errorf("user not found")
		}
		return model.UserIncidental{}, err
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return model.UserIncidental{}, err
	}

	return currentUser, nil
}

func FetchLatestUser(id int) (model.UserIncidental, error) {
	db, err := sql.Open("mysql", config.ConnectionStr)
	if err != nil {
		return model.UserIncidental{}, err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return model.UserIncidental{}, err
	}

	query := `
	SELECT 
	user_incidental.id AS id,
	user_incidental.uid AS uid,
	user_incidental.name AS name,
	user_incidental.bio AS bio,
	user_incidental.profile AS profile,
	user_incidental.status AS status,
	user_incidental.popularity AS popularity

	FROM user_incidental
	WHERE user_incidental.id = ? 
	`

	var currentUser model.UserIncidental
	err = db.QueryRow(query, id).Scan(
		&currentUser.Id,
		&currentUser.Uid,
		&currentUser.Name,
		&currentUser.Bio,
		&currentUser.Profile,
		&currentUser.Status,
		&currentUser.Popularity,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.UserIncidental{}, fmt.Errorf("user not found")
		}
		return model.UserIncidental{}, err
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return model.UserIncidental{}, err
	}

	return currentUser, nil
}

func UpdateProfile(modify_info model.UserModifyProfile) error {
	db, err := sql.Open("mysql", config.ConnectionStr)
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	query := `
	UPDATE user_incidental
	SET profile = ?
	WHERE user_incidental.id = ? 
	`

	_, err = tx.Exec(query, modify_info.Profile, modify_info.Id)

	if err != nil {
		return err
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func UpdateName(modify_info model.UserModifyName) error {
	db, err := sql.Open("mysql", config.ConnectionStr)
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	query := `
	UPDATE user_incidental
	SET name = ?
	WHERE user_incidental.id = ? 
	`

	_, err = tx.Exec(query, modify_info.Name, modify_info.Id)

	if err != nil {
		return err
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func UpdateBio(modify_info model.UserModifyBio) error {
	db, err := sql.Open("mysql", config.ConnectionStr)
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	query := `
	UPDATE user_incidental
	SET bio = ?
	WHERE user_incidental.id = ? 
	`

	_, err = tx.Exec(query, modify_info.Bio, modify_info.Id)

	if err != nil {
		return err
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
