package utils

import (
	"database/sql"
	_ "go-sql-driver/mysql"
	"log"
)

func CreateTables() {
	db, err := sql.Open("mysql", "sa:x123@(192.168.101.4:3306)/Blog?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	query := `
CREATE TABLE if not exists users(
	id INT AUTO_INCREMENT,
	username TINYTEXT NOT NULL,
	password TINYTEXT NOT NULL,
	PRIMARY KEY(id)
);

CREATE TABLE if not exists user_incidental(
	uid TEXT, //主键，主要用于被搜索，生成后不可更改
	id INT NOT NULL,
	name VARCHAR(255), //昵称，也可用于被搜索，允许可以更改
    bio TEXT, //个性签名
    profile TEXT, //头像图片地址
    status BOOL, //登录状态
    popularity DOUBLE, //主页受欢迎程度，用于优先搜索
    createdAt DATE //创建时间
	FOREIGN KEY (id) REFERENCES users(user.id)
	 ON DELETE restrict 
	 ON UPDATE cascade
);

CREATE TABLE if not exists articles(
	id INT AUTO_INCREMENT,
	username TEXT NOT NULL,
	password TEXT NOT NULL,
	PRIMARY KEY(id)
);
`
	if _, err := db.Exec(query); err != nil {
		log.Fatal(err)
	}

}
